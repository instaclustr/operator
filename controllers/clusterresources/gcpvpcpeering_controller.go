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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// GCPVPCPeeringReconciler reconciles a GCPVPCPeering object
type GCPVPCPeeringReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GCPVPCPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	gcp := &v1beta1.GCPVPCPeering{}
	err := r.Client.Get(ctx, req.NamespacedName, gcp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "GCP VPC Peering resource is not found", "request", req)
			return ctrl.Result{}, nil
		}
		l.Error(err, "Unable to fetch GCP VPC Peering", "request", req)
		return ctrl.Result{}, err
	}

	switch gcp.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, gcp, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, gcp, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, gcp, l)
	default:
		l.Info("Event isn't handled",
			"project ID", gcp.Spec.PeerProjectID,
			"network name", gcp.Spec.PeerVPCNetworkName,
			"request", req,
			"event", gcp.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}
}

func (r *GCPVPCPeeringReconciler) handleCreateCluster(
	ctx context.Context,
	gcp *v1beta1.GCPVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	if gcp.Status.ID == "" {
		l.Info(
			"Creating GCP VPC Peering resource",
			"project ID", gcp.Spec.PeerProjectID,
			"network name", gcp.Spec.PeerVPCNetworkName,
		)

		gcpStatus, err := r.API.CreatePeering(instaclustr.GCPPeeringEndpoint, &gcp.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create GCP VPC Peering resource",
				"resource spec", gcp.Spec,
			)
			r.EventRecorder.Eventf(
				gcp, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			gcp, models.Normal, models.Created,
			"Resource creation request is sent. Resource ID: %s",
			gcpStatus.ID,
		)

		patch := gcp.NewPatch()
		gcp.Status.PeeringStatus = *gcpStatus
		err = r.Status().Patch(ctx, gcp, patch)
		if err != nil {
			l.Error(err, "Cannot patch GCP VPC Peering resource status",
				"project ID", gcp.Spec.PeerProjectID,
				"network name", gcp.Spec.PeerVPCNetworkName,
				"metadata", gcp.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				gcp, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		controllerutil.AddFinalizer(gcp, models.DeletionFinalizer)
		gcp.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, gcp, patch)
		if err != nil {
			l.Error(err, "Cannot patch GCP VPC Peering resource metadata",
				"project ID", gcp.Spec.PeerProjectID,
				"network name", gcp.Spec.PeerVPCNetworkName,
				"metadata", gcp.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				gcp, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		l.Info(
			"GCP VPC Peering resource was created",
			"id", gcpStatus.ID,
			"project ID", gcp.Spec.PeerProjectID,
			"network name", gcp.Spec.PeerVPCNetworkName,
		)
	}
	err := r.startGCPVPCPeeringStatusJob(gcp)
	if err != nil {
		l.Error(err, "Cannot start GCP VPC Peering checker status job",
			"id", gcp.Status.ID)
		r.EventRecorder.Eventf(
			gcp, models.Warning, models.CreationFailed,
			"Resource status check job is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	r.EventRecorder.Eventf(
		gcp, models.Normal, models.Created,
		"Resource status check job is started",
	)

	return ctrl.Result{}, nil
}

func (r *GCPVPCPeeringReconciler) handleUpdateCluster(
	ctx context.Context,
	gcp *v1beta1.GCPVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	l.Info("Update is not implemented")

	return ctrl.Result{}, nil
}

func (r *GCPVPCPeeringReconciler) handleDeleteCluster(
	ctx context.Context,
	gcp *v1beta1.GCPVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	status, err := r.API.GetPeeringStatus(gcp.Status.ID, instaclustr.GCPPeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get GCP VPC Peering status from the Instaclustr API",
			"id", status.ID,
			"project ID", gcp.Spec.PeerProjectID,
			"network name", gcp.Spec.PeerVPCNetworkName,
		)
		r.EventRecorder.Eventf(
			gcp, models.Warning, models.FetchFailed,
			"Resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	if status != nil {
		err = r.API.DeletePeering(gcp.Status.ID, instaclustr.GCPPeeringEndpoint)
		if err != nil {
			l.Error(err, "Cannot update GCP VPC Peering resource status",
				"id", gcp.Status.ID,
				"project ID", gcp.Spec.PeerProjectID,
				"network ame", gcp.Spec.PeerVPCNetworkName,
				"metadata", gcp.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				gcp, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
		r.EventRecorder.Eventf(
			gcp, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent",
		)
	}

	patch := gcp.NewPatch()
	r.Scheduler.RemoveJob(gcp.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(gcp, models.DeletionFinalizer)
	gcp.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, gcp, patch)
	if err != nil {
		l.Error(err, "Cannot patch GCP VPC Peering resource metadata",
			"id", gcp.Status.ID,
			"project ID", gcp.Spec.PeerProjectID,
			"network ame", gcp.Spec.PeerVPCNetworkName,
			"metadata", gcp.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			gcp, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("GCP VPC Peering has been deleted",
		"id", gcp.Status.ID,
		"project ID", gcp.Spec.PeerProjectID,
		"network name", gcp.Spec.PeerVPCNetworkName,
		"data centre ID", gcp.Spec.DataCentreID,
		"status", gcp.Status.PeeringStatus,
	)

	r.EventRecorder.Eventf(
		gcp, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return ctrl.Result{}, nil
}

func (r *GCPVPCPeeringReconciler) startGCPVPCPeeringStatusJob(gcpPeering *v1beta1.GCPVPCPeering) error {
	job := r.newWatchStatusJob(gcpPeering)

	err := r.Scheduler.ScheduleJob(gcpPeering.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *GCPVPCPeeringReconciler) newWatchStatusJob(gcpPeering *v1beta1.GCPVPCPeering) scheduler.Job {
	l := log.Log.WithValues("component", "GCPVPCPeeringStatusJob")
	return func() error {
		ctx := context.Background()

		key := client.ObjectKeyFromObject(gcpPeering)
		err := r.Get(ctx, key, gcpPeering)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", key,
				)

				r.Scheduler.RemoveJob(gcpPeering.GetJobID(scheduler.StatusChecker))

				return nil
			}

			return err
		}

		instaPeeringStatus, err := r.API.GetPeeringStatus(gcpPeering.Status.ID, instaclustr.GCPPeeringEndpoint)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(ctx, gcpPeering)
			}

			l.Error(err, "Cannot get GCP VPC Peering Status from Inst API", "id", gcpPeering.Status.ID)
			return err
		}

		if !arePeeringStatusesEqual(instaPeeringStatus, &gcpPeering.Status.PeeringStatus) {
			l.Info("GCP VPC Peering status of k8s is different from Instaclustr. Reconcile statuses..",
				"status from Inst API", instaPeeringStatus,
				"status", gcpPeering.Status)

			patch := gcpPeering.NewPatch()
			gcpPeering.Status.PeeringStatus = *instaPeeringStatus
			err := r.Status().Patch(ctx, gcpPeering, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *GCPVPCPeeringReconciler) handleExternalDelete(ctx context.Context, key *v1beta1.GCPVPCPeering) error {
	l := log.FromContext(ctx)

	patch := key.NewPatch()
	key.Status.StatusCode = models.DeletedStatus
	err := r.Status().Patch(ctx, key, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(key, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(key.GetJobID(scheduler.StatusChecker))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPVPCPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.GCPVPCPeering{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.GCPVPCPeering)
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
