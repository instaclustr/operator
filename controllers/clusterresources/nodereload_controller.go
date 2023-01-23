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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	convertorsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

// NodeReloadReconciler reconciles a NodeReload object
type NodeReloadReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=nodereloads/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeReload object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NodeReloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var nrs clusterresourcesv1alpha1.NodeReload
	err := r.Client.Get(ctx, req.NamespacedName, &nrs)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Node Reload resource is not found", "request", req)
			return reconcile.Result{}, nil
		}
		l.Error(err, "Unable to fetch Node Reload", "request", req)
		return reconcile.Result{}, err
	}

	if len(nrs.Spec.Nodes) == 0 {
		err = r.Client.Delete(ctx, &nrs)
		if err != nil {
			l.Error(err,
				"Cannot delete Node Reload resource from K8s cluster",
				"Node Reload spec", nrs.Spec,
			)
			return models.ReconcileRequeue, nil
		}
		l.Info(
			"Nodes were reloaded, resource was deleted",
			"Node Reload spec", nrs.Spec,
		)
		return models.ReconcileResult, nil
	}
	patch := nrs.NewPatch()

	if nrs.Status.NodeInProgress.NodeID == "" && nrs.Status.NodeInProgress.Bundle == "" {
		nodeInProgress := &modelsv1.NodeReload{
			Bundle: nrs.Spec.Nodes[len(nrs.Spec.Nodes)-1].Bundle,
			NodeID: nrs.Spec.Nodes[len(nrs.Spec.Nodes)-1].NodeID,
		}
		nrs.Status.NodeInProgress = clusterresourcesv1alpha1.Node{
			NodeID: nodeInProgress.NodeID,
			Bundle: nodeInProgress.Bundle,
		}

		err = r.API.CreateNodeReload(nodeInProgress.Bundle, nodeInProgress.NodeID, nodeInProgress)
		if err != nil {
			l.Error(err,
				"Cannot start Node Reload process",
				"bundle", nodeInProgress.Bundle,
				"nodeID", nodeInProgress.NodeID,
			)
			return models.ReconcileRequeue, nil
		}

		err = r.Status().Patch(ctx, &nrs, patch)
		if err != nil {
			l.Error(err,
				"Cannot patch Node Reload status",
				"bundle", nrs.Status.NodeInProgress.Bundle,
				"nodeID", nrs.Status.NodeInProgress.NodeID,
			)
			return models.ReconcileRequeue, nil
		}
	}

	nodeReloadStatus, err := r.API.GetNodeReloadStatus(nrs.Status.NodeInProgress.Bundle, nrs.Status.NodeInProgress.NodeID)
	if err != nil {
		l.Error(err,
			"Cannot get Node Reload status",
			"bundle", nrs.Status.NodeInProgress.Bundle,
			"nodeID", nrs.Status.NodeInProgress.NodeID,
		)
		return models.ReconcileRequeue, nil
	}

	nrs.Status = convertorsv1.NodeReloadStatusFromInstAPI(nrs.Status.NodeInProgress, nodeReloadStatus)
	err = r.Status().Patch(ctx, &nrs, patch)
	if err != nil {
		l.Error(err,
			"Cannot patch Node Reload status",
			"bundle", nrs.Status.NodeInProgress.Bundle,
			"nodeID", nrs.Status.NodeInProgress.NodeID,
		)
		return models.ReconcileRequeue, nil
	}

	if nrs.Status.CurrentOperationStatus[0].Status != "COMPLETED" {
		l.Info("Node Reload operation is not completed yet, please wait a few minutes",
			"bundle", nrs.Status.NodeInProgress.Bundle,
			"nodeID", nrs.Status.NodeInProgress.NodeID,
			"status", nrs.Status.CurrentOperationStatus,
		)
		return models.ReconcileRequeue, nil
	}
	nrs.Status.NodeInProgress = clusterresourcesv1alpha1.Node{
		NodeID: "",
		Bundle: "",
	}
	err = r.Status().Patch(ctx, &nrs, patch)
	if err != nil {
		l.Error(err,
			"Cannot patch Node Reload status",
			"bundle", nrs.Status.NodeInProgress.Bundle,
			"nodeID", nrs.Status.NodeInProgress.NodeID,
		)
		return models.ReconcileRequeue, nil
	}

	nrs.Spec.Nodes = nrs.Spec.Nodes[:len(nrs.Spec.Nodes)-1]
	err = r.Patch(ctx, &nrs, patch)
	if err != nil {
		l.Error(err, "Cannot patch Node Reload cluster",
			"spec", nrs.Spec,
		)
		return models.ReconcileRequeue, nil
	}

	return models.ReconcileResult, nil
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
