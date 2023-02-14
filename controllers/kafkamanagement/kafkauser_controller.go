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
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
)

const (
	kafkaUserField = ".spec.kafkaUserSecretName"
)

// KafkaUserReconciler reconciles a KafkaUser object
type KafkaUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/finalizers,verbs=update
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=secrets,verbs=get;list;watch;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var kafkaUser kafkamanagementv1alpha1.KafkaUser
	err := r.Client.Get(ctx, req.NamespacedName, &kafkaUser)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "KafkaUser resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "Unable to fetch KafkaUser", "request", req)
		return reconcile.Result{}, err
	}

	switch kafkaUser.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateKafkaUser(ctx, &kafkaUser, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateKafkaUser(ctx, &kafkaUser, l), nil

	case models.DeletingEvent:
		return r.handleDeleteKafkaUser(ctx, &kafkaUser, l), nil
	case models.SecretEvent:
		kafkaUserSecret := &v1.Secret{}
		kafkaUserSecretNamespacedName := types.NamespacedName{
			Name:      kafkaUser.Spec.KafkaUserSecretName,
			Namespace: kafkaUser.Spec.KafkaUserSecretNamespace,
		}
		err = r.Client.Get(ctx, kafkaUserSecretNamespacedName, kafkaUserSecret)
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot get KafkaUser credentials. Secret is not found",
				"request", req,
			)
			return models.ReconcileRequeue, nil
		}
		return r.handleUpdateKafkaUser(ctx, &kafkaUser, l), nil
	default:
		l.Info("Unhandled event", "annotations", kafkaUser.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *KafkaUserReconciler) handleCreateKafkaUser(
	ctx context.Context,
	kafkaUser *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	if kafkaUser.Status.ID == "" {
		l.Info(
			"Creating KafkaUser resource",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
		)

		kafkaUserSpec, err := convertors.KafkaUserToAPIv2(&kafkaUser.Spec, &ctx, r.Client)
		if err != nil {
			l.Error(
				err, "Cannot convert KafkaUser Spec",
				"KafkaUser Spec", kafkaUserSpec,
			)
			return models.ReconcileRequeue
		}

		kafkaUserStatus, err := r.API.CreateKafkaUser(instaclustr.KafkaUserEndpoint, kafkaUserSpec)
		if err != nil {
			l.Error(
				err, "Cannot create KafkaUser resource",
				"KafkaUser resource spec", kafkaUser.Spec,
			)
			return models.ReconcileRequeue
		}

		patch := kafkaUser.NewPatch()
		kafkaUser.Status = *kafkaUserStatus
		err = r.Status().Patch(ctx, kafkaUser, patch)
		if err != nil {
			l.Error(err, "Cannot patch KafkaUser resource status",
				"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
				"initial Permissions", kafkaUser.Spec.InitialPermissions,
				"KafkaUser Options", kafkaUser.Spec.Options,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(kafkaUser, models.DeletionFinalizer)
		kafkaUser.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, kafkaUser, patch)
		if err != nil {
			l.Error(err, "Cannot patch KafkaUser resource metadata",
				"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
				"initial Permissions", kafkaUser.Spec.InitialPermissions,
				"KafkaUser Options", kafkaUser.Spec.Options,
				"KafkaUser metadata", kafkaUser.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"KafkaUser resource was created",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
		)
	}

	return reconcile.Result{}
}

func (r *KafkaUserReconciler) handleUpdateKafkaUser(
	ctx context.Context,
	kafkaUser *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	kafkaUserSpec, err := convertors.KafkaUserToAPIv2(&kafkaUser.Spec, &ctx, r.Client)
	if err != nil {
		l.Error(
			err, "Cannot convert KafkaUser Spec",
			"KafkaUser Spec", kafkaUserSpec,
		)
		return models.ReconcileRequeue
	}

	err = r.API.UpdateKafkaUser(kafkaUser.Status.ID, instaclustr.KafkaUserEndpoint, kafkaUserSpec)
	if err != nil {
		l.Error(err, "Cannot update KafkaUser",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
		)
	}

	patch := kafkaUser.NewPatch()
	kafkaUser.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, kafkaUser, patch)
	if err != nil {
		l.Error(err, "Cannot patch KafkaUser resource metadata",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
			"KafkaUser metadata", kafkaUser.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("KafkaUser resource has been updated",
		"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
		"initial Permissions", kafkaUser.Spec.InitialPermissions,
		"KafkaUser Options", kafkaUser.Spec.Options,
	)

	return reconcile.Result{}
}

func (r *KafkaUserReconciler) handleDeleteKafkaUser(
	ctx context.Context,
	kafkaUser *kafkamanagementv1alpha1.KafkaUser,
	l logr.Logger,
) reconcile.Result {
	patch := kafkaUser.NewPatch()
	err := r.Patch(ctx, kafkaUser, patch)
	if err != nil {
		l.Error(err, "Cannot patch KafkaUser resource metadata",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
			"KafkaUser metadata", kafkaUser.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetKafkaUserStatus(kafkaUser.Status.ID, instaclustr.KafkaUserEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get KafkaUser status from the Instaclustr API",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteKafkaUser(kafkaUser.Status.ID, instaclustr.KafkaUserEndpoint)
		if err != nil {
			l.Error(err, "Cannot update KafkaUser resource statuss",
				"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
				"initial Permissions", kafkaUser.Spec.InitialPermissions,
				"KafkaUser Options", kafkaUser.Spec.Options,
				"KafkaUser metadata", kafkaUser.ObjectMeta,
			)
			return models.ReconcileRequeue
		}
		return models.ReconcileRequeue
	}

	controllerutil.RemoveFinalizer(kafkaUser, models.DeletionFinalizer)
	kafkaUser.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kafkaUser, patch)
	if err != nil {
		l.Error(err, "Cannot patch KafkaUser resource metadata",
			"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
			"initial Permissions", kafkaUser.Spec.InitialPermissions,
			"KafkaUser Options", kafkaUser.Spec.Options,
			"KafkaUser metadata", kafkaUser.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("KafkaUser has been deleted",
		"Kafka Cluster ID", kafkaUser.Spec.ClusterID,
		"initial Permissions", kafkaUser.Spec.InitialPermissions,
		"KafkaUser Options", kafkaUser.Spec.Options,
	)

	return reconcile.Result{}
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
				if event.ObjectOld.GetGeneration() == event.ObjectNew.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
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
