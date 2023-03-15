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
	"errors"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
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

	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
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

	user := &kafkamanagementv1alpha1.KafkaUser{}
	err := r.Client.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka User resource is not found", "request", req)
			return models.ExitReconcile, nil
		}
		l.Error(err, "Unable to fetch Kafka User", "request", req)
		return models.ReconcileRequeue, nil
	}

	switch user.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateKafkaUser(ctx, user, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateKafkaUser(ctx, user, l), nil

	case models.DeletingEvent:
		return r.handleDeleteKafkaUser(ctx, user, l), nil
	case models.SecretEvent:
		kafkaUserSecret := &v1.Secret{}
		kafkaUserSecretNamespacedName := types.NamespacedName{
			Name:      user.Spec.KafkaUserSecretName,
			Namespace: user.Spec.KafkaUserSecretNamespace,
		}
		err = r.Client.Get(ctx, kafkaUserSecretNamespacedName, kafkaUserSecret)
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot get Kafka User credentials. Secret is not found",
				"request", req,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.FetchFailed,
				"Fetch user credentials secret is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}
		return r.handleUpdateKafkaUser(ctx, user, l), nil
	default:
		l.Info("Unhandled event", "annotations", user.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *KafkaUserReconciler) handleCreateKafkaUser(
	ctx context.Context,
	user *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	if user.Status.ID == "" {
		l.Info(
			"Creating Kafka User resource",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
		)

		iKafkaUser := user.Spec.ToInstAPI()
		username, password, err := r.getKafkaUserCredsFromSecret(user.Spec)
		if err != nil {
			l.Error(
				err, "Cannot get Kafka User creds from secret",
				"kafka user spec", iKafkaUser,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.FetchFailed,
				"Fetch user credentials from secret is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		iKafkaUser.Username = username
		iKafkaUser.Password = password
		kafkaUserStatus, err := r.API.CreateKafkaUser(instaclustr.KafkaUserEndpoint, iKafkaUser)
		if err != nil {
			l.Error(
				err, "Cannot create Kafka User resource",
				"kafka user resource spec", user.Spec,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			user, models.Normal, models.Created,
			"Resource creation request is sent. User ID: %s",
			kafkaUserStatus.ID,
		)

		patch := user.NewPatch()
		user.Status = *kafkaUserStatus
		err = r.Status().Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka User resource status",
				"kafka cluster ID", user.Spec.ClusterID,
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(user, models.DeletionFinalizer)
		user.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka User resource metadata",
				"kafka cluster ID", user.Spec.ClusterID,
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Kafka User resource was created",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
		)
	}

	return models.ExitReconcile
}

func (r *KafkaUserReconciler) handleUpdateKafkaUser(
	ctx context.Context,
	user *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	iKafkaUser := user.Spec.ToInstAPI()
	username, password, err := r.getKafkaUserCredsFromSecret(user.Spec)
	if err != nil {
		l.Error(
			err, "Cannot get Kafka User creds from secret",
			"kafka user spec", iKafkaUser,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user credentials from secret is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	iKafkaUser.Username = username
	iKafkaUser.Password = password
	err = r.API.UpdateKafkaUser(user.Status.ID, iKafkaUser)
	if err != nil {
		l.Error(err, "Cannot update Kafka User",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.UpdateFailed,
			"Resource update on the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	patch := user.NewPatch()
	user.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, user, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka User resource metadata",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
			"kafka user metadata", user.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka User resource has been updated",
		"kafka cluster ID", user.Spec.ClusterID,
		"initial permissions", user.Spec.InitialPermissions,
		"kafka user options", user.Spec.Options,
	)

	return models.ExitReconcile
}

func (r *KafkaUserReconciler) handleDeleteKafkaUser(
	ctx context.Context,
	user *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	patch := user.NewPatch()
	err := r.Patch(ctx, user, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka User resource metadata",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
			"kafka user metadata", user.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetKafkaUserStatus(user.Status.ID, instaclustr.KafkaUserEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Kafka User status from the Instaclustr API",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteKafkaUser(user.Status.ID, instaclustr.KafkaUserEndpoint)
		if err != nil {
			l.Error(err, "cannot update Kafka User resource statuss",
				"kafka cluster ID", user.Spec.ClusterID,
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		r.EventRecorder.Eventf(
			user, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)
	}

	controllerutil.RemoveFinalizer(user, models.DeletionFinalizer)
	user.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, user, patch)
	if err != nil {
		l.Error(err, "cannot patch Kafka User resource metadata",
			"kafka cluster ID", user.Spec.ClusterID,
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
			"kafka user metadata", user.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka User has been deleted",
		"kafka cluster ID", user.Spec.ClusterID,
		"initial permissions", user.Spec.InitialPermissions,
		"kafka user options", user.Spec.Options,
	)

	r.EventRecorder.Eventf(
		user, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return models.ExitReconcile
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&kafkamanagementv1alpha1.KafkaUser{},
		kafkaUserField,
		func(rawObj client.Object,
		) []string {

			kafkaUser := rawObj.(*kafkamanagementv1alpha1.KafkaUser)
			if kafkaUser.Spec.KafkaUserSecretName == "" {
				return nil
			}
			return []string{kafkaUser.Spec.KafkaUserSecretName}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkamanagementv1alpha1.KafkaUser{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*kafkamanagementv1alpha1.KafkaUser)
				if newObj.DeletionTimestamp != nil {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if event.ObjectOld.GetGeneration() == newObj.Generation {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).Owns(&v1.Secret{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findSecretObjects),
		).
		Complete(r)
}

func (r *KafkaUserReconciler) findSecretObjects(secret client.Object) []reconcile.Request {
	kafkaUserList := &kafkamanagementv1alpha1.KafkaUserList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(kafkaUserField, secret.GetName()),
		Namespace:     secret.GetNamespace(),
	}
	err := r.List(context.TODO(), kafkaUserList, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(kafkaUserList.Items))
	for i, item := range kafkaUserList.Items {
		patch := item.NewPatch()
		item.GetAnnotations()[models.ResourceStateAnnotation] = models.SecretEvent
		err = r.Patch(context.TODO(), &item, patch)
		if err != nil {
			return []reconcile.Request{}
		}
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *KafkaUserReconciler) getKafkaUserCredsFromSecret(
	kafkaUserSpec kafkamanagementv1alpha1.KafkaUserSpec,
) (string, string, error) {
	kafkaUserSecret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      kafkaUserSpec.KafkaUserSecretName,
		Namespace: kafkaUserSpec.KafkaUserSecretNamespace,
	}

	err := r.Get(context.TODO(), kafkaUserSecretNamespacedName, kafkaUserSecret)
	if err != nil {
		return "", "", err
	}

	username := kafkaUserSecret.Data[models.Username]
	password := kafkaUserSecret.Data[models.Password]

	return string(username[:len(username)-1]), string(password[:len(password)-1]), nil
}
