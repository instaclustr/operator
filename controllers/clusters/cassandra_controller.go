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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	apiv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    *instaclustr.Client
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

	var cassandraCluster clustersv1alpha1.Cassandra
	err := r.Client.Get(ctx, req.NamespacedName, &cassandraCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Error(
				err, "unable to fetch Cassandra Cluster",
				"Cassandra cluster cluster spec", cassandraCluster.Spec,
			)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	}

	cassandraFinalizerName := "cassandra-insta-finalizer/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if cassandraCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(&cassandraCluster, cassandraFinalizerName) {
			controllerutil.AddFinalizer(&cassandraCluster, cassandraFinalizerName)
			err := r.Update(ctx, &cassandraCluster)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&cassandraCluster, cassandraFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			err := r.API.DeleteCassandraCluster(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
			if err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			l.Info("Cassandra cluster has been deleted",
				"Cassandra cluster spec", cassandraCluster.Spec,
			)

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&cassandraCluster, cassandraFinalizerName)
			err = r.Update(ctx, &cassandraCluster)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if cassandraCluster.Status.ID == "" {
		l.Info(
			"Cassandra Cluster ID not found, creating Cassandra cluster",
			"Cluster name", cassandraCluster.Spec.Name,
			"Data centres", cassandraCluster.Spec.DataCentres,
		)

		cassandraSpec := apiv2.CassandraToInstAPI(&cassandraCluster.Spec)
		id, err := r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandraSpec)
		if err != nil {
			l.Error(
				err, "cannot create Cassandra cluster",
				"Cassandra cluster spec", cassandraCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		cassandraCluster.Status.ID = id
	}

	if cassandraCluster.Status.ID != "" && cassandraCluster.Status.Status == "RUNNING" {

		cassandraDCs, err := r.API.GetCassandraDCs(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(
				err, "cannot get Cassandra Data Centres from the Instaclustr API",
				"Cassandra cluster spec", cassandraCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		result := apiv2.CompareCassandraDCs(cassandraCluster.Spec.DataCentres, cassandraDCs)
		if result != nil {
			err = r.API.UpdateCassandraCluster(cassandraCluster.Status.ID,
				instaclustr.CassandraEndpoint,
				result,
			)
			if err != nil {
				l.Error(
					err, "cannot update Cassandra cluster",
					"Cassandra cluster spec", cassandraCluster.Spec,
				)
				return reconcile.Result{}, err
			}

			l.Info("Cassandra cluster has been updated",
				"Cluster name", cassandraCluster.Spec.Name,
				"Cluster ID", cassandraCluster.Status.ID,
				"Kind", cassandraCluster.Kind,
				"API Version", cassandraCluster.APIVersion,
				"Namespace", cassandraCluster.Namespace,
			)
		}
	}

	currentClusterStatus, err := r.API.GetClusterStatus(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil {
		l.Error(
			err, "cannot get Cassandra cluster status from the Instaclustr API",
			"Cassandra cluster spec", cassandraCluster.Spec,
		)
		return reconcile.Result{}, err
	}

	cassandraCluster.Status.ClusterStatus = *currentClusterStatus
	err = r.Status().Update(context.Background(), &cassandraCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	l.Info(
		"Cassandra cluster status has been updated",
		"Cluster name", cassandraCluster.Spec.Name,
		"Cluster ID", cassandraCluster.Status.ID,
		"Kind", cassandraCluster.Kind,
		"API Version", cassandraCluster.APIVersion,
		"Namespace", cassandraCluster.Namespace,
		"Cluster Status", cassandraCluster.Status.Status,
	)

	return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cassandra{}).
		Complete(r)
}
