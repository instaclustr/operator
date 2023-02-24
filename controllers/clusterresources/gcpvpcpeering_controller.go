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

// GCPVPCPeeringReconciler reconciles a GCPVPCPeering object
type GCPVPCPeeringReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GCPVPCPeering object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GCPVPCPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var gcp clusterresourcesv1alpha1.GCPVPCPeering
	err := r.Client.Get(ctx, req.NamespacedName, &gcp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "GCP VPC Peering resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "unable to fetch GCP VPC Peering", "request", req)
		return reconcile.Result{}, err
	}

	switch gcp.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &gcp, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &gcp, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &gcp, l), nil
	default:
		l.Info("event isn't handled",
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
			"Request", req,
			"event", gcp.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *GCPVPCPeeringReconciler) handleCreateCluster(
	ctx context.Context,
	gcp *clusterresourcesv1alpha1.GCPVPCPeering,
	l logr.Logger,
) reconcile.Result {

	if gcp.Status.ID == "" {
		l.Info(
			"Creating GCP VPC Peering resource",
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
		)

		gcpStatus, err := r.API.CreatePeering(instaclustr.GCPPeeringEndpoint, &gcp.Spec)
		if err != nil {
			l.Error(
				err, "cannot create GCP VPC Peering resource",
				"GCP VPC Peering resource spec", gcp.Spec,
			)
			return models.ReconcileRequeue
		}

		patch := gcp.NewPatch()
		gcp.Status.PeeringStatus = *gcpStatus
		err = r.Status().Patch(ctx, gcp, patch)
		if err != nil {
			l.Error(err, "cannot patch GCP VPC Peering resource status",
				"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
				"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
				"GCP VPC Peering metadata", gcp.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(gcp, models.DeletionFinalizer)
		gcp.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, gcp, patch)
		if err != nil {
			l.Error(err, "cannot patch GCP VPC Peering resource metadata",
				"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
				"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
				"GCP VPC Peering metadata", gcp.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"GCP VPC Peering resource was created",
			"GCP Peering ID", gcpStatus.ID,
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
		)
	}
	err := r.startGCPVPCPeeringStatusJob(gcp)
	if err != nil {
		l.Error(err, "cannot start GCP VPC Peering checker status job",
			"GCP VPC Peering ID", gcp.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *GCPVPCPeeringReconciler) handleUpdateCluster(
	ctx context.Context,
	gcp *clusterresourcesv1alpha1.GCPVPCPeering,
	l logr.Logger,
) reconcile.Result {
	l.Info("Update is not implemented")

	return reconcile.Result{}
}

func (r *GCPVPCPeeringReconciler) handleDeleteCluster(
	ctx context.Context,
	gcp *clusterresourcesv1alpha1.GCPVPCPeering,
	l logr.Logger,
) reconcile.Result {

	patch := gcp.NewPatch()
	err := r.Patch(ctx, gcp, patch)
	if err != nil {
		l.Error(err, "cannot patch GCP VPC Peering resource metadata",
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
			"GCP VPC Peering metadata", gcp.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetPeeringStatus(gcp.Status.ID, instaclustr.GCPPeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get GCP VPC Peering status from the Instaclustr API",
			"GCP Peering ID", status.ID,
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		r.Scheduler.RemoveJob(gcp.GetJobID(scheduler.StatusChecker))
		err = r.API.DeletePeering(gcp.Status.ID, instaclustr.GCPPeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot update GCP VPC Peering resource statuss",
				"GCP VPC Peering ID", gcp.Status.ID,
				"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
				"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
				"GCP VPC Peering metadata", gcp.ObjectMeta,
			)
			return models.ReconcileRequeue

		}
		return models.ReconcileRequeue
	}

	controllerutil.RemoveFinalizer(gcp, models.DeletionFinalizer)
	gcp.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, gcp, patch)
	if err != nil {
		l.Error(err, "cannot patch GCP VPC Peering resource metadata",
			"GCP Peering ID", gcp.Status.ID,
			"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
			"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
			"GCP VPC Peering metadata", gcp.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("GCP VPC Peering has been deleted",
		"GCP VPC Peering ID", gcp.Status.ID,
		"GCP VPC Peering Project ID", gcp.Spec.PeerProjectID,
		"GCP VPC Peering Network Name", gcp.Spec.PeerVPCNetworkName,
		"GCP VPC Peering Data Centre ID", gcp.Spec.DataCentreID,
		"GCP VPC Peering Status", gcp.Status.PeeringStatus,
	)

	return reconcile.Result{}
}

func (r *GCPVPCPeeringReconciler) startGCPVPCPeeringStatusJob(gcpPeering *clusterresourcesv1alpha1.GCPVPCPeering) error {
	job := r.newWatchStatusJob(gcpPeering)

	err := r.Scheduler.ScheduleJob(gcpPeering.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *GCPVPCPeeringReconciler) newWatchStatusJob(gcpPeering *clusterresourcesv1alpha1.GCPVPCPeering) scheduler.Job {
	l := log.Log.WithValues("component", "GCPVPCPeeringStatusJob")
	return func() error {
		instaPeeringStatus, err := r.API.GetPeeringStatus(gcpPeering.Status.ID, instaclustr.GCPPeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot get GCP VPC Peering Status from Inst API", "GCP VPC Peering ID", gcpPeering.Status.ID)
			return err
		}

		if !isPeeringStatusesEqual(instaPeeringStatus, &gcpPeering.Status.PeeringStatus) {
			l.Info("GCP VPC Peering status of k8s is different from Instaclustr. Reconcile statuses..",
				"GCP VPC Peering Status from Inst API", instaPeeringStatus,
				"GCP VPC Peering Status", gcpPeering.Status)

			patch := gcpPeering.NewPatch()
			gcpPeering.Status.PeeringStatus = *instaPeeringStatus
			err := r.Status().Patch(context.Background(), gcpPeering, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPVPCPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.GCPVPCPeering{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
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
