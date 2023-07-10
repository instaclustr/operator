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
	"sigs.k8s.io/controller-runtime/pkg/log"

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

		l.Error(err, "Cannot get Cassandra user secret", "request", req)
		r.EventRecorder.Eventf(u, models.Warning, models.NotFound,
			"Cannot get user secret. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	patch := u.NewPatch()

	username, password, err := getUserCreds(s)
	if err != nil {
		l.Error(err, "Cannot get the Cassandra user credentials from the secret",
			"secret name", s.Name,
			"secret namespace", s.Namespace)
		r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
			"Cannot get the Cassandra user credentials from the secret. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	for clusterID, event := range u.Status.ClustersEvents {
		if event == models.CreatingEvent {
			err = r.API.CreateUser(u.ToInstAPI(username, password), clusterID, instaclustr.CassandraBundleUser)
			if err != nil {
				l.Error(err, "Cannot create a user for the Cassandra cluster",
					"cluster ID", clusterID,
					"username", username)
				r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
					"Cannot create user. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			event = models.Created
			u.Status.ClustersEvents[clusterID] = event

			err = r.Status().Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user status")
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			l.Info("User has been created", "username", username)
			r.EventRecorder.Eventf(u, models.Normal, models.Created,
				"User has been created for a cluster. Cluster ID: %s, username: %s",
				clusterID, username)

			finalizerNeeded := controllerutil.AddFinalizer(s, models.DeletionFinalizer)
			if finalizerNeeded {
				err = r.Update(ctx, s)
				if err != nil {
					l.Error(err, "Cannot update Cassandra user secret",
						"secret name", s.Name,
						"secret namespace", s.Namespace)
					r.EventRecorder.Eventf(u, models.Warning, models.UpdatedEvent,
						"Cannot assign Cassandra user to a k8s secret. Reason: %v", err)

					return models.ReconcileRequeue, nil
				}
			}

			controllerutil.AddFinalizer(u, models.DeletionUserFinalizer+clusterID)
			err = r.Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user resource",
					"secret name", s.Name,
					"secret namespace", s.Namespace)
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			continue
		}

		if event == models.DeletingEvent {
			err = r.API.DeleteUser(username, clusterID, instaclustr.CassandraBundleUser)
			if err != nil {
				l.Error(err, "Cannot delete Cassandra user")
				r.EventRecorder.Eventf(u, models.Warning, models.DeletingEvent,
					"Cannot delete user. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			l.Info("User has been deleted for cluster", "username", username,
				"cluster ID", clusterID)
			r.EventRecorder.Eventf(u, models.Normal, models.Deleted,
				"User has been deleted for a cluster. Cluster ID: %s, username: %s",
				clusterID, username)

			delete(u.Status.ClustersEvents, clusterID)

			err = r.Status().Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user status")
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			controllerutil.RemoveFinalizer(u, models.DeletionUserFinalizer+clusterID)
			err = r.Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra user resource",
					"secret name", s.Name,
					"secret namespace", s.Namespace)
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			continue
		}
	}

	if u.DeletionTimestamp != nil {
		err = r.handleDeleteUser(ctx, l, s, u)
		if err != nil {
			return models.ReconcileRequeue, nil
		}
	}

	return models.ExitReconcile, nil
}

func (r *CassandraUserReconciler) handleDeleteUser(
	ctx context.Context,
	l logr.Logger,
	s *k8sCore.Secret,
	u *v1beta1.CassandraUser,
) error {
	username, _, err := getUserCreds(s)
	if err != nil {
		l.Error(err, "Cannot get user credentials")
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

	l.Info("Cassandra user has been deleted", "username", username)

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
func (r *CassandraUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.CassandraUser{}).
		Complete(r)
}
