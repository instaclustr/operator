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

// ClusterNetworkFirewallRuleReconciler reconciles a ClusterNetworkFirewallRule object
type ClusterNetworkFirewallRuleReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterNetworkFirewallRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterNetworkFirewallRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	firewallRule := &clusterresourcesv1alpha1.ClusterNetworkFirewallRule{}
	err := r.Client.Get(ctx, req.NamespacedName, firewallRule)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Cluster network firewall rule resource is not found",
				"resource name", req.NamespacedName,
			)
			return reconcile.Result{}, nil
		}

		l.Error(err, "Unable to fetch cluster network firewall rule")
		return reconcile.Result{}, err
	}

	switch firewallRule.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		reconcileResult := r.HandleCreateFirewallRule(firewallRule, &l, ctx, &req)
		return reconcileResult, nil
	case models.UpdatingEvent:
		reconcileResult := r.HandleUpdateFirewallRule(firewallRule, &l, ctx, &req)
		return reconcileResult, nil
	case models.DeletingEvent:
		reconcileResult := r.HandleDeleteFirewallRule(firewallRule, &l, ctx, &req)
		return reconcileResult, nil
	default:
		l.Info("UNKNOWN EVENT",
			"Cluster network firewall rule spec", firewallRule.Spec,
		)
		return reconcile.Result{}, err
	}
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleCreateFirewallRule(
	firewallRule *clusterresourcesv1alpha1.ClusterNetworkFirewallRule,
	l *logr.Logger,
	ctx context.Context,
	req *ctrl.Request,
) reconcile.Result {
	if firewallRule.Status.ID == "" {
		l.Info(
			"Creating cluster network firewall rule",
			"cluster ID", firewallRule.Spec.ClusterID,
			"type", firewallRule.Spec.Type,
		)

		patch := firewallRule.NewPatch()

		firewallRuleStatus, err := r.API.CreateFirewallRule(instaclustr.FirewallRuleEndpoint, &firewallRule.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create cluster network firewall rule",
				"spec", firewallRule.Spec,
			)
			return models.ReconcileRequeue
		}

		firewallRule.Status = *firewallRuleStatus

		err = r.Status().Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster network firewall rule status ", "ID", firewallRule.Status.ID)
			return models.ReconcileRequeue
		}

		firewallRule.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(firewallRule, models.DeletionFinalizer)

		err = r.Patch(ctx, firewallRule, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster network firewall rule",
				"cluster ID", firewallRule.Spec.ClusterID,
				"type", firewallRule.Spec.Type,
			)
			return models.ReconcileRequeue
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
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleUpdateFirewallRule(
	firewallRule *clusterresourcesv1alpha1.ClusterNetworkFirewallRule,
	l *logr.Logger,
	ctx context.Context,
	req *ctrl.Request,
) reconcile.Result {
	l.Info("Cluster network firewall rule update is not implemented",
		"firewall rule ID", firewallRule.Spec.ClusterID,
		"type", firewallRule.Spec.Type,
	)

	return reconcile.Result{}
}

func (r *ClusterNetworkFirewallRuleReconciler) HandleDeleteFirewallRule(
	firewallRule *clusterresourcesv1alpha1.ClusterNetworkFirewallRule,
	l *logr.Logger,
	ctx context.Context,
	req *ctrl.Request,
) reconcile.Result {
	controllerutil.RemoveFinalizer(firewallRule, models.DeletionFinalizer)
	err := r.Update(ctx, firewallRule)
	if err != nil {
		l.Error(err, "Cannot update cluster network firewall rule resource after finalizer removal",
			"cluster ID", firewallRule.Spec.ClusterID,
			"status", firewallRule.Status,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster network firewall rule deletion is not implemented",
		"cluster ID", firewallRule.Spec.ClusterID,
		"status", firewallRule.Status,
	)

	return reconcile.Result{}
}

func (r *ClusterNetworkFirewallRuleReconciler) startFirewallRuleStatusJob(firewallRule *clusterresourcesv1alpha1.ClusterNetworkFirewallRule) error {
	job := r.newWatchStatusJob(firewallRule)

	err := r.Scheduler.ScheduleJob(firewallRule.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterNetworkFirewallRuleReconciler) newWatchStatusJob(firewallRule *clusterresourcesv1alpha1.ClusterNetworkFirewallRule) scheduler.Job {
	l := log.Log.WithValues("component", "FirewallRuleStatusJob")
	return func() error {
		instaFirewallRuleStatus, err := r.API.GetFirewallRuleStatus(firewallRule.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get cluster network firewall rule status from Inst API", "firewall rule ID", firewallRule.Status.ID)
			return err
		}

		if !isFirewallRuleStatusesEqual(instaFirewallRuleStatus, &firewallRule.Status) {
			l.Info("Cluster network firewall rule status of k8s is different from Instaclustr. Reconcile statuses..",
				"firewall rule Status from Inst API", instaFirewallRuleStatus,
				"firewall rule status", firewallRule.Status)
			firewallRule.Status = *instaFirewallRuleStatus
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
		For(&clusterresourcesv1alpha1.ClusterNetworkFirewallRule{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					models.ResourceStateAnnotation: models.UpdatingEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.GenericEvent})
				return true
			},
		})).
		Complete(r)
}