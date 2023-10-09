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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
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
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch an AWS endpoint service principal resource")

		return models.ReconcileRequeue, err
	}

	// Handle resource deletion
	if principal.DeletionTimestamp != nil {
		err = r.handleDelete(ctx, l, principal)
		if err != nil {
			return models.ReconcileRequeue, err
		}

		return models.ExitReconcile, nil
	}

	// Handle resource creation
	if principal.Status.ID == "" {
		err = r.handleCreate(ctx, l, principal)
		if err != nil {
			return models.ReconcileRequeue, nil
		}

		return models.ExitReconcile, nil
	}

	return models.ExitReconcile, nil
}

func (r *AWSEndpointServicePrincipalReconciler) handleCreate(ctx context.Context, l logr.Logger, principal *clusterresourcesv1beta1.AWSEndpointServicePrincipal) error {
	b, err := r.API.CreateAWSEndpointServicePrincipal(principal.Spec)
	if err != nil {
		l.Error(err, "failed to create an AWS endpoint service principal resource on Instaclustr")
		r.EventRecorder.Eventf(principal, models.Warning, models.CreationFailed,
			"Failed to create an AWS endpoint service principal on Instaclustr. Reason: %v", err,
		)

		return err
	}

	patch := principal.NewPatch()
	err = json.Unmarshal(b, &principal.Status)
	if err != nil {
		l.Error(err, "failed to parse an AWS endpoint service principal resource response from Instaclustr")
		r.EventRecorder.Eventf(principal, models.Warning, models.ConversionFailed,
			"Failed to parse an AWS endpoint service principal resource response from Instaclustr. Reason: %v", err,
		)

		return err
	}

	err = r.Status().Patch(ctx, principal, patch)
	if err != nil {
		l.Error(err, "failed to patch an AWS endpoint service principal resource status with its ID")
		r.EventRecorder.Eventf(principal, models.Warning, models.PatchFailed,
			"Failed to patch an AWS endpoint service principal resource with its ID. Reason: %v", err,
		)

		return err
	}

	controllerutil.AddFinalizer(principal, models.DeletionFinalizer)
	err = r.Patch(ctx, principal, patch)
	if err != nil {
		l.Error(err, "failed to patch an AWS endpoint service principal resource with finalizer")
		r.EventRecorder.Eventf(principal, models.Warning, models.PatchFailed,
			"Failed to patch an AWS endpoint service principal resource with finalizer. Reason: %v", err,
		)

		return err
	}

	l.Info("AWS endpoint service principal resource has been created")
	r.EventRecorder.Event(principal, models.Normal, models.Created,
		"AWS endpoint service principal resource has been created",
	)

	return nil
}

func (r *AWSEndpointServicePrincipalReconciler) handleDelete(ctx context.Context, logger logr.Logger, resource *clusterresourcesv1beta1.AWSEndpointServicePrincipal) error {
	err := r.API.DeleteAWSEndpointServicePrincipal(resource.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "failed to delete an AWS endpoint service principal resource on Instaclustr")
		r.EventRecorder.Eventf(resource, models.Warning, models.DeletionFailed,
			"Failed to delete an AWS endpoint service principal on Instaclustr. Reason: %v", err,
		)

		return err
	}

	patch := resource.NewPatch()
	controllerutil.RemoveFinalizer(resource, models.DeletionFinalizer)
	err = r.Patch(ctx, resource, patch)
	if err != nil {
		logger.Error(err, "failed to delete finalizer an AWS endpoint service principal resource")
		r.EventRecorder.Eventf(resource, models.Warning, models.PatchFailed,
			"Failed to delete finalizer from an AWS endpoint service principal resource. Reason: %v", err,
		)

		return err
	}

	logger.Info("AWS endpoint service principal resource has been deleted")

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSEndpointServicePrincipalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1beta1.AWSEndpointServicePrincipal{}).
		Complete(r)
}
