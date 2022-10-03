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
	"github.com/instaclustr/operator/pkg/instaclustr/apiv2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    *apiv2.ClientV2
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cassandra object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var cassandraCluster clustersv2alpha1.Cassandra
	err := r.Client.Get(ctx, req.NamespacedName, &cassandraCluster)
	if err != nil {
		l.Error(
			err, "unable to fetch Cassandra Cluster",
			"Cassandra cluster cluster spec", cassandraCluster.Spec,
		)
		return reconcile.Result{}, err
	}

	if cassandraCluster.Status.ClusterID == "" {
		l.Info(
			"Cassandra Cluster ID not found, creating Cassandra cluster",
			"Cluster name", cassandraCluster.Spec.Name,
			"Data centres", cassandraCluster.Spec.DataCentres,
		)

		id, err := r.API.Client.CreateCluster(instaclustr.CassandraEndpoint, cassandraCluster.Spec)
		if err != nil {
			l.Error(
				err, "cannot create Cassandra cluster",
				"Cassandra cluster spec", cassandraCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		currentClusterStatus, err := r.API.GetCassandraClusterStatus(id)
		if err != nil {
			l.Error(
				err, "cannot get Cassandra cluster status from the Instaclustr API",
				"Cassandra cluster spec", cassandraCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		cassandraCluster.Status = *currentClusterStatus
		err = r.Status().Update(context.Background(), &cassandraCluster)
		if err != nil {
			return reconcile.Result{}, err
		}
		l.Info(
			"Cassandra resource has been created",
			"Cluster name", cassandraCluster.Spec.Name,
			"cluster ID", id,
			"Kind", cassandraCluster.Kind,
			"API Version", cassandraCluster.APIVersion,
		)

	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv2alpha1.Cassandra{}).
		Complete(r)
}
