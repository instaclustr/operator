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

package clusterresources

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

// RedisUserReconciler reconciles a RedisUser object
type RedisUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=redisusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=redisusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=redisusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RedisUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	user := &v1beta1.RedisUser{}

	err := r.Get(ctx, req.NamespacedName, user)

	if err != nil {
		if k8sErrors.IsNotFound(err) {
			l.Info("Redis user resource is not found",
				"request", req,
			)

			return ctrl.Result{}, nil
		}

		l.Error(err, "Cannot fetch Redis user resource",
			"request", req,
		)
		return ctrl.Result{}, err
	}

	secret := &k8sCore.Secret{}

	err = r.Get(ctx, types.NamespacedName{
		Namespace: user.Spec.SecretRef.Namespace,
		Name:      user.Spec.SecretRef.Name,
	}, secret)
	if err != nil {
		l.Error(err, "Cannot get Redis user password secret",
			"secret name", user.Spec.SecretRef.Name,
			"secret namespace", user.Spec.SecretRef.Namespace,
			"user ID", user.Status.ID,
		)

		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch default user password secret is failed. Reason: %v",
			err,
		)

		return ctrl.Result{}, err
	}

	username, password, err := getUserCreds(secret)
	if err != nil {
		l.Error(err, "Cannot get user credentials", "user", username)
		r.EventRecorder.Eventf(user, models.Warning, models.CreatingEvent,
			"Cannot get user. Reason: %v", err)

		return ctrl.Result{}, err
	}

	if controllerutil.AddFinalizer(secret, user.GetDeletionFinalizer()) {
		err = r.Update(ctx, secret)
		if err != nil {
			l.Error(err, "Cannot update Redis user secret with deletion finalizer",
				"secret name", secret.Name,
				"secret namespace", secret.Namespace)
			r.EventRecorder.Eventf(user, models.Warning, models.UpdatedEvent,
				"Cannot assign k8s secret to a Redis user. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	patch := user.NewPatch()
	if controllerutil.AddFinalizer(user, user.GetDeletionFinalizer()) {
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Patch is failed. Cannot set finalizer to the user")
			r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
				"Patch is failed. Cannot set finalizer to the user. Reason: %v", err)
			return ctrl.Result{}, err
		}
	}

	for clusterID, event := range user.Status.ClustersEvents {
		userID := fmt.Sprintf(instaclustr.RedisUserIDFmt, clusterID, username)

		if event == models.CreatingEvent {
			_, err = r.API.GetRedisUser(userID)
			if err != nil && !errors.Is(err, instaclustr.NotFound) {
				l.Error(err, "Cannot check if the user already exists on the cluster",
					"cluster ID", clusterID,
					"username", username)
				r.EventRecorder.Eventf(user, models.Warning, models.CreatingEvent,
					"Cannot create user. Reason: %v", err)

				return ctrl.Result{}, err
			}

			if errors.Is(err, instaclustr.NotFound) {
				_, err = r.API.CreateRedisUser(user.Spec.ToInstAPI(password, clusterID, username))
				if err != nil {
					l.Error(err, "Cannot create a user for the Redis cluster",
						"cluster ID", clusterID,
						"username", username)
					r.EventRecorder.Eventf(user, models.Warning, models.CreatingEvent,
						"Cannot create user. Reason: %v", err)

					return ctrl.Result{}, err
				}
			}

			user.Status.ClustersEvents[clusterID] = models.Created
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Redis user status", "user", user.Name)
				r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return ctrl.Result{}, err
			}

			l.Info("User has been created for a Redis cluster", "username", username)
			r.EventRecorder.Eventf(user, models.Normal, models.Created,
				"User has been created for a Redis cluster. Cluster ID: %s, username: %s",
				clusterID, username)

			continue
		}

		if event == models.DeletingEvent {
			userID := fmt.Sprintf(instaclustr.RedisUserIDFmt, clusterID, username)
			err = r.API.DeleteRedisUser(userID)
			if err != nil && !errors.Is(err, instaclustr.NotFound) {
				l.Error(err, "Cannot delete Redis user from the cluster.",
					"cluster ID", clusterID, "user ID", userID)
				r.EventRecorder.Eventf(user, models.Warning, models.DeletingEvent,
					"Cannot delete Redis user from the cluster. Cluster ID: %s. Reason: %v", clusterID, err)

				return ctrl.Result{}, err
			}

			l.Info("User has been deleted for cluster", "username", username,
				"cluster ID", clusterID)
			r.EventRecorder.Eventf(user, models.Normal, models.Deleted,
				"User has been deleted for a cluster. Cluster ID: %s, username: %s",
				clusterID, username)

			delete(user.Status.ClustersEvents, clusterID)

			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Redis user status", "user", user.Name)
				r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return ctrl.Result{}, err
			}

			continue
		}

		if event == models.ClusterDeletingEvent {
			err = r.detachUserFromDeletedCluster(ctx, clusterID, user, l)
			if err != nil {
				l.Error(err, "Cannot detach Redis user resource",
					"secret name", secret.Name,
					"secret namespace", secret.Namespace)
				r.EventRecorder.Eventf(user, models.Warning, models.DeletionFailed,
					"Resource detach is failed. Reason: %v", err)

				return ctrl.Result{}, err
			}
			continue
		}

		err = r.API.UpdateRedisUser(user.ToInstAPIUpdate(password, userID))
		if err != nil && !errors.Is(err, instaclustr.NotFound) {
			l.Error(err, "Cannot update redis user password",
				"secret name", user.Spec.SecretRef.Name,
				"secret namespace", user.Spec.SecretRef.Namespace,
				"username", username,
				"cluster ID", clusterID,
				"user ID", userID,
			)

			r.EventRecorder.Eventf(
				user, models.Warning, models.UpdateFailed,
				"Resource update on the Instacluter is failed. Reason: %v",
				err,
			)

			return ctrl.Result{}, err
		}
		if errors.Is(err, instaclustr.NotFound) {
			l.Info("Cannot update redis user password, the user doesn't exist on the given cluster",
				"secret name", user.Spec.SecretRef.Name,
				"secret namespace", user.Spec.SecretRef.Namespace,
				"username", username,
				"cluster ID", clusterID,
				"user ID", userID,
			)

			r.EventRecorder.Eventf(user, models.Warning, models.UpdateFailed,
				"Given user doesn`t exist on the cluster ID: %v", clusterID)
		}

		l.Info("Redis user has been updated",
			"secret name", user.Spec.SecretRef.Name,
			"secret namespace", user.Spec.SecretRef.Namespace,
			"username", username,
			"cluster ID", clusterID,
			"user ID", userID,
		)
	}

	if user.DeletionTimestamp != nil {
		err = r.handleDeleteUser(ctx, l, secret, user)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RedisUserReconciler) detachUserFromDeletedCluster(
	ctx context.Context,
	clusterID string,
	user *v1beta1.RedisUser,
	logger logr.Logger,
) error {

	patch := user.NewPatch()
	delete(user.Status.ClustersEvents, clusterID)

	err := r.Status().Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot detach clusterID from the Redis user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Detaching clusterID from the Redis user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	err = r.Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot delete finalizer from the Redis user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Deleting finalizer from the Redis user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	logger.Info("Redis user has been deleted from the cluster",
		"cluster ID", clusterID,
	)
	r.EventRecorder.Eventf(
		user, models.Normal, models.Deleted,
		"Redis user resource has been deleted from the cluster (clusterID: %v)", clusterID,
	)

	return nil
}

func (r *RedisUserReconciler) handleDeleteUser(
	ctx context.Context,
	l logr.Logger,
	s *k8sCore.Secret,
	u *v1beta1.RedisUser,
) error {
	username, _, err := getUserCreds(s)
	if err != nil {
		l.Error(err, "Cannot get user credentials", "user", u.Name)
		r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
			"Cannot get user credentials. Reason: %v", err)

		return err
	}

	for clusterID, event := range u.Status.ClustersEvents {
		if event == models.Created || event == models.CreatingEvent {
			l.Error(models.ErrUserStillExist, instaclustr.MsgDeleteUser,
				"username", username, "cluster ID", clusterID)
			r.EventRecorder.Event(u, models.Warning, models.DeletingEvent, instaclustr.MsgDeleteUser)

			return models.ErrUserStillExist
		}
	}

	controllerutil.RemoveFinalizer(s, u.GetDeletionFinalizer())
	err = r.Update(ctx, s)
	if err != nil {
		l.Error(err, "Cannot remove finalizer from secret", "secret name", s.Name)
		r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
			"Patch is failed. Cannot remove finalizer from secret. Reason: %v", err)

		return err
	}

	patch := u.NewPatch()
	controllerutil.RemoveFinalizer(u, u.GetDeletionFinalizer())
	err = r.Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot remove finalizer from user", "user finalizer", u.Finalizers)
		r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
			"Patch is failed. Cannot remove finalizer from user. Reason: %v", err)

		return err
	}

	l.Info("Redis User resource was deleted")

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.RedisUser{}).
		Complete(r)
}
