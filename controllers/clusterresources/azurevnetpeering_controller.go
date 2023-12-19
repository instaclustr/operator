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

// AzureVNetPeeringReconciler reconciles a AzureVNetPeering object
type AzureVNetPeeringReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=azurevnetpeerings/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *AzureVNetPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	azure := &v1beta1.AzureVNetPeering{}
	err := r.Client.Get(ctx, req.NamespacedName, azure)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Azure VNet Peering resource is not found", "request", req)
			return ctrl.Result{}, nil
		}
		l.Error(err, "unable to fetch Azure VNet Peering", "request", req)
		return ctrl.Result{}, err
	}

	switch azure.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreatePeering(ctx, azure, l)
	case models.UpdatingEvent:
		return r.handleUpdatePeering(ctx, azure, &l)
	case models.DeletingEvent:
		return r.handleDeletePeering(ctx, azure, &l)
	default:
		l.Info("event isn't handled",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			"Request", req,
			"event", azure.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}
}

func (r *AzureVNetPeeringReconciler) handleCreatePeering(
	ctx context.Context,
	azure *v1beta1.AzureVNetPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	if azure.Status.ID == "" {
		var cdcID string
		var err error
		if azure.Spec.ClusterRef != nil {
			cdcID, err = GetDataCentreID(r.Client, ctx, azure.Spec.ClusterRef)
			if err != nil {
				l.Error(err, "Cannot get CDCID",
					"Cluster reference", azure.Spec.ClusterRef,
				)
				return ctrl.Result{}, err
			}
			l.Info(
				"Creating Azure VNet Peering resource from the cluster reference",
				"cluster reference", azure.Spec.ClusterRef,
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			)
		} else {
			cdcID = azure.Spec.DataCentreID
			l.Info(
				"Creating Azure VNet Peering resource",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			)
		}

		azureStatus, err := r.API.CreateAzureVNetPeering(&azure.Spec, cdcID)
		if err != nil {
			l.Error(
				err, "cannot create Azure VNet Peering resource",
				"Azure VNet Peering resource spec", azure.Spec,
			)
			r.EventRecorder.Eventf(
				azure, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			azure, models.Normal, models.Created,
			"Resource creation request is sent. Resource ID: %s",
			azureStatus.ID,
		)

		patch := azure.NewPatch()
		azure.Status.PeeringStatus = *azureStatus
		azure.Status.CDCID = cdcID
		err = r.Status().Patch(ctx, azure, patch)
		if err != nil {
			l.Error(err, "cannot patch Azure VNet Peering resource status",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
			)
			r.EventRecorder.Eventf(
				azure, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
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
			r.EventRecorder.Eventf(
				azure, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
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
		r.EventRecorder.Eventf(
			azure, models.Warning, models.CreationFailed,
			"Resource status check job is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	r.EventRecorder.Eventf(
		azure, models.Normal, models.Created,
		"Resource status check job is started",
	)

	return ctrl.Result{}, nil
}

func (r *AzureVNetPeeringReconciler) handleUpdatePeering(
	ctx context.Context,
	azure *v1beta1.AzureVNetPeering,
	l *logr.Logger,
) (ctrl.Result, error) {
	l.Info("Update is not implemented")

	return ctrl.Result{}, nil
}

func (r *AzureVNetPeeringReconciler) handleDeletePeering(
	ctx context.Context,
	azure *v1beta1.AzureVNetPeering,
	l *logr.Logger,
) (ctrl.Result, error) {
	status, err := r.API.GetPeeringStatus(azure.Status.ID, instaclustr.AzurePeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Azure VNet Peering status from the Instaclustr API",
			"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
			"AD Object ID", azure.Spec.PeerADObjectID,
			"Resource Group", azure.Spec.PeerResourceGroup,
			"Vnet Name", azure.Spec.PeerVirtualNetworkName,
		)
		r.EventRecorder.Eventf(
			azure, models.Warning, models.FetchFailed,
			"Resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	if status != nil {
		r.Scheduler.RemoveJob(azure.GetJobID(scheduler.StatusChecker))
		err = r.API.DeletePeering(azure.Status.ID, instaclustr.AzurePeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot update Azure VNet Peering resource status",
				"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
				"AD Object ID", azure.Spec.PeerADObjectID,
				"Resource Group", azure.Spec.PeerResourceGroup,
				"Vnet Name", azure.Spec.PeerVirtualNetworkName,
				"Azure VNet Peering metadata", azure.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				azure, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			azure, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent",
		)
	}

	patch := azure.NewPatch()
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
		r.EventRecorder.Eventf(
			azure, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("Azure VNet Peering has been deleted",
		"Azure VNet Peering ID", azure.Status.ID,
		"Azure Subscription ID", azure.Spec.PeerSubscriptionID,
		"AD Object ID", azure.Spec.PeerADObjectID,
		"Resource Group", azure.Spec.PeerResourceGroup,
		"Vnet Name", azure.Spec.PeerVirtualNetworkName,
		"Azure VNet Peering Status", azure.Status.PeeringStatus,
	)

	r.EventRecorder.Eventf(
		azure, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return ctrl.Result{}, nil
}

func (r *AzureVNetPeeringReconciler) startAzureVNetPeeringStatusJob(azurePeering *v1beta1.AzureVNetPeering,
) error {
	job := r.newWatchStatusJob(azurePeering)

	err := r.Scheduler.ScheduleJob(azurePeering.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AzureVNetPeeringReconciler) newWatchStatusJob(azureVNetPeering *v1beta1.AzureVNetPeering,
) scheduler.Job {
	l := log.Log.WithValues("component", "AzureVNetPeeringStatusJob")
	return func() error {
		ctx := context.Background()

		key := client.ObjectKeyFromObject(azureVNetPeering)
		err := r.Get(ctx, key, azureVNetPeering)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", key,
				)

				r.Scheduler.RemoveJob(azureVNetPeering.GetJobID(scheduler.StatusChecker))

				return nil
			}

			return err
		}

		instaPeeringStatus, err := r.API.GetPeeringStatus(azureVNetPeering.Status.ID, instaclustr.AzurePeeringEndpoint)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(ctx, azureVNetPeering)
			}

			l.Error(err, "cannot get Azure VNet Peering Status from Inst API", "Azure VNet Peering ID", azureVNetPeering.Status.ID)
			return err
		}

		if !arePeeringStatusesEqual(instaPeeringStatus, &azureVNetPeering.Status.PeeringStatus) {
			l.Info("Azure VNet Peering status of k8s is different from Instaclustr. Reconcile statuses..",
				"Azure VNet Peering Status from Inst API", instaPeeringStatus,
				"Azure VNet Peering Status", azureVNetPeering.Status)

			patch := azureVNetPeering.NewPatch()
			azureVNetPeering.Status.PeeringStatus = *instaPeeringStatus
			err := r.Status().Patch(ctx, azureVNetPeering, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *AzureVNetPeeringReconciler) handleExternalDelete(ctx context.Context, key *v1beta1.AzureVNetPeering) error {
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
func (r *AzureVNetPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.AzureVNetPeering{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.AzureVNetPeering)
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
