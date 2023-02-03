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
	"strconv"

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

		logger.Error(err, "Unable to fetch Redis cluster",
			"resource name", req.NamespacedName,
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
			"cluster name", redisCluster.Spec.Name,
			"request", req,
		)
		return models.ReconcileResult, err
	default:
		logger.Info("Unknown event isn't handled",
			"cluster name", redisCluster.Spec.Name,
			"request", req,
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
		if redisCluster.Spec.HasRestore() {
			logger.Info(
				"Creating Redis cluster from backup",
				"original cluster ID", redisCluster.Spec.RestoreFrom.ClusterID,
			)

			id, err := r.API.RestoreRedisCluster(redisCluster.Spec.RestoreFrom)
			if err != nil {
				logger.Error(
					err, "Cannot create Redis cluster from backup",
					"original cluster ID", redisCluster.Spec.RestoreFrom.ClusterID,
				)
				return models.ReconcileRequeue
			}

			patch := redisCluster.NewPatch()
			redisCluster.Status.ID = id
			err = r.Status().Patch(ctx, redisCluster, patch)
			if err != nil {
				logger.Error(err, "Cannot update Redis cluster status",
					"cluster name", redisCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}
		} else {
			logger.Info(
				"Creating Redis cluster",
				"cluster name", redisCluster.Spec.Name,
				"data centres", redisCluster.Spec.DataCentres,
			)

			redisSpec := redisCluster.Spec.ToInstAPIv2()

			id, err := r.API.CreateCluster(instaclustr.RedisEndpoint, redisSpec)
			if err != nil {
				logger.Error(
					err, "Cannot create Redis cluster",
					"cluster manifest", redisCluster.Spec,
				)
				return models.ReconcileRequeue
			}

			patch := redisCluster.NewPatch()
			redisCluster.Status.ID = id
			err = r.Status().Patch(ctx, redisCluster, patch)
			if err != nil {
				logger.Error(err, "Cannot update Redis cluster status",
					"cluster name", redisCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}
		}
	}

	controllerutil.AddFinalizer(redisCluster, models.DeletionFinalizer)
	redisCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
	redisCluster.Annotations[models.DeletionConfirmed] = models.False

	err := r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster",
			"cluster name", redisCluster.Spec.Name,
			"cluster metadata", redisCluster.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	err = r.startClusterStatusJob(redisCluster)
	if err != nil {
		logger.Error(err, "Cannot start cluster status job",
			"redis cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.startClusterBackupsJob(redisCluster)
	if err != nil {
		logger.Error(err, "Cannot start Redis cluster backups check job",
			"cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info(
		"Redis resource has been created",
		"cluster name", redisCluster.Name,
		"cluster ID", redisCluster.Status.ID,
		"kind", redisCluster.Kind,
		"api version", redisCluster.APIVersion,
		"namespace", redisCluster.Namespace,
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
			err, "Cannot get Redis cluster status from the Instaclustr API",
			"cluster name", redisCluster.Spec.Name,
			"cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if redisInstClusterStatus.Status != models.RunningStatus {
		logger.Info("Cluster is not ready to update",
			"cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNumber(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot reconcile data centres number",
			"cluster name", redisCluster.Spec.Name,
			"data centres", redisCluster.Spec.DataCentres,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNodeSize(redisInstClusterStatus, redisCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot reconcile data centres node size",
			"cluster name", redisCluster.Spec.Name,
			"cluster status", redisInstClusterStatus.Status,
			"current node size", redisInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"new node size", redisCluster.Spec.DataCentres[0].NodeSize,
		)

		return models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(redisCluster)
	if err != nil {
		logger.Error(err, "Cannot update description and twoFactorDelete",
			"cluster name", redisCluster.Spec.Name,
			"cluster status", redisCluster.Status,
			"two factor delete", redisCluster.Spec.TwoFactorDelete,
		)

		return models.ReconcileRequeue
	}

	redisCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster after update",
			"cluster name", redisCluster.Spec.Name,
			"cluster metadata", redisCluster.ObjectMeta,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Redis cluster was updated",
		"cluster name", redisCluster.Spec.Name,
	)

	return models.ReconcileResult
}

func (r *RedisReconciler) HandleDeleteCluster(
	ctx context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) reconcile.Result {
	status, err := r.API.GetClusterStatus(redisCluster.Status.ID, instaclustr.RedisEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get Redis cluster status from Instaclustr",
			"cluster ID", redisCluster.Status.ID,
			"cluster Name", redisCluster.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	if len(redisCluster.Spec.TwoFactorDelete) != 0 &&
		redisCluster.Annotations[models.DeletionConfirmed] != models.True {
		logger.Info("Redis cluster deletion is not confirmed",
			"cluster ID", redisCluster.Status.ID,
			"cluster name", redisCluster.Spec.Name,
		)

		redisCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.patchClusterMetadata(ctx, redisCluster, logger)
		if err != nil {
			logger.Error(err, "Cannot patch Redis cluster metadata after finalizer removal",
				"cluster name", redisCluster.Spec.Name,
				"cluster ID", redisCluster.Status.ID,
			)

			return models.ReconcileRequeue
		}

		return models.ReconcileResult
	}

	if status != nil {
		err = r.API.DeleteCluster(redisCluster.Status.ID, instaclustr.RedisEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete Redis cluster",
				"cluster name", redisCluster.Spec.Name,
				"cluster status", redisCluster.Status.Status,
			)

			return models.ReconcileRequeue
		}

		logger.Info("Redis cluster is being deleted",
			"cluster name", redisCluster.Spec.Name,
			"cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(redisCluster.GetJobID(scheduler.BackupsChecker))

	logger.Info("Deleting cluster backup resources",
		"cluster ID", redisCluster.Status.ID,
	)

	err = r.deleteBackups(ctx, redisCluster.Status.ID, redisCluster.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", redisCluster.Status.ID,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster backup resources were deleted",
		"cluster ID", redisCluster.Status.ID,
	)

	r.Scheduler.RemoveJob(redisCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(redisCluster, models.DeletionFinalizer)
	redisCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.patchClusterMetadata(ctx, redisCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster metadata after finalizer removal",
			"cluster name", redisCluster.Spec.Name,
			"cluster ID", redisCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Redis cluster was deleted",
		"cluster name", redisCluster.Spec.Name,
		"cluster ID", redisCluster.Status.ID,
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

func (r *RedisReconciler) startClusterBackupsJob(cluster *clustersv1alpha1.Redis) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
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
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instStatus, err := r.API.GetClusterStatus(cluster.Status.ID, instaclustr.RedisEndpoint)
		if err != nil {
			l.Error(err, "Cannot get Redis cluster status",
				"cluster ID", cluster.Status.ID,
			)

			return err
		}

		patch := cluster.NewPatch()
		if !areStatusesEqual(instStatus, &cluster.Status.ClusterStatus) {
			l.Info("Updating Redis cluster status",
				"new status", instStatus,
				"old status", cluster.Status.ClusterStatus,
			)

			cluster.Status.ClusterStatus = *instStatus
			err = r.Status().Patch(context.Background(), cluster, patch)
			if err != nil {
				l.Error(err, "Cannot patch Redis cluster",
					"cluster name", cluster.Spec.Name,
					"status", cluster.Status.Status,
				)

				return err
			}
		}

		if instStatus.CurrentClusterOperationStatus == models.NoOperation {
			instSpec, err := r.API.GetRedisSpec(cluster.Status.ID, instaclustr.RedisEndpoint)
			if err != nil {
				l.Error(err, "Cannot get Redis cluster spec from Instaclustr API",
					"cluster ID", cluster.Status.ID,
				)

				return err
			}

			if !cluster.Spec.IsSpecEqual(instSpec) {
				cluster.Spec.SetSpecFromInst(instSpec)
				err = r.Patch(context.Background(), cluster, patch)
				if err != nil {
					l.Error(err, "Cannot patch Redis cluster spec",
						"cluster ID", cluster.Status.ID,
					)

					return err
				}

				l.Info("Redis cluster spec has been updated",
					"cluster ID", cluster.Status.ID,
				)
			}
		}

		return nil
	}
}

func (r *RedisReconciler) newWatchBackupsJob(cluster *clustersv1alpha1.Redis) scheduler.Job {
	l := log.Log.WithValues("component", "redisBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("Redis cluster is being deleted. Backups check job skipped",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Redis cluster backups",
				"cluster name", cluster.Spec.Name,
				"clusterID", cluster.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.RedisClusterKind)

		k8sBackupList, err := r.listClusterBackups(ctx, cluster.Status.ID, cluster.Namespace)
		if err != nil {
			return err
		}

		k8sBackups := map[int]*clusterresourcesv1alpha1.ClusterBackup{}
		unassignedBackups := []*clusterresourcesv1alpha1.ClusterBackup{}
		for _, k8sBackup := range k8sBackupList.Items {
			if k8sBackup.Status.Start != 0 {
				k8sBackups[k8sBackup.Status.Start] = &k8sBackup
				continue
			}
			if k8sBackup.Annotations[models.StartTimestampAnnotation] != "" {
				patch := k8sBackup.NewPatch()
				k8sBackup.Status.Start, err = strconv.Atoi(k8sBackup.Annotations[models.StartTimestampAnnotation])
				if err != nil {
					return err
				}

				err = r.Status().Patch(ctx, &k8sBackup, patch)
				if err != nil {
					return err
				}

				k8sBackups[k8sBackup.Status.Start] = &k8sBackup
				continue
			}

			unassignedBackups = append(unassignedBackups, &k8sBackup)
		}

		for start, instBackup := range instBackupEvents {
			if _, exists := k8sBackups[start]; exists {
				if k8sBackups[start].Status.End != 0 {
					continue
				}

				patch := k8sBackups[start].NewPatch()
				k8sBackups[start].Status.UpdateStatus(instBackup)
				err = r.Status().Patch(ctx, k8sBackups[start], patch)
				if err != nil {
					return err
				}

				l.Info("Backup resource was updated",
					"backup resource name", k8sBackups[start].Name,
				)
				continue
			}

			if len(unassignedBackups) != 0 {
				backupToAssign := unassignedBackups[len(unassignedBackups)-1]
				unassignedBackups = unassignedBackups[:len(unassignedBackups)-1]
				patch := backupToAssign.NewPatch()
				backupToAssign.Status.Start = instBackup.Start
				backupToAssign.Status.UpdateStatus(instBackup)
				err = r.Status().Patch(context.TODO(), backupToAssign, patch)
				if err != nil {
					return err
				}
				continue
			}

			backupSpec := cluster.NewBackupSpec(start)
			err = r.Create(ctx, backupSpec)
			if err != nil {
				return err
			}
			l.Info("Found new backup on Instaclustr. New backup resource was created",
				"backup resource name", backupSpec.Name,
			)
		}

		return nil
	}
}

func (r *RedisReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1alpha1.ClusterBackupList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err := r.Client.List(ctx, backupsList, listOpts...)
	if err != nil {
		return nil, err
	}

	return backupsList, nil
}

func (r *RedisReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
	backupsList, err := r.listClusterBackups(ctx, clusterID, namespace)
	if err != nil {
		return err
	}

	if len(backupsList.Items) == 0 {
		return nil
	}

	backupType := &clusterresourcesv1alpha1.ClusterBackup{}
	opts := []client.DeleteAllOfOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err = r.DeleteAllOf(ctx, backupType, opts...)
	if err != nil {
		return err
	}

	for _, backup := range backupsList.Items {
		patch := backup.NewPatch()
		controllerutil.RemoveFinalizer(&backup, models.DeletionFinalizer)
		err = r.Patch(ctx, &backup, patch)
		if err != nil {
			return err
		}
	}

	return nil
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
