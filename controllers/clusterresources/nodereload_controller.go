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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

// NodeReloadReconciler reconciles a NodeReload object
type NodeReloadReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

const (
	nodeReloadOperationStatusCompleted = "COMPLETED"
)

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NodeReloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	nrs := &v1beta1.NodeReload{}
	err := r.Client.Get(ctx, req.NamespacedName, nrs)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Node Reload resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "Unable to fetch Node Reload", "request", req)
		return reconcile.Result{}, err
	}

	patch := nrs.NewPatch()

	if len(nrs.Status.PendingNodes)+len(nrs.Status.CompletedNodes)+len(nrs.Status.FailedNodes) == 0 {
		nrs.Status.PendingNodes = nrs.Spec.Nodes
		err := r.Status().Patch(ctx, nrs, patch)
		if err != nil {
			l.Error(err, "Failed to patch pending nodes to the resource")
			r.EventRecorder.Eventf(nrs, models.Warning, models.PatchFailed,
				"Failed to patch pending nodes to the resource. Reason: %w", err,
			)

			return reconcile.Result{}, err
		}
	}

	if len(nrs.Status.PendingNodes) == 0 && nrs.Status.NodeInProgress == nil {
		r.EventRecorder.Eventf(
			nrs, models.Normal, models.UpdatedEvent,
			"The controller has finished working",
		)
		l.Info(
			"The controller has finished working",
			"completed nodes", nrs.Status.CompletedNodes,
			"failed nodes", nrs.Status.FailedNodes,
		)

		return reconcile.Result{}, nil
	}

	if nrs.Status.NodeInProgress == nil {
		nodeInProgress := nrs.Status.PendingNodes[0]

		err := r.API.CreateNodeReload(nodeInProgress)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleNodeNotFound(ctx, nodeInProgress, nrs)
			}

			l.Error(err, "Failed to trigger node reload", "node ID", nodeInProgress.ID)
			r.EventRecorder.Eventf(nrs, models.Warning, models.CreationFailed,
				"Failed to trigger node reload. Reason: %w", err,
			)

			return reconcile.Result{}, err
		}

		nrs.Status.NodeInProgress = nodeInProgress
		if len(nrs.Status.PendingNodes) > 0 {
			nrs.Status.PendingNodes = nrs.Status.PendingNodes[1:]
		}

		err = r.Status().Patch(ctx, nrs, patch)
		if err != nil {
			l.Error(err, "Failed to patch node in progress",
				"node ID", nodeInProgress.ID,
			)
			r.EventRecorder.Eventf(nrs, models.Warning, models.PatchFailed,
				"Failed to patch node in progress. Reason: %w", err,
			)

			return reconcile.Result{}, err
		}
	}

	nodeReloadStatus, err := r.API.GetNodeReloadStatus(nrs.Status.NodeInProgress.ID)
	if err != nil {
		if errors.Is(err, instaclustr.NotFound) {
			return r.handleNodeNotFound(ctx, nrs.Status.NodeInProgress, nrs)
		}

		l.Error(err, "Failed to fetch node reload status from Instaclustr",
			"node ID", nrs.Status.NodeInProgress.ID,
		)
		r.EventRecorder.Eventf(nrs, models.Warning, models.FetchFailed,
			"Failed to fetch node reload status from Instaclustr. Reason: %w", err,
		)

		return reconcile.Result{}, err
	}

	nrs.Status.CurrentOperationStatus = &v1beta1.Operation{
		OperationID:  nodeReloadStatus.OperationID,
		TimeCreated:  nodeReloadStatus.TimeCreated,
		TimeModified: nodeReloadStatus.TimeModified,
		Status:       nodeReloadStatus.Status,
		Message:      nodeReloadStatus.Message,
	}

	err = r.Status().Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err, "Failed to patch current operation status",
			"node ID", nrs.Status.NodeInProgress.ID,
			"currentOperationStatus", nodeReloadStatus,
		)
		r.EventRecorder.Eventf(nrs, models.Warning, models.FetchFailed,
			"Failed to patch current operation status. Reason: %w", err,
		)

		return reconcile.Result{}, err
	}

	if nodeReloadStatus.Status != nodeReloadOperationStatusCompleted {
		l.Info("Node Reload operation is not completed yet, please wait a few minutes",
			"status", nrs.Status,
		)

		return models.ReconcileRequeue, nil
	}

	l.Info("The node has been successfully reloaded",
		"status", nrs.Status,
	)
	r.EventRecorder.Eventf(nrs, models.Normal, models.UpdatedEvent,
		"Node %s has been successfully reloaded", nrs.Status,
	)

	patch = nrs.NewPatch()

	nrs.Status.CompletedNodes = append(nrs.Status.CompletedNodes, nrs.Status.NodeInProgress)
	nrs.Status.CurrentOperationStatus, nrs.Status.NodeInProgress = nil, nil

	err = r.Status().Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err, "Failed to patch completed nodes")
		r.EventRecorder.Eventf(nrs, models.Warning, models.PatchFailed,
			"Failed to patch completed nodes. Reason: %w", err,
		)

		return reconcile.Result{}, err
	}

	return models.ImmediatelyRequeue, nil
}

func (r *NodeReloadReconciler) handleNodeNotFound(ctx context.Context, node *v1beta1.Node, nrs *v1beta1.NodeReload) (reconcile.Result, error) {
	l := log.FromContext(ctx)

	patch := nrs.NewPatch()

	if nrs.Status.NodeInProgress == nil && len(nrs.Status.PendingNodes) > 0 {
		nrs.Status.PendingNodes = nrs.Status.PendingNodes[1:]
	}

	nrs.Status.FailedNodes = append(nrs.Status.FailedNodes, node)
	nrs.Status.CurrentOperationStatus, nrs.Status.NodeInProgress = nil, nil

	err := r.Status().Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err, "Cannot patch failed node")
		r.EventRecorder.Event(nrs, models.Warning, models.PatchFailed,
			"Cannot patch failed node",
		)

		return reconcile.Result{}, err
	}

	l.Error(err, "Node is not found on the Instaclustr side",
		"nodeID", node.ID,
	)
	r.EventRecorder.Eventf(nrs, models.Warning, models.FetchFailed,
		"Node %s is not found on Instaclustr", node.ID,
	)

	return models.ImmediatelyRequeue, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay,
				ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&v1beta1.NodeReload{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return event.ObjectNew.GetGeneration() != event.ObjectOld.GetGeneration()
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Complete(r)
}
