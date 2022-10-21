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
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Redis object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	redisCluster := &clustersv1alpha1.Redis{}
	err := r.Client.Get(ctx, req.NamespacedName, redisCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch Redis cluster",
			"Resource name", req.NamespacedName,
		)
		return reconcile.Result{}, err
	}

	redisAnnotations := redisCluster.Annotations
	switch redisAnnotations[models.CurrentEventAnnotation] {
	case models.CreateEvent:
		err = r.HandleCreateCluster(redisCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster manifest", redisCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	case models.UpdateEvent:
		reconcileResult, err := r.HandleUpdateCluster(redisCluster, &logger, &ctx, &req)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterNotRunning) {
				logger.Info("Cluster is not ready to update",
					"Cluster name", redisCluster.Spec.Name,
					"Cluster status", redisCluster.Status.ClusterStatus,
					"Reason", err)
				return *reconcileResult, nil
			}
			logger.Error(err, "cannot update Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return *reconcileResult, nil
	case models.DeleteEvent:
		err = r.HandleDeleteCluster(redisCluster, &logger, &ctx, &req)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterIsBeingDeleted) {
				logger.Info("Cluster is being deleted",
					"Cluster name", redisCluster.Spec.Name,
					"Cluster status", redisCluster.Status.Status,
				)
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: models.Requeue60,
				}, nil
			}

			logger.Error(err, "cannot delete Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster status", redisCluster.Status.Status,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	default:
		logger.Info("UNKNOWN EVENT",
			"Cluster name", redisCluster.Spec.Name,
		)
		return reconcile.Result{}, err
	}
}

func (r *RedisReconciler) HandleCreateCluster(
	redisCluster *clustersv1alpha1.Redis,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	logger.Info(
		"Creating Redis cluster",
		"Cluster name", redisCluster.Spec.Name,
		"Data centres", redisCluster.Spec.DataCentres,
	)

	redisSpec := r.ToInstAPIv1(&redisCluster.Spec)

	id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, redisSpec)
	if err != nil {
		logger.Error(
			err, "cannot create Redis cluster",
			"Cluster manifest", redisCluster.Spec,
		)
		return err
	}

	redisCluster.Annotations[models.PreviousEventAnnotation] = redisCluster.Annotations[models.CurrentEventAnnotation]
	redisCluster.Annotations[models.CurrentEventAnnotation] = models.UpdateEvent
	redisCluster.Annotations[models.IsClusterCreatedAnnotation] = "true"
	redisCluster.SetFinalizers([]string{models.DeletionFinalizer})

	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch Redis cluster",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster metadata", redisCluster.ObjectMeta,
		)
		return err
	}

	redisCluster.Status.ID = id
	err = r.Status().Update(*ctx, redisCluster)
	if err != nil {
		logger.Error(err, "cannot update Redis cluster status",
			"Cluster name", redisCluster.Spec.Name,
		)
		return err
	}

	logger.Info(
		"Redis resource has been created",
		"Cluster name", redisCluster.Name,
		"Cluster ID", redisCluster.Status.ID,
		"Kind", redisCluster.Kind,
		"Api version", redisCluster.APIVersion,
		"Namespace", redisCluster.Namespace,
	)

	return nil
}

func (r *RedisReconciler) HandleUpdateCluster(
	redisCluster *clustersv1alpha1.Redis,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) (*reconcile.Result, error) {
	redisInstClusterStatus, err := r.API.GetClusterStatus(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get Redis cluster status from the Instaclustr API",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster ID", redisCluster.Status.ID,
		)

		return &reconcile.Result{}, err
	}

	if redisInstClusterStatus.Status != modelsv1.RunningStatus {
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: models.Requeue60,
		}, instaclustr.ClusterNotRunning
	}

	err = r.reconcileDataCentresNumber(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot reconcile data centres number",
			"Cluster name", redisCluster.Spec.Name,
			"Data centres", redisCluster.Spec.DataCentres,
		)

		return nil, err
	}

	reconcileResult, err := r.reconcileDataCentresNodeSize(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster status", redisInstClusterStatus.Status,
			"Current node size", redisInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"New node size", redisCluster.Spec.DataCentres[0].NodeSize,
		)
		return reconcileResult, err
	}
	if reconcileResult != nil {
		return reconcileResult, nil
	}

	err = r.updateDescriptionAndTwoFactorDelete(redisCluster)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster status", redisCluster.Status,
			"Two factor delete", redisCluster.Spec.TwoFactorDelete,
		)
		return &reconcile.Result{}, err
	}

	redisCluster.Annotations[models.PreviousEventAnnotation] = redisCluster.Annotations[models.CurrentEventAnnotation]
	redisCluster.Annotations[models.CurrentEventAnnotation] = ""
	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch Redis cluster after update",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster metadata", redisCluster.ObjectMeta,
		)
		return &reconcile.Result{}, err
	}

	logger.Info("Redis cluster was updated",
		"Cluster name", redisCluster.Spec.Name,
	)

	return &reconcile.Result{}, nil
}

func (r *RedisReconciler) HandleDeleteCluster(
	redisCluster *clustersv1alpha1.Redis,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	status, err := r.API.GetClusterStatus(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		return err
	}

	if status != nil {
		err = r.API.DeleteCluster(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster status", redisCluster.Status.Status,
			)

			return err
		}
		return instaclustr.ClusterIsBeingDeleted
	}

	controllerutil.RemoveFinalizer(redisCluster, models.DeletionFinalizer)
	err = r.Update(*ctx, redisCluster)
	if err != nil {
		logger.Error(
			err, "cannot update Redis cluster CRD after finalizer removal",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster ID", redisCluster.Status.ID,
		)
		return err
	}

	logger.Info("Redis cluster was deleted",
		"Cluster name", redisCluster.Spec.Name,
		"Cluster ID", redisCluster.Status.ID,
	)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Redis{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.DeleteEvent})
					return true
				}

				annotations := event.Object.GetAnnotations()
				if annotations[models.IsClusterCreatedAnnotation] == "true" {
					event.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.UpdateEvent})
					return true
				}

				event.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.CreateEvent})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.DeleteEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					models.CurrentEventAnnotation: models.UpdateEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.GenericEvent})
				return true
			},
		})).
		Complete(r)
}
