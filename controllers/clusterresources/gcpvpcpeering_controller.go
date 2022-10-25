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
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	clusterresourcesv2alpha1 "github.com/instaclustr/operator/apis/clusterresources/v2alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	GCPPeeringFinalizer = "gcppeering-insta-finalizer/finalizer"
)

// GCPVPCPeeringReconciler reconciles a GCPVPCPeering object
type GCPVPCPeeringReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
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

	gcpPeering := &clusterresourcesv2alpha1.GCPVPCPeering{}
	err := r.Client.Get(ctx, req.NamespacedName, gcpPeering)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		l.Error(err, "unable to fetch GCP VPC Peering resource")
		return reconcile.Result{}, err
	}

	awsPeeringAnnotations := gcpPeering.Annotations
	switch awsPeeringAnnotations[models.CurrentEventAnnotation] {
	case models.CreateEvent:
		err = r.HandleCreateCluster(gcpPeering, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot create GCP VPC Peering resource",
				"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
				"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
				"Data Centre ID", gcpPeering.Spec.CDCID,
				"Subnets", gcpPeering.Spec.PeerSubnets,
			)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	case models.UpdateEvent:
		reconcileResult, err := r.HandleUpdateCluster(gcpPeering, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot update GCP VPC Peering resource",
				"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
				"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
				"Data Centre ID", gcpPeering.Spec.CDCID,
				"Subnets", gcpPeering.Spec.PeerSubnets,
			)
			return reconcile.Result{}, err
		}
		return *reconcileResult, nil
	case models.DeleteEvent:
		err = r.HandleDeleteCluster(gcpPeering, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot delete GCP VPC Peering resource",
				"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
				"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
				"Data Centre ID", gcpPeering.Spec.CDCID,
				"Subnets", gcpPeering.Spec.PeerSubnets,
			)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	default:
		l.Info("UNKNOWN EVENT")
		return reconcile.Result{}, err
	}
}

func (r *GCPVPCPeeringReconciler) HandleCreateCluster(
	gcpPeering *clusterresourcesv2alpha1.GCPVPCPeering,
	l *logr.Logger,
	ctx *context.Context,
) error {

	l.Info(
		"Creating GCP VPC Peering resource",
		"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
		"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
		"Data Centre ID", gcpPeering.Spec.CDCID,
		"Subnets", gcpPeering.Spec.PeerSubnets,
	)

	awsStatus, err := r.API.CreateGCPPeering(instaclustr.GCPEndpoint, &gcpPeering.Spec)
	if err != nil {
		l.Error(
			err, "cannot create GCP VPC Peering resource",
			"GCP VPC Peering resource spec", gcpPeering.Spec,
		)
		return err
	}

	gcpPeering.Annotations[models.PreviousEventAnnotation] = gcpPeering.Annotations[models.CurrentEventAnnotation]
	gcpPeering.Annotations[models.CurrentEventAnnotation] = models.UpdateEvent

	err = r.patchPeeringMetadata(ctx, gcpPeering, l)
	if err != nil {
		l.Error(err, "cannot patch GCP VPC Peering resource metadata",
			"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
			"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
			"Data Centre ID", gcpPeering.Spec.CDCID,
			"Subnets", gcpPeering.Spec.PeerSubnets,
			"Cluster ID", gcpPeering.Status.ID,
			"Cluster metadata", gcpPeering.ObjectMeta,
		)
		return err
	}

	gcpPeering.Status = *awsStatus
	err = r.Status().Update(*ctx, gcpPeering)
	if err != nil {
		l.Error(err, "cannot update GCP VPC Peering resource statuss",
			"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
			"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
			"Data Centre ID", gcpPeering.Spec.CDCID,
			"Subnets", gcpPeering.Spec.PeerSubnets,
			"Cluster ID", gcpPeering.Status.ID,
			"Cluster metadata", gcpPeering.ObjectMeta,
		)
		return err
	}

	l.Info(
		"GCP VPC Peering resource was created",
		"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
		"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
		"Data Centre ID", gcpPeering.Spec.CDCID,
		"Subnets", gcpPeering.Spec.PeerSubnets,
	)
	return nil
}

func (r *GCPVPCPeeringReconciler) HandleUpdateCluster(
	gcpPeering *clusterresourcesv2alpha1.GCPVPCPeering,
	l *logr.Logger,
	ctx *context.Context,
) (*reconcile.Result, error) {

	l.Info("GCP VPC Peering creation is not allowed yet",
		"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
		"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
		"Data Centre ID", gcpPeering.Spec.CDCID,
		"Subnets", gcpPeering.Spec.PeerSubnets,
		"Cluster ID", gcpPeering.Status.ID,
		"AWS VPC Peering Status", gcpPeering.Status.VPCPeeringStatus,
	)

	return &reconcile.Result{}, nil
}

func (r *GCPVPCPeeringReconciler) HandleDeleteCluster(
	gcpPeering *clusterresourcesv2alpha1.GCPVPCPeering,
	l *logr.Logger,
	ctx *context.Context,
) error {

	err := r.API.DeleteGCPPeering(gcpPeering.Status.ID, instaclustr.GCPEndpoint)
	if err != nil {
		l.Error(err, "cannot update GCP VPC Peering resource statuss",
			"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
			"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
			"Data Centre ID", gcpPeering.Spec.CDCID,
			"Subnets", gcpPeering.Spec.PeerSubnets,
			"Cluster ID", gcpPeering.Status.ID,
			"Cluster metadata", gcpPeering.ObjectMeta,
		)
		return err
	}

	l.Info("AWS VPC Peering has been deleted",
		"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
		"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
		"Data Centre ID", gcpPeering.Spec.CDCID,
		"Subnets", gcpPeering.Spec.PeerSubnets,
		"Cluster ID", gcpPeering.Status.ID,
		"AWS VPC Peering Status", gcpPeering.Status.VPCPeeringStatus,
	)

	controllerutil.RemoveFinalizer(gcpPeering, GCPPeeringFinalizer)
	err = r.Update(*ctx, gcpPeering)
	if err != nil {
		l.Error(
			err, "cannot update GCP VPC Peering CRD",
			"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
			"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
			"Data Centre ID", gcpPeering.Spec.CDCID,
			"Subnets", gcpPeering.Spec.PeerSubnets,
		)
		return err
	}
	return nil
}

func (r *GCPVPCPeeringReconciler) patchPeeringMetadata(
	ctx *context.Context,
	gcpPeering *clusterresourcesv2alpha1.GCPVPCPeering,
	l *logr.Logger,
) error {
	patchRequest := []*clusterresourcesv2alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(gcpPeering.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clusterresourcesv2alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(gcpPeering.Finalizers)
	if err != nil {
		return err
	}

	finzlizersPatch := &clusterresourcesv2alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(*ctx, gcpPeering, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	l.Info("GCP VPC Peering resource patched",
		"GCP Peer Project ID", gcpPeering.Spec.PeerProjectID,
		"Peer VPC Network Name", gcpPeering.Spec.PeerVPCNetworkName,
		"Data Centre ID", gcpPeering.Spec.CDCID,
		"Subnets", gcpPeering.Spec.PeerSubnets,
		"Finalizers", gcpPeering.Finalizers,
		"Annotations", gcpPeering.Annotations,
	)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPVPCPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv2alpha1.GCPVPCPeering{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.CreateEvent})
				event.Object.SetFinalizers([]string{GCPPeeringFinalizer})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.DeleteEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					models.CurrentEventAnnotation: models.UpdateEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{models.CurrentEventAnnotation: models.UnknownEvent})
				return true
			},
		})).
		Complete(r)
}
