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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// MaintenanceEventsReconciler reconciles a MaintenanceEvents object
type MaintenanceEventsReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
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

	me := &clusterresourcesv1alpha1.MaintenanceEvents{}
	err := r.Client.Get(ctx, req.NamespacedName, me)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("MaintenanceEvents resource is not found",
				"request", req,
			)

			return models.ReconcileResult, nil
		}

		l.Error(err, "Cannot get MaintenanceEvents resource",
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

			return models.ReconcileRequeue, nil
		}

		r.Scheduler.RemoveJob(me.GetJobID(scheduler.StatusChecker))
		controllerutil.RemoveFinalizer(me, models.DeletionFinalizer)
		err = r.Patch(ctx, me, patch)
		if err != nil {
			l.Error(err, "Cannot patch MaintenanceEvents resource",
				"resource", me,
			)

			return models.ReconcileRequeue, nil
		}

		return models.ReconcileResult, nil
	}

	err = r.reconcileMaintenanceEventsReschedules(me)
	if err != nil {
		l.Error(err, "Cannot reconcile MaintenanceEvents Reschedules",
			"maintenance events spec", me.Spec,
		)

		return models.ReconcileRequeue, nil
	}

	err = r.reconcileExclusionWindows(me)
	if err != nil {
		l.Error(err, "Cannot reconcile Exclusion Windows",
			"maintenance events spec", me.Spec,
		)

		return models.ReconcileRequeue, nil
	}

	err = r.Status().Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch MaintenanceEvents status",
			"maintenance events status", me.Status,
		)

		return models.ReconcileRequeue, nil
	}

	controllerutil.AddFinalizer(me, models.DeletionFinalizer)
	err = r.Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch MaintenanceEvents status",
			"maintenance events  status", me.Status,
		)

		return models.ReconcileRequeue, nil
	}

	err = r.startMaintenanceEventStatusJob(me)
	if err != nil {
		l.Error(err, "Cannot start MaintenanceEvents status job",
			"maintenance events spec", me.Spec,
		)

		return models.ReconcileRequeue, nil
	}

	l.Info("MaintenanceEvents resource was reconciled",
		"maintenance events spec", me.Spec,
		"maintenance events status", me.Status,
	)

	return models.ReconcileResult, nil
}

func (r *MaintenanceEventsReconciler) startMaintenanceEventStatusJob(me *clusterresourcesv1alpha1.MaintenanceEvents) error {
	job := r.newWatchStatusJob(me)

	err := r.Scheduler.ScheduleJob(me.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *MaintenanceEventsReconciler) newWatchStatusJob(me *clusterresourcesv1alpha1.MaintenanceEvents) scheduler.Job {
	l := log.Log.WithValues("component", "MaintenanceEventStatusJob")
	return func() error {
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      me.Name,
			Namespace: me.Namespace,
		}, me)
		if err != nil {
			l.Error(err, "Cannot get MaintenanceEventss resource",
				"maintenance events resource", me,
			)

			return err
		}

		var updated bool
		patch := me.NewPatch()
		instMEventsStatuses, err := r.API.GetMaintenanceEventsStatuses(me.Spec.ClusterID)
		if err != nil {
			l.Error(err, "Cannot get MaintenanceEvents statuses",
				"maintenance events resource", me,
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
				"maintenance events resource", me,
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
				l.Error(err, "Cannot get MaintenanceEvents resource status",
					"maintenance events resource", me,
				)

				return err
			}

			l.Info("MaintenanceEvents resource status was updated",
				"maintenance events resource", me,
			)
		}

		return nil
	}
}

func (r *MaintenanceEventsReconciler) reconcileMaintenanceEventsReschedules(mEvents *clusterresourcesv1alpha1.MaintenanceEvents) error {
	var updatedMEventsStatuses []*clusterresourcesv1alpha1.MaintenanceEventStatus
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

func (r *MaintenanceEventsReconciler) reconcileExclusionWindows(mEvents *clusterresourcesv1alpha1.MaintenanceEvents) error {
	instWindowsStatuses, err := r.API.GetExclusionWindowsStatuses(mEvents.Spec.ClusterID)
	if err != nil {
		return err
	}

	instWindowsMap := map[clusterresourcesv1alpha1.ExclusionWindowSpec]string{}
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

	var windowsStatus []*clusterresourcesv1alpha1.ExclusionWindowStatus
	for k8sWindow := range k8sWindowsMap {
		if _, exists := instWindowsMap[k8sWindow]; !exists {
			k8sWindowsMap[k8sWindow], err = r.API.CreateExclusionWindow(mEvents.Spec.ClusterID, k8sWindow)
			if err != nil {
				return err
			}
		}

		windowsStatus = append(windowsStatus, &clusterresourcesv1alpha1.ExclusionWindowStatus{
			ID:                  k8sWindowsMap[k8sWindow],
			ExclusionWindowSpec: k8sWindow,
		})
	}

	mEvents.Status.ExclusionWindowsStatuses = windowsStatus

	return nil
}

func (r *MaintenanceEventsReconciler) deleteExclusionWindows(mEvents *clusterresourcesv1alpha1.MaintenanceEvents) error {
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
		For(&clusterresourcesv1alpha1.MaintenanceEvents{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return !(event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration())
			},
		})).Complete(r)
}
