/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusters

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	convertorsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

// OpenSearchReconciler reconciles a OpenSearch object
type OpenSearchReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenSearch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *OpenSearchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	openSearch := &clustersv1alpha1.OpenSearch{}
	err := r.Client.Get(ctx, req.NamespacedName, openSearch)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch OpenSearch cluster")
		return reconcile.Result{}, err
	}

	openSearchAnnotations := openSearch.Annotations
	switch openSearchAnnotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		err = r.HandleCreateCluster(openSearch, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create OpenSearch cluster",
				"Cluster name", openSearch.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	case models.UpdatingEvent:
		reconcileResult, err := r.HandleUpdateCluster(openSearch, &logger, &ctx, &req)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterNotRunning) {
				logger.Info("OpenSearch cluster is not ready to update",
					"Cluster name", openSearch.Spec.Name,
					"Cluster status", openSearch.Status.ClusterStatus,
					"Reason", err)
				return *reconcileResult, nil
			}

			logger.Error(err, "cannot update OpenSearch cluster",
				"Cluster name", openSearch.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return *reconcileResult, nil
	case models.DeletingEvent:
		err = r.HandleDeleteCluster(openSearch, &logger, &ctx, &req)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterIsBeingDeleted) {
				logger.Info("OpenSearch cluster is being deleted",
					"Cluster name", openSearch.Spec.Name,
					"Cluster status", openSearch.Status.Status,
				)
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: models.Requeue60,
				}, nil
			}

			logger.Error(err, "cannot delete OpenSearch cluster",
				"Cluster name", openSearch.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	default:
		logger.Info("UNKNOWN EVENT",
			"Cluster manifest", openSearch.Spec,
		)
		return reconcile.Result{}, err
	}
}

func (r *OpenSearchReconciler) HandleCreateCluster(
	openSearch *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	logger.Info(
		"Creating OpenSearch cluster",
		"Cluster name", openSearch.Spec.Name,
		"Data centres", openSearch.Spec.DataCentres,
	)

	openSearchSpec := convertorsv1.OpenSearchToInstAPI(&openSearch.Spec)

	id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, openSearchSpec)
	if err != nil {
		logger.Error(
			err, "cannot create OpenSearch cluster",
			"Cluster name", openSearch.Spec.Name,
			"Cluster manifest", openSearch.Spec,
		)
		return err
	}

	openSearch.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent

	err = r.patchClusterMetadata(ctx, openSearch, logger)
	if err != nil {
		logger.Error(err, "cannot patch OpenSearch cluster",
			"Cluster name", openSearch.Spec.Name,
			"Cluster metadata", openSearch.ObjectMeta,
		)
		return err
	}

	openSearch.Status.ID = id
	err = r.Status().Update(*ctx, openSearch)
	if err != nil {
		logger.Error(err, "cannot update OpenSearch cluster status",
			"Cluster name", openSearch.Spec.Name,
			"Cluster status", openSearch.Status,
		)
		return err
	}

	logger.Info(
		"OpenSearch resource has been created",
		"Cluster name", openSearch.Name,
		"Cluster ID", openSearch.Status.ID,
		"Kind", openSearch.Kind,
		"Api version", openSearch.APIVersion,
		"Namespace", openSearch.Namespace,
	)

	return nil
}

func (r *OpenSearchReconciler) HandleUpdateCluster(
	openSearch *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) (*reconcile.Result, error) {
	openSearchInstClusterStatus, err := r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get OpenSearch cluster status from the Instaclustr API",
			"Cluster name", openSearch.Spec.Name,
			"Cluster ID", openSearch.Status.ID,
		)

		return &reconcile.Result{}, err
	}

	if openSearchInstClusterStatus.Status != modelsv1.RunningStatus {
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: models.Requeue60,
		}, instaclustr.ClusterNotRunning
	}

	err = r.reconcileDataCentresNodeSize(openSearchInstClusterStatus, openSearch, logger)
	if errors.Is(err, instaclustr.StatusPreconditionFailed) {
		logger.Info("OpenSearch cluster is not ready to resize",
			"Cluster name", openSearch.Spec.Name,
			"Cluster status", openSearchInstClusterStatus.Status,
			"Reason", err,
		)
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: models.Requeue60,
		}, nil
	}
	if errors.Is(err, instaclustr.HasActiveResizeOperation) {
		logger.Info("Cluster is not ready to resize",
			"Cluster name", openSearch.Spec.Name,
			"New data centre manifest", openSearch.Spec.DataCentres[0],
			"Reason", err,
		)
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: models.Requeue60,
		}, nil
	}
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", openSearch.Spec.Name,
			"Cluster status", openSearchInstClusterStatus.Status,
		)
		return &reconcile.Result{}, err
	}

	err = r.updateDescriptionAndTwoFactorDelete(openSearch)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", openSearch.Spec.Name,
			"Cluster status", openSearchInstClusterStatus.Status,
		)
		return &reconcile.Result{}, err
	}

	openSearch.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	openSearchInstClusterStatus, err = r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get OpenSearch cluster status from the Instaclustr API",
			"Cluster name", openSearch.Spec.Name,
			"Cluster ID", openSearch.Status.ID,
		)
		return &reconcile.Result{}, err
	}

	err = r.patchClusterMetadata(ctx, openSearch, logger)
	if err != nil {
		logger.Error(err, "cannot patch OpenSearch metadata",
			"Cluster name", openSearch.Spec.Name,
			"Cluster metadata", openSearch.ObjectMeta,
		)
		return &reconcile.Result{}, err
	}

	openSearch.Status.ClusterStatus = *openSearchInstClusterStatus
	err = r.Status().Update(*ctx, openSearch)
	if err != nil {
		logger.Error(err, "cannot update OpenSearch cluster status",
			"Cluster name", openSearch.Spec.Name,
			"Cluster status", openSearch.Status,
		)
		return &reconcile.Result{}, err
	}

	logger.Info("OpenSearch cluster was updated",
		"Cluster name", openSearch.Spec.Name,
		"Cluster status", openSearch.Status,
	)

	return &reconcile.Result{}, nil
}

func (r *OpenSearchReconciler) HandleDeleteCluster(
	openSearch *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	status, err := r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		return err
	}

	if status != nil {
		err = r.API.DeleteCluster(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete OpenSearch cluster",
				"Cluster name", openSearch.Spec.Name,
				"Cluster status", openSearch.Status.Status,
			)

			return err
		}
		return instaclustr.ClusterIsBeingDeleted
	}

	controllerutil.RemoveFinalizer(openSearch, models.DeletionFinalizer)
	err = r.Update(*ctx, openSearch)
	if err != nil {
		logger.Error(
			err, "cannot update OpenSearch cluster CRD after finalizer removal",
			"Cluster name", openSearch.Spec.Name,
			"Cluster ID", openSearch.Status.ID,
		)
		return err
	}

	logger.Info("OpenSearch cluster was deleted",
		"Cluster name", openSearch.Spec.Name,
		"Cluster ID", openSearch.Status.ID,
	)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.OpenSearch{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				annotations := event.Object.GetAnnotations()
				if annotations[models.ResourceStateAnnotation] != "" {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
					return true
				}

				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					models.ResourceStateAnnotation: models.UpdatingEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.GenericEvent})
				return true
			},
		})).
		Complete(r)
}
