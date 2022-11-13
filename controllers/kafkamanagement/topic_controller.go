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
	"fmt"

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

// TopicReconciler reconciles a TopicName object
type TopicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TopicName object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var topic kafkamanagementv1alpha1.Topic
	err := r.Client.Get(ctx, req.NamespacedName, &topic)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Kafka topic is not found", "request", req)
			return reconcile.Result{}, nil
		}

		l.Error(err, "unable to fetch Kafka topic", "request", req)
		return reconcile.Result{}, err
	}

	switch topic.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &topic, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &topic, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &topic, l), nil

	case models.GenericEvent:
		l.Info("event isn't handled", "Topic name", topic.Spec.TopicName, "request", req,
			"event", topic.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *TopicReconciler) handleCreateCluster(
	ctx context.Context,
	topic *kafkamanagementv1alpha1.Topic,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Creation Event")

	if topic.Status.ID == "" {
		l.Info("Creating Kafka topic",
			"Topic name", topic.Spec.TopicName)

		patch := topic.NewPatch()
		var err error

		err = r.API.CreateKafkaTopic(instaclustr.KafkaTopicEndpoint, topic)
		if err != nil {
			l.Error(err, "cannot create Kafka topic", "spec", topic.Spec)
			return models.ReconcileRequeue
		}
		l.Info("Kafka topic has been created", "cluster ID", topic.Status.ID)

		err = r.Status().Patch(ctx, topic, patch)
		if err != nil {
			l.Error(err, "cannot patch Kafka topic from the Instaclustr API",
				"spec", topic.Spec)
			return models.ReconcileRequeue
		}

		topic.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(topic, models.DeletionFinalizer)

		err = r.Patch(ctx, topic, patch)
		if err != nil {
			l.Error(err, "cannot patch kafka topic after create op", "kafka topic name", topic.Spec.TopicName)
			return models.ReconcileRequeue
		}
	}

	return reconcile.Result{}
}

func (r *TopicReconciler) handleUpdateCluster(
	ctx context.Context,
	t *kafkamanagementv1alpha1.Topic,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Update Event")

	patch := t.NewPatch()
	url := fmt.Sprintf(instaclustr.KafkaTopicConfigsUpdateEndpoint, t.Status.ID)

	err := r.API.UpdateKafkaTopic(url, t)
	if err != nil {
		l.Error(err, "Unable to update topic, got error from Instaclustr",
			"Cluster name", t.Spec.TopicName,
			"Cluster status", t.Status,
		)
		return models.ReconcileRequeue
	}

	l.Info("kafka topic has been updated",
		"topic configs to be updated", t.Spec.TopicConfigs,
		"topic status", t.Status)

	err = r.Status().Patch(ctx, t, patch)
	if err != nil {
		l.Error(err, "cannot patch Kafka topic management after update op",
			"spec", t.Spec)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *TopicReconciler) handleDeleteCluster(
	ctx context.Context,
	topic *kafkamanagementv1alpha1.Topic,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Deletion Event")

	status, err := r.API.GetClusterStatus(topic.Status.ID, instaclustr.KafkaTopicEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "cannot get Kafka topic",
			"topic name", topic.Spec.TopicName,
			"topic id", topic.Status.ID)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteKafkaTopic(instaclustr.KafkaTopicEndpoint, topic.Status.ID)
		if err != nil {
			l.Error(err, "cannot delete kafka topic",
				"topic name", topic.Spec.TopicName,
				"topic ID", topic.Status.ID)
			return models.ReconcileRequeue
		}
	}

	patch := topic.NewPatch()
	controllerutil.RemoveFinalizer(topic, models.DeletionFinalizer)
	topic.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, topic, patch)
	if err != nil {
		l.Error(err, "cannot patch remove finalizer from kafka",
			"Cluster name", topic.Spec.TopicName)
		return models.ReconcileRequeue
	}

	l.Info("Kafka topic has been deleted",
		"Topic name", topic.Spec.TopicName)

	return reconcile.Result{}
}

// confirmDeletion confirms if resource is deleting and set appropriate annotation.
func confirmDeletion(obj client.Object) {
	if obj.GetDeletionTimestamp() != nil {
		obj.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
		return
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kafkamanagementv1alpha1.Topic{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
				confirmDeletion(event.ObjectNew)
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.GenericEvent})
				return true
			},
		})).Complete(r)
}
