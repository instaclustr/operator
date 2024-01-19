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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	redis := &v1beta1.Redis{}
	err := r.Client.Get(ctx, req.NamespacedName, redis)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Redis cluster resource is not found",
				"resource name", req.NamespacedName,
			)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch Redis cluster",
			"resource name", req.NamespacedName,
		)
		return models.ExitReconcile, nil
	}

	switch redis.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, redis, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, redis, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, redis, l)
	case models.GenericEvent:
		l.Info("Redis generic event isn't handled",
			"cluster name", redis.Spec.Name,
			"request", req,
		)
		return models.ExitReconcile, nil
	default:
		l.Info("Unknown event isn't handled",
			"cluster name", redis.Spec.Name,
			"request", req,
		)
		return models.ExitReconcile, nil
	}
}

func (r *RedisReconciler) handleCreateCluster(
	ctx context.Context,
	redis *v1beta1.Redis,
	l logr.Logger,
) (reconcile.Result, error) {
	var err error
	if redis.Status.ID == "" {
		var id string
		if redis.Spec.HasRestore() {
			l.Info(
				"Creating Redis cluster from backup",
				"original cluster ID", redis.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreCluster(redis.RestoreInfoToInstAPI(redis.Spec.RestoreFrom), models.RedisAppKind)
			if err != nil {
				l.Error(
					err, "Cannot restore Redis cluster from backup",
					"original cluster ID", redis.Spec.RestoreFrom.ClusterID,
				)
				r.EventRecorder.Eventf(
					redis, models.Warning, models.CreationFailed,
					"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			l.Info(
				"Redis cluster was created from backup",
				"original cluster ID", redis.Spec.RestoreFrom.ClusterID,
			)

			r.EventRecorder.Eventf(
				redis, models.Normal, models.Created,
				"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
				redis.Spec.RestoreFrom.ClusterID,
				id,
			)
		} else {
			l.Info(
				"Creating Redis cluster",
				"cluster name", redis.Spec.Name,
				"data centres", redis.Spec.DataCentres,
			)

			id, err = r.API.CreateCluster(instaclustr.RedisEndpoint, redis.Spec.ToInstAPI())
			if err != nil {
				l.Error(
					err, "Cannot create Redis cluster",
					"cluster manifest", redis.Spec,
				)
				r.EventRecorder.Eventf(
					redis, models.Warning, models.CreationFailed,
					"Cluster creation on the Instaclustr is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			l.Info(
				"Redis cluster was created",
				"cluster ID", id,
				"cluster name", redis.Spec.Name,
			)
			r.EventRecorder.Eventf(
				redis, models.Normal, models.Created,
				"Cluster creation request is sent. Cluster ID: %s",
				id,
			)
		}

		patch := redis.NewPatch()
		redis.Status.ID = id
		err = r.Status().Patch(ctx, redis, patch)
		if err != nil {
			l.Error(err, "Cannot patch Redis cluster status",
				"cluster name", redis.Spec.Name)
			r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		l.Info("Redis resource has been created",
			"cluster name", redis.Name,
			"cluster ID", redis.Status.ID,
			"api version", redis.APIVersion)
	}

	patch := redis.NewPatch()
	controllerutil.AddFinalizer(redis, models.DeletionFinalizer)
	redis.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		l.Error(err, "Cannot patch Redis cluster",
			"cluster name", redis.Spec.Name,
			"cluster metadata", redis.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			redis, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if redis.Status.State != models.DeletedStatus {
		err = r.startClusterStatusJob(redis)
		if err != nil {
			l.Error(err, "Cannot start cluster status job",
				"redis cluster ID", redis.Status.ID,
			)

			r.EventRecorder.Eventf(
				redis, models.Warning, models.CreationFailed,
				"Cluster status job creation is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			redis, models.Normal, models.Created,
			"Cluster status check job is started",
		)

		if redis.Spec.OnPremisesSpec != nil {
			iData, err := r.API.GetRedis(redis.Status.ID)
			if err != nil {
				l.Error(err, "Cannot get cluster from the Instaclustr API",
					"cluster name", redis.Spec.Name,
					"data centres", redis.Spec.DataCentres,
					"cluster ID", redis.Status.ID,
				)
				r.EventRecorder.Eventf(
					redis, models.Warning, models.FetchFailed,
					"Cluster fetch from the Instaclustr API is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}
			iRedis, err := redis.FromInstAPI(iData)
			if err != nil {
				l.Error(
					err, "Cannot convert cluster from the Instaclustr API",
					"cluster name", redis.Spec.Name,
					"cluster ID", redis.Status.ID,
				)
				r.EventRecorder.Eventf(
					redis, models.Warning, models.ConversionFailed,
					"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			bootstrap := newOnPremisesBootstrap(
				r.Client,
				redis,
				r.EventRecorder,
				iRedis.Status.ClusterStatus,
				redis.Spec.OnPremisesSpec,
				newExposePorts(redis.GetExposePorts()),
				redis.GetHeadlessPorts(),
				redis.Spec.PrivateNetworkCluster,
			)

			err = handleCreateOnPremisesClusterResources(ctx, bootstrap)
			if err != nil {
				l.Error(
					err, "Cannot create resources for on-premises cluster",
					"cluster spec", redis.Spec.OnPremisesSpec,
				)
				r.EventRecorder.Eventf(
					redis, models.Warning, models.CreationFailed,
					"Resources creation for on-premises cluster is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			err = r.startClusterOnPremisesIPsJob(redis, bootstrap)
			if err != nil {
				l.Error(err, "Cannot start on-premises cluster IPs check job",
					"cluster ID", redis.Status.ID,
				)

				r.EventRecorder.Eventf(
					redis, models.Warning, models.CreationFailed,
					"On-premises cluster IPs check job is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			l.Info(
				"On-premises resources have been created",
				"cluster name", redis.Spec.Name,
				"on-premises Spec", redis.Spec.OnPremisesSpec,
				"cluster ID", redis.Status.ID,
			)
			return models.ExitReconcile, nil
		}

		err = r.startClusterBackupsJob(redis)
		if err != nil {
			l.Error(err, "Cannot start Redis cluster backups check job",
				"cluster ID", redis.Status.ID,
			)

			r.EventRecorder.Eventf(
				redis, models.Warning, models.CreationFailed,
				"Cluster backups job creation is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			redis, models.Normal, models.Created,
			"Cluster backups check job is started",
		)

		if redis.Spec.UserRefs != nil {
			err = r.startUsersCreationJob(redis)

			if err != nil {
				l.Error(err, "Failed to start user creation job")
				r.EventRecorder.Eventf(redis, models.Warning, models.CreationFailed,
					"User creation job is failed. Reason: %v", err,
				)
				return reconcile.Result{}, err
			}

			r.EventRecorder.Event(redis, models.Normal, models.Created,
				"Cluster user creation job is started")
		}
	}

	l.Info(
		"Redis resource has been created",
		"cluster name", redis.Name,
		"cluster ID", redis.Status.ID,
		"kind", redis.Kind,
		"api version", redis.APIVersion,
		"namespace", redis.Namespace,
	)

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) startUsersCreationJob(cluster *v1beta1.Redis) error {
	job := r.newUsersCreationJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.UserCreator), scheduler.UserCreationInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisReconciler) handleUpdateCluster(
	ctx context.Context,
	redis *v1beta1.Redis,
	req ctrl.Request,
	l logr.Logger,
) (reconcile.Result, error) {
	iData, err := r.API.GetRedis(redis.Status.ID)
	if err != nil {
		l.Error(
			err, "Cannot get Redis cluster from the Instaclustr API",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.FetchFailed,
			"Fetch cluster from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	iRedis, err := redis.FromInstAPI(iData)
	if err != nil {
		l.Error(
			err, "Cannot convert Redis cluster from the Instaclustr API",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.ConversionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if redis.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return r.handleExternalChanges(redis, iRedis, l)
	}

	if redis.Spec.ClusterSettingsNeedUpdate(iRedis.Spec.Cluster) {
		l.Info("Updating cluster settings",
			"instaclustr description", iRedis.Spec.Description,
			"instaclustr two factor delete", iRedis.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(redis.Status.ID, redis.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", redis.Status.ID, "cluster spec", redis.Spec)
			r.EventRecorder.Eventf(redis, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !redis.Spec.IsEqual(iRedis.Spec) {
		l.Info("Update request to Instaclustr API has been sent",
			"spec data centres", redis.Spec.DataCentres,
			"resize settings", redis.Spec.ResizeSettings,
		)

		err = r.API.UpdateRedis(redis.Status.ID, redis.Spec.DCsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update Redis cluster data centres",
				"cluster name", redis.Spec.Name,
				"cluster status", redis.Status,
				"data centres", redis.Spec.DataCentres)

			r.EventRecorder.Eventf(redis, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	err = handleUsersChanges(ctx, r.Client, r, redis)
	if err != nil {
		l.Error(err, "Failed to handle users changes")
		r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
			"Handling users changes is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	patch := redis.NewPatch()
	redis.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		l.Error(err, "Cannot patch Redis cluster after update",
			"cluster name", redis.Spec.Name,
			"cluster metadata", redis.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", redis.Spec.Name,
		"cluster ID", redis.Status.ID,
		"data centres", redis.Spec.DataCentres,
	)

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) handleExternalChanges(redis, iRedis *v1beta1.Redis, l logr.Logger) (reconcile.Result, error) {
	patch := redis.NewPatch()

	if !redis.Spec.IsEqual(iRedis.Spec) {
		redis.Annotations[models.ExternalChangesAnnotation] = models.True

		l.Info(msgExternalChanges,
			"specification of k8s resource", redis.Spec,
			"data from Instaclustr ", iRedis.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(redis.Spec, iRedis.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iRedis.Spec, "k8s resource spec", redis.Spec)
			return models.ExitReconcile, nil
		}

		r.EventRecorder.Eventf(redis, models.Warning, models.ExternalChanges, msgDiffSpecs)
	} else {
		redis.Annotations[models.ExternalChangesAnnotation] = ""
		l.Info("External changes have been reconciled", "resource ID", redis.Status.ID)
		r.EventRecorder.Event(redis, models.Normal, models.ExternalChanges, "External changes have been reconciled")
	}

	err := r.Patch(context.Background(), redis, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", redis.Spec.Name, "cluster ID", redis.Status.ID)

		r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) handleDeleteCluster(
	ctx context.Context,
	redis *v1beta1.Redis,
	l logr.Logger,
) (reconcile.Result, error) {

	_, err := r.API.GetRedis(redis.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Redis cluster status from Instaclustr",
			"cluster ID", redis.Status.ID,
			"cluster name", redis.Spec.Name,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.FetchFailed,
			"Fetch cluster from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID)

		err = r.API.DeleteCluster(redis.Status.ID, instaclustr.RedisEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Redis cluster",
				"cluster name", redis.Spec.Name,
				"cluster status", redis.Status.State,
			)

			r.EventRecorder.Eventf(
				redis, models.Warning, models.DeletionFailed,
				"Cluster deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(redis, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if redis.Spec.TwoFactorDelete != nil {
			patch := redis.NewPatch()

			redis.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			redis.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, redis, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", redis.Spec.Name,
					"cluster state", redis.Status.State)
				r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", redis.Status.ID)

			r.EventRecorder.Event(redis, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
		if redis.Spec.OnPremisesSpec != nil {
			err = deleteOnPremResources(ctx, r.Client, redis.Status.ID, redis.Namespace)
			if err != nil {
				l.Error(err, "Cannot delete cluster on-premises resources",
					"cluster ID", redis.Status.ID)
				r.EventRecorder.Eventf(redis, models.Warning, models.DeletionFailed,
					"Cluster on-premises resources deletion is failed. Reason: %v", err)
				return reconcile.Result{}, err
			}

			l.Info("Cluster on-premises resources are deleted",
				"cluster ID", redis.Status.ID)
			r.EventRecorder.Eventf(redis, models.Normal, models.Deleted,
				"Cluster on-premises resources are deleted")
			r.Scheduler.RemoveJob(redis.GetJobID(scheduler.OnPremisesIPsChecker))
		}
	}

	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.StatusChecker))
	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.BackupsChecker))

	l.Info("Deleting cluster backup resources",
		"cluster ID", redis.Status.ID,
	)

	err = r.deleteBackups(ctx, redis.Status.ID, redis.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", redis.Status.ID,
		)
		r.EventRecorder.Eventf(
			redis, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	r.EventRecorder.Eventf(
		redis, models.Normal, models.Deleted,
		"Cluster backup resources are deleted",
	)

	l.Info("Cluster backup resources are deleted",
		"cluster ID", redis.Status.ID,
	)

	err = detachUsers(ctx, r.Client, r, redis)
	if err != nil {
		l.Error(err, "Failed to detach users from the cluster")
		r.EventRecorder.Eventf(redis, models.Warning, models.DeletionFailed,
			"Detaching users from the cluster is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	patch := redis.NewPatch()
	controllerutil.RemoveFinalizer(redis, models.DeletionFinalizer)
	redis.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		l.Error(err, "Cannot patch Redis cluster metadata after finalizer removal",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, redis.Name, redis.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Redis cluster expose service",
			"cluster ID", redis.Status.ID,
			"cluster name", redis.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Redis cluster was deleted",
		"cluster name", redis.Spec.Name,
		"cluster ID", redis.Status.ID,
	)

	r.EventRecorder.Eventf(
		redis, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) startClusterOnPremisesIPsJob(redis *v1beta1.Redis, b *onPremisesBootstrap) error {
	job := newWatchOnPremisesIPsJob(redis.Kind, b)

	err := r.Scheduler.ScheduleJob(redis.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisReconciler) startClusterStatusJob(cluster *v1beta1.Redis) error {
	job := r.newWatchStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisReconciler) startClusterBackupsJob(cluster *v1beta1.Redis) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisReconciler) newUsersCreationJob(redis *v1beta1.Redis) scheduler.Job {
	l := log.Log.WithValues("component", "redisUsersCreationJob")

	return func() error {
		ctx := context.Background()

		err := r.Get(ctx, types.NamespacedName{
			Namespace: redis.Namespace,
			Name:      redis.Name,
		}, redis)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if redis.Status.State != models.RunningStatus {
			l.Info("User creation job is scheduled")
			r.EventRecorder.Eventf(redis, models.Normal, models.CreationFailed,
				"User creation job is scheduled, cluster is not in the running state",
			)
			return nil
		}

		err = handleUsersChanges(ctx, r.Client, r, redis)
		if err != nil {
			l.Error(err, "Failed to create users")
			r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
				"Creating users is failed. Reason: %w", err,
			)
			return err
		}

		l.Info("User creation job successfully finished")
		r.EventRecorder.Eventf(redis, models.Normal, models.Created,
			"User creation job successfully finished",
		)

		r.Scheduler.RemoveJob(redis.GetJobID(scheduler.UserCreator))

		return nil
	}
}

func (r *RedisReconciler) newWatchStatusJob(redis *v1beta1.Redis) scheduler.Job {
	l := log.Log.WithValues("component", "redisStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(redis)
		err := r.Get(context.Background(), namespacedName, redis)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(redis.GetJobID(scheduler.UserCreator))
			r.Scheduler.RemoveJob(redis.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(redis.GetJobID(scheduler.StatusChecker))
			return nil
		}

		iData, err := r.API.GetRedis(redis.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if redis.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), redis, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), redis)
			}

			l.Error(err, "Cannot get Redis cluster status from Instaclustr",
				"cluster ID", redis.Status.ID,
			)

			return err
		}

		iRedis, err := redis.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert Redis cluster status from Instaclustr",
				"cluster ID", redis.Status.ID,
			)

			return err
		}

		if !areStatusesEqual(&iRedis.Status.ClusterStatus, &redis.Status.ClusterStatus) {
			l.Info("Updating Redis cluster status",
				"new status", iRedis.Status,
				"old status", redis.Status,
			)

			areDCsEqual := areDataCentresEqual(iRedis.Status.ClusterStatus.DataCentres, redis.Status.ClusterStatus.DataCentres)

			patch := redis.NewPatch()
			redis.Status.ClusterStatus = iRedis.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), redis, patch)
			if err != nil {
				l.Error(err, "Cannot patch Redis cluster",
					"cluster name", redis.Spec.Name,
					"status", redis.Status.State,
				)

				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iRedis.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					redis.Name,
					redis.Namespace,
					redis.Spec.PrivateNetworkCluster,
					nodes,
					models.RedisConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iRedis.Status.CurrentClusterOperationStatus == models.NoOperation &&
			redis.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!redis.Spec.IsEqual(iRedis.Spec) {

			k8sData, err := removeRedundantFieldsFromSpec(redis.Spec, "userRefs")
			if err != nil {
				l.Error(err, "Cannot remove redundant fields from k8s Spec")
				return err
			}

			l.Info(msgExternalChanges, "instaclustr data", iRedis.Spec, "k8s resource spec", string(k8sData))

			patch := redis.NewPatch()
			redis.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), redis, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", redis.Spec.Name, "cluster state", redis.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(redis.Spec, iRedis.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iRedis.Spec, "k8s resource spec", redis.Spec)
				return err
			}
			r.EventRecorder.Eventf(redis, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), redis)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", redis.Spec.Name,
				"cluster ID", redis.Status.ID,
			)
			return err
		}

		if redis.Status.State == models.RunningStatus && redis.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := redis.NewPatch()
			for _, dc := range redis.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", redis.Spec.Name,
						"cluster ID", redis.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), redis, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", redis.Spec.Name,
						"cluster ID", redis.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *RedisReconciler) newWatchBackupsJob(cluster *v1beta1.Redis) scheduler.Job {
	l := log.Log.WithValues("component", "redisBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		instBackups, err := r.API.GetClusterBackups(cluster.Status.ID, models.ClusterKindsMap[cluster.Kind])
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
			l.Error(err, "Cannot get list of Redis cluster backups",
				"cluster name", cluster.Spec.Name,
				"clusterID", cluster.Status.ID,
			)

			return err
		}

		k8sBackups := map[int]*clusterresourcesv1beta1.ClusterBackup{}
		unassignedBackups := []*clusterresourcesv1beta1.ClusterBackup{}
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

func (r *RedisReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1beta1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1beta1.ClusterBackupList{}
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

	backupType := &clusterresourcesv1beta1.ClusterBackup{}
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

func (r *RedisReconciler) NewUserResource() userObject {
	return &clusterresourcesv1beta1.RedisUser{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{RateLimiter: r.RateLimiter}).
		For(&v1beta1.Redis{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				newObj := event.ObjectNew.(*v1beta1.Redis)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
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
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Complete(r)
}

func (r *RedisReconciler) reconcileMaintenanceEvents(ctx context.Context, redis *v1beta1.Redis) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(redis.Status.ID)
	if err != nil {
		return err
	}

	if !redis.Status.AreMaintenanceEventStatusesEqual(iMEStatuses) {
		patch := redis.NewPatch()
		redis.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, redis, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", redis.Status.ID,
			"events", redis.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *RedisReconciler) handleExternalDelete(ctx context.Context, redis *v1beta1.Redis) error {
	l := log.FromContext(ctx)

	patch := redis.NewPatch()
	redis.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, redis, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(redis, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.StatusChecker))

	return nil
}
