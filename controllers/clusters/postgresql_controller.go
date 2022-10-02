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

package clusters

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.APIv1
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgreSQL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pgCluster *clustersv1alpha1.PostgreSQL
	err := r.Client.Get(ctx, req.NamespacedName, pgCluster)
	if err != nil {
		logger.Error(err, "unable to fetch PostgreSQL cluster")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if pgCluster.Status.ClusterID == "" {
		logger.Info(
			"PostgreSQL Cluster ID not found, creating PostgreSQL cluster",
			"Cluster name", pgCluster.Spec.ClusterName,
			"Data centres", pgCluster.Spec.DataCentres,
		)

		id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, pgCluster.Spec)
		if err != nil {
			logger.Error(
				err, "cannot create PostgreSQL cluster",
				"PostgreSQL manifest", pgCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		pgCluster.Status.ClusterID = id

		logger.Info(
			"PostgreSQL resource has been created",
			"Cluster name", pgCluster.Name,
			"Cluster ID", id,
			"Kind", pgCluster.Kind,
			"Api version", pgCluster.APIVersion,
		)
	}

	pgInstaCluster, err := r.API.GetPostgreSQLCluster(instaclustr.ClustersEndpoint, pgCluster.Status.ClusterID)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.ClusterName,
			"Cluster ID", pgCluster.Status.ClusterID,
		)
		return reconcile.Result{}, err
	}

	pgCluster.Status = *pgInstaCluster

	err = r.Status().Update(context.Background(), pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster CRD",
			"Cluster name", pgCluster.Spec.ClusterName,
			"Cluster ID", pgCluster.Status.ClusterID,
		)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.PostgreSQL{}).
		Complete(r)
}
