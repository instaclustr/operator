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

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	apiv1convertors "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
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
		if k8serrors.IsNotFound(err) {
			logger.Info("PostgreSQL custom resource is not found",
				"Resource name", req.NamespacedName,
			)
			return reconcile.Result{}, nil
		}

		logger.Error(err, "unable to fetch PostgreSQL cluster",
			"Resource name", req.NamespacedName,
		)
		return reconcile.Result{}, err
	}

	pgAnnotations := pgCluster.Annotations
	switch pgAnnotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		err = r.HandleCreateCluster(&ctx, pgCluster, &logger, &req)
		if err != nil {
			logger.Error(err, "cannot create PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster manifest", pgCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	case models.UpdatingEvent:
		reconcileResult, err := r.HandleUpdateCluster(&ctx, pgCluster, &logger, &req)
		if err != nil {
			logger.Error(err, "cannot update PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}

		return *reconcileResult, nil
	case models.DeletingEvent:
		err = r.HandleDeleteCluster(pgCluster, &logger, &ctx, &req)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterIsBeingDeleted) {
				logger.Info("PostgreSQL cluster is still deleting",
					"Cluster name", pgCluster.Spec.Name,
					"Cluster ID", pgCluster.Status.ID,
					"Status", pgCluster.Status.Status,
				)

				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: models.Requeue60,
				}, nil
			}

			logger.Error(err, "cannot delete PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster status", pgCluster.Status.Status,
			)
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, err
	default:
		logger.Info("UNKNOWN EVENT",
			"Cluster name", pgCluster.Spec.Name,
		)
		return reconcile.Result{}, err
	}
}

func (r *PostgreSQLReconciler) HandleCreateCluster(
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	req *ctrl.Request,
) error {
	logger.Info(
		"Creating PostgreSQL cluster",
		"Cluster name", pgCluster.Spec.Name,
		"Data centres", pgCluster.Spec.DataCentres,
	)

	pgSpec := apiv1convertors.PgToInstAPI(&pgCluster.Spec)

	id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, pgSpec)
	if err != nil {
		logger.Error(
			err, "cannot create PostgreSQL cluster",
			"PostgreSQL manifest", pgCluster.Spec,
		)
		return err
	}

	pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
	pgCluster.SetFinalizers([]string{models.DeletionFinalizer})

	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL cluster",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster metadata", pgCluster.ObjectMeta,
		)
		return err
	}

	patch := pgCluster.NewPatch()
	pgCluster.Status.ID = id
	err = r.Status().Patch(*ctx, pgCluster, patch)
	if err != nil {
		return err
	}

	err = r.startClusterStatusJob(pgCluster)
	if err != nil {
		logger.Error(err, "cannot start cluster status job",
			"PostgreSQL cluster ID", pgCluster.Status.ID)
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
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	req *ctrl.Request,
) (*reconcile.Result, error) {
	pgInstClusterStatus, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileResult, err
	}

	err = r.reconcileDataCentresNodeSize(pgInstClusterStatus, pgCluster, logger)
	if errors.Is(err, instaclustr.ClusterNotRunning) || errors.Is(err, instaclustr.StatusPreconditionFailed) {
		logger.Info("cluster is not ready to resize",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Reason", err,
		)
		return &models.ReconcileRequeue, nil
	}
	if errors.Is(err, instaclustr.IncorrectNodeSize) {
		logger.Info("cannot downsize node type",
			"Cluster name", pgCluster.Spec.Name,
			"Current node size", pgInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"New node size", pgCluster.Spec.DataCentres[0].NodeSize,
			"Reason", err,
		)
		return &models.ReconcileRequeue, nil
	}
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Current node size", pgInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"New node size", pgCluster.Spec.DataCentres[0].NodeSize,
		)
		return &models.ReconcileResult, err
	}
	logger.Info("Data centres were reconciled",
		"Cluster name", pgCluster.Spec.Name,
	)

	err = r.reconcileClusterConfigurations(
		pgCluster.Status.ID,
		pgInstClusterStatus.Status,
		pgCluster.Spec.ClusterConfigurations,
		logger,
	)
	if err != nil {
		if errors.Is(err, instaclustr.ClusterNotRunning) {
			logger.Info("cluster is not ready to update cluster configurations",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster status", pgInstClusterStatus.Status,
				"Reason", err,
			)
			return &models.ReconcileRequeue, nil
		}

		logger.Error(err, "cannot reconcile cluster configurations",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Configurations", pgCluster.Spec.ClusterConfigurations,
		)
		return &models.ReconcileResult, err
	}

	err = r.updateDescriptionAndTwoFactorDelete(pgCluster)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Two factor delete", pgCluster.Spec.TwoFactorDelete,
		)
		return &models.ReconcileResult, err
	}

	pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	pgInstClusterStatus, err = r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return &models.ReconcileResult, err
	}

	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL metadata",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster metadata", pgCluster.ObjectMeta,
		)
		return &models.ReconcileResult, err
	}

	patch := pgCluster.NewPatch()
	pgCluster.Status.ClusterStatus = *pgInstClusterStatus
	err = r.Status().Patch(*ctx, pgCluster, patch)
	if err != nil {
		return &models.ReconcileResult, err
	}

	logger.Info("PostgreSQL cluster was updated",
		"Cluster name", pgCluster.Spec.Name,
		"Cluster status", pgCluster.Status.Status,
	)

	return &models.ReconcileResult, nil
}

func (r *PostgreSQLReconciler) HandleDeleteCluster(
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
	ctx *context.Context,
	req *ctrl.Request,
) error {
	status, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		return err
	}

	if status != nil {
		err = r.API.DeleteCluster(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster status", pgCluster.Status.Status,
			)

			return err
		}
		return instaclustr.ClusterIsBeingDeleted
	}

	r.Scheduler.RemoveJob(pgCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(pgCluster, models.DeletionFinalizer)
	pgCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Update(*ctx, pgCluster)
	if err != nil {
		logger.Error(
			err, "cannot update PostgreSQL cluster CRD after finalizer removal",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)
		return err
	}

	logger.Info("PostgreSQL cluster was deleted",
		"Cluster name", pgCluster.Spec.Name,
		"Cluster ID", pgCluster.Status.ID,
	)

	return nil
}

func (r *PostgreSQLReconciler) startClusterStatusJob(cluster *clustersv1alpha1.PostgreSQL) error {
	job := r.newWatchStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) newWatchStatusJob(cluster *clustersv1alpha1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLStatusClusterJob")

	return func() error {
		instStatus, err := r.API.GetClusterStatus(cluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			l.Error(err, "cannot get PostgreSQL cluster status",
				"Cluster name", cluster.Spec.Name,
				"ClusterID", cluster.Status.ID,
			)
			return err
		}

		if !isStatusesEqual(&cluster.Status.ClusterStatus, instStatus) {
			l.Info("Updating PostgreSQL cluster status",
				"New status", instStatus,
				"Old status", cluster.Status.ClusterStatus,
			)

			patch := cluster.NewPatch()
			cluster.Status.ClusterStatus = *instStatus
			err = r.Status().Patch(context.Background(), cluster, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.PostgreSQL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				annotations := event.Object.GetAnnotations()
				if annotations[models.ResourceStateAnnotation] != "" {
					event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
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
