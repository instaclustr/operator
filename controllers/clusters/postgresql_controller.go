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
	previousEventAnnotation = "instaclustr.con/previousEvent"
	currentEventAnnotation  = "instaclustr.con/currentEvent"
	createEvent             = "create"
	updateEvent             = "update"
	deleteEvent             = "delete"
	unknownEvent            = "unknown"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    *instaclustr.Client
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

	pgCluster := &clustersv1alpha1.PostgreSQL{}
	err := r.Client.Get(ctx, req.NamespacedName, pgCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch PostgreSQL cluster")
		return reconcile.Result{}, err
	}

	pgAnnotations := pgCluster.Annotations
	switch pgAnnotations[currentEventAnnotation] {
	case createEvent:
		err = r.HandleCreateCluster(pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot create PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	case updateEvent:
		reconcileResult, err := r.HandleUpdateCluster(pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot update PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return *reconcileResult, nil
	case deleteEvent:
		err = r.HandleDeleteCluster(pgCluster, &logger, &ctx, &req)
		if err != nil {
			logger.Error(err, "cannot delete PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	case unknownEvent:
		logger.Info("UNKNOWN EVENT")

		return reconcile.Result{}, err
	default:
		return reconcile.Result{}, err
	}
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

	pgCluster.Annotations[previousEventAnnotation] = pgCluster.Annotations[currentEventAnnotation]
	pgCluster.Annotations[currentEventAnnotation] = updateEvent

	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL cluster",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster metadata", pgCluster.ObjectMeta,
		)
		return err
	}

	pgCluster.Status.ID = id
	err = r.Status().Update(*ctx, pgCluster)
	if err != nil {
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

	return nil
}

func (r *PostgreSQLReconciler) HandleUpdateCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) (*reconcile.Result, error) {
	pgInstClusterStatus, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &reconcile.Result{}, err
	}

	err = r.reconcileDataCentresNodeSize(pgInstClusterStatus, pgCluster, logger)
	if err == instaclustr.ClusterNotRunning || err == instaclustr.StatusPreconditionFailed {
		logger.Error(err, "cluster is not ready to resize",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
		)
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: instaclustr.Requeue60,
		}, nil
	}
	if err == instaclustr.IncorrectNodeSize {
		logger.Error(err, "cannot downsize node type")
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: instaclustr.Requeue60,
		}, nil
	}
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
		)
		return &reconcile.Result{}, err
	}

	err = r.reconcileClusterConfigurations(
		pgCluster.Status.ID,
		pgInstClusterStatus.Status,
		pgCluster.Spec.ClusterConfigurations,
		logger,
	)
	if err != nil {
		if err == instaclustr.ClusterNotRunning {
			logger.Error(err, "cluster is not ready to update cluster configurations",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster status", pgInstClusterStatus.Status,
			)
			return &reconcile.Result{
				Requeue:      true,
				RequeueAfter: instaclustr.Requeue60,
			}, nil
		}

		logger.Error(err, "cannot reconcile cluster configurations",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
		)
		return &reconcile.Result{}, err
	}

	err = r.updateDescriptionAndTwoFactorDelete(pgCluster)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
		)
		return &reconcile.Result{}, err
	}

	pgCluster.Annotations[previousEventAnnotation] = pgCluster.Annotations[currentEventAnnotation]
	pgCluster.Annotations[currentEventAnnotation] = ""
	pgInstClusterStatus, err = r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return &reconcile.Result{}, err
	}

	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL metadata",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster metadata", pgCluster.ObjectMeta,
		)
		return &reconcile.Result{}, err
	}

	pgCluster.Status.ClusterStatus = *pgInstClusterStatus
	err = r.Status().Update(*ctx, pgCluster)
	if err != nil {
		return &reconcile.Result{}, err
	}

	logger.Info("PostgreSQL cluster was updated",
		"Cluster name", pgCluster.Spec.Name,
	)

	return &reconcile.Result{}, nil
}

func (r *PostgreSQLReconciler) HandleDeleteCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {

	// Cluster deletion logic

	logger.Info("PostgreSQL cluster was deleted",
		"Cluster name", pgCluster.Spec.Name,
		"Cluster id", pgCluster.Status.ID,
	)

	controllerutil.RemoveFinalizer(pgCluster, "finalizer")
	err := r.Update(*ctx, pgCluster)
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
				event.Object.SetAnnotations(map[string]string{currentEventAnnotation: createEvent})
				event.Object.SetFinalizers([]string{"finalizer"})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{currentEventAnnotation: deleteEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					currentEventAnnotation: updateEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{currentEventAnnotation: unknownEvent})
				return true
			},
		})).
		Complete(r)
}
