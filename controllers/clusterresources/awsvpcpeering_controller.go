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

// AWSVPCPeeringReconciler reconciles a AWSVPCPeering object
type AWSVPCPeeringReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AWSVPCPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	aws := &v1beta1.AWSVPCPeering{}
	err := r.Client.Get(ctx, req.NamespacedName, aws)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "AWS VPC Peering resource is not found", "request", req)
			return ctrl.Result{}, nil
		}
		l.Error(err, "unable to fetch AWS VPC Peering", "request", req)
		return ctrl.Result{}, err
	}

	switch aws.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreatePeering(ctx, aws, l)
	case models.UpdatingEvent:
		return r.handleUpdatePeering(ctx, aws, l)
	case models.DeletingEvent:
		return r.handleDeletePeering(ctx, aws, l)
	default:
		l.Info("event isn't handled",
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
			"Request", req,
			"event", aws.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}
}

func (r *AWSVPCPeeringReconciler) handleCreatePeering(
	ctx context.Context,
	aws *v1beta1.AWSVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	if aws.Status.ID == "" {
		var cdcID string
		var err error
		if aws.Spec.ClusterRef != nil {
			cdcID, err = GetDataCentreID(r.Client, ctx, aws.Spec.ClusterRef)
			if err != nil {
				l.Error(err, "Cannot get CDCID",
					"Cluster reference", aws.Spec.ClusterRef,
				)
				return ctrl.Result{}, err
			}
			l.Info(
				"Creating AWS VPC Peering resource from the cluster reference",
				"cluster reference", aws.Spec.ClusterRef,
				"cdcID ID", cdcID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"Region", aws.Spec.PeerRegion,
			)
		} else {
			cdcID = aws.Spec.DataCentreID
			l.Info(
				"Creating AWS VPC Peering resource",
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"Region", aws.Spec.PeerRegion,
			)
		}

		awsStatus, err := r.API.CreateAWSVPCPeering(&aws.Spec, cdcID)
		if err != nil {
			l.Error(
				err, "cannot create AWS VPC Peering resource",
				"AWS VPC Peering resource spec", aws.Spec,
			)
			r.EventRecorder.Eventf(
				aws, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			aws, models.Normal, models.Created,
			"Resource creation request is sent. Resource ID: %s",
			awsStatus.ID,
		)

		patch := aws.NewPatch()
		aws.Status.PeeringStatus = *awsStatus
		aws.Status.CDCID = cdcID
		err = r.Status().Patch(ctx, aws, patch)
		if err != nil {
			l.Error(err, "cannot patch AWS VPC Peering resource status",
				"AWS Peering ID", awsStatus.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"Region", aws.Spec.PeerRegion,
				"AWS VPC Peering metadata", aws.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				aws, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		controllerutil.AddFinalizer(aws, models.DeletionFinalizer)
		aws.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, aws, patch)
		if err != nil {
			l.Error(err, "cannot patch AWS VPC Peering resource metadata",
				"AWS Peering ID", awsStatus.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"Region", aws.Spec.PeerRegion,
				"AWS VPC Peering metadata", aws.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				aws, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		l.Info(
			"AWS VPC Peering resource was created",
			"AWS Peering ID", awsStatus.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
		)
	}
	err := r.startAWSVPCPeeringStatusJob(aws)
	if err != nil {
		l.Error(err, "cannot start AWS VPC Peering checker status job",
			"AWS VPC Peering ID", aws.Status.ID)
		r.EventRecorder.Eventf(
			aws, models.Warning, models.CreationFailed,
			"Resource status check job is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	r.EventRecorder.Eventf(
		aws, models.Normal, models.Created,
		"Resource status check job is started",
	)

	return ctrl.Result{}, nil
}

func (r *AWSVPCPeeringReconciler) handleUpdatePeering(
	ctx context.Context,
	aws *v1beta1.AWSVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	instaAWSPeering, err := r.API.GetAWSVPCPeering(aws.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get AWS VPC Peering from Instaclutr",
			"AWS VPC Peering ID", aws.Status.ID,
		)
		r.EventRecorder.Eventf(aws, models.Warning, models.UpdateFailed,
			"Cannot get AWS VPC Peering from Instaclutr. Reason: %v",
		)

		return ctrl.Result{}, err
	}

	if aws.Annotations[models.ExternalChangesAnnotation] == models.True {
		if !subnetsEqual(instaAWSPeering.PeerSubnets, aws.Spec.PeerSubnets) {
			l.Info("The resource specification still differs from the Instaclustr resource specification, please reconcile it manually",
				"AWS VPC ID", aws.Status.ID,
				"k8s peerSubnets", aws.Spec.PeerSubnets,
				"instaclustr peerSubnets", instaAWSPeering.PeerSubnets,
			)
			r.EventRecorder.Eventf(aws, models.Warning, models.UpdateFailed,
				"The resource specification still differs from the Instaclustr resource specification, please reconcile it manually.",
			)

			return ctrl.Result{}, nil
		}

		patch := aws.NewPatch()
		delete(aws.Annotations, models.ExternalChangesAnnotation)
		err = r.Patch(ctx, aws, patch)
		if err != nil {
			l.Error(err, "Cannot delete external changes annotation from the resource",
				"AWS VPC Peering ID", aws.Status.ID,
			)
			r.EventRecorder.Eventf(
				aws, models.Warning, models.PatchFailed,
				"Deleting external changes annotation is failed. Reason: %v",
				err,
			)

			return ctrl.Result{}, err
		}

		l.Info("External changes of the k8s resource specification was fixed",
			"AWS VPC Peering ID", aws.Status.ID,
		)
		r.EventRecorder.Eventf(aws, models.Normal, models.ExternalChanges,
			"External changes of the k8s resource specification was fixed",
		)

		return ctrl.Result{}, nil
	}

	err = r.API.UpdatePeering(aws.Status.ID, instaclustr.AWSPeeringEndpoint, &aws.Spec)
	if err != nil {
		l.Error(err, "cannot update AWS VPC Peering",
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
			"Subnets", aws.Spec.PeerSubnets,
		)
		r.EventRecorder.Eventf(
			aws, models.Warning, models.UpdateFailed,
			"Resource update on the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	patch := aws.NewPatch()
	aws.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, aws, patch)
	if err != nil {
		l.Error(err, "cannot patch AWS VPC Peering resource metadata",
			"AWS Peering ID", aws.Status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
			"AWS VPC Peering metadata", aws.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			aws, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("AWS VPC Peering resource has been updated",
		"AWS VPC Peering ID", aws.Status.ID,
		"AWS Account ID", aws.Spec.PeerAWSAccountID,
		"VPC ID", aws.Spec.PeerVPCID,
		"Region", aws.Spec.PeerRegion,
		"AWS VPC Peering Data Centre ID", aws.Status.CDCID,
		"AWS VPC Peering Status", aws.Status.PeeringStatus,
	)

	return ctrl.Result{}, nil
}

func (r *AWSVPCPeeringReconciler) handleDeletePeering(
	ctx context.Context,
	aws *v1beta1.AWSVPCPeering,
	l logr.Logger,
) (ctrl.Result, error) {
	status, err := r.API.GetPeeringStatus(aws.Status.ID, instaclustr.AWSPeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get AWS VPC Peering status from the Instaclustr API",
			"AWS Peering ID", status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
		)
		r.EventRecorder.Eventf(
			aws, models.Warning, models.FetchFailed,
			"Resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	if status != nil {
		err = r.API.DeletePeering(aws.Status.ID, instaclustr.AWSPeeringEndpoint)
		if err != nil {
			l.Error(err, "cannot update AWS VPC Peering resource status",
				"AWS Peering ID", aws.Status.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"Region", aws.Spec.PeerRegion,
				"AWS VPC Peering metadata", aws.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				aws, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
		r.EventRecorder.Eventf(
			aws, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent",
		)
		return ctrl.Result{}, err
	}

	r.Scheduler.RemoveJob(aws.GetJobID(scheduler.SyncJob))

	patch := aws.NewPatch()
	controllerutil.RemoveFinalizer(aws, models.DeletionFinalizer)
	aws.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, aws, patch)
	if err != nil {
		l.Error(err, "cannot patch AWS VPC Peering resource metadata",
			"AWS Peering ID", aws.Status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"Region", aws.Spec.PeerRegion,
			"AWS VPC Peering metadata", aws.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			aws, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("AWS VPC Peering has been deleted",
		"AWS VPC Peering ID", aws.Status.ID,
		"VPC ID", aws.Spec.PeerVPCID,
		"Region", aws.Spec.PeerRegion,
		"AWS VPC Peering Data Centre ID", aws.Status.CDCID,
		"AWS VPC Peering Status", aws.Status.PeeringStatus,
	)

	return ctrl.Result{}, nil
}

func (r *AWSVPCPeeringReconciler) startAWSVPCPeeringStatusJob(awsPeering *v1beta1.AWSVPCPeering) error {
	job := r.newWatchStatusJob(awsPeering)

	err := r.Scheduler.ScheduleJob(awsPeering.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AWSVPCPeeringReconciler) newWatchStatusJob(awsPeering *v1beta1.AWSVPCPeering) scheduler.Job {
	l := log.Log.WithValues("component", "AWSVPCPeeringStatusJob")
	return func() error {
		ctx := context.Background()

		namespacedName := client.ObjectKeyFromObject(awsPeering)
		err := r.Get(ctx, namespacedName, awsPeering)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName,
			)

			r.Scheduler.RemoveJob(awsPeering.GetJobID(scheduler.SyncJob))

			return nil
		}

		instaAWSPeering, err := r.API.GetAWSVPCPeering(awsPeering.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(ctx, awsPeering)
			}

			l.Error(err, "cannot get AWS VPC Peering Status from Inst API",
				"AWS VPC Peering ID", awsPeering.Status.ID,
			)
			return err
		}

		instaPeeringStatus := v1beta1.PeeringStatus{
			ID:         instaAWSPeering.ID,
			StatusCode: instaAWSPeering.StatusCode,
		}

		if !arePeeringStatusesEqual(&instaPeeringStatus, &awsPeering.Status.PeeringStatus) {
			l.Info("AWS VPC Peering status of k8s is different from Instaclustr. Reconcile statuses..",
				"AWS VPC Peering Status from Inst API", instaPeeringStatus,
				"AWS VPC Peering Status", awsPeering.Status)

			patch := awsPeering.NewPatch()
			awsPeering.Status.PeeringStatus = instaPeeringStatus
			err := r.Status().Patch(ctx, awsPeering, patch)
			if err != nil {
				return err
			}
		}

		if awsPeering.Status.StatusCode == models.AWSVPCPeeringStatusCodeDeleted {
			l.Info("The AWSPeering was deleted on AWS, stopping job...")
			r.EventRecorder.Event(awsPeering, models.Warning, models.DeletedEvent,
				"The AWSPeering was deleted on AWS, stopping job...",
			)

			r.Scheduler.RemoveJob(awsPeering.GetJobID(scheduler.SyncJob))

			return nil
		}

		if awsPeering.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			awsPeering.Annotations[models.ExternalChangesAnnotation] != models.True &&
			!subnetsEqual(instaAWSPeering.PeerSubnets, awsPeering.Spec.PeerSubnets) {
			l.Info("The k8s resource specification doesn't match the specification of Instaclustr, please change it manually",
				"k8s peerSubnets", instaAWSPeering.PeerSubnets,
				"instaclutr peerSubnets", awsPeering.Spec.PeerSubnets,
			)

			patch := awsPeering.NewPatch()
			awsPeering.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(ctx, awsPeering, patch)
			if err != nil {
				l.Error(err, "Cannot patch the resource with external changes annotation")

				return err
			}

			r.EventRecorder.Event(awsPeering, models.Warning, models.ExternalChanges,
				"k8s spec doesn't match spec of Instaclutr, please change it manually",
			)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSVPCPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.AWSVPCPeering{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
				}
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.AWSVPCPeering)
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

func (r *AWSVPCPeeringReconciler) handleExternalDelete(ctx context.Context, key *v1beta1.AWSVPCPeering) error {
	l := log.FromContext(ctx)

	patch := key.NewPatch()
	key.Status.StatusCode = models.DeletedStatus
	err := r.Status().Patch(ctx, key, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(key, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(key.GetJobID(scheduler.SyncJob))

	return nil
}
