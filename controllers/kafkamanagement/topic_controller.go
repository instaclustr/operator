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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

// TopicReconciler reconciles a TopicName object
type TopicReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=topics/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	topic := &v1beta1.Topic{}
	err := r.Client.Get(ctx, req.NamespacedName, topic)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Kafka topic is not found", "request", req)
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch Kafka topic", "request", req)
		return ctrl.Result{}, err
	}

	switch topic.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateTopic(ctx, topic, l)

	case models.UpdatingEvent:
		return r.handleUpdateTopic(ctx, topic, l)

	case models.DeletingEvent:
		return r.handleDeleteTopic(ctx, topic, l)

	case models.GenericEvent:
		l.Info("Event isn't handled", "topic name", topic.Spec.TopicName, "request", req,
			"event", topic.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TopicReconciler) handleCreateTopic(
	ctx context.Context,
	topic *v1beta1.Topic,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Creation Event")

	if topic.Status.ID == "" {
		l.Info("Creating Kafka topic",
			"topic name", topic.Spec.TopicName)

		patch := topic.NewPatch()
		var err error

		err = r.API.CreateKafkaTopic(instaclustr.KafkaTopicEndpoint, topic)
		if err != nil {
			l.Error(err, "Cannot create Kafka topic", "spec", topic.Spec)
			r.EventRecorder.Eventf(
				topic, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
		l.Info("Kafka topic has been created", "cluster ID", topic.Status.ID)

		r.EventRecorder.Eventf(
			topic, models.Normal, models.Created,
			"Resource creation request is sent. Topic ID: %s",
			topic.Status.ID,
		)

		err = r.Status().Patch(ctx, topic, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka topic from the Instaclustr API",
				"spec", topic.Spec)
			r.EventRecorder.Eventf(
				topic, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		topic.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(topic, models.DeletionFinalizer)

		err = r.Patch(ctx, topic, patch)
		if err != nil {
			l.Error(err, "Cannot patch kafka topic after create op", "kafka topic name", topic.Spec.TopicName)
			r.EventRecorder.Eventf(
				topic, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *TopicReconciler) handleUpdateTopic(
	ctx context.Context,
	t *v1beta1.Topic,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Update Event")

	patch := t.NewPatch()
	url := fmt.Sprintf(instaclustr.KafkaTopicConfigsUpdateEndpoint, t.Status.ID)

	err := r.API.UpdateKafkaTopic(url, t)
	if err != nil {
		l.Error(err, "Unable to update topic, got error from Instaclustr",
			"cluster name", t.Spec.TopicName,
			"cluster status", t.Status,
		)
		r.EventRecorder.Eventf(
			t, models.Warning, models.UpdateFailed,
			"Resource update on the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("Kafka topic has been updated",
		"topic configs to be updated", t.Spec.TopicConfigs,
		"topic status", t.Status)

	err = r.Status().Patch(ctx, t, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka topic management after update op",
			"spec", t.Spec)
		r.EventRecorder.Eventf(
			t, models.Warning, models.PatchFailed,
			"Resource status patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TopicReconciler) handleDeleteTopic(
	ctx context.Context,
	topic *v1beta1.Topic,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Deletion Event")

	_, err := r.API.GetTopicStatus(topic.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka topic",
			"topic name", topic.Spec.TopicName,
			"topic id", topic.Status.ID)
		r.EventRecorder.Eventf(
			topic, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	if !errors.Is(err, instaclustr.NotFound) {
		err = r.API.DeleteKafkaTopic(instaclustr.KafkaTopicEndpoint, topic.Status.ID)
		if err != nil {
			l.Error(err, "Cannot delete kafka topic",
				"topic name", topic.Spec.TopicName,
				"topic ID", topic.Status.ID)
			r.EventRecorder.Eventf(
				topic, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
		r.EventRecorder.Eventf(
			topic, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)
	}

	patch := topic.NewPatch()
	controllerutil.RemoveFinalizer(topic, models.DeletionFinalizer)
	topic.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, topic, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from kafka",
			"cluster name", topic.Spec.TopicName)
		r.EventRecorder.Eventf(
			topic, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("Kafka topic has been deleted",
		"topic name", topic.Spec.TopicName)

	r.EventRecorder.Eventf(
		topic, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return ctrl.Result{}, nil
}

// confirmDeletion confirms if resource is deleting and set appropriate annotation.
func confirmDeletion(obj client.Object) {
	if obj.GetDeletionTimestamp() != nil {
		obj.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
		return
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&v1beta1.Topic{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.Topic)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if newObj.Generation == event.ObjectOld.GetGeneration() {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
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
