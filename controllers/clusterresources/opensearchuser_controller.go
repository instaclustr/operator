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

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// OpenSearchUserReconciler reconciles a OpenSearchUser object
type OpenSearchUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OpenSearchUserReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	user := &clusterresourcesv1beta1.OpenSearchUser{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			logger.Info("OpenSearch user resource is not found",
				"request", req,
			)
			return models.ExitReconcile, nil
		}
		logger.Error(err, "Cannot fetch OpenSearch user resource",
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
		if k8sErrors.IsNotFound(err) {
			logger.Info("OpenSearch user secret is not found",
				"request", req,
			)
			r.EventRecorder.Event(
				user, models.Warning, models.NotFound,
				"Secret is not found, please create a new secret or set an actual reference",
			)
			return models.ReconcileRequeue, nil
		}

		logger.Error(err, "Cannot get OpenSearch user's secret")
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"User's secret fetching has been failed. Reason: %v", err,
		)

		return models.ReconcileRequeue, nil
	}

	patch := client.MergeFrom(secret.DeepCopy())
	if controllerutil.AddFinalizer(secret, user.GetDeletionFinalizer()) {
		err = r.Patch(ctx, secret, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch user's secret with deletion finalizer")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Patching secret with deletion finalizer has been failed. Reason: %v", err,
			)
			return models.ReconcileRequeue, nil
		}
	}

	patch = user.NewPatch()
	if controllerutil.AddFinalizer(user, user.GetDeletionFinalizer()) {
		err = r.Patch(ctx, user, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch user with deletion finalizer")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Patching OpenSearch user with deletion finalizer has been failed. Reason: %v", err,
			)
			return models.ReconcileRequeue, nil
		}
	}

	errorOccurred := false

	for clusterID, event := range user.Status.ClustersEvents {
		switch event {
		case models.CreatingEvent:
			err = r.createUser(ctx, clusterID, user, secret, logger)
		case models.DeletingEvent:
			err = r.deleteUser(ctx, clusterID, user, secret, logger)
		case models.ClusterDeletingEvent:
			err = r.detachUserFromDeletedCluster(ctx, clusterID, user, logger)
		default:
			if event != models.Created && event != models.DeletedEvent {
				logger.Info("unhandled event has been occurred", "event", event)
			}
		}
		if err != nil {
			errorOccurred = true
		}
	}

	if errorOccurred {
		return models.ReconcileRequeue, nil
	}

	if user.DeletionTimestamp != nil {
		if user.Status.ClustersEvents != nil {
			logger.Error(models.ErrUserStillExist, "please remove the user from the cluster specification")
			r.EventRecorder.Event(user, models.Warning, models.DeletingEvent,
				"The user is still attached to cluster, please remove the user from the cluster specification.",
			)

			return models.ExitReconcile, nil
		}

		patch = client.MergeFrom(secret.DeepCopy())
		controllerutil.RemoveFinalizer(secret, user.GetDeletionFinalizer())
		err = r.Patch(ctx, secret, patch)
		if err != nil {
			logger.Error(err, "Cannot delete finalizer from the user's secret")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Deleting finalizer from the user's secret has been failed. Reason: %v", err,
			)
			return models.ReconcileRequeue, nil
		}

		patch = user.NewPatch()
		controllerutil.RemoveFinalizer(user, user.GetDeletionFinalizer())
		err = r.Patch(ctx, user, patch)
		if err != nil {
			logger.Error(err, "Cannot delete finalizer from the OpenSearch user resource")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Deleting finalizer  from the OpenSearch user resource has been failed. Reason: %v", err,
			)
			return models.ReconcileRequeue, nil
		}

		logger.Info("The user resource has been deleted")
	}

	return models.ExitReconcile, nil
}

func (r *OpenSearchUserReconciler) createUser(
	ctx context.Context,
	clusterID string,
	user *clusterresourcesv1beta1.OpenSearchUser,
	secret *k8sCore.Secret,
	logger logr.Logger,
) error {
	username, password, err := getUserCreds(secret)
	if err != nil {
		logger.Error(err, "Cannot get user's credentials during creating user on the cluster")
		r.EventRecorder.Eventf(
			user, models.Warning, models.CreatingEvent,
			"Cannot get user's credentials during creating user on the cluster. Reason: %v", err,
		)
		return err
	}

	err = r.API.CreateUser(user.ToInstaAPI(username, password), clusterID, models.OpenSearchAppKind)
	if err != nil {
		logger.Error(err, "Cannot create OpenSearch user on Instaclustr",
			"username", username,
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.CreationFailed,
			"OpenSearch user creating on Instaclustr has been failed. Reason: %v", err,
		)
		return err
	}

	patch := user.NewPatch()
	if controllerutil.AddFinalizer(user, getDeletionUserFinalizer(clusterID)) {
		err = r.Patch(ctx, user, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch user resource with deletion finalizer",
				"cluster ID", clusterID,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patching with deletion finalizer has been failed ). Reason: %v",
				err,
			)
			return err
		}
	}

	patch = user.NewPatch()
	user.Status.ClustersEvents[clusterID] = models.Created
	err = r.Status().Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot patch user resource with created state",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Resource patching with created state has been failed. Reason: %v", err,
		)
		return err
	}

	logger.Info("OpenSearch user has been created",
		"cluster ID", clusterID,
	)
	r.EventRecorder.Eventf(user, models.Normal, models.Created,
		"OpenSearch user resource has been created on the cluster with ID: %v",
		clusterID,
	)

	return nil
}

func (r *OpenSearchUserReconciler) deleteUser(
	ctx context.Context,
	clusterID string,
	user *clusterresourcesv1beta1.OpenSearchUser,
	secret *k8sCore.Secret,
	logger logr.Logger,
) error {
	username, _, err := getUserCreds(secret)
	if err != nil {
		logger.Error(err, "Cannot get user's credentials during deleting")
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletingEvent,
			"Resource deleting has been failed. Reason: %v", err,
		)
		return err
	}

	err = r.API.DeleteUser(username, clusterID, models.OpenSearchAppKind)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot delete OpenSearch user resource from Instaclustr",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletionFailed,
			"Resource deletion on Instaclustr has been failed. Reason: %v",
			err,
		)
		return err
	}

	r.EventRecorder.Eventf(
		user, models.Normal, models.DeletionStarted,
		"Resource deletion request has been sent to the Instaclustr API.",
	)

	patch := user.NewPatch()
	delete(user.Status.ClustersEvents, clusterID)
	err = r.Status().Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot delete clusterID from the OpenSearch user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Deleting clusterID from the OpenSearch user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	patch = user.NewPatch()
	controllerutil.RemoveFinalizer(user, getDeletionUserFinalizer(clusterID))
	err = r.Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot delete finalizer from the OpenSearch user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Deleting finalizer from the OpenSearch user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	logger.Info("OpenSearch user has been deleted from the cluster",
		"cluster ID", clusterID,
	)
	r.EventRecorder.Eventf(
		user, models.Normal, models.Deleted,
		"OpenSearchUser resource has been deleted from the cluster (clusterID: %v)", clusterID,
	)

	return nil
}

func (r *OpenSearchUserReconciler) detachUserFromDeletedCluster(
	ctx context.Context,
	clusterID string,
	user *clusterresourcesv1beta1.OpenSearchUser,
	logger logr.Logger,
) error {
	patch := user.NewPatch()
	delete(user.Status.ClustersEvents, clusterID)
	err := r.Status().Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot detach clusterID from the OpenSearch user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Detaching clusterID from the OpenSearch user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	patch = user.NewPatch()
	controllerutil.RemoveFinalizer(user, getDeletionUserFinalizer(clusterID))
	err = r.Patch(ctx, user, patch)
	if err != nil {
		logger.Error(err, "Cannot delete finalizer from the OpenSearch user resource",
			"cluster ID", clusterID,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Deleting finalizer from the OpenSearch user resource has been failed. Reason: %v",
			err,
		)
		return err
	}

	logger.Info("OpenSearch user has been deleted from the cluster",
		"cluster ID", clusterID,
	)
	r.EventRecorder.Eventf(
		user, models.Normal, models.Deleted,
		"OpenSearchUser resource has been deleted from the cluster (clusterID: %v)", clusterID,
	)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1beta1.OpenSearchUser{}).
		Complete(r)
}
