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

package kafkamanagement

import (
	"context"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

const (
	kafkaUserField = ".spec.kafkaUserSecretName"
)

// KafkaUserReconciler reconciles a KafkaUser object
type KafkaUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/finalizers,verbs=update
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=secrets,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	user := &v1beta1.KafkaUser{}
	err := r.Client.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka user resource is not found", "request", req)
			return ctrl.Result{}, nil
		}
		l.Error(err, "Unable to fetch Kafka user", "request", req)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user is failed. Reason: %v",
			err,
		)

		return ctrl.Result{}, err
	}

	secret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      user.Spec.SecretRef.Name,
		Namespace: user.Spec.SecretRef.Namespace,
	}
	err = r.Client.Get(ctx, kafkaUserSecretNamespacedName, secret)
	if err != nil {
		l.Error(err, "Cannot get secret for kafka user.", "request", req)
		r.EventRecorder.Eventf(user, models.Warning, models.FetchFailed,
			"Fetch user credentials secret is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	username, password, err := r.getKafkaUserCredsFromSecret(user.Spec)
	if err != nil {
		l.Error(err, "Cannot get Kafka user creds from secret", "kafka user spec", user.Spec)
		r.EventRecorder.Eventf(user, models.Warning, models.FetchFailed,
			"Fetch user credentials from secret is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	if controllerutil.AddFinalizer(secret, user.GetDeletionFinalizer()) {
		err = r.Update(ctx, secret)
		if err != nil {
			l.Error(err, "Cannot update Kafka user secret",
				"secret name", secret.Name,
				"secret namespace", secret.Namespace)
			r.EventRecorder.Eventf(user, models.Warning, models.UpdatedEvent,
				"Cannot assign Kafka user to a k8s secret. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	patch := user.NewPatch()
	if controllerutil.AddFinalizer(user, user.GetDeletionFinalizer()) {
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user with finalizer")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Patching Kafka user with finalizer has been failed. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	for clusterID, clusterEvent := range user.Status.ClustersEvents {
		if clusterEvent == models.CreatingEvent {
			l.Info(
				"Creating Kafka user",
				"initial permissions", user.Spec.InitialPermissions,
			)
			iKafkaUser := user.Spec.ToInstAPI(clusterID, username, password)
			_, err = r.API.CreateKafkaUser(instaclustr.KafkaUserEndpoint, iKafkaUser)
			if err != nil {
				l.Error(
					err, "Cannot create Kafka User",
					"kafka user resource spec", user.Spec,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.CreationFailed,
					"Resource creation on the Instaclustr is failed. Reason: %v",
					err,
				)

				return ctrl.Result{}, err
			}

			r.EventRecorder.Eventf(
				user, models.Normal, models.Created,
				"Resource creation request is sent. user ID: %s",
				user.GetID(clusterID, username),
			)

			annots := user.GetAnnotations()
			if annots == nil {
				user.SetAnnotations(make(map[string]string))
			}

			user.Status.ClustersEvents[clusterID] = models.CreatedEvent
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return ctrl.Result{}, err
			}

			l.Info(
				"Kafka user was created",
				"initial permissions", user.Spec.InitialPermissions,
			)

			continue
		}

		if clusterEvent == models.DeletingEvent {
			userID := user.GetID(clusterID, username)
			err = r.API.DeleteKafkaUser(userID, instaclustr.KafkaUserEndpoint)
			if err != nil {
				l.Error(err, "cannot delete Kafka user",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.DeletionFailed,
					"Resource deletion on the Instaclustr is failed. Reason: %v",
					err,
				)

				return ctrl.Result{}, err
			}

			r.EventRecorder.Eventf(
				user, models.Normal, models.DeletionStarted,
				"Resource deletion request is sent to the Instaclustr API.",
			)

			patch = user.NewPatch()
			delete(user.Status.ClustersEvents, clusterID)
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return ctrl.Result{}, err
			}

			l.Info("Kafka user has been deleted",
				"initial permissions", user.Spec.InitialPermissions,
			)

			r.EventRecorder.Eventf(user, models.Normal, models.Deleted,
				"User has been deleted for a cluster, username: %s, clusterID: %s.",
				username, clusterID)

			continue
		}

		if clusterEvent == models.ClusterDeletingEvent {
			delete(user.Status.ClustersEvents, clusterID)
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return ctrl.Result{}, err
			}

			l.Info("Kafka user has been detached",
				"initial permissions", user.Spec.InitialPermissions,
			)
			r.EventRecorder.Eventf(
				user, models.Normal, models.Deleted,
				"User is detached from cluster",
			)

			continue
		}

		iKafkaUser := user.Spec.ToInstAPI(clusterID, username, password)
		userID := user.GetID(clusterID, username)
		err = r.API.UpdateKafkaUser(userID, iKafkaUser)
		if err != nil {
			l.Error(err, "Cannot update Kafka user",
				"initial permissions", user.Spec.InitialPermissions,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.UpdateFailed,
				"Resource update on the Instaclustr API is failed. Reason: %v",
				err,
			)

			return ctrl.Result{}, err
		}

		user.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user resource metadata",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)

			return ctrl.Result{}, err
		}

		user.Status.ClustersEvents[clusterID] = models.UpdatedEvent
		err = r.Status().Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user resource status",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)

			return ctrl.Result{}, err
		}

		l.Info("Kafka user resource has been updated",
			"initial permissions", user.Spec.InitialPermissions,
		)

		continue
	}

	if user.DeletionTimestamp != nil {
		for clusterID, clusterEvent := range user.Status.ClustersEvents {
			if clusterEvent == models.Created || clusterEvent == models.CreatingEvent || clusterEvent == models.UpdatingEvent || clusterEvent == models.UpdatedEvent {
				l.Error(models.ErrUserStillExist, instaclustr.MsgDeleteUser,
					"username", username, "cluster ID", clusterID)
				r.EventRecorder.Event(user, models.Warning, models.DeletingEvent, instaclustr.MsgDeleteUser)

				return ctrl.Result{}, err
			}
		}

		controllerutil.RemoveFinalizer(user, user.GetDeletionFinalizer())
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot delete finalizer from the Kafka user resource")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Deleting finalizer from the Kafka user resource has been failed. Reason: %v", err,
			)

			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(secret, user.GetDeletionFinalizer())
		err = r.Update(ctx, secret)
		if err != nil {
			l.Error(err, "Cannot remove finalizer from secret", "secret name", secret.Name)

			r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		l.Info("Kafka user resource has been deleted", "username", username)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1beta1.KafkaUser{},
		kafkaUserField,
		func(rawObj client.Object,
		) []string {

			kafkaUser := rawObj.(*v1beta1.KafkaUser)
			if kafkaUser.Spec.SecretRef.Name == "" {
				return nil
			}

			return []string{kafkaUser.Spec.SecretRef.Name}
		}); err != nil {

		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&v1beta1.KafkaUser{}).
		Owns(&v1.Secret{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findSecretObjects),
		).
		Complete(r)
}

func (r *KafkaUserReconciler) findSecretObjects(secret client.Object) []ctrl.Request {
	kafkaUserList := &v1beta1.KafkaUserList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(kafkaUserField, secret.GetName()),
		Namespace:     secret.GetNamespace(),
	}
	err := r.List(context.TODO(), kafkaUserList, listOps)
	if err != nil {
		return []ctrl.Request{}
	}

	requests := make([]ctrl.Request, len(kafkaUserList.Items))
	for i, item := range kafkaUserList.Items {
		patch := item.NewPatch()
		annotations := item.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
			item.SetAnnotations(annotations)
		}
		annotations[models.ResourceStateAnnotation] = models.SecretEvent
		err = r.Patch(context.TODO(), &item, patch)
		if err != nil {
			return []ctrl.Request{}
		}
		requests[i] = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *KafkaUserReconciler) getKafkaUserCredsFromSecret(
	kafkaUserSpec v1beta1.KafkaUserSpec,
) (string, string, error) {
	kafkaUserSecret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      kafkaUserSpec.SecretRef.Name,
		Namespace: kafkaUserSpec.SecretRef.Namespace,
	}

	err := r.Get(context.TODO(), kafkaUserSecretNamespacedName, kafkaUserSecret)
	if err != nil {
		return "", "", err
	}

	username := kafkaUserSecret.Data[models.Username]
	password := kafkaUserSecret.Data[models.Password]

	if len(username) == 0 || len(password) == 0 {
		return "", "", models.ErrMissingSecretKeys
	}

	return string(username[:len(username)-1]), string(password[:len(password)-1]), nil
}
