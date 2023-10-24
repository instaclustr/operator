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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=redis/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	redis := &v1beta1.Redis{}
	err := r.Client.Get(ctx, req.NamespacedName, redis)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Redis cluster resource is not found",
				"resource name", req.NamespacedName,
			)
			return models.ExitReconcile, nil
		}

		logger.Error(err, "Unable to fetch Redis cluster",
			"resource name", req.NamespacedName,
		)
		return models.ExitReconcile, nil
	}

	switch redis.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, redis, logger)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, redis, logger)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, redis, logger)
	case models.GenericEvent:
		logger.Info("Redis generic event isn't handled",
			"cluster name", redis.Spec.Name,
			"request", req,
		)
		return models.ExitReconcile, nil
	default:
		logger.Info("Unknown event isn't handled",
			"cluster name", redis.Spec.Name,
			"request", req,
		)
		return models.ExitReconcile, nil
	}
}

func (r *RedisReconciler) handleCreateCluster(
	ctx context.Context,
	redis *v1beta1.Redis,
	logger logr.Logger,
) (reconcile.Result, error) {
	var err error
	if redis.Status.ID == "" {
		var id string
		if redis.Spec.HasRestore() {
			logger.Info(
				"Creating Redis cluster from backup",
				"original cluster ID", redis.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreCluster(redis.RestoreInfoToInstAPI(redis.Spec.RestoreFrom), models.RedisAppKind)
			if err != nil {
				logger.Error(
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

			logger.Info(
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
			logger.Info(
				"Creating Redis cluster",
				"cluster name", redis.Spec.Name,
				"data centres", redis.Spec.DataCentres,
			)

			id, err = r.API.CreateCluster(instaclustr.RedisEndpoint, redis.Spec.ToInstAPI())
			if err != nil {
				logger.Error(
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

			logger.Info(
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
			logger.Error(err, "Cannot patch Redis cluster status",
				"cluster name", redis.Spec.Name)
			r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		logger.Info("Redis resource has been created",
			"cluster name", redis.Name,
			"cluster ID", redis.Status.ID,
			"api version", redis.APIVersion)
	}

	patch := redis.NewPatch()
	controllerutil.AddFinalizer(redis, models.DeletionFinalizer)
	redis.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster",
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
			logger.Error(err, "Cannot start cluster status job",
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

		err = r.startClusterBackupsJob(redis)
		if err != nil {
			logger.Error(err, "Cannot start Redis cluster backups check job",
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
				logger.Error(err, "Failed to start user creation job")
				r.EventRecorder.Eventf(redis, models.Warning, models.CreationFailed,
					"User creation job is failed. Reason: %v", err,
				)
				return reconcile.Result{}, err
			}

			r.EventRecorder.Event(redis, models.Normal, models.Created,
				"Cluster user creation job is started")
		}
	}

	logger.Info(
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
	logger logr.Logger,
) (reconcile.Result, error) {
	iData, err := r.API.GetRedis(redis.Status.ID)
	if err != nil {
		logger.Error(
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
		logger.Error(
			err, "Cannot convert Redis cluster from the Instaclustr API",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID,
		)

		r.EventRecorder.Eventf(
			redis, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if redis.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(redis, iRedis, logger)
	}

	if redis.Spec.ClusterSettingsNeedUpdate(iRedis.Spec.Cluster) {
		logger.Info("Updating cluster settings",
			"instaclustr description", iRedis.Spec.Description,
			"instaclustr two factor delete", iRedis.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(redis.Status.ID, redis.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			logger.Error(err, "Cannot update cluster settings",
				"cluster ID", redis.Status.ID, "cluster spec", redis.Spec)
			r.EventRecorder.Eventf(redis, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !redis.Spec.IsEqual(iRedis.Spec) {
		logger.Info("Update request to Instaclustr API has been sent",
			"spec data centres", redis.Spec.DataCentres,
			"resize settings", redis.Spec.ResizeSettings,
		)

		err = r.API.UpdateRedis(redis.Status.ID, redis.Spec.DCsUpdateToInstAPI())
		if err != nil {
			logger.Error(err, "Cannot update Redis cluster data centres",
				"cluster name", redis.Spec.Name,
				"cluster status", redis.Status,
				"data centres", redis.Spec.DataCentres,
			)

			r.EventRecorder.Eventf(
				redis, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v",
				err,
			)

			patch := redis.NewPatch()
			redis.Annotations[models.UpdateQueuedAnnotation] = models.True
			err = r.Patch(ctx, redis, patch)
			if err != nil {
				logger.Error(err, "Cannot patch metadata",
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
			return reconcile.Result{}, err
		}
	}

	patch := redis.NewPatch()
	redis.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	redis.Annotations[models.UpdateQueuedAnnotation] = ""
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster after update",
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

	logger.Info(
		"Cluster has been updated",
		"cluster name", redis.Spec.Name,
		"cluster ID", redis.Status.ID,
		"data centres", redis.Spec.DataCentres,
	)

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) handleCreateUsers(
	ctx context.Context,
	redis *v1beta1.Redis,
	l logr.Logger,
	userRefs *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: userRefs.Namespace,
		Name:      userRefs.Name,
	}

	u := &clusterresourcesv1beta1.RedisUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Redis user is not found", "request", req)
			r.EventRecorder.Event(u, models.Warning, "Not Found",
				"User is not found, create a new one Redis User or provide right userRef.")

			return err
		}

		l.Error(err, "Cannot get Redis user secret", "request", req)
		r.EventRecorder.Eventf(u, models.Warning, "Cannot Get",
			"Cannot get Redis user secret. Reason: %v", err)

		return err
	}

	if _, exist := u.Status.ClustersEvents[redis.Status.ID]; exist {
		l.Info("User is already existing on the cluster",
			"user reference", userRefs)
		r.EventRecorder.Eventf(redis, models.Normal, models.CreationFailed,
			"User is already existing on the cluster. User reference: %v", userRefs)

		return nil
	}

	patch := u.NewPatch()

	if u.Status.ClustersEvents == nil {
		u.Status.ClustersEvents = make(map[string]string)
	}

	u.Status.ClustersEvents[redis.Status.ID] = models.CreatingEvent

	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the Redis User status with the CreatingEvent",
			"cluster name", redis.Spec.Name, "cluster ID", redis.Status.ID)
		r.EventRecorder.Eventf(redis, models.Warning, models.CreationFailed,
			"Cannot add Redis User to the cluster. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for creation", "username", u.Name)

	return nil
}

func (r *RedisReconciler) handleExternalChanges(redis, iRedis *v1beta1.Redis, l logr.Logger) (reconcile.Result, error) {
	if !redis.Spec.IsEqual(iRedis.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", redis.Spec,
			"data from Instaclustr ", iRedis.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(redis.Spec, iRedis.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iRedis.Spec, "k8s resource spec", redis.Spec)
			return models.ExitReconcile, nil
		}
		r.EventRecorder.Eventf(redis, models.Warning, models.ExternalChanges, msgDiffSpecs)
		return models.ExitReconcile, nil
	}

	patch := redis.NewPatch()

	redis.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), redis, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", redis.Spec.Name, "cluster ID", redis.Status.ID)

		r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info("External changes have been reconciled", "resource ID", redis.Status.ID)
	r.EventRecorder.Event(redis, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) handleDeleteCluster(
	ctx context.Context,
	redis *v1beta1.Redis,
	logger logr.Logger,
) (reconcile.Result, error) {

	_, err := r.API.GetRedis(redis.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get Redis cluster status from Instaclustr",
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
		logger.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", redis.Spec.Name,
			"cluster ID", redis.Status.ID)

		err = r.API.DeleteCluster(redis.Status.ID, instaclustr.RedisEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete Redis cluster",
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
				logger.Error(err, "Cannot patch cluster resource",
					"cluster name", redis.Spec.Name,
					"cluster state", redis.Status.State)
				r.EventRecorder.Eventf(redis, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return reconcile.Result{}, err
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", redis.Status.ID)

			r.EventRecorder.Event(redis, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.StatusChecker))
	r.Scheduler.RemoveJob(redis.GetJobID(scheduler.BackupsChecker))

	logger.Info("Deleting cluster backup resources",
		"cluster ID", redis.Status.ID,
	)

	err = r.deleteBackups(ctx, redis.Status.ID, redis.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete cluster backup resources",
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

	logger.Info("Cluster backup resources are deleted",
		"cluster ID", redis.Status.ID,
	)

	for _, ref := range redis.Spec.UserRefs {
		err = r.detachUserResource(ctx, logger, redis, ref)
		if err != nil {
			logger.Error(err, "Cannot detach Redis user",
				"cluster name", redis.Spec.Name,
				"cluster status", redis.Status.State,
			)

			r.EventRecorder.Eventf(
				redis, models.Warning, models.DeletionFailed,
				"Cluster detaching on the Instaclustr is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}
	}

	patch := redis.NewPatch()
	controllerutil.RemoveFinalizer(redis, models.DeletionFinalizer)
	redis.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, redis, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Redis cluster metadata after finalizer removal",
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
		logger.Error(err, "Cannot delete Redis cluster expose service",
			"cluster ID", redis.Status.ID,
			"cluster name", redis.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	logger.Info("Redis cluster was deleted",
		"cluster name", redis.Spec.Name,
		"cluster ID", redis.Status.ID,
	)

	r.EventRecorder.Eventf(
		redis, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *RedisReconciler) detachUserResource(
	ctx context.Context,
	l logr.Logger,
	redis *v1beta1.Redis,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.RedisUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Redis user is not found", "request", req)
			r.EventRecorder.Eventf(redis, models.Warning, models.NotFound,
				"User resource is not found, please provide correct userRef."+
					"Current provided reference: %v", uRef)
			return nil
		}

		l.Error(err, "Cannot get Redis user", "user", u.Spec)
		r.EventRecorder.Eventf(redis, models.Warning, models.DeletionFailed,
			"Cannot get Redis user. User reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[redis.Status.ID]; !exist {
		return nil
	}

	patch := u.NewPatch()
	u.Status.ClustersEvents[redis.Status.ID] = models.ClusterDeletingEvent
	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the Redis user status with the ClusterDeletingEvent",
			"cluster name", redis.Spec.Name, "cluster ID", redis.Status.ID)
		r.EventRecorder.Eventf(redis, models.Warning, models.DeletionFailed,
			"Cannot patch the Redis user status with the ClusterDeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("The user has been detached from the cluster")

	return nil
}

func (r *RedisReconciler) handleUserEvent(
	newObj *v1beta1.Redis,
	oldUsers []*v1beta1.UserReference,
) {
	ctx := context.TODO()
	l := log.FromContext(ctx)

	for _, newUser := range newObj.Spec.UserRefs {
		var exist bool

		for _, oldUser := range oldUsers {

			if *newUser == *oldUser {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		err := r.handleCreateUsers(ctx, newObj, l, newUser)
		if err != nil {
			l.Error(err, "Cannot create Redis user in predicate", "user", newUser)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot create user. Reason: %v", err)
		}

		oldUsers = append(oldUsers, newUser)
	}

	for _, oldUser := range oldUsers {
		var exist bool

		for _, newUser := range newObj.Spec.UserRefs {

			if *oldUser == *newUser {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		err := r.handleUsersDelete(ctx, l, newObj, oldUser)
		if err != nil {
			l.Error(err, "Cannot delete Redis user", "user", oldUser)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot delete user from cluster. Reason: %v", err)
		}
	}
}

func (r *RedisReconciler) handleUsersDelete(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Redis,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.RedisUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Redis user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User is not found, create a new one Redis User or provide correct userRef."+
					"Current provided reference: %v", uRef)
			return nil
		}

		l.Error(err, "Cannot get Redis user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot get Redis user. User reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; !exist {
		l.Info("User is not existing on the cluster",
			"user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.DeletionFailed,
			"User is not existing on the cluster. User reference: %v", uRef)

		return nil
	}

	patch := u.NewPatch()

	u.Status.ClustersEvents[c.Status.ID] = models.DeletingEvent

	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the Redis User status with the DeletingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot patch the Redis User status with the DeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for deletion", "username", u.Name)

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
	logger := log.Log.WithValues("component", "redisUsersCreationJob")

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
			logger.Info("User creation job is scheduled")
			r.EventRecorder.Eventf(redis, models.Normal, models.CreationFailed,
				"User creation job is scheduled, cluster is not in the running state",
			)
			return nil
		}

		for _, ref := range redis.Spec.UserRefs {
			err = r.handleCreateUsers(ctx, redis, logger, ref)
			if err != nil {
				logger.Error(err, "Failed to create a user for the cluster",
					"user ref", ref,
				)
				r.EventRecorder.Eventf(redis, models.Warning, models.CreationFailed,
					"Failed to create a user for the cluster. Reason: %v", err,
				)
				return err
			}
		}

		logger.Info("User creation job successfully finished")
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
			redis.Annotations[models.UpdateQueuedAnnotation] != models.True &&
			!redis.Spec.IsEqual(iRedis.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iRedis.Spec, "k8s resource spec", redis.Spec)

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

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
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

				oldObj := event.ObjectOld.(*v1beta1.Redis)

				r.handleUserEvent(newObj, oldObj.Spec.UserRefs)

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
