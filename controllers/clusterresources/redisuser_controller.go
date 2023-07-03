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

	k8sCore "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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

	passwordSecret := &k8sCore.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: user.Spec.PasswordSecretNamespace,
		Name:      user.Spec.PasswordSecretName,
	}, passwordSecret)
	if err != nil {
		l.Error(err, "Cannot get Redis user password secret",
			"secret name", user.Spec.PasswordSecretName,
			"secret namespace", user.Spec.PasswordSecretNamespace,
			"username", user.Spec.Username,
			"cluster ID", user.Spec.ClusterID,
			"user ID", user.Status.ID,
		)

		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch default user password secret is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue, nil
	}

	if user.DeletionTimestamp != nil {
		err = r.API.DeleteRedisUser(user.Status.ID)
		if err != nil && !errors.Is(err, instaclustr.NotFound) {
			l.Error(err, "Cannot delete Redis user",
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
				"user ID", user.Status.ID,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		r.EventRecorder.Eventf(
			user, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)

		patch := user.NewPatch()
		controllerutil.RemoveFinalizer(user, models.DeletionFinalizer)
		passwordSecret.Labels[models.RedisUserNamespaceLabel] = ""
		passwordSecret.Labels[models.ControlledByLabel] = ""
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Redis user resource",
				"secret name", user.Spec.PasswordSecretName,
				"secret namespace", user.Spec.PasswordSecretNamespace,
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
				"status", user.Status,
			)

			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		l.Info("Redis user is deleted",
			"secret name", user.Spec.PasswordSecretName,
			"secret namespace", user.Spec.PasswordSecretNamespace,
			"username", user.Spec.Username,
			"cluster ID", user.Spec.ClusterID,
			"user ID", user.Status.ID,
		)

		r.EventRecorder.Eventf(
			user, models.Normal, models.Deleted,
			"Resource is deleted",
		)

		return models.ExitReconcile, nil
	}

	password := r.getPassword(passwordSecret)
	if user.Status.ID == "" {
		patch := user.NewPatch()
		if len(passwordSecret.Labels) == 0 {
			passwordSecret.Labels = map[string]string{}
		}
		passwordSecret.Labels[models.RedisUserNamespaceLabel] = user.Namespace
		passwordSecret.Labels[models.ControlledByLabel] = user.Name
		err = r.Update(ctx, passwordSecret)
		if err != nil {
			l.Error(err, "Cannot update Redis user password secret",
				"secret name", user.Spec.PasswordSecretName,
				"secret namespace", user.Spec.PasswordSecretNamespace,
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
				"user ID", user.Status.ID,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.UpdateFailed,
				"Resource update is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		user.Status.ID, err = r.API.CreateRedisUser(user.Spec.ToInstAPI(password))
		if err != nil {
			l.Error(err, "Cannot create Redis user",
				"secret name", user.Spec.PasswordSecretName,
				"secret namespace", user.Spec.PasswordSecretNamespace,
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
			)

			r.EventRecorder.Eventf(
				user, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		r.EventRecorder.Eventf(
			user, models.Normal, models.Created,
			"Resource creation request is sent. User ID: %s",
			user.Status.ID,
		)

		err = r.Status().Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Redis user status",
				"secret name", user.Spec.PasswordSecretName,
				"secret namespace", user.Spec.PasswordSecretNamespace,
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
				"status", user.Status,
			)

			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		controllerutil.AddFinalizer(user, models.DeletionFinalizer)
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Redis user resource",
				"secret name", user.Spec.PasswordSecretName,
				"secret namespace", user.Spec.PasswordSecretNamespace,
				"username", user.Spec.Username,
				"cluster ID", user.Spec.ClusterID,
				"status", user.Status,
			)

			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		l.Info("Redis user is created",
			"secret name", user.Spec.PasswordSecretName,
			"secret namespace", user.Spec.PasswordSecretNamespace,
			"username", user.Spec.Username,
			"cluster ID", user.Spec.ClusterID,
			"user ID", user.Status.ID,
		)

		return models.ExitReconcile, nil
	}

	err = r.API.UpdateRedisUser(user.ToInstAPIUpdate(password))
	if err != nil {
		l.Error(err, "Cannot update redis user",
			"secret name", user.Spec.PasswordSecretName,
			"secret namespace", user.Spec.PasswordSecretNamespace,
			"username", user.Spec.Username,
			"cluster ID", user.Spec.ClusterID,
			"user ID", user.Status.ID,
		)

		r.EventRecorder.Eventf(
			user, models.Warning, models.UpdateFailed,
			"Resource update on the Instacluter is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue, nil
	}

	l.Info("Redis user is updated",
		"secret name", user.Spec.PasswordSecretName,
		"secret namespace", user.Spec.PasswordSecretNamespace,
		"username", user.Spec.Username,
		"cluster ID", user.Spec.ClusterID,
		"user ID", user.Status.ID,
	)

	return models.ExitReconcile, nil
}

func (r *RedisUserReconciler) getPassword(secret *k8sCore.Secret) string {
	password := secret.Data[models.Password]

	return string(password[:len(password)-1])
}

func (r *RedisUserReconciler) newUserReconcileRequest(secretObj client.Object) []reconcile.Request {
	user := &v1beta1.RedisUser{}
	userNamespacedName := types.NamespacedName{
		Namespace: secretObj.GetLabels()[models.RedisUserNamespaceLabel],
		Name:      secretObj.GetLabels()[models.ControlledByLabel],
	}

	err := r.Get(context.TODO(), userNamespacedName, user)
	if err != nil {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: userNamespacedName,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.RedisUser{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return event.ObjectNew.GetGeneration() != event.ObjectOld.GetGeneration()
			},
		})).
		Watches(
			&source.Kind{Type: &k8sCore.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.newUserReconcileRequest),
		).
		Complete(r)
}
