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

	var me clusterresourcesv1alpha1.MaintenanceEvents
	err := r.Client.Get(ctx, req.NamespacedName, &me)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Maintenance Event resource is not found", "request", req)
			return models.ReconcileResult, nil
		}
		l.Error(err, "unable to fetch Maintenance Event", "request", req)
		return models.ReconcileResult, nil
	}

	switch me.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateMaintenanceEvent(ctx, &me, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateMaintenanceEvent(ctx, &me, l), nil

	case models.DeletingEvent:
		return r.handleDeleteMaintenanceEvent(ctx, &me, l), nil
	default:
		l.Info("Event isn't handled",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
			"request", req,
			"event", me.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *MaintenanceEventsReconciler) handleCreateMaintenanceEvent(
	ctx context.Context,
	me *clusterresourcesv1alpha1.MaintenanceEvents,
	l logr.Logger,
) reconcile.Result {

	if me.Status.ID == "" {
		l.Info(
			"Creating Exclusion Window",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
		)

		meStatus, err := r.API.CreateExclusionWindow(instaclustr.ExclusionWindowEndpoint, &me.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create Exclusion Window",
				"exclusion window resource spec", me.Spec,
			)
			return models.ReconcileRequeue
		}

		patch := me.NewPatch()
		me.Status = *meStatus
		err = r.Status().Patch(ctx, me, patch)
		if err != nil {
			l.Error(err, "Cannot patch Exclusion Window status",
				"cluster ID", me.Spec.ClusterID,
				"day of week", me.Spec.DayOfWeek,
				"start hour", me.Spec.StartHour,
				"duration in hours", me.Spec.DurationInHours,
				"maintenance Event metadata", me.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(me, models.DeletionFinalizer)
		me.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, me, patch)
		if err != nil {
			l.Error(err, "Cannot patch Exclusion Window resource metadata",
				"cluster ID", me.Spec.ClusterID,
				"day of week", me.Spec.DayOfWeek,
				"start hour", me.Spec.StartHour,
				"duration in hours", me.Spec.DurationInHours,
				"maintenance Event metadata", me.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Exclusion Window was created",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
		)
	}
	err := r.startMaintenanceEventStatusJob(me)
	if err != nil {
		l.Error(err, "Cannot start Maintenance Event checker status job",
			"exclusion window ID", me.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *MaintenanceEventsReconciler) handleUpdateMaintenanceEvent(
	ctx context.Context,
	me *clusterresourcesv1alpha1.MaintenanceEvents,
	l logr.Logger,
) reconcile.Result {
	err := r.API.UpdateMaintenanceEvent(&me.Spec, instaclustr.MaintenanceEventEndpoint)
	if err != nil {
		l.Error(err, "Cannot update Maintenance Event",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
		)
		return models.ReconcileRequeue
	}

	patch := me.NewPatch()
	me.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch Maintenance Event resource metadata",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
			"maintenance Event metadata", me.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("Maintenance Event resource has been updated",
		"cluster ID", me.Spec.ClusterID,
		"day of week", me.Spec.DayOfWeek,
		"start hour", me.Spec.StartHour,
		"duration in hours", me.Spec.DurationInHours,
	)

	return reconcile.Result{}
}

func (r *MaintenanceEventsReconciler) handleDeleteMaintenanceEvent(
	ctx context.Context,
	me *clusterresourcesv1alpha1.MaintenanceEvents,
	l logr.Logger,
) reconcile.Result {
	patch := me.NewPatch()
	err := r.Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patch Exclusion Window resource metadata",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
			"maintenance Event metadata", me.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetExclusionWindowStatus(me.Spec.ClusterID, instaclustr.ExclusionWindowEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Exclusion Window status from the Instaclustr API",
			"Cluster EventID", me.Spec.ClusterID,
			"Day of week", me.Spec.DayOfWeek,
			"Start hour", me.Spec.StartHour,
			"Duration in hours", me.Spec.DurationInHours,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteExclusionWindow(&me.Status, instaclustr.ExclusionWindowEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Exclusion Window status",
				"cluster ID", me.Spec.ClusterID,
				"day of week", me.Spec.DayOfWeek,
				"start hour", me.Spec.StartHour,
				"duration in hours", me.Spec.DurationInHours,
				"maintenance Event metadata", me.ObjectMeta,
			)
			return models.ReconcileRequeue
		}
	}

	r.Scheduler.RemoveJob(me.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(me, models.DeletionFinalizer)
	me.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, me, patch)
	if err != nil {
		l.Error(err, "Cannot patchExclusion Window resource metadata",
			"cluster ID", me.Spec.ClusterID,
			"day of week", me.Spec.DayOfWeek,
			"start hour", me.Spec.StartHour,
			"duration in hours", me.Spec.DurationInHours,
			"maintenance Event metadata", me.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("Exclusion Window has been deleted",
		"cluster ID", me.Spec.ClusterID,
		"day of week", me.Spec.DayOfWeek,
		"start hour", me.Spec.StartHour,
		"duration in hours", me.Spec.DurationInHours,
	)

	return reconcile.Result{}
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

		instaMEStatus, err := r.API.GetMaintenanceEventStatus(me.Spec.ClusterID, instaclustr.MaintenanceEventEndpoint)
		if err != nil {
			l.Error(err, "Cannot get Maintenance Event Status from Inst API", "maintenance event ID", me.Status.ID)
			return err
		}

		if !isMaintenanceEventStatusesEqual(instaMEStatus.MaintenanceEvents, me.Status.MaintenanceEvents) {
			l.Info("Maintenance Event status of k8s is different from Instaclustr. Reconcile statuses..",
				"maintenance event Status from Inst API", instaMEStatus,
				"maintenance event Status", me.Status)

			patch := me.NewPatch()
			me.Status.MaintenanceEvents = instaMEStatus.MaintenanceEvents
			err := r.Status().Patch(context.Background(), me, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MaintenanceEventsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.MaintenanceEvents{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				// for operator reboots
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
		})).Complete(r)
}
