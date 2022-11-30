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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
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
			return models.ReconcileResult, nil
		}

		logger.Error(err, "unable to fetch Redis cluster",
			"Resource name", req.NamespacedName,
		)
		return models.ReconcileResult, nil
	}

	switch redisCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, redisCluster, logger), nil
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, redisCluster, logger), nil
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, redisCluster, logger), err
	case models.GenericEvent:
		logger.Info("Redis generic event isn't handled",
			"Cluster name", redisCluster.Spec.Name,
			"Request", req,
		)
		return models.ReconcileResult, err
	default:
		logger.Info("Unknown event isn't handled",
			"Cluster name", redisCluster.Spec.Name,
			"Request", req,
		)
		return models.ReconcileResult, err
	}
}

func (r *RedisReconciler) HandleCreateCluster(
	ctx context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) reconcile.Result {
	if redisCluster.Status.ID == "" {
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
			return models.ReconcileRequeue
		}

		redisCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		redisCluster.Annotations[models.DeletionConfirmed] = models.False
		redisCluster.Finalizers = append(redisCluster.Finalizers, models.DeletionFinalizer)

		err = r.patchClusterMetadata(ctx, redisCluster, logger)
		if err != nil {
			logger.Error(err, "cannot patch Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster metadata", redisCluster.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		patch := redisCluster.NewPatch()
		redisCluster.Status.ID = id
		err = r.Status().Patch(ctx, redisCluster, patch)
		if err != nil {
			logger.Error(err, "cannot update Redis cluster status",
				"Cluster name", redisCluster.Spec.Name,
			)
			return models.ReconcileRequeue
		}
	}

	err := r.startClusterStatusJob(redisCluster)
	if err != nil {
		logger.Error(err, "cannot start cluster status job",
			"Redis cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info(
		"Redis resource has been created",
		"Cluster name", redisCluster.Name,
		"Cluster ID", redisCluster.Status.ID,
		"Kind", redisCluster.Kind,
		"Api version", redisCluster.APIVersion,
		"Namespace", redisCluster.Namespace,
	)

	return models.ReconcileResult
}

func (r *RedisReconciler) HandleUpdateCluster(
	ctx context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) reconcile.Result {
	redisInstClusterStatus, err := r.API.GetClusterStatus(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get Redis cluster status from the Instaclustr API",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if redisInstClusterStatus.Status != models.RunningStatus {
		logger.Info("Cluster is not ready to update",
			"Cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNumber(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot reconcile data centres number",
			"Cluster name", redisCluster.Spec.Name,
			"Data centres", redisCluster.Spec.DataCentres,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNodeSize(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster status", redisInstClusterStatus.Status,
			"Current node size", redisInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"New node size", redisCluster.Spec.DataCentres[0].NodeSize,
		)

		return models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(redisCluster)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster status", redisCluster.Status,
			"Two factor delete", redisCluster.Spec.TwoFactorDelete,
		)

		return models.ReconcileRequeue
	}

	redisCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch Redis cluster after update",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster metadata", redisCluster.ObjectMeta,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Redis cluster was updated",
		"Cluster name", redisCluster.Spec.Name,
	)

	return models.ReconcileResult
}

func (r *RedisReconciler) HandleDeleteCluster(
	ctx context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) reconcile.Result {
	status, err := r.API.GetClusterStatus(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "cannot get Redis cluster status from Instaclustr",
			"Cluster ID", redisCluster.Status.ID,
			"Cluster Name", redisCluster.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	if len(redisCluster.Spec.TwoFactorDelete) != 0 &&
		redisCluster.Annotations[models.DeletionConfirmed] != models.True {
		logger.Info("Redis cluster deletion is not confirmed",
			"Cluster ID", redisCluster.Status.ID,
			"Cluster name", redisCluster.Spec.Name,
		)

		redisCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.patchClusterMetadata(ctx, redisCluster, logger)
		if err != nil {
			logger.Error(err, "cannot patch Redis cluster metadata after finalizer removal",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster ID", redisCluster.Status.ID,
			)

			return models.ReconcileRequeue
		}

		return models.ReconcileResult
	}

	if status != nil {
		err = r.API.DeleteCluster(redisCluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete Redis cluster",
				"Cluster name", redisCluster.Spec.Name,
				"Cluster status", redisCluster.Status.Status,
			)

			return models.ReconcileRequeue
		}

		logger.Info("Redis cluster is being deleted",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(redisCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(redisCluster, models.DeletionFinalizer)
	redisCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch Redis cluster metadata after finalizer removal",
			"Cluster name", redisCluster.Spec.Name,
			"Cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Redis cluster was deleted",
		"Cluster name", redisCluster.Spec.Name,
		"Cluster ID", redisCluster.Status.ID,
	)

	return models.ReconcileResult
}

func (r *RedisReconciler) startClusterStatusJob(cluster *clustersv1alpha1.Redis) error {
	job := r.newWatchStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisReconciler) newWatchStatusJob(cluster *clustersv1alpha1.Redis) scheduler.Job {
	l := log.Log.WithValues("component", "redisStatusClusterJob")
	return func() error {
		err := r.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("Redis cluster is being deleted. Status check job skipped",
				"Cluster name", cluster.Spec.Name,
				"Cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instStatus, err := r.API.GetClusterStatus(cluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			l.Error(err, "cannot get Redis cluster status",
				"ClusterID", cluster.Status.ID,
			)
			return err
		}

		if !isStatusesEqual(instStatus, &cluster.Status.ClusterStatus) {
			l.Info("Updating Redis cluster status",
				"New status", instStatus,
				"Old status", cluster.Status.ClusterStatus,
			)

			patch := cluster.NewPatch()
			cluster.Status.ClusterStatus = *instStatus
			err = r.Status().Patch(context.Background(), cluster, patch)
			if err != nil {
				l.Error(err, "cannot patch Redis cluster",
					"Cluster name", cluster.Spec.Name,
					"Status", cluster.Status.Status,
				)
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Redis{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				newObj := event.Object.(*clustersv1alpha1.Redis)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*clustersv1alpha1.Redis)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				if newObj.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).
		Complete(r)
}
