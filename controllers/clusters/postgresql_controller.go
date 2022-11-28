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
	"strconv"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
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
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=get;list;create;update;patch;deletecollection;delete

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
			return models.ReconcileResult, nil
		}

		logger.Error(err, "unable to fetch PostgreSQL cluster",
			"Resource name", req.NamespacedName,
		)
		return models.ReconcileResult, err
	}

	switch pgCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		reconcileResult := r.HandleCreateCluster(&ctx, pgCluster, &logger)
		return *reconcileResult, nil
	case models.UpdatingEvent:
		reconcileResult := r.HandleUpdateCluster(&ctx, pgCluster, &logger)
		return *reconcileResult, nil
	case models.DeletingEvent:
		reconcileResult := r.HandleDeleteCluster(&ctx, pgCluster, &logger)
		return *reconcileResult, nil
	case models.GenericEvent:
		logger.Info("PostgreSQL resource generic event isn't handled",
			"Cluster name", pgCluster.Spec.Name,
			"Request", req,
			"Event", pgCluster.Annotations[models.ResourceStateAnnotation],
		)
		return models.ReconcileResult, nil
	default:
		logger.Info("PostgreSQL resource event isn't handled",
			"Cluster name", pgCluster.Spec.Name,
			"Request", req,
			"Event", pgCluster.Annotations[models.ResourceStateAnnotation],
		)
		return models.ReconcileResult, nil
	}
}

func (r *PostgreSQLReconciler) HandleCreateCluster(
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) *reconcile.Result {
	if pgCluster.Status.ID == "" {
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
			return &models.ReconcileRequeue
		}

		pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
		pgCluster.Annotations[models.DeletionConfirmed] = models.False
		pgCluster.Finalizers = append(pgCluster.Finalizers, models.DeletionFinalizer)
		err = r.patchClusterMetadata(ctx, pgCluster, logger)
		if err != nil {
			logger.Error(err, "cannot patch PostgreSQL resource metadata",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster metadata", pgCluster.ObjectMeta,
			)
			return &models.ReconcileRequeue
		}

		patch := pgCluster.NewPatch()
		pgCluster.Status.ID = id
		err = r.Status().Patch(*ctx, pgCluster, patch)
		if err != nil {
			logger.Error(err, "cannot patch PostgreSQL resource status",
				"Cluster name", pgCluster.Spec.Name,
				"Status", pgCluster.Status,
			)

			return &models.ReconcileRequeue
		}

		logger.Info(
			"PostgreSQL resource has been created",
			"Cluster name", pgCluster.Name,
			"Cluster ID", pgCluster.Status.ID,
			"Kind", pgCluster.Kind,
			"Api version", pgCluster.APIVersion,
			"Namespace", pgCluster.Namespace,
		)
	}

	err := r.startClusterStatusJob(pgCluster)
	if err != nil {
		logger.Error(err, "cannot start PostgreSQL cluster status check job",
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	err = r.startClusterBackupsJob(pgCluster)
	if err != nil {
		logger.Error(err, "cannot start PostgreSQL cluster backups check job",
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	return &reconcile.Result{Requeue: true}
}

func (r *PostgreSQLReconciler) HandleUpdateCluster(
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) *reconcile.Result {
	pgInstClusterStatus, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	if pgInstClusterStatus.Status != models.RunningStatus {
		logger.Info("cluster is not ready to update",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Reason", instaclustr.ClusterNotRunning,
		)

		return &models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNodeSize(pgInstClusterStatus, pgCluster, logger)
	if errors.Is(err, instaclustr.StatusPreconditionFailed) {
		logger.Info("cluster is not ready to resize",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Reason", err,
		)

		return &models.ReconcileRequeue
	}
	if err != nil {
		logger.Error(err, "cannot reconcile data centres node size",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Current node size", pgInstClusterStatus.DataCentres[0].Nodes[0].Size,
			"New node size", pgCluster.Spec.DataCentres[0].NodeSize,
		)

		return &models.ReconcileRequeue
	}

	logger.Info("Data centres were reconciled",
		"Cluster name", pgCluster.Spec.Name,
	)

	err = r.reconcileClusterConfigurations(pgCluster.Status.ID, pgCluster.Spec.ClusterConfigurations, logger)
	if err != nil {
		logger.Error(err, "cannot reconcile cluster configurations",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Configurations", pgCluster.Spec.ClusterConfigurations,
		)

		return &models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(pgCluster)
	if err != nil {
		logger.Error(err, "cannot update description and twoFactorDelete",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster status", pgInstClusterStatus.Status,
			"Two factor delete", pgCluster.Spec.TwoFactorDelete,
			"Description", pgCluster.Spec.Description,
		)

		return &models.ReconcileRequeue
	}

	pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	pgInstClusterStatus, err = r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get PostgreSQL cluster status from the Instaclustr API",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL resource metadata",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster metadata", pgCluster.ObjectMeta,
		)

		return &models.ReconcileRequeue
	}

	patch := pgCluster.NewPatch()
	pgCluster.Status.ClusterStatus = *pgInstClusterStatus
	err = r.Status().Patch(*ctx, pgCluster, patch)
	if err != nil {
		logger.Error(err, "cannot patch PostgreSQL resource status",
			"Cluster name", pgCluster.Spec.Name,
			"Status", pgCluster.Status,
		)

		return &models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was updated",
		"Cluster name", pgCluster.Spec.Name,
		"Cluster status", pgCluster.Status.Status,
	)

	return &models.ReconcileResult
}

func (r *PostgreSQLReconciler) HandleDeleteCluster(
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) *reconcile.Result {
	status, err := r.API.GetClusterStatus(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "cannot get PostgreSQL cluster status",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	if status != nil {
		if len(pgCluster.Spec.TwoFactorDelete) != 0 &&
			pgCluster.Annotations[models.DeletionConfirmed] != models.True {
			pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			err = r.patchClusterMetadata(ctx, pgCluster, logger)
			if err != nil {
				logger.Error(err, "cannot patch PostgreSQL cluster metadata",
					"Cluster name", pgCluster.Spec.Name,
				)

				return &models.ReconcileRequeue
			}

			logger.Info("PostgreSQL cluster deletion is not confirmed",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster ID", pgCluster.Status.ID,
				"Confirmation annotation", models.DeletionConfirmed,
				"Annotation value", pgCluster.Annotations[models.DeletionConfirmed],
			)

			return &models.ReconcileRequeue
		}

		err = r.API.DeleteCluster(pgCluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete PostgreSQL cluster",
				"Cluster name", pgCluster.Spec.Name,
				"Cluster status", pgCluster.Status.Status,
			)

			return &models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster is being deleted",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
			"Status", pgCluster.Status.Status,
		)

		return &models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(pgCluster.GetJobID(scheduler.BackupsChecker))

	logger.Info("Deleting cluster backup resources",
		"Cluster ID", pgCluster.Status.ID,
	)

	err = r.deleteBackups(ctx, pgCluster.Status.ID, pgCluster.Namespace)
	if err != nil {
		logger.Error(err, "cannot delete cluster backup resources",
			"Cluster ID", pgCluster.Status.ID,
		)
		return &models.ReconcileRequeue
	}

	logger.Info("Cluster backup resources were deleted",
		"Cluster ID", pgCluster.Status.ID,
	)

	r.Scheduler.RemoveJob(pgCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(pgCluster, models.DeletionFinalizer)
	pgCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(
			err, "cannot patch PostgreSQL resource metadata after finalizer removal",
			"Cluster name", pgCluster.Spec.Name,
			"Cluster ID", pgCluster.Status.ID,
		)

		return &models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was deleted",
		"Cluster name", pgCluster.Spec.Name,
		"Cluster ID", pgCluster.Status.ID,
	)

	return &models.ReconcileResult
}

func (r *PostgreSQLReconciler) startClusterStatusJob(cluster *clustersv1alpha1.PostgreSQL) error {
	job := r.newWatchStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) startClusterBackupsJob(cluster *clustersv1alpha1.PostgreSQL) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) newWatchStatusJob(cluster *clustersv1alpha1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLStatusClusterJob")

	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			l.Error(err, "cannot get PosgtreSQL custom resource",
				"Resource name", cluster.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("PostgreSQL cluster is being deleted. Status check job skipped",
				"Cluster name", cluster.Spec.Name,
				"Cluster ID", cluster.Status.ID,
			)

			return nil
		}

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

func (r *PostgreSQLReconciler) newWatchBackupsJob(cluster *clustersv1alpha1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("PostgreSQL cluster is being deleted. Backups check job skipped",
				"Cluster name", cluster.Spec.Name,
				"Cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "cannot get PostgreSQL cluster backups",
				"Cluster name", cluster.Spec.Name,
				"ClusterID", cluster.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.PgBackupEventType)

		k8sBackupList, err := r.listClusterBackups(&ctx, cluster.Status.ID, cluster.Namespace)
		if err != nil {
			return err
		}

		k8sBackups := map[int]*clusterresourcesv1alpha1.ClusterBackup{}
		unassignedBackups := []*clusterresourcesv1alpha1.ClusterBackup{}
		for _, k8sBackup := range k8sBackupList.Items {
			if k8sBackup.Status.Start != 0 {
				k8sBackups[k8sBackup.Status.Start] = &k8sBackup
				continue
			}
			if k8sBackup.Annotations[models.StartAnnotation] != "" {
				patch := k8sBackup.NewPatch()
				k8sBackup.Status.Start, err = strconv.Atoi(k8sBackup.Annotations[models.StartAnnotation])
				if err != nil {
					return err
				}

				err = r.Status().Patch(ctx, &k8sBackup, patch)
				if err != nil {
					return err
				}

				k8sBackups[k8sBackup.Status.Start] = &k8sBackup
				continue
			}

			unassignedBackups = append(unassignedBackups, &k8sBackup)
		}

		for start, instBackup := range instBackupEvents {
			if _, exists := k8sBackups[start]; exists {
				if k8sBackups[start].Status.End != 0 {
					continue
				}

				patch := k8sBackups[start].NewPatch()
				k8sBackups[start].Status.UpdateStatus(instBackup)
				err = r.Status().Patch(ctx, k8sBackups[start], patch)
				if err != nil {
					return err
				}

				l.Info("Backup resource was updated",
					"Backup resource name", k8sBackups[start].Name,
				)
				continue
			}

			if len(unassignedBackups) != 0 {
				backupToAssign := unassignedBackups[len(unassignedBackups)-1]
				unassignedBackups = unassignedBackups[:len(unassignedBackups)-1]
				patch := backupToAssign.NewPatch()
				backupToAssign.Status.Start = instBackup.Start
				backupToAssign.Status.UpdateStatus(instBackup)
				err = r.Status().Patch(context.TODO(), backupToAssign, patch)
				if err != nil {
					return err
				}
				continue
			}

			backupSpec := cluster.NewBackupSpec(start)
			err = r.Create(ctx, backupSpec)
			if err != nil {
				return err
			}
			l.Info("Found new backup on Instaclustr. New backup resource was created",
				"Backup resource name", backupSpec.Name,
			)
		}

		return nil
	}
}

func (r *PostgreSQLReconciler) listClusterBackups(ctx *context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1alpha1.ClusterBackupList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err := r.Client.List(*ctx, backupsList, listOpts...)
	if err != nil {
		return nil, err
	}

	return backupsList, nil
}

func (r *PostgreSQLReconciler) deleteBackups(ctx *context.Context, clusterID, namespace string) error {
	backupsList, err := r.listClusterBackups(ctx, clusterID, namespace)
	if err != nil {
		return err
	}

	if len(backupsList.Items) == 0 {
		return nil
	}

	backupType := &clusterresourcesv1alpha1.ClusterBackup{}
	opts := []client.DeleteAllOfOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err = r.DeleteAllOf(*ctx, backupType, opts...)
	if err != nil {
		return err
	}

	for _, backup := range backupsList.Items {
		patch := backup.NewPatch()
		controllerutil.RemoveFinalizer(&backup, models.DeletionFinalizer)
		err = r.Patch(*ctx, &backup, patch)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.PostgreSQL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*clustersv1alpha1.PostgreSQL)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).
		Owns(&clusterresourcesv1alpha1.ClusterBackup{}).
		Complete(r)
}
