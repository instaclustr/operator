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

// MirrorReconciler reconciles a Mirror object
type MirrorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mirror object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *MirrorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	mirror := &kafkamanagementv1alpha1.Mirror{}
	err := r.Client.Get(ctx, req.NamespacedName, mirror)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Kafka mirror is not found", "request", req)
			return reconcile.Result{}, nil
		}

		l.Error(err, "unable to fetch Kafka mirror", "request", req)
		return reconcile.Result{}, err
	}

	switch mirror.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, mirror, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, mirror, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, mirror, l), nil

	case models.GenericEvent:
		l.Info("event isn't handled",
			"Kafka Connect ID to mirror", mirror.Spec.KafkaConnectClusterID,
			"request", req, "event", mirror.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *MirrorReconciler) handleCreateCluster(
	ctx context.Context,
	mirror *kafkamanagementv1alpha1.Mirror,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Creation Event")

	if mirror.Status.ID == "" {
		l.Info("Creating Kafka mirror",
			"Kafka Connect ID to mirror", mirror.Spec.KafkaConnectClusterID)

		patch := mirror.NewPatch()
		var err error

		err = r.API.CreateKafkaMirror(instaclustr.KafkaMirrorEndpoint, mirror)
		if err != nil {
			l.Error(err, "cannot create Kafka mirror", "spec", mirror.Spec)
			return models.ReconcileRequeue
		}
		l.Info("Kafka mirror has been created", "Mirror ID", mirror.Status.ID)

		err = r.Status().Patch(ctx, mirror, patch)
		if err != nil {
			l.Error(err, "cannot patch Kafka mirror from the Instaclustr API",
				"spec", mirror.Spec)
			return models.ReconcileRequeue
		}

		mirror.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(mirror, models.DeletionFinalizer)

		err = r.Patch(ctx, mirror, patch)
		if err != nil {
			l.Error(err, "cannot patch kafka mirror after create op", "kafka mirror name", mirror.Spec.KafkaConnectClusterID)
			return models.ReconcileRequeue
		}
	}

	return reconcile.Result{}
}

func (r *MirrorReconciler) handleUpdateCluster(
	ctx context.Context,
	t *kafkamanagementv1alpha1.Mirror,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Update Event")

	patch := t.NewPatch()

	err := r.API.UpdateKafkaMirror(instaclustr.KafkaMirrorEndpoint, t)
	if err != nil {
		l.Error(err, "Unable to update mirror, got error from Instaclustr",
			"Cluster name", t.Spec.KafkaConnectClusterID,
			"Cluster ID", t.Status.ID,
		)
		return models.ReconcileRequeue
	}
	l.Info("kafka mirror has been updated")

	err = r.Status().Patch(ctx, t, patch)
	if err != nil {
		l.Error(err, "cannot patch Kafka mirror management after update op",
			"spec", t.Spec)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *MirrorReconciler) handleDeleteCluster(
	ctx context.Context,
	mirror *kafkamanagementv1alpha1.Mirror,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Deletion Event")

	status, err := r.API.GetClusterStatus(mirror.Status.ID, instaclustr.KafkaMirrorEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "cannot get Kafka mirror",
			"mirror name", mirror.Spec.KafkaConnectClusterID,
			"mirror id", mirror.Status.ID)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteKafkaMirror(instaclustr.KafkaMirrorEndpoint, mirror.Status.ID)
		if err != nil {
			l.Error(err, "cannot delete kafka mirror",
				"mirror name", mirror.Spec.KafkaConnectClusterID,
				"mirror ID", mirror.Status.ID)
			return models.ReconcileRequeue
		}
	}

	patch := mirror.NewPatch()
	controllerutil.RemoveFinalizer(mirror, models.DeletionFinalizer)
	mirror.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, mirror, patch)
	if err != nil {
		l.Error(err, "cannot patch remove finalizer from kafka",
			"Cluster name", mirror.Spec.KafkaConnectClusterID)
		return models.ReconcileRequeue
	}

	l.Info("Kafka mirror has been deleted",
		"Mirror name", mirror.Spec.KafkaConnectClusterID)

	return reconcile.Result{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MirrorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkamanagementv1alpha1.Mirror{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
				confirmDeletion(event.ObjectNew)
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).Complete(r)
}