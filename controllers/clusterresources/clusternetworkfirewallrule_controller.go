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

const (
	statusDELETED = "DELETED"
)

// ClusterNetworkFirewallRuleReconciler reconciles a ClusterNetworkFirewallRule object
type ClusterNetworkFirewallRuleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterNetworkFirewallRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	firewallRule := &v1beta1.ClusterNetworkFirewallRule{}
	err := r.Client.Get(ctx, req.NamespacedName, firewallRule)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Cluster network firewall rule resource is not found",
				"resource name", req.NamespacedName,
			)
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch cluster network firewall rule")
		return ctrl.Result{}, err
	}

	switch firewallRule.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateFirewallRule(ctx, firewallRule, &l)
	case models.UpdatingEvent:
		return r.HandleUpdateFirewallRule(ctx, firewallRule, &l)
	case models.DeletingEvent:
		return r.HandleDeleteFirewallRule(ctx, firewallRule, &l)
	case models.GenericEvent:
		l.Info("Cluster network firewall rule event isn't handled",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
			"request", req,
			"event", firewallRule.Annotations[models.ResourceStateAnnotation])
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleCreateFirewallRule(
	ctx context.Context,
	firewallRule *v1beta1.ClusterNetworkFirewallRule,
	l *logr.Logger,
) (ctrl.Result, error) {
	if firewallRule.Status.ID == "" {
		l.Info(
			"Creating cluster network firewall rule",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)

		patch := firewallRule.NewPatch()

		firewallRuleStatus, err := r.API.CreateFirewallRule(instaclustr.ClusterNetworkFirewallRuleEndpoint, &firewallRule.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create cluster network firewall rule",
				"spec", firewallRule.Spec,
			)
			r.EventRecorder.Eventf(
				firewallRule, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			firewallRule, models.Normal, models.Created,
			"Resource creation request is sent",
		)

		firewallRule.Status.FirewallRuleStatus = *firewallRuleStatus

		err = r.Status().Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster network firewall rule status ", "ID", firewallRule.Status.ID)
			r.EventRecorder.Eventf(
				firewallRule, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		firewallRule.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(firewallRule, models.DeletionFinalizer)

		err = r.Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster network firewall rule",
				"cluster ID", firewallRule.Spec.ClusterID,
				"type", firewallRule.Spec.Type,
			)
			r.EventRecorder.Eventf(
				firewallRule, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		l.Info(
			"Cluster network firewall rule resource has been created",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)
	}

	err := r.startFirewallRuleStatusJob(firewallRule)
	if err != nil {
		l.Error(err, "Cannot start cluster network firewall rule status checker job",
			"firewall rule ID", firewallRule.Status.ID)
		r.EventRecorder.Eventf(
			firewallRule, models.Warning, models.CreationFailed,
			"Resource status job creation is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	r.EventRecorder.Eventf(
		firewallRule, models.Normal, models.Created,
		"Resource status check job is started",
	)

	return ctrl.Result{}, nil
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleUpdateFirewallRule(
	ctx context.Context,
	firewallRule *v1beta1.ClusterNetworkFirewallRule,
	l *logr.Logger,
) (ctrl.Result, error) {
	l.Info("Cluster network firewall rule update is not implemented",
		"firewall rule ID", firewallRule.Spec.ClusterID,
		"type", firewallRule.Spec.Type,
	)

	return ctrl.Result{}, nil
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleDeleteFirewallRule(
	ctx context.Context,
	firewallRule *v1beta1.ClusterNetworkFirewallRule,
	l *logr.Logger,
) (ctrl.Result, error) {
	patch := firewallRule.NewPatch()
	err := r.Patch(ctx, firewallRule, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster network firewall rule metadata",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)
		r.EventRecorder.Eventf(
			firewallRule, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	status, err := r.API.GetFirewallRuleStatus(firewallRule.Status.ID, instaclustr.ClusterNetworkFirewallRuleEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get cluster network firewall rule status from the Instaclustr API",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)
		r.EventRecorder.Eventf(
			firewallRule, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	if status != nil && status.Status != statusDELETED {
		err = r.API.DeleteFirewallRule(firewallRule.Status.ID, instaclustr.ClusterNetworkFirewallRuleEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete cluster network firewall rule",
				"rule ID", firewallRule.Status.ID,
				"cluster ID", firewallRule.Spec.ClusterID,
				"type", firewallRule.Spec.Type,
			)
			r.EventRecorder.Eventf(
				firewallRule, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			firewallRule, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)
	}

	r.Scheduler.RemoveJob(firewallRule.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(firewallRule, models.DeletionFinalizer)
	firewallRule.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, firewallRule, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster network firewall rule metadata",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
			"status", firewallRule.Status,
		)
		r.EventRecorder.Eventf(
			firewallRule, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	l.Info("Cluster network firewall rule has been deleted",
		"cluster ID", firewallRule.Spec.ClusterID,
		"type", firewallRule.Spec.Type,
		"status", firewallRule.Status,
	)

	r.EventRecorder.Eventf(
		firewallRule, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return ctrl.Result{}, nil
}

func (r *ClusterNetworkFirewallRuleReconciler) startFirewallRuleStatusJob(firewallRule *v1beta1.ClusterNetworkFirewallRule) error {
	job := r.newWatchStatusJob(firewallRule)

	err := r.Scheduler.ScheduleJob(firewallRule.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterNetworkFirewallRuleReconciler) newWatchStatusJob(firewallRule *v1beta1.ClusterNetworkFirewallRule) scheduler.Job {
	l := log.Log.WithValues("component", "FirewallRuleStatusJob")
	return func() error {
		instaFirewallRuleStatus, err := r.API.GetFirewallRuleStatus(firewallRule.Status.ID, instaclustr.ClusterNetworkFirewallRuleEndpoint)
		if err != nil {
			l.Error(err, "Cannot get cluster network firewall rule status from Inst API", "firewall rule ID", firewallRule.Status.ID)
			return err
		}

		if !areFirewallRuleStatusesEqual(instaFirewallRuleStatus, &firewallRule.Status.FirewallRuleStatus) {
			l.Info("Cluster network firewall rule status of k8s is different from Instaclustr. Reconcile statuses..",
				"firewall rule Status from Inst API", instaFirewallRuleStatus,
				"firewall rule status", firewallRule.Status)
			firewallRule.Status.FirewallRuleStatus = *instaFirewallRuleStatus
			err := r.Status().Update(context.Background(), firewallRule)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterNetworkFirewallRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.ClusterNetworkFirewallRule{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.ClusterNetworkFirewallRule)
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
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).
		Complete(r)
}
