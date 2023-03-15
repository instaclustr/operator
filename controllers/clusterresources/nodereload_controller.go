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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// NodeReloadReconciler reconciles a NodeReload object
type NodeReloadReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

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

	nrs := &clusterresourcesv1alpha1.NodeReload{}
	err := r.Client.Get(ctx, req.NamespacedName, nrs)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Node Reload resource is not found", "request", req)
			return models.ExitReconcile, nil
		}
		l.Error(err, "Unable to fetch Node Reload", "request", req)
		return models.ReconcileRequeue, err
	}

	if len(nrs.Spec.Nodes) == 0 {
		err = r.Client.Delete(ctx, nrs)
		if err != nil {
			l.Error(err,
				"Cannot delete Node Reload resource from K8s cluster",
				"Node Reload spec", nrs.Spec,
			)
			r.EventRecorder.Eventf(
				nrs, models.Warning, models.DeletionFailed,
				"Resource deletion is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}
		r.EventRecorder.Eventf(
			nrs, models.Normal, models.DeletionStarted,
			"Resource is deleted.",
		)
		l.Info(
			"Nodes were reloaded, resource was deleted",
			"Node Reload spec", nrs.Spec,
		)
		return models.ExitReconcile, nil
	}
	patch := nrs.NewPatch()

	if nrs.Status.NodeInProgress.ID == "" {
		nodeInProgress := &clusterresourcesv1alpha1.Node{
			ID: nrs.Spec.Nodes[len(nrs.Spec.Nodes)-1].ID,
		}
		nrs.Status.NodeInProgress.ID = nodeInProgress.ID

		err = r.API.CreateNodeReload(nodeInProgress)
		if err != nil {
			l.Error(err,
				"Cannot start Node Reload process",
				"nodeID", nodeInProgress.ID,
			)
			r.EventRecorder.Eventf(
				nrs, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}

		r.EventRecorder.Eventf(
			nrs, models.Normal, models.Created,
			"Resource creation request is sent. Node ID: %s",
			nrs.Status.NodeInProgress.ID,
		)

		err = r.Status().Patch(ctx, nrs, patch)
		if err != nil {
			l.Error(err,
				"Cannot patch Node Reload status",
				"nodeID", nrs.Status.NodeInProgress,
			)
			r.EventRecorder.Eventf(
				nrs, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue, nil
		}
	}

	nodeReloadStatus, err := r.API.GetNodeReloadStatus(nrs.Status.NodeInProgress.ID)
	if err != nil {
		l.Error(err,
			"Cannot get Node Reload status",
			"nodeID", nrs.Status.NodeInProgress,
		)
		r.EventRecorder.Eventf(
			nrs, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	nrs.Status = *nrs.Status.FromInstAPI(nodeReloadStatus)
	err = r.Status().Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err,
			"Cannot patch Node Reload status",
			"nodeID", nrs.Status.NodeInProgress,
		)
		r.EventRecorder.Eventf(
			nrs, models.Warning, models.PatchFailed,
			"Resource status patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	if nrs.Status.CurrentOperationStatus.Status != "COMPLETED" {
		l.Info("Node Reload operation is not completed yet, please wait a few minutes",
			"nodeID", nrs.Status.NodeInProgress,
			"status", nrs.Status,
		)
		return models.ReconcileRequeue, nil
	}

	nrs.Status.NodeInProgress.ID = ""
	err = r.Status().Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err,
			"Cannot patch Node Reload status",
			"nodeID", nrs.Status.NodeInProgress,
		)
		r.EventRecorder.Eventf(
			nrs, models.Warning, models.PatchFailed,
			"Resource status patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	nrs.Spec.Nodes = nrs.Spec.Nodes[:len(nrs.Spec.Nodes)-1]
	err = r.Patch(ctx, nrs, patch)
	if err != nil {
		l.Error(err, "Cannot patch Node Reload cluster",
			"spec", nrs.Spec,
		)
		r.EventRecorder.Eventf(
			nrs, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue, nil
	}

	return models.ExitReconcile, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1alpha1.NodeReload{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return event.ObjectNew.GetGeneration() != event.ObjectOld.GetGeneration()
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Complete(r)
}
