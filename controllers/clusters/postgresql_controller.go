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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	apiv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1"
)

const (
	lastEventAnnotation = "instaclustr.con/lastEvent"
	createEvent         = "create"
	updateEvent         = "update"
	deleteEvent         = "delete"
	unknownEvent        = "unknown"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
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

	var pgCluster clustersv1alpha1.PostgreSQL
	err := r.Client.Get(ctx, req.NamespacedName, &pgCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch PostgreSQL cluster")
		return reconcile.Result{}, err
	}

	pgAnnotations := pgCluster.Annotations
	switch pgAnnotations[lastEventAnnotation] {
	case createEvent:
		err = r.HandleCreateCluster(&pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create PostgreSQL cluster")
			return reconcile.Result{}, err
		}

	case updateEvent:
		err = r.HandleUpdateCluster(&pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create PostgreSQL cluster")
			return reconcile.Result{}, err
		}

	case deleteEvent:
		err = r.HandleDeleteCluster(&pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create PostgreSQL cluster")
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	case unknownEvent:
		logger.Info("UNKNOWN EVENT")
		return reconcile.Result{}, err
	}

	// status checker logic
	pgInstaCluster, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != client.IgnoreNotFound(err) {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return reconcile.Result{}, err
	}

	err = r.Client.Get(ctx, req.NamespacedName, &pgCluster)
	if err != nil {
		logger.Error(err, "unable to fetch PostgreSQL cluster")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	pgCluster.Status.ClusterStatus = pgInstaCluster.ClusterStatus

	err = r.Status().Update(context.Background(), &pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster status by status checker",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *PostgreSQLReconciler) HandleCreateCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	logger.Info(
		"Creating PostgreSQL cluster",
		"Cluster name", pgCluster.Spec.Name,
		"Data centres", pgCluster.Spec.DataCentres,
	)

	pgSpec := apiv1.PgToInstAPI(&pgCluster.Spec)

	id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, pgSpec)
	if err != nil {
		logger.Error(
			err, "cannot create PostgreSQL cluster",
			"PostgreSQL manifest", pgCluster.Spec,
		)
		return err
	}

	err = r.Update(context.Background(), pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster CRD after creation",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return err
	}

	err = r.Client.Get(*ctx, req.NamespacedName, pgCluster)
	if err != nil {
		logger.Error(err, "unable to fetch PostgreSQL cluster")
		return client.IgnoreNotFound(err)
	}

	pgCluster.Status.ID = id
	err = r.Status().Update(context.Background(), pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster status after creation",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return err
	}

	logger.Info(
		"PostgreSQL resource has been created",
		"Cluster name", pgCluster.Name,
		"Cluster ID", pgCluster.Status.ID,
		"Kind", pgCluster.Kind,
		"Api version", pgCluster.APIVersion,
		"Namespace", pgCluster.Namespace,
	)

	return err
}

func (r *PostgreSQLReconciler) HandleUpdateCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {

	// Some logic

	err := r.Status().Update(context.Background(), pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster status",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) HandleDeleteCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {

	// Some logic

	controllerutil.RemoveFinalizer(pgCluster, "finalizer")
	err := r.Update(context.Background(), pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster CRD",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.PostgreSQL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{lastEventAnnotation: createEvent})
				event.Object.SetFinalizers([]string{"finalizer"})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{lastEventAnnotation: deleteEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					lastEventAnnotation: updateEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{lastEventAnnotation: unknownEvent})
				return true
			},
		})).
		Complete(r)
}
