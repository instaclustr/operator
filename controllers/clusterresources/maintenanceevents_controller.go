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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// MaintenanceEventsReconciler reconciles a MaintenanceEvents object
type MaintenanceEventsReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=maintenanceevents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=maintenanceevents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=maintenanceevents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MaintenanceEvents object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *MaintenanceEventsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	me := &v1beta1.MaintenanceEvents{}
	err := r.Client.Get(ctx, req.NamespacedName, me)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Maintenance Event resource is not found",
				"request", req,
			)
			return models.ExitReconcile, nil
		}
		l.Error(err, "Cannot get Maintenance Event resource",
			"request", req,
		)
		return models.ReconcileRequeue, nil
	}

	if len(me.Spec.MaintenanceEventsReschedules) == 0 {
		err = r.Client.Delete(ctx, me)
		if err != nil {
			l.Error(err,
				"Cannot delete Maintenance Events resource from K8s cluster",
				"Maintenance Events spec", me.Spec,
			)
			r.EventRecorder.Eventf(
				me, models.Warning, models.DeletionFailed,
				"Resource deletion is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}
		r.EventRecorder.Eventf(
			me, models.Normal, models.DeletionStarted,
			"Resource is deleted.",
		)
		l.Info(
			"Maintenance Events were rescheduled, resource was deleted",
			"Maintenance Events spec", me.Spec,
		)
		return models.ExitReconcile, nil
	}

	patch := me.NewPatch()
	mEvents, err := r.API.GetMaintenanceEvents(me.Spec.ClusterID, models.UpcomingME)
	if err != nil {
		l.Error(err,
			"Cannot get Maintenance Event list",
			"rescheduledEvent", me.Status.RescheduledEvent,
		)
		r.EventRecorder.Eventf(
			me, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	for _, meStatus := range mEvents {
		if me.Status.RescheduledEvent.MaintenanceEventId == "" && !meStatus.IsFinalized {
			mEvent := &v1beta1.MaintenanceEventReschedule{
				MaintenanceEventId: me.Spec.MaintenanceEventsReschedules[len(me.Spec.MaintenanceEventsReschedules)-1].MaintenanceEventId,
			}
			me.Status.RescheduledEvent.MaintenanceEventId = mEvent.MaintenanceEventId

			err = r.API.RescheduleMaintenanceEvent(mEvent)
			if err != nil {
				l.Error(err,
					"Cannot start Maintenance Event reschedule process",
					"Maintenance Event ID", mEvent.MaintenanceEventId,
				)
				r.EventRecorder.Eventf(
					me, models.Warning, models.CreationFailed,
					"Resource creation on the Instaclustr is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue, nil
			}

			r.EventRecorder.Eventf(
				me, models.Normal, models.Created,
				"Resource reschedule request is sent. Maintenance Event ID: %s",
				me.Status.RescheduledEvent.MaintenanceEventId,
			)

			err = r.Status().Patch(ctx, me, patch)
			if err != nil {
				l.Error(err,
					"Cannot patch Maintenance Event status",
					"Maintenance Event ID", me.Status.RescheduledEvent.MaintenanceEventId,
				)
				r.EventRecorder.Eventf(
					me, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue, nil
			}

			me.Status.RescheduledEvent.MaintenanceEventId = ""
			err = r.Status().Patch(ctx, me, patch)
			if err != nil {
				l.Error(err,
					"Cannot patch MaintenanceEvent status",
					"Maintenance Event ID", me.Status.RescheduledEvent.MaintenanceEventId,
				)
				r.EventRecorder.Eventf(
					me, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue, nil
			}

			me.Spec.MaintenanceEventsReschedules = me.Spec.MaintenanceEventsReschedules[:len(me.Spec.MaintenanceEventsReschedules)-1]
			err = r.Patch(ctx, me, patch)
			if err != nil {
				l.Error(err, "Cannot patch Maintenance Event",
					"spec", me.Spec,
				)
				r.EventRecorder.Eventf(
					me, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue, nil
			}
		}
	}
	return models.ExitReconcile, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MaintenanceEventsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.MaintenanceEvents{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return !(event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration())
			},
		})).Complete(r)
}
