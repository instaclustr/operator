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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// ExclusionWindowReconciler reconciles a ExclusionWindow object
type ExclusionWindowReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=exclusionwindows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=exclusionwindows/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=exclusionwindows/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExclusionWindow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ExclusionWindowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	ew := &v1beta1.ExclusionWindow{}
	err := r.Client.Get(ctx, req.NamespacedName, ew)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Exclusion Window resource is not found", "request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Cannot get Exclusion Window resource", "request", req)
		return models.ReconcileRequeue, nil
	}

	switch ew.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateWindow(ctx, ew, l), nil
	case models.DeletingEvent:
		return r.handleDeleteWindow(ctx, ew, l), nil
	default:
		l.Info("event isn't handled",
			"Cluster ID", ew.Spec.ClusterID,
			"Exclusion Window Spec", ew.Spec,
			"Request", req,
			"event", ew.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *ExclusionWindowReconciler) handleCreateWindow(
	ctx context.Context,
	ew *v1beta1.ExclusionWindow,
	l logr.Logger,
) reconcile.Result {
	if ew.Status.ID == "" {
		l.Info(
			"Creating Exclusion Window resource",
			"Cluster ID", ew.Spec.ClusterID,
			"Exclusion Window Spec", ew.Spec,
		)

		id, err := r.API.CreateExclusionWindow(ew.Spec.ClusterID, &ew.Spec)
		if err != nil {
			l.Error(
				err, "cannot create Exclusion Window resource",
				"Exclusion Window resource spec", ew.Spec,
			)
			r.EventRecorder.Eventf(
				ew, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		r.EventRecorder.Eventf(
			ew, models.Normal, models.Created,
			"Resource creation request is sent. Resource ID: %s",
			id,
		)

		patch := ew.NewPatch()
		ew.Status.ID = id
		err = r.Status().Patch(ctx, ew, patch)
		if err != nil {
			l.Error(err, "cannot patch Exclusion Window resource status after creation",
				"Cluster ID", ew.Spec.ClusterID,
				"Exclusion Window Spec", ew.Spec,
				"Exclusion Window metadata", ew.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				ew, models.Warning, models.PatchFailed,
				"Status patch is failed after resource creation. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(ew, models.DeletionFinalizer)
		ew.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, ew, patch)
		if err != nil {
			l.Error(err, "cannot patch Exclusion Window resource metadata with created event",
				"Cluster ID", ew.Spec.ClusterID,
				"Exclusion Window Spec", ew.Spec,
				"Exclusion Window metadata", ew.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				ew, models.Warning, models.PatchFailed,
				"Resource patch is failed after resource creation. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Exclusion Window resource was created",
			"Cluster ID", ew.Spec.ClusterID,
			"Exclusion Window Spec", ew.Spec,
		)
	}
	return models.ExitReconcile
}

func (r *ExclusionWindowReconciler) handleDeleteWindow(
	ctx context.Context,
	ew *v1beta1.ExclusionWindow,
	l logr.Logger,
) reconcile.Result {
	status, err := r.API.GetExclusionWindowsStatus(ew.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Exclusion Window status from the Instaclustr API",
			"Cluster ID", ew.Spec.ClusterID,
			"Exclusion Window Spec", ew.Spec,
		)
		r.EventRecorder.Eventf(
			ew, models.Warning, models.FetchFailed,
			"Resource fetch from the Instaclustr API is failed while deletion. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if status != "" {
		err = r.API.DeleteExclusionWindow(ew.Status.ID)
		if err != nil {
			l.Error(err, "cannot delete Exclusion Window resource",
				"Cluster ID", ew.Spec.ClusterID,
				"Exclusion Window Spec", ew.Spec,
				"Exclusion Window metadata", ew.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				ew, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		r.EventRecorder.Eventf(
			ew, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent",
		)
	}

	patch := ew.NewPatch()
	controllerutil.RemoveFinalizer(ew, models.DeletionFinalizer)
	ew.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, ew, patch)
	if err != nil {
		l.Error(err, "cannot patch Exclusion Window resource metadata with deleted event",
			"Cluster ID", ew.Spec.ClusterID,
			"Exclusion Window Spec", ew.Spec,
			"Exclusion Window metadata", ew.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			ew, models.Warning, models.PatchFailed,
			"Resource patch is failed while deletion. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Exclusion Window has been deleted",
		"Cluster ID", ew.Spec.ClusterID,
		"Exclusion Window Spec", ew.Spec,
		"Exclusion Window Status", ew.Status,
	)

	r.EventRecorder.Eventf(
		ew, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return models.ExitReconcile
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExclusionWindowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.ExclusionWindow{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.ExclusionWindow)
				if newObj.DeletionTimestamp != nil {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if newObj.Generation == event.ObjectOld.GetGeneration() {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
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
