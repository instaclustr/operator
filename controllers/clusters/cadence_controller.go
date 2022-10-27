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
	"github.com/instaclustr/operator/pkg/models"
)

// CadenceReconciler reconciles a Cadence object
type CadenceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cadence object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CadenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	cadenceCluster := &clustersv1alpha1.Cadence{}
	err := r.Client.Get(ctx, req.NamespacedName, cadenceCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Cadence resource is not found",
				"Resource name", req.NamespacedName,
			)
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch Cadence resource",
			"Resource name", req.NamespacedName,
		)
		return reconcile.Result{}, err
	}

	cadenceAnnotations := cadenceCluster.Annotations
	switch cadenceAnnotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		reconcileResult := r.HandleCreateCluster(cadenceCluster, &logger, &ctx, &req)
		return *reconcileResult, nil
	case models.UpdatingEvent:
		reconcileResult := r.HandleUpdateCluster(cadenceCluster, &logger, &ctx, &req)
		return *reconcileResult, nil
	case models.DeletingEvent:
		reconcileResult := r.HandleDeleteCluster(cadenceCluster, &logger, &ctx, &req)
		return *reconcileResult, nil
	default:
		logger.Info("UNKNOWN EVENT",
			"Cluster name", cadenceCluster.Spec.Name,
		)
		return reconcile.Result{}, nil
	}
}

func (r *CadenceReconciler) HandleCreateCluster(
	cadenceCluster *clustersv1alpha1.Cadence,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) *reconcile.Result {
	if cadenceCluster.Status.ID == "" {
		logger.Info(
			"Creating Cadence cluster",
			"Cluster name", cadenceCluster.Spec.Name,
			"Data centres", cadenceCluster.Spec.DataCentres,
		)

		cadenceAPISpec, err := cadenceCluster.Spec.ToInstAPIv1(ctx, r.Client)
		if err != nil {
			logger.Error(err, "cannot convert Cadence cluster manifest to API spec",
				"Cluster manifest", cadenceCluster.Spec,
			)
			return &models.ReconcileRequeue60
		}

		id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, cadenceAPISpec)
		if err != nil {
			logger.Error(
				err, "cannot create Cadence cluster",
				"Cadence manifest", cadenceCluster.Spec,
			)
			return &models.ReconcileRequeue60
		}

		cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
		cadenceCluster.SetFinalizers([]string{models.DeletionFinalizer})

		patch, err := cadenceCluster.NewClusterMetadataPatch(ctx, logger)
		if err != nil {
			logger.Error(err, "cannot create Cadence cluster metadata patch",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster metadata", cadenceCluster.ObjectMeta,
			)
			return &models.ReconcileRequeue60
		}

		err = r.Client.Patch(*ctx, cadenceCluster, *patch)
		if err != nil {
			logger.Error(err, "cannot patch Cadence cluster",
				"Cluster name", cadenceCluster.Spec.Name,
				"Patch", *patch,
			)
			return &models.ReconcileRequeue60
		}

		cadenceCluster.Status.ID = id
		err = r.Status().Update(*ctx, cadenceCluster)
		if err != nil {
			logger.Error(err, "cannot update Cadence cluster status",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceCluster.Status,
			)
			return &models.ReconcileRequeue60
		}

		logger.Info(
			"Cadence resource has been created",
			"Cluster name", cadenceCluster.Name,
			"Cluster ID", cadenceCluster.Status.ID,
			"Kind", cadenceCluster.Kind,
			"Api version", cadenceCluster.APIVersion,
			"Namespace", cadenceCluster.Namespace,
		)
	}

	// TODO start status checker job

	return &reconcile.Result{Requeue: true}
}

func (r *CadenceReconciler) HandleUpdateCluster(
	cadenceCluster *clustersv1alpha1.Cadence,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) *reconcile.Result {
	logger.Info("Cadence cluster update is not implemented",
		"Cluster name", cadenceCluster.Spec.Name,
		"Cluster status", cadenceCluster.Status.Status,
	)

	return &reconcile.Result{}
}

func (r *CadenceReconciler) HandleDeleteCluster(
	cadenceCluster *clustersv1alpha1.Cadence,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) *reconcile.Result {
	controllerutil.RemoveFinalizer(cadenceCluster, models.DeletionFinalizer)
	err := r.Update(*ctx, cadenceCluster)
	if err != nil {
		logger.Error(err, "cannot update Cadence cluster resource after finalizer removal",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster ID", cadenceCluster.Status.ID,
		)
		return &models.ReconcileRequeue60
	}

	logger.Info("Cadence deletion is not implemented",
		"Cluster name", cadenceCluster.Spec.Name,
		"Cluster ID", cadenceCluster.Status.ID,
	)

	return &reconcile.Result{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CadenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cadence{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
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
