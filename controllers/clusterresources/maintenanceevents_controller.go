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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	patch := me.NewPatch()
	if me.DeletionTimestamp != nil {
		err = r.deleteExclusionWindows(me)
		if err != nil {
			l.Error(err, "Cannot delete Exclusion Windows",
				"resource", me,
			)

			r.EventRecorder.Eventf(
				me, models.Warning, models.DeletionFailed,
				"Exclusion windows deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		r.EventRecorder.Eventf(
			me, models.Normal, models.DeletionStarted,
			"Exclusion windows deletion request is sent to the Instaclustr API.",
		)

		r.Scheduler.RemoveJob(me.GetJobID(scheduler.StatusChecker))
		controllerutil.RemoveFinalizer(me, models.DeletionFinalizer)
		err = r.Patch(ctx, me, patch)
		if err != nil {
			l.Error(err, "Cannot patch Maintenance Event resource",
				"resource", me,
			)

			r.EventRecorder.Eventf(
				me, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		return models.ExitReconcile, nil
	}

	err = r.reconcileMaintenanceEventsReschedules(me)
	if err != nil {
		l.Error(err, "Cannot reconcile Maintenance Events Reschedules",
			"spec", me.Spec,
		)

		r.EventRecorder.Eventf(
			me, models.Warning, models.UpdateFailed,
			"Maintenance events update is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	err = r.reconcileExclusionWindows(me)
	if err != nil {
		l.Error(err, "Cannot reconcile Exclusion Windows",
			"spec", me.Spec,
		)

		r.EventRecorder.Eventf(
			me, models.Warning, models.UpdateFailed,
			"Exclusion windows update is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	err = r.Status().Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch Maintenance Event status",
			"status", me.Status,
		)

		r.EventRecorder.Eventf(
			me, models.Warning, models.PatchFailed,
			"Resource status patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	controllerutil.AddFinalizer(me, models.DeletionFinalizer)
	err = r.Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch Maintenance Event status",
			"status", me.Status,
		)

		r.EventRecorder.Eventf(
			me, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	err = r.startMaintenanceEventStatusJob(me)
	if err != nil {
		l.Error(err, "Cannot start Maintenance Event status job",
			"spec", me.Spec,
		)

		r.EventRecorder.Eventf(
			me, models.Warning, models.CreationFailed,
			"Resource status check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	r.EventRecorder.Eventf(
		me, models.Normal, models.Created,
		"Cluster backups check job is started",
	)

	l.Info("Maintenance Event resource was reconciled",
		"spec", me.Spec,
		"status", me.Status,
	)

	return models.ExitReconcile, nil
}

func (r *MaintenanceEventsReconciler) startMaintenanceEventStatusJob(me *v1beta1.MaintenanceEvents) error {
	job := r.newWatchStatusJob(me)

	err := r.Scheduler.ScheduleJob(me.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *MaintenanceEventsReconciler) newWatchStatusJob(me *v1beta1.MaintenanceEvents) scheduler.Job {
	l := log.Log.WithValues("component", "MaintenanceEventStatusJob")
	return func() error {
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      me.Name,
			Namespace: me.Namespace,
		}, me)
		if err != nil {
			l.Error(err, "Cannot get Maintenance Events resource",
				"resource", me,
			)

			return err
		}

		var updated bool
		patch := me.NewPatch()
		instMEventsStatuses, err := r.API.GetMaintenanceEventsStatuses(me.Spec.ClusterID)
		if err != nil {
			l.Error(err, "Cannot get Maintenance Events statuses",
				"resource", me,
			)

			return err
		}

		if !me.AreMEventsStatusesEqual(instMEventsStatuses) {
			me.Status.EventsStatuses = instMEventsStatuses
			updated = true
		}

		instWindowsStatuses, err := r.API.GetExclusionWindowsStatuses(me.Spec.ClusterID)
		if err != nil {
			l.Error(err, "Cannot get Exclusion Windows statuses",
				"resource", me,
			)

			return err
		}

		if !me.AreExclusionWindowsStatusesEqual(instWindowsStatuses) {
			me.Status.ExclusionWindowsStatuses = instWindowsStatuses
			updated = true
		}

		if updated {
			err = r.Status().Patch(context.TODO(), me, patch)
			if err != nil {
				l.Error(err, "Cannot get Maintenance Events resource status",
					"resource", me,
				)

				return err
			}

			l.Info("Maintenance Events resource status was updated",
				"resource", me,
			)
		}

		return nil
	}
}

func (r *MaintenanceEventsReconciler) reconcileMaintenanceEventsReschedules(mEvents *v1beta1.MaintenanceEvents) error {
	var updatedMEventsStatuses []*v1beta1.MaintenanceEventStatus
	instMEvents, err := r.API.GetMaintenanceEventsStatuses(mEvents.Spec.ClusterID)
	if err != nil {
		return err
	}

	for _, k8sMEvent := range mEvents.Spec.MaintenanceEventsReschedules {
		for _, instMEvent := range instMEvents {
			if instMEvent.IsFinalized {
				updatedMEventsStatuses = append(updatedMEventsStatuses, instMEvent)

				break
			}

			if k8sMEvent.ScheduleID == instMEvent.ID &&
				(k8sMEvent.ScheduledStartTime != instMEvent.ScheduledStartTime) {
				updatedMEventStatus, err := r.API.UpdateMaintenanceEvent(*k8sMEvent)
				if err != nil {
					return err
				}

				updatedMEventsStatuses = append(updatedMEventsStatuses, updatedMEventStatus)
			}
		}
	}

	mEvents.Status.EventsStatuses = updatedMEventsStatuses

	return nil
}

func (r *MaintenanceEventsReconciler) reconcileExclusionWindows(mEvents *v1beta1.MaintenanceEvents) error {
	instWindowsStatuses, err := r.API.GetExclusionWindowsStatuses(mEvents.Spec.ClusterID)
	if err != nil {
		return err
	}

	instWindowsMap := map[v1beta1.ExclusionWindowSpec]string{}
	for _, instWindow := range instWindowsStatuses {
		instWindowsMap[instWindow.ExclusionWindowSpec] = instWindow.ID
	}

	k8sWindowsMap := mEvents.Spec.NewExclusionWindowMap()

	for instWindow, id := range instWindowsMap {
		if _, exists := k8sWindowsMap[instWindow]; !exists {
			err = r.API.DeleteExclusionWindow(id)
			if err != nil {
				return err
			}

			continue
		}

		k8sWindowsMap[instWindow] = id
	}

	var windowsStatus []*v1beta1.ExclusionWindowStatus
	for k8sWindow := range k8sWindowsMap {
		if _, exists := instWindowsMap[k8sWindow]; !exists {
			k8sWindowsMap[k8sWindow], err = r.API.CreateExclusionWindow(mEvents.Spec.ClusterID, k8sWindow)
			if err != nil {
				return err
			}
		}

		windowsStatus = append(windowsStatus, &v1beta1.ExclusionWindowStatus{
			ID:                  k8sWindowsMap[k8sWindow],
			ExclusionWindowSpec: k8sWindow,
		})
	}

	mEvents.Status.ExclusionWindowsStatuses = windowsStatus

	return nil
}

func (r *MaintenanceEventsReconciler) deleteExclusionWindows(mEvents *v1beta1.MaintenanceEvents) error {
	for _, window := range mEvents.Status.ExclusionWindowsStatuses {
		err := r.API.DeleteExclusionWindow(window.ID)
		if err != nil {
			return err
		}
	}

	return nil
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
