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
	"fmt"

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
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

			return models.ExitReconcile, nil
		}

		l.Error(err, "Cannot fetch Redis user resource",
			"request", req,
		)
		return models.ReconcileRequeue, nil
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

		return models.ReconcileRequeue, nil
	}

	username, password, err := getUserCreds(secret)
	if err != nil {
		l.Error(err, "Cannot get user credentials", "user", username)
		r.EventRecorder.Eventf(user, models.Warning, models.CreatingEvent,
			"Cannot get user. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	for clusterID, event := range user.Status.ClustersEvents {
		if event == models.CreatingEvent {
			patch := user.NewPatch()

			_, err = r.API.CreateRedisUser(user.Spec.ToInstAPI(password, clusterID, username))
			if err != nil {
				l.Error(err, "Cannot create a user for the Redis cluster",
					"cluster ID", clusterID,
					"username", username)
				r.EventRecorder.Eventf(user, models.Warning, models.CreatingEvent,
					"Cannot create user. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			event = models.Created
			user.Status.ClustersEvents[clusterID] = event

			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Redis user status", "user", user.Name)
				r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			l.Info("User has been created for a Redis cluster", "username", username)
			r.EventRecorder.Eventf(user, models.Normal, models.Created,
				"User has been created for a Redis cluster. Cluster ID: %s, username: %s",
				clusterID, username)

			finalizerNeeded := controllerutil.AddFinalizer(secret, models.DeletionFinalizer)
			if finalizerNeeded {
				err = r.Update(ctx, secret)
				if err != nil {
					l.Error(err, "Cannot update Cassandra user secret",
						"secret name", secret.Name,
						"secret namespace", secret.Namespace)
					r.EventRecorder.Eventf(user, models.Warning, models.UpdatedEvent,
						"Cannot assign Cassandra user to a k8s secret. Reason: %v", err)

					return models.ReconcileRequeue, nil
				}
			}

			controllerutil.AddFinalizer(user, user.DeletionUserFinalizer(clusterID, username))
			err = r.Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user resource",
					"secret name", secret.Name,
					"secret namespace", secret.Namespace)
				r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			continue
		}

		if event == models.DeletingEvent {
			patch := user.NewPatch()

			userID := fmt.Sprintf(instaclustr.RedisUserIDFmt, clusterID, user.Namespace)
			err = r.API.DeleteRedisUser(userID)
			if err != nil {
				l.Error(err, "Cannot delete Redis user", "user", user.Name)
				r.EventRecorder.Eventf(user, models.Warning, models.DeletingEvent,
					"Cannot delete user. Reason: %v", err)

				return models.ReconcileRequeue, nil
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

				return models.ReconcileRequeue, nil
			}

			controllerutil.RemoveFinalizer(user, user.DeletionUserFinalizer(clusterID, user.Namespace))
			err = r.Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user resource",
					"secret name", secret.Name,
					"secret namespace", secret.Namespace)
				r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			continue
		}

		userID := fmt.Sprintf(instaclustr.RedisUserIDFmt, clusterID, username)
		err = r.API.UpdateRedisUser(user.ToInstAPIUpdate(password, userID))
		if err != nil {
			l.Error(err, "Cannot update redis user",
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

			return models.ReconcileRequeue, nil
		}

		l.Info("Redis user has been updated",
			"secret name", user.Spec.SecretRef.Name,
			"secret namespace", user.Spec.SecretRef.Namespace,
			"username", username,
			"cluster ID", clusterID,
			"user ID", userID,
		)

		return models.ExitReconcile, nil
	}

	if user.DeletionTimestamp != nil {
		err = r.handleDeleteUser(ctx, l, secret, user)
		if err != nil {
			return models.ReconcileRequeue, nil
		}
	}

	return models.ExitReconcile, nil
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
			l.Error(models.ErrUserStillExist, "please remove the user from the cluster specification",
				"username", username, "cluster ID", clusterID)
			r.EventRecorder.Event(u, models.Warning, models.DeletingEvent,
				"The user is still attached to cluster, please remove the user from the cluster specification.")

			return models.ErrUserStillExist
		}
	}

	controllerutil.RemoveFinalizer(s, models.DeletionFinalizer)
	err = r.Update(ctx, s)
	if err != nil {
		l.Error(err, "Cannot remove finalizer from secret", "secret name", s.Name)

		r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v", err)

		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.RedisUser{}).
		Complete(r)
}
