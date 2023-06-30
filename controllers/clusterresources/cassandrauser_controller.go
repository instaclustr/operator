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

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// CassandraUserReconciler reconciles a CassandraUser object
type CassandraUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=cassandrausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=cassandrausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=cassandrausers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CassandraUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	u := &v1beta1.CassandraUser{}
	err := r.Get(ctx, req.NamespacedName, u)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			l.Info("Cassandra user resource is not found", "request", req)

			return models.ExitReconcile, nil
		}
		l.Error(err, "Cannot fetch Cassandra user resource", "request", req)

		return models.ReconcileRequeue, nil
	}

	s := &k8sCore.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: u.Spec.SecretRef.Namespace,
		Name:      u.Spec.SecretRef.Name,
	}, s)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			l.Info("Cassandra user secret is not found", "request", req)
			r.EventRecorder.Event(u, models.Warning, models.NotFound,
				"Secret is not found, please create a new secret or set an actual reference")
			return models.ReconcileRequeue, nil
		}

		l.Error(err, "Cannot get Cassandra user secret", "user", u.Name)

		return models.ReconcileRequeue, nil
	}

	if s.Labels == nil {
		s.Labels = map[string]string{}
	}

	if u.DeletionTimestamp != nil {
		err = r.handleDeleteUser(ctx, l, s, u)
		if err != nil {
			return models.ReconcileRequeue, nil
		}

		return models.ExitReconcile, nil
	}

	if u.GetAnnotations()[models.ResourceStateAnnotation] == models.SecretEvent {
		l.Info("Secret of Cassandra user has been updated", "secret reference", u.Spec.SecretRef)

		if s.DeletionTimestamp != nil {
			err = r.handleDeleteUser(ctx, l, s, u)
			if err != nil {
				return models.ReconcileRequeue, nil
			}

			return models.ExitReconcile, nil
		}

		patch := u.NewPatch()
		u.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatedSecret
		err = r.Patch(ctx, u, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra user on secret update", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		return models.ExitReconcile, nil
	}

	if u.Status.ClusterID != "" && u.Status.State != models.Created {
		patch := u.NewPatch()

		username, password, err := r.getUserCreds(s)
		if err != nil {
			l.Error(err, "Cannot get user credentials", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
				"Cannot get user credentials. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		iu := u.ToInstAPI(username, password)
		err = r.API.CreateUser(iu, u.Status.ClusterID, instaclustr.CassandraBundleUser)
		if err != nil {
			l.Error(err, "Cannot create Cassandra user", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
				"Cannot create user. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		u.Status.State = models.Created

		err = r.Status().Patch(ctx, u, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra user status", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		l.Info("User has been created", "username", iu.Username)
		r.EventRecorder.Eventf(u, models.Normal, models.Created,
			"User has been created for a cluster. Cluster ID: %s, username: %s", u.Status.ClusterID, iu.Username)

		return models.ExitReconcile, nil
	}

	if s.Labels[models.CassandraUserNamespaceLabel] == "" || s.Labels[models.ControlledByLabel] == "" {
		s.Labels[models.CassandraUserNamespaceLabel] = u.Namespace
		s.Labels[models.ControlledByLabel] = u.Name
		controllerutil.AddFinalizer(s, models.DeletionFinalizer)

		err = r.Update(ctx, s)
		if err != nil {
			l.Error(err, "Cannot update Cassandra user secret", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.UpdatedEvent,
				"Cannot assign Cassandra user to a k8s secret. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		patch := u.NewPatch()

		finalizerNeeded := controllerutil.AddFinalizer(u, models.DeletionFinalizer)
		if finalizerNeeded {
			err = r.Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user resource", "user", u.Name)
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}
		}

		l.Info("User has been attached to the secret", "user", u.Name)
		r.EventRecorder.Event(u, models.Normal, models.Created, "User has been attached to the secret")
	}

	return models.ExitReconcile, nil
}

func (r *CassandraUserReconciler) getUserCreds(secret *k8sCore.Secret) (username, password string, err error) {
	password = string(secret.Data[models.Password])
	username = string(secret.Data[models.Username])

	if len(username) == 0 || len(password) == 0 {
		return "", "", models.ErrMissingSecretKeys
	}

	return username[:len(username)-1], password[:len(password)-1], nil
}

func (r *CassandraUserReconciler) handleDeleteUser(
	ctx context.Context,
	l logr.Logger,
	s *k8sCore.Secret,
	u *v1beta1.CassandraUser,
) error {
	username, _, err := r.getUserCreds(s)
	if err != nil {
		l.Error(err, "Cannot get user credentials", "user", u.Name)
		r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
			"Cannot get user credentials. Reason: %v", err)

		return err
	}

	if u.Status.ClusterID != "" {
		err := r.API.DeleteUser(username, u.Status.ClusterID, instaclustr.CassandraBundleUser)
		if err != nil {
			l.Error(err, "Cannot delete user", "user", u.Name)
			r.EventRecorder.Eventf(u, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v", err)
			return err
		}
	}
	l.Info("Cassandra user has been deleted", "username", username)

	p := u.NewPatch()

	controllerutil.RemoveFinalizer(u, models.DeletionFinalizer)

	err = r.Patch(context.Background(), u, p)
	if err != nil {
		l.Error(err, "Cannot patch Cassandra user resource", "user", u.Name)

		r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v", err)
		return err
	}

	s.Labels[models.CassandraUserNamespaceLabel] = ""
	s.Labels[models.ControlledByLabel] = ""
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

func (r *CassandraUserReconciler) newUserReconcileRequest(secret client.Object) []reconcile.Request {
	userNamespacedName := types.NamespacedName{
		Namespace: secret.GetLabels()[models.CassandraUserNamespaceLabel],
		Name:      secret.GetLabels()[models.ControlledByLabel],
	}

	user := &v1beta1.CassandraUser{}

	err := r.Get(context.Background(), userNamespacedName, user)
	if err != nil {
		return nil
	}

	patch := user.NewPatch()

	annots := user.GetAnnotations()
	if annots == nil {
		annots = map[string]string{}
	}
	annots[models.ResourceStateAnnotation] = models.SecretEvent

	err = r.Patch(context.Background(), user, patch)
	if err != nil {
		return []reconcile.Request{}
	}

	return []reconcile.Request{{NamespacedName: userNamespacedName}}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.CassandraUser{}).
		Watches(&source.Kind{Type: &k8sCore.Secret{}}, handler.EnqueueRequestsFromMapFunc(r.newUserReconcileRequest)).
		Complete(r)
}
