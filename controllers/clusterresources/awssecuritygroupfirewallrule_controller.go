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

// AWSSecurityGroupFirewallRuleReconciler reconciles a AWSSecurityGroupFirewallRule object
type AWSSecurityGroupFirewallRuleReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awssecuritygroupfirewallrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awssecuritygroupfirewallrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awssecuritygroupfirewallrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AWSSecurityGroupFirewallRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AWSSecurityGroupFirewallRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	firewallRule := &clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule{}
	err := r.Client.Get(ctx, req.NamespacedName, firewallRule)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("AWSSecurityGroupFirewallRule resource is not found",
				"resource name", req.NamespacedName,
			)
			return reconcile.Result{}, nil
		}

		l.Error(err, "Unable to fetch AWSSecurityGroupFirewallRule")
		return reconcile.Result{}, err
	}

	switch firewallRule.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		reconcileResult := r.HandleCreateFirewallRule(ctx, firewallRule, &l)
		return reconcileResult, nil
	case models.UpdatingEvent:
		reconcileResult := r.HandleUpdateFirewallRule(ctx, firewallRule, &l)
		return reconcileResult, nil
	case models.DeletingEvent:
		reconcileResult := r.HandleDeleteFirewallRule(ctx, firewallRule, &l)
		return reconcileResult, nil
	case models.GenericEvent:
		l.Info("AWSSecurityGroupFirewallRule event isn't handled",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
			"request", req,
			"event", firewallRule.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *AWSSecurityGroupFirewallRuleReconciler) HandleCreateFirewallRule(
	ctx context.Context,
	firewallRule *clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule,
	l *logr.Logger,
) reconcile.Result {
	if firewallRule.Status.ID == "" {
		l.Info(
			"Creating AWSSecurityGroupFirewallRule",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)

		patch := firewallRule.NewPatch()

		firewallRuleStatus, err := r.API.CreateFirewallRule(instaclustr.AWSSecurityGroupFirewallRuleEndpoint, &firewallRule.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create AWSSecurityGroupFirewallRule",
				"spec", firewallRule.Spec,
			)
			return models.ReconcileRequeue
		}

		firewallRule.Status.FirewallRuleStatus = *firewallRuleStatus

		err = r.Status().Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWSSecurityGroupFirewallRule status ", "ID", firewallRule.Status.ID)
			return models.ReconcileRequeue
		}

		firewallRule.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(firewallRule, models.DeletionFinalizer)

		err = r.Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWSSecurityGroupFirewallRule",
				"cluster ID", firewallRule.Spec.ClusterID,
				"type", firewallRule.Spec.Type,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"AWSSecurityGroupFirewallRule resource has been created",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)
	}

	err := r.startFirewallRuleStatusJob(firewallRule)
	if err != nil {
		l.Error(err, "Cannot start AWSSecurityGroupFirewallRule status checker job",
			"firewall rule ID", firewallRule.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *AWSSecurityGroupFirewallRuleReconciler) HandleUpdateFirewallRule(
	ctx context.Context,
	firewallRule *clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule,
	l *logr.Logger,
) reconcile.Result {
	l.Info("AWSSecurityGroupFirewallRule update is not implemented",
		"firewall rule ID", firewallRule.Spec.ClusterID,
		"type", firewallRule.Spec.Type,
	)

	return reconcile.Result{}
}

func (r *AWSSecurityGroupFirewallRuleReconciler) HandleDeleteFirewallRule(
	ctx context.Context,
	firewallRule *clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule,
	l *logr.Logger,
) reconcile.Result {
	patch := firewallRule.NewPatch()
	err := r.Patch(ctx, firewallRule, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWSSecurityGroupFirewallRule metadata",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)

		return models.ReconcileRequeue
	}

	status, err := r.API.GetFirewallRuleStatus(firewallRule.Status.ID, instaclustr.AWSSecurityGroupFirewallRuleEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get AWSSecurityGroupFirewallRule status from the Instaclustr API",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)

		return models.ReconcileRequeue
	}

	if status != nil && status.Status != models.StatusDELETED {
		err = r.API.DeleteFirewallRule(firewallRule.Status.ID, instaclustr.AWSSecurityGroupFirewallRuleEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete AWSSecurityGroupFirewallRule",
				"rule ID", firewallRule.Status.ID,
				"cluster ID", firewallRule.Spec.ClusterID,
				"type", firewallRule.Spec.Type,
			)

			return models.ReconcileRequeue
		}
	}

	r.Scheduler.RemoveJob(firewallRule.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(firewallRule, models.DeletionFinalizer)
	firewallRule.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent

	err = r.Patch(ctx, firewallRule, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWSSecurityGroupFirewallRule metadata",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
			"status", firewallRule.Status,
		)

		return models.ReconcileRequeue
	}

	l.Info("AWSSecurityGroupFirewallRule has been deleted",
		"cluster ID", firewallRule.Spec.ClusterID,
		"type", firewallRule.Spec.Type,
		"status", firewallRule.Status,
	)

	return models.ReconcileResult
}

func (r *AWSSecurityGroupFirewallRuleReconciler) startFirewallRuleStatusJob(firewallRule *clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule) error {
	job := r.newWatchStatusJob(firewallRule)

	err := r.Scheduler.ScheduleJob(firewallRule.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AWSSecurityGroupFirewallRuleReconciler) newWatchStatusJob(firewallRule *clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule) scheduler.Job {
	l := log.Log.WithValues("component", "FirewallRuleStatusJob")
	return func() error {
		instaFirewallRuleStatus, err := r.API.GetFirewallRuleStatus(firewallRule.Status.ID, instaclustr.AWSSecurityGroupFirewallRuleEndpoint)
		if err != nil {
			l.Error(err, "Cannot get AWSSecurityGroupFirewallRule status from Inst API", "firewall rule ID", firewallRule.Status.ID)
			return err
		}

		if !isFirewallRuleStatusesEqual(instaFirewallRuleStatus, &firewallRule.Status.FirewallRuleStatus) {
			l.Info("AWSSecurityGroupFirewallRule status of k8s is different from Instaclustr. Reconcile statuses..",
				"firewall rule Status from Inst API", instaFirewallRuleStatus,
				"firewall rule status", firewallRule.Status)
			patch := firewallRule.NewPatch()
			firewallRule.Status.FirewallRuleStatus = *instaFirewallRuleStatus
			err := r.Status().Patch(context.Background(), firewallRule, patch)
			if err != nil {
				return err
			}

			if instaFirewallRuleStatus.Status == models.StatusDELETED {
				r.Scheduler.RemoveJob(firewallRule.GetJobID(scheduler.StatusChecker))
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSSecurityGroupFirewallRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.AWSSecurityGroupFirewallRule{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).
		Complete(r)
}
