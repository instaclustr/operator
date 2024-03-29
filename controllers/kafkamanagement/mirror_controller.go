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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// MirrorReconciler reconciles a Mirror object
type MirrorReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=mirrors/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *MirrorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	mirror := &v1beta1.Mirror{}
	err := r.Client.Get(ctx, req.NamespacedName, mirror)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Kafka mirror is not found", "request", req)
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch Kafka mirror", "request", req)
		return ctrl.Result{}, err
	}

	switch mirror.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateMirror(ctx, mirror, l)
	case models.UpdatingEvent:
		return r.handleUpdateMirror(ctx, mirror, l)
	case models.DeletingEvent:
		return r.handleDeleteMirror(ctx, mirror, l)
	case models.GenericEvent:
		l.Info("Event isn't handled",
			"kafka Connect ID to mirror", mirror.Spec.KafkaConnectClusterID,
			"request", req, "event", mirror.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *MirrorReconciler) handleCreateMirror(
	ctx context.Context,
	mirror *v1beta1.Mirror,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Creation Event")

	if mirror.Status.ID == "" {
		l.Info("Creating kafka Mirror", "Kafka mirror spec", mirror.Spec)
		iStatus, err := r.API.CreateKafkaMirror(&mirror.Spec)
		if err != nil {
			l.Error(err, "Cannot create Kafka mirror", "spec", mirror.Spec)
			r.EventRecorder.Eventf(mirror, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v", err)
			return ctrl.Result{}, err
		}
		l.Info("Kafka mirror has been created", "mirror ID", iStatus.ID)
		r.EventRecorder.Eventf(mirror, models.Normal, models.Created,
			"Kafka mirror creation request is sent to Instaclustr API. Mirror ID: %s", iStatus.ID)

		patch := mirror.NewPatch()
		mirror.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(mirror, models.DeletionFinalizer)
		err = r.Patch(ctx, mirror, patch)
		if err != nil {
			l.Error(err, "Cannot patch kafka mirror after create",
				"kafka mirror connector name", iStatus.ConnectorName)
			r.EventRecorder.Eventf(mirror, models.Warning, models.PatchFailed,
				"Resource patch is failed after request to Instaclustr was sent. Reason: %v", err)
			return ctrl.Result{}, err
		}

		patch = mirror.NewPatch()
		mirror.Status = *iStatus
		err = r.Status().Patch(ctx, mirror, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka mirror status from the Instaclustr API",
				"spec", mirror.Spec, "status", iStatus)
			r.EventRecorder.Eventf(mirror, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v", err)
			return ctrl.Result{}, err
		}
	}

	err := r.startClusterStatusJob(mirror)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"mirror cluster ID", mirror.Status.ID)
		r.EventRecorder.Eventf(mirror, models.Warning, models.CreationFailed,
			"Mirror status job creation is failed. Reason: %v", err)
		return ctrl.Result{}, err
	}
	r.EventRecorder.Event(mirror, models.Normal, models.Created,
		"mirror status check job is started")

	return ctrl.Result{}, nil
}

func (r *MirrorReconciler) handleUpdateMirror(
	ctx context.Context,
	mirror *v1beta1.Mirror,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Update Event")

	iMirror, err := r.API.GetMirrorStatus(mirror.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get Kafka mirror from Instaclustr", "mirror ID", mirror.Status.ID)
		r.EventRecorder.Eventf(mirror, models.Warning, models.UpdateFailed,
			"Resource update on the Instaclustr API is failed. Reason: %v", err)
		return ctrl.Result{}, err
	}

	if mirror.Spec.TargetLatency != iMirror.TargetLatency {
		err = r.API.UpdateKafkaMirror(mirror.Status.ID, mirror.Spec.TargetLatency)
		if err != nil {
			l.Error(err, "Unable to update kafka mirror",
				"kafka connect ID", mirror.Spec.KafkaConnectClusterID, "mirror ID", mirror.Status.ID)
			r.EventRecorder.Eventf(mirror, models.Warning, models.UpdateFailed,
				"Failed to sent a update request to the Instaclustr API. Reason: %v", err)
			return ctrl.Result{}, err
		}

		l.Info("Kafka mirror has been updated",
			"kafka connect ID", mirror.Spec.KafkaConnectClusterID,
			"mirror ID", mirror.Status.ID,
			"target latency", mirror.Spec.TargetLatency,
			"connector name", mirror.Status.ConnectorName)

		return ctrl.Result{}, nil
	}

	patch := mirror.NewPatch()
	mirror.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, mirror, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka mirror management after update",
			"mirror ID", mirror.Status.ID,
			"kafka connect", mirror.Spec.KafkaConnectClusterID)
		r.EventRecorder.Eventf(mirror, models.Warning, models.PatchFailed,
			"Resource status patch is failed. Reason: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MirrorReconciler) handleDeleteMirror(
	ctx context.Context,
	mirror *v1beta1.Mirror,
	l logr.Logger,
) (ctrl.Result, error) {
	l = l.WithName("Deletion Event")

	iMirror, err := r.API.GetMirrorStatus(mirror.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka mirror",
			"kafka connect ID", mirror.Spec.KafkaConnectClusterID,
			"mirror id", mirror.Status.ID)
		r.EventRecorder.Eventf(mirror, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v", err)
		return ctrl.Result{}, err
	}

	if iMirror != nil {
		err = r.API.DeleteKafkaMirror(mirror.Status.ID)
		if err != nil {
			l.Error(err, "Cannot delete kafka mirror",
				"mirror name", mirror.Spec.KafkaConnectClusterID,
				"mirror ID", mirror.Status.ID)
			r.EventRecorder.Eventf(mirror, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v", err)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(mirror, models.Normal, models.DeletionStarted,
			"Mirror deletion request is sent to the Instaclustr API.")
	}

	patch := mirror.NewPatch()
	r.Scheduler.RemoveJob(mirror.GetJobID(scheduler.SyncJob))
	mirror.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	controllerutil.RemoveFinalizer(mirror, models.DeletionFinalizer)
	err = r.Patch(ctx, mirror, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from kafka",
			"cluster name", mirror.Spec.KafkaConnectClusterID)
		r.EventRecorder.Eventf(mirror, models.Warning, models.PatchFailed,
			"Mirror patch is failed after deletion. Reason: %v", err)
		return ctrl.Result{}, err
	}

	l.Info("Kafka mirror has been deleted",
		"kafka connect ID", mirror.Spec.KafkaConnectClusterID,
		"mirror ID", mirror.Status.ID)

	r.EventRecorder.Eventf(mirror, models.Normal, models.Deleted,
		"Resource is deleted")

	return ctrl.Result{}, nil
}

func (r *MirrorReconciler) startClusterStatusJob(mirror *v1beta1.Mirror) error {
	job := r.newWatchStatusJob(mirror)

	err := r.Scheduler.ScheduleJob(mirror.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *MirrorReconciler) newWatchStatusJob(mirror *v1beta1.Mirror) scheduler.Job {
	l := log.Log.WithValues("component", "mirrorStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(mirror)
		err := r.Get(context.Background(), namespacedName, mirror)
		if k8serrors.IsNotFound(err) {
			l.Info("Mirror is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.EventRecorder.Eventf(mirror, models.Normal, models.Deleted,
				"Mirror is not found in the k8s cluster. Closing Instaclustr status sync.")

			r.Scheduler.RemoveJob(mirror.GetJobID(scheduler.SyncJob))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get mirror resource", "resource name", mirror.Name)
			r.EventRecorder.Eventf(mirror, models.Warning, models.FetchFailed,
				"Cannot get mirror resource, request: %v", namespacedName)
			return err
		}

		iMirror, err := r.API.GetMirrorStatus(mirror.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				l.Info("Mirror is not found in Instaclustr. Deleting resource.",
					"mirror ID", mirror.Status.ID,
					"kafka connect ID", mirror.Spec.KafkaConnectClusterID,
					"namespaced name", namespacedName)
				r.EventRecorder.Event(mirror, models.Normal, models.Deleted,
					"Mirror is not found in Instaclustr. Deleting resource.")

				err = r.Delete(context.Background(), mirror)
				if err != nil {
					l.Error(err, "Cannot delete mirror resource after not found in the Instaclustr",
						"mirror namespacedName", namespacedName, "resource ID", mirror.Status.ID)
					r.EventRecorder.Event(mirror, models.Normal, models.Deleted,
						"Cannot delete mirror resource after not found in the Instaclustr")

					return err
				}

				return nil
			}

			l.Error(err, "Cannot get cluster from Instaclustr", "mirror ID", mirror.Status.ID)
			return err
		}

		if mirror.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			mirror.Spec.TargetLatency != iMirror.TargetLatency {
			l.Info("k8s Kafka Mirror target latency is different from Instaclustr. Reconcile target latency..",
				"mirror ID", mirror.Status.ID,
				"kafka connect ID", mirror.Spec.KafkaConnectClusterID,
				"Instaclustr target latency", iMirror.TargetLatency,
				"k8s target latency", mirror.Spec.TargetLatency)
			r.EventRecorder.Eventf(mirror, models.Warning, models.ExternalChanges,
				"There are external changes on the Instaclustr console. Please reconcile the specification manually. "+
					"targetLatency from k8s: %v, targetLatency from instaclustr: %v",
				mirror.Spec.TargetLatency, iMirror.TargetLatency)
		}

		if !mirror.Status.IsEqual(iMirror) {
			l.Info("k8s Kafka Mirror is different from Instaclustr. Reconcile statuses..",
				"Instaclustr data", iMirror,
				"k8s operator data", mirror.Status)

			patch := mirror.NewPatch()
			mirror.Status = *iMirror
			err = r.Status().Patch(context.Background(), mirror, patch)
			if err != nil {
				l.Error(err, "Cannot patch Mirror status",
					"mirror ID name", mirror.Status.ID)
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MirrorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&v1beta1.Mirror{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.Mirror)

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
