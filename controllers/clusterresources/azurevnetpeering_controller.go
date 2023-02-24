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

// AzureVNetPeeringReconciler reconciles a AzureVNetPeering object
type AzureVNetPeeringReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AzureVNetPeering object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *AzureVNetPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var azure clusterresourcesv1alpha1.AzureVNetPeering
	err := r.Client.Get(ctx, req.NamespacedName, &azure)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Azure VNet Peering resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "unable to fetch Azure VNet Peering", "request", req)
		return reconcile.Result{}, err
	}

	switch azure.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &azure, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &azure, &l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &azure, &l), nil
	default:
		l.Info("event isn't handled",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			"Request", req,
			"event", azure.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *AzureVNetPeeringReconciler) handleCreateCluster(
	ctx context.Context,
	azure *clusterresourcesv1alpha1.AzureVNetPeering,
	l logr.Logger,
) reconcile.Result {

	if azure.Status.ID == "" {
		l.Info(
			"Creating Azure VNet Peering resource",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
		)

		azureStatus, err := r.API.CreatePeering(instaclustr.AzurePeeringEndpoint, &azure.Spec)
		if err != nil {
			l.Error(
				err, "cannot create Azure VNet Peering resource",
				"Azure VNet Peering resource spec", azure.Spec,
			)
			return models.ReconcileRequeue
		}

		patch := azure.NewPatch()
		azure.Status.PeeringStatus = *azureStatus
		err = r.Status().Patch(ctx, azure, patch)
		if err != nil {
			l.Error(err, "cannot patch Azure VNet Peering resource status",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(azure, models.DeletionFinalizer)
		azure.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, azure, patch)
		if err != nil {
			l.Error(err, "cannot patch Azure VNet Peering resource metadata",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
				"Azure VNet Peering metadata", azure.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Azure VNet Peering resource was created",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			"Peer Subnets", azure.Spec.PeerSubnets,
		)
	}
	err := r.startAzureVNetPeeringStatusJob(azure)
	if err != nil {
		l.Error(err, "cannot start Azure VNet Peering checker status job",
			"Azure VNet Peering ID", azure.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *AzureVNetPeeringReconciler) handleUpdateCluster(
	ctx context.Context,
	azure *clusterresourcesv1alpha1.AzureVNetPeering,
	l *logr.Logger,
) reconcile.Result {
	l.Info("Update is not implemented")

	return reconcile.Result{}
}

func (r *AzureVNetPeeringReconciler) handleDeleteCluster(
	ctx context.Context,
	azure *clusterresourcesv1alpha1.AzureVNetPeering,
	l *logr.Logger,
) reconcile.Result {

	patch := azure.NewPatch()
	err := r.Patch(ctx, azure, patch)
	if err != nil {
		l.Error(err, "cannot patch Azure VNet Peering resource metadata",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			"Azure VNet Peering metadata", azure.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetPeeringStatus(azure.Status.ID, instaclustr.AzurePeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Azure VNet Peering status from the Instaclustr API",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		r.Scheduler.RemoveJob(azure.GetJobID(scheduler.StatusChecker))
		err = r.API.DeletePeering(azure.Status.ID, instaclustr.AzurePeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot update Azure VNet Peering resource statuss",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
				"Azure VNet Peering metadata", azure.ObjectMeta,
			)
			return models.ReconcileRequeue

		}
		return models.ReconcileRequeue
	}

	controllerutil.RemoveFinalizer(azure, models.DeletionFinalizer)
	azure.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, azure, patch)
	if err != nil {
		l.Error(err, "cannot patch Azure VNet Peering resource metadata",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			"Azure VNet Peering metadata", azure.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("Azure VNet Peering has been deleted",
		"Azure VNet Peering ID", azure.Status.ID,
		"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
		"AD Object ID", azure.Spec.PeerADObjectID,
		"Resource Group", azure.Spec.PeerResourceGroup,
		"Vnet Name", azure.Spec.PeerVirtualNetworkName,
		"Azure VNet Peering Status", azure.Status.PeeringStatus,
	)

	return reconcile.Result{}
}

func (r *AzureVNetPeeringReconciler) startAzureVNetPeeringStatusJob(azurePeering *clusterresourcesv1alpha1.AzureVNetPeering,
) error {
	job := r.newWatchStatusJob(azurePeering)

	err := r.Scheduler.ScheduleJob(azurePeering.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AzureVNetPeeringReconciler) newWatchStatusJob(azureVNetPeering *clusterresourcesv1alpha1.AzureVNetPeering,
) scheduler.Job {
	l := log.Log.WithValues("component", "AzureVNetPeeringStatusJob")
	return func() error {
		instaPeeringStatus, err := r.API.GetPeeringStatus(azureVNetPeering.Status.ID, instaclustr.AzurePeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot get Azure VNet Peering Status from Inst API", "Azure VNet Peering ID", azureVNetPeering.Status.ID)
			return err
		}

		if !isPeeringStatusesEqual(instaPeeringStatus, &azureVNetPeering.Status.PeeringStatus) {
			l.Info("Azure VNet Peering status of k8s is different from Instaclustr. Reconcile statuses..",
				"Azure VNet Peering Status from Inst API", instaPeeringStatus,
				"Azure VNet Peering Status", azureVNetPeering.Status)

			patch := azureVNetPeering.NewPatch()
			azureVNetPeering.Status.PeeringStatus = *instaPeeringStatus
			err := r.Status().Patch(context.Background(), azureVNetPeering, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AzureVNetPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.AzureVNetPeering{}, builder.WithPredicates(predicate.Funcs{
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
