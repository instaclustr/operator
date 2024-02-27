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
	"errors"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// AWSEndpointServicePrincipalReconciler reconciles a AWSEndpointServicePrincipal object
type AWSEndpointServicePrincipalReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsendpointserviceprincipals,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsendpointserviceprincipals/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsendpointserviceprincipals/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *AWSEndpointServicePrincipalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	principal := &clusterresourcesv1beta1.AWSEndpointServicePrincipal{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, principal)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch an AWS endpoint service principal resource")

		return ctrl.Result{}, err
	}

	if principal.DeletionTimestamp != nil {
		return r.handleDelete(ctx, l, principal)
	}

	if principal.Status.ID == "" {
		return r.handleCreate(ctx, l, principal)
	}

	return ctrl.Result{}, nil
}

func (r *AWSEndpointServicePrincipalReconciler) handleCreate(ctx context.Context, l logr.Logger, principal *clusterresourcesv1beta1.AWSEndpointServicePrincipal) (ctrl.Result, error) {
	var cdcID string
	var err error
	if principal.Spec.ClusterRef != nil {
		cdcID, err = GetDataCentreID(r.Client, ctx, principal.Spec.ClusterRef)
		if err != nil {
			l.Error(err, "Cannot get CDCID",
				"Cluster reference", principal.Spec.ClusterRef,
			)
			return ctrl.Result{}, err
		}
		l.Info(
			"Creating AWS Endpoint Service Principal resource from the cluster reference",
			"cluster reference", principal.Spec.ClusterRef,
		)
	} else {
		cdcID = principal.Spec.ClusterDataCenterID
		l.Info(
			"Creating AWS Endpoint Service Principal resource",
			"principal", principal.Spec,
		)
	}

	b, err := r.API.CreateAWSEndpointServicePrincipal(principal.Spec, cdcID)
	if err != nil {
		l.Error(err, "failed to create an AWS endpoint service principal resource on Instaclustr")
		r.EventRecorder.Eventf(principal, models.Warning, models.CreationFailed,
			"Failed to create an AWS endpoint service principal on Instaclustr. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	patch := principal.NewPatch()
	err = json.Unmarshal(b, &principal.Status)
	if err != nil {
		l.Error(err, "failed to parse an AWS endpoint service principal resource response from Instaclustr")
		r.EventRecorder.Eventf(principal, models.Warning, models.ConversionFailed,
			"Failed to parse an AWS endpoint service principal resource response from Instaclustr. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	principal.Status.CDCID = cdcID
	err = r.Status().Patch(ctx, principal, patch)
	if err != nil {
		l.Error(err, "failed to patch an AWS endpoint service principal resource status with its ID")
		r.EventRecorder.Eventf(principal, models.Warning, models.PatchFailed,
			"Failed to patch an AWS endpoint service principal resource with its ID. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	controllerutil.AddFinalizer(principal, models.DeletionFinalizer)
	err = r.Patch(ctx, principal, patch)
	if err != nil {
		l.Error(err, "failed to patch an AWS endpoint service principal resource with finalizer")
		r.EventRecorder.Eventf(principal, models.Warning, models.PatchFailed,
			"Failed to patch an AWS endpoint service principal resource with finalizer. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	l.Info("AWS endpoint service principal resource has been created")
	r.EventRecorder.Event(principal, models.Normal, models.Created,
		"AWS endpoint service principal resource has been created",
	)

	err = r.startWatchStatusJob(ctx, principal)
	if err != nil {
		l.Error(err, "failed to start status checker job")
		r.EventRecorder.Eventf(principal, models.Warning, models.CreationFailed,
			"Failed to start status checker job. Reason: %w", err,
		)

		return ctrl.Result{}, err
	}
	r.EventRecorder.Eventf(principal, models.Normal, models.Created,
		"Status check job %s has been started", principal.GetJobID(scheduler.SyncJob),
	)

	return ctrl.Result{}, nil
}

func (r *AWSEndpointServicePrincipalReconciler) handleDelete(ctx context.Context, logger logr.Logger, resource *clusterresourcesv1beta1.AWSEndpointServicePrincipal) (ctrl.Result, error) {
	err := r.API.DeleteAWSEndpointServicePrincipal(resource.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "failed to delete an AWS endpoint service principal resource on Instaclustr")
		r.EventRecorder.Eventf(resource, models.Warning, models.DeletionFailed,
			"Failed to delete an AWS endpoint service principal on Instaclustr. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	patch := resource.NewPatch()
	controllerutil.RemoveFinalizer(resource, models.DeletionFinalizer)
	err = r.Patch(ctx, resource, patch)
	if err != nil {
		logger.Error(err, "failed to delete finalizer an AWS endpoint service principal resource")
		r.EventRecorder.Eventf(resource, models.Warning, models.PatchFailed,
			"Failed to delete finalizer from an AWS endpoint service principal resource. Reason: %v", err,
		)

		return ctrl.Result{}, err
	}

	logger.Info("AWS endpoint service principal resource has been deleted")

	return ctrl.Result{}, nil
}

func (r *AWSEndpointServicePrincipalReconciler) startWatchStatusJob(ctx context.Context, resource *clusterresourcesv1beta1.AWSEndpointServicePrincipal) error {
	job := r.newWatchStatusJob(ctx, resource)
	return r.Scheduler.ScheduleJob(resource.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
}

func (r *AWSEndpointServicePrincipalReconciler) newWatchStatusJob(ctx context.Context, principal *clusterresourcesv1beta1.AWSEndpointServicePrincipal) scheduler.Job {
	l := log.FromContext(ctx, "components", "WatchStatusJob")

	return func() error {
		key := client.ObjectKeyFromObject(principal)
		err := r.Get(ctx, key, principal)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", key,
				)

				r.Scheduler.RemoveJob(principal.GetJobID(scheduler.SyncJob))

				return nil
			}

			return err
		}

		_, err = r.API.GetAWSEndpointServicePrincipal(principal.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(ctx, principal)
			}

			return err
		}

		return nil
	}
}

func (r *AWSEndpointServicePrincipalReconciler) handleExternalDelete(ctx context.Context, principal *clusterresourcesv1beta1.AWSEndpointServicePrincipal) error {
	l := log.FromContext(ctx)

	patch := principal.NewPatch()
	principal.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, principal, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(principal, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(principal.GetJobID(scheduler.SyncJob))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSEndpointServicePrincipalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&clusterresourcesv1beta1.AWSEndpointServicePrincipal{}).
		Complete(r)
}
