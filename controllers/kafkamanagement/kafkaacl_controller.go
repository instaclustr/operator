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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// KafkaACLReconciler reconciles a KafkaACL object
type KafkaACLReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkaacls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkaacls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkaacls/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *KafkaACLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var kafkaACL kafkamanagementv1alpha1.KafkaACL
	err := r.Client.Get(ctx, req.NamespacedName, &kafkaACL)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Kafka ACL resource is not found", "request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch kafka ACL", "request", req)
		return models.ReconcileRequeue, nil
	}

	switch kafkaACL.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateKafkaACL(ctx, &kafkaACL, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateKafkaACL(ctx, &kafkaACL, l), nil

	case models.DeletingEvent:
		return r.handleDeleteKafkaACL(ctx, &kafkaACL, l), nil
	default:
		l.Info("Event isn't handled",
			"cluster ID", kafkaACL.Spec.ClusterID,
			"user query", kafkaACL.Spec.UserQuery,
			"request", req,
			"event", kafkaACL.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *KafkaACLReconciler) handleCreateKafkaACL(
	ctx context.Context,
	acl *kafkamanagementv1alpha1.KafkaACL,
	l logr.Logger,
) reconcile.Result {
	if acl.Status.ID == "" {
		l.Info("Creating kafka ACL",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)

		kafkaACLStatus, err := r.API.CreateKafkaACL(instaclustr.KafkaACLEndpoint, &acl.Spec)
		if err != nil {
			l.Error(err, "Cannot create kafka ACL",
				"kafka ACL resource spec", acl.Spec,
			)
			r.EventRecorder.Eventf(
				acl, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			acl, models.Normal, models.Created,
			"Resource creation request is sent. Resource ID: %s",
			kafkaACLStatus.ID,
		)

		patch := acl.NewPatch()
		acl.Status = *kafkaACLStatus
		err = r.Status().Patch(ctx, acl, patch)
		if err != nil {
			l.Error(err, "Cannot patch kafka ACL status",
				"cluster ID", acl.Spec.ClusterID,
				"user query", acl.Spec.UserQuery,
			)
			r.EventRecorder.Eventf(
				acl, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(acl, models.DeletionFinalizer)
		acl.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, acl, patch)
		if err != nil {
			l.Error(err, "Cannot patch kafka ACL metadata",
				"cluster ID", acl.Spec.ClusterID,
				"user query", acl.Spec.UserQuery,
			)
			r.EventRecorder.Eventf(
				acl, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Kafka ACL was created",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
	}

	return models.ExitReconcile
}

func (r *KafkaACLReconciler) handleUpdateKafkaACL(
	ctx context.Context,
	acl *kafkamanagementv1alpha1.KafkaACL,
	l logr.Logger,
) reconcile.Result {
	err := r.API.UpdateKafkaACL(acl.Status.ID, instaclustr.KafkaACLEndpoint, &acl.Spec)
	if err != nil {
		l.Error(err, "Cannot update kafka ACL",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
		r.EventRecorder.Eventf(
			acl, models.Warning, models.UpdateFailed,
			"Resource update on the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	patch := acl.NewPatch()
	acl.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, acl, patch)
	if err != nil {
		l.Error(err, "Cannot patch kafka ACL metadata",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
		r.EventRecorder.Eventf(
			acl, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka ACL has been updated",
		"cluster ID", acl.Spec.ClusterID,
		"user query", acl.Spec.UserQuery,
	)
	return models.ExitReconcile
}

func (r *KafkaACLReconciler) handleDeleteKafkaACL(
	ctx context.Context,
	acl *kafkamanagementv1alpha1.KafkaACL,
	l logr.Logger,
) reconcile.Result {
	patch := acl.NewPatch()
	err := r.Patch(ctx, acl, patch)
	if err != nil {
		l.Error(err, "Cannot patch kafka ACL metadata",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
		r.EventRecorder.Eventf(
			acl, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetKafkaACLStatus(acl.Status.ID, instaclustr.KafkaACLEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get kafka ACL status from the Instaclustr API",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
		r.EventRecorder.Eventf(
			acl, models.Warning, models.FetchFailed,
			"Resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteKafkaACL(acl.Status.ID, instaclustr.KafkaACLEndpoint)
		if err != nil {
			l.Error(err, "Cannot update kafka ACL status",
				"cluster ID", acl.Spec.ClusterID,
				"user query", acl.Spec.UserQuery,
			)
			r.EventRecorder.Eventf(
				acl, models.Warning, models.DeletionFailed,
				"Resource deletion is failed on the Instaclustr. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		r.EventRecorder.Eventf(
			acl, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)
	}

	controllerutil.RemoveFinalizer(acl, models.DeletionFinalizer)
	acl.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, acl, patch)
	if err != nil {
		l.Error(err, "Cannot patch kafka ACL metadata",
			"cluster ID", acl.Spec.ClusterID,
			"user query", acl.Spec.UserQuery,
		)
		r.EventRecorder.Eventf(
			acl, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka ACL has been deleted",
		"cluster ID", acl.Spec.ClusterID,
		"user query", acl.Spec.UserQuery,
	)

	r.EventRecorder.Eventf(
		acl, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return models.ExitReconcile
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaACLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkamanagementv1alpha1.KafkaACL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*kafkamanagementv1alpha1.KafkaACL)
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
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).Complete(r)
}
