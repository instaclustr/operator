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
	"fmt"

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
)

// OpenSearchEgressRulesReconciler reconciles a OpenSearchEgressRules object
type OpenSearchEgressRulesReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchegressrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchegressrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=opensearchegressrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenSearchEgressRules object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OpenSearchEgressRulesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rule := &clusterresourcesv1beta1.OpenSearchEgressRules{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, rule)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch OpenSearch Egress Rules resource")

		return ctrl.Result{}, err
	}

	// It`s handling resource deletion
	if rule.DeletionTimestamp != nil {
		err = r.handleDelete(ctx, l, rule)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// It`s handling resource creation
	if rule.Status.ID == "" {
		err = r.handleCreate(ctx, l, rule)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *OpenSearchEgressRulesReconciler) handleCreate(ctx context.Context, l logr.Logger, rule *clusterresourcesv1beta1.OpenSearchEgressRules) error {
	patch := rule.NewPatch()

	if rule.Status.ID == "" {
		rule.Status.ID = fmt.Sprintf("%s~%s~%s", rule.Spec.ClusterID, rule.Spec.Source, rule.Spec.OpenSearchBindingID)
	}

	_, err := r.API.GetOpenSearchEgressRule(rule.Status.ID)
	if !errors.Is(err, instaclustr.NotFound) && err != nil {
		l.Error(err, "failed to get OpenSearch Egress Rule resource from Instaclustr")
		r.EventRecorder.Eventf(rule, models.Warning, models.CreationFailed,
			"Failed to get OpenSearch Egress Rule from Instaclustr. Reason: %v", err,
		)

		return err
	}

	if errors.Is(err, instaclustr.NotFound) {
		rule.Status.ID, err = r.API.CreateOpenSearchEgressRules(rule)
		if err != nil {
			l.Error(err, "failed to create OpenSearch Egress Rule resource on Instaclustr")
			r.EventRecorder.Eventf(rule, models.Warning, models.CreationFailed,
				"Failed to create OpenSearch Egress Rule on Instaclustr. Reason: %v", err,
			)

			return err
		}

	}

	err = r.Status().Patch(ctx, rule, patch)
	if err != nil {
		l.Error(err, "failed to patch OpenSearch Egress Rule status with its id")
		r.EventRecorder.Eventf(rule, models.Warning, models.PatchFailed,
			"Failed to patch OpenSearch Egress Rule with its id. Reason: %v", err,
		)

		return err
	}

	controllerutil.AddFinalizer(rule, models.DeletionFinalizer)
	err = r.Patch(ctx, rule, patch)
	if err != nil {
		l.Error(err, "failed to patch OpenSearch Egress Rule with finalizer")
		r.EventRecorder.Eventf(rule, models.Warning, models.PatchFailed,
			"Failed to patch OpenSearch Egress Rule with finalizer. Reason: %v", err,
		)

		return err
	}

	l.Info("OpenSearch Egress Rule has been created")
	r.EventRecorder.Event(rule, models.Normal, models.Created,
		"OpenSearch Egress Rule has been created",
	)

	return nil
}

func (r *OpenSearchEgressRulesReconciler) handleDelete(ctx context.Context, logger logr.Logger, rule *clusterresourcesv1beta1.OpenSearchEgressRules) error {
	_, err := r.API.GetOpenSearchEgressRule(rule.Status.ID)
	if !errors.Is(err, instaclustr.NotFound) && err != nil {
		logger.Error(err, "failed to get OpenSearch Egress Rule resource from Instaclustr")
		r.EventRecorder.Eventf(rule, models.Warning, models.CreationFailed,
			"Failed to get OpenSearch Egress Rule from Instaclustr. Reason: %v", err,
		)

		return err
	}

	if !errors.Is(err, instaclustr.NotFound) {
		err = r.API.DeleteOpenSearchEgressRule(rule.Status.ID)
		if err != nil && !errors.Is(err, instaclustr.NotFound) {
			logger.Error(err, "failed to delete OpenSearch Egress Rule on Instaclustr")
			r.EventRecorder.Eventf(rule, models.Warning, models.DeletionFailed,
				"Failed to delete OpenSearch Egress Rule on Instaclustr. Reason: %v", err,
			)

			return err
		}
	}

	patch := rule.NewPatch()
	controllerutil.RemoveFinalizer(rule, models.DeletionFinalizer)
	err = r.Patch(ctx, rule, patch)
	if err != nil {
		logger.Error(err, "failed to delete finalizer OpenSearch Egress Rule")
		r.EventRecorder.Eventf(rule, models.Warning, models.PatchFailed,
			"Failed to delete finalizer from OpenSearch Egress Rule. Reason: %v", err,
		)

		return err
	}

	logger.Info("OpenSearch Egress Rule has been deleted")

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchEgressRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&clusterresourcesv1beta1.OpenSearchEgressRules{}).
		Complete(r)
}
