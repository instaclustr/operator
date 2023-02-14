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

// AWSVPCPeeringReconciler reconciles a AWSVPCPeering object
type AWSVPCPeeringReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsvpcpeerings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AWSVPCPeering object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AWSVPCPeeringReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var aws clusterresourcesv1alpha1.AWSVPCPeering
	err := r.Client.Get(ctx, req.NamespacedName, &aws)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "AWSVPCPeering resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "Unable to fetch AWSVPCPeering", "request", req)
		return reconcile.Result{}, err
	}

	switch aws.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &aws, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &aws, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &aws, l), nil
	default:
		l.Info("Event isn't handled",
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
			"request", req,
			"event", aws.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *AWSVPCPeeringReconciler) handleCreateCluster(
	ctx context.Context,
	aws *clusterresourcesv1alpha1.AWSVPCPeering,
	l logr.Logger,
) reconcile.Result {

	if aws.Status.ID == "" {
		l.Info(
			"Creating AWSVPCPeering resource",
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
		)

		awsStatus, err := r.API.CreatePeering(instaclustr.AWSPeeringEndpoint, &aws.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create AWSVPCPeering resource",
				"AWSVPCPeering resource spec", aws.Spec,
			)
			return models.ReconcileRequeue
		}

		patch := aws.NewPatch()
		aws.Status.PeeringStatus = *awsStatus
		err = r.Status().Patch(ctx, aws, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWSVPCPeering resource status",
				"AWS Peering ID", awsStatus.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"region", aws.Spec.PeerRegion,
				"AWSVPCPeering metadata", aws.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(aws, models.DeletionFinalizer)
		aws.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, aws, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWSVPCPeering resource metadata",
				"AWS Peering ID", awsStatus.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"region", aws.Spec.PeerRegion,
				"AWSVPCPeering metadata", aws.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"AWSVPCPeering resource was created",
			"AWS Peering ID", awsStatus.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
		)
	}
	err := r.startAWSVPCPeeringStatusJob(aws)
	if err != nil {
		l.Error(err, "Cannot start AWSVPCPeering checker status job",
			"AWSVPCPeering ID", aws.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *AWSVPCPeeringReconciler) handleUpdateCluster(
	ctx context.Context,
	aws *clusterresourcesv1alpha1.AWSVPCPeering,
	l logr.Logger,
) reconcile.Result {
	err := r.API.UpdatePeering(aws.Status.ID, instaclustr.AWSPeeringEndpoint, &aws.Spec)
	if err != nil {
		l.Error(err, "Cannot update AWSVPCPeering",
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
			"subnets", aws.Spec.PeerSubnets,
		)
	}

	patch := aws.NewPatch()
	aws.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, aws, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWSVPCPeering resource metadata",
			"AWS Peering ID", aws.Status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
			"AWSVPCPeering metadata", aws.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("AWSVPCPeering resource has been updated",
		"AWSVPCPeering ID", aws.Status.ID,
		"AWS Account ID", aws.Spec.PeerAWSAccountID,
		"VPC ID", aws.Spec.PeerVPCID,
		"region", aws.Spec.PeerRegion,
		"AWSVPCPeering Data Centre ID", aws.Spec.DataCentreID,
		"AWSVPCPeering Status", aws.Status.PeeringStatus,
	)

	return reconcile.Result{}
}

func (r *AWSVPCPeeringReconciler) handleDeleteCluster(
	ctx context.Context,
	aws *clusterresourcesv1alpha1.AWSVPCPeering,
	l logr.Logger,
) reconcile.Result {

	patch := aws.NewPatch()
	err := r.Patch(ctx, aws, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWSVPCPeering resource metadata",
			"AWS Peering ID", aws.Status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
			"AWSVPCPeering metadata", aws.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetPeeringStatus(aws.Status.ID, instaclustr.AWSPeeringEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get AWSVPCPeering status from the Instaclustr API",
			"AWS Peering ID", status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		r.Scheduler.RemoveJob(aws.GetJobID(scheduler.StatusChecker))
		err = r.API.DeletePeering(aws.Status.ID, instaclustr.AWSPeeringEndpoint)
		if err != nil {
			l.Error(err, "Cannot update AWSVPCPeering resource statuss",
				"AWS Peering ID", aws.Status.ID,
				"AWS Account ID", aws.Spec.PeerAWSAccountID,
				"VPC ID", aws.Spec.PeerVPCID,
				"region", aws.Spec.PeerRegion,
				"AWSVPCPeering metadata", aws.ObjectMeta,
			)
			return models.ReconcileRequeue

		}
		return models.ReconcileRequeue
	}

	controllerutil.RemoveFinalizer(aws, models.DeletionFinalizer)
	aws.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, aws, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWSVPCPeering resource metadata",
			"AWS Peering ID", aws.Status.ID,
			"AWS Account ID", aws.Spec.PeerAWSAccountID,
			"VPC ID", aws.Spec.PeerVPCID,
			"region", aws.Spec.PeerRegion,
			"AWSVPCPeering metadata", aws.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("AWSVPCPeering has been deleted",
		"AWSVPCPeering ID", aws.Status.ID,
		"VPC ID", aws.Spec.PeerVPCID,
		"region", aws.Spec.PeerRegion,
		"AWSVPCPeering Data Centre ID", aws.Spec.DataCentreID,
		"AWSVPCPeering Status", aws.Status.PeeringStatus,
	)

	return reconcile.Result{}
}

func (r *AWSVPCPeeringReconciler) startAWSVPCPeeringStatusJob(awsPeering *clusterresourcesv1alpha1.AWSVPCPeering) error {
	job := r.newWatchStatusJob(awsPeering)

	err := r.Scheduler.ScheduleJob(awsPeering.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AWSVPCPeeringReconciler) newWatchStatusJob(awsPeering *clusterresourcesv1alpha1.AWSVPCPeering) scheduler.Job {
	l := log.Log.WithValues("component", "AWSVPCPeeringStatusJob")
	return func() error {
		instaPeeringStatus, err := r.API.GetPeeringStatus(awsPeering.Status.ID, instaclustr.AWSPeeringEndpoint)
		if err != nil {
			l.Error(err, "Cannot get AWSVPCPeering Status from Inst API", "AWSVPCPeering ID", awsPeering.Status.ID)
			return err
		}

		if !isPeeringStatusesEqual(instaPeeringStatus, &awsPeering.Status.PeeringStatus) {
			l.Info("AWSVPCPeering status of k8s is different from Instaclustr. Reconcile statuses..",
				"AWSVPCPeering Status from Inst API", instaPeeringStatus,
				"AWSVPCPeering Status", awsPeering.Status)

			patch := awsPeering.NewPatch()
			awsPeering.Status.PeeringStatus = *instaPeeringStatus
			err := r.Status().Patch(context.Background(), awsPeering, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSVPCPeeringReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.AWSVPCPeering{}, builder.WithPredicates(predicate.Funcs{
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
