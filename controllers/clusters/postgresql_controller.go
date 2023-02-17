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
	"encoding/json"
	"errors"
	"strconv"

	"github.com/go-logr/logr"
	k8sCore "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
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
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;create;delete;update

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
				"resource name", req.NamespacedName,
			)
			return models.ExitReconcile, nil
		}

		logger.Error(err, "Unable to fetch PostgreSQL cluster",
			"resource name", req.NamespacedName,
		)
		return models.ExitReconcile, err
	}

	switch pgCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, pgCluster, logger), nil
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, pgCluster, logger), nil
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, pgCluster, logger), nil
	case models.GenericEvent:
		logger.Info("PostgreSQL resource generic event isn't handled",
			"cluster name", pgCluster.Spec.Name,
			"request", req,
			"event", pgCluster.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	default:
		logger.Info("PostgreSQL resource event isn't handled",
			"cluster name", pgCluster.Spec.Name,
			"request", req,
			"event", pgCluster.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *PostgreSQLReconciler) HandleCreateCluster(
	ctx context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	var id string
	var err error

	patch := pgCluster.NewPatch()
	if pgCluster.Status.DefaultUserSecretName == "" {
		secretName, err := pgCluster.GetUserSecretName(ctx, r.Client)
		if err != nil {
			logger.Error(err, "Cannot get PostgreSQL secret name",
				"cluster name", pgCluster.Spec.Name,
				"cluster ID", pgCluster.Status.ID,
			)

			return models.ReconcileRequeue
		}

		pgCluster.Status.DefaultUserSecretName = secretName

		if secretName == "" {
			secret := pgCluster.NewUserSecret()
			err = r.Client.Create(ctx, secret)
			if err != nil {
				logger.Error(err, "Cannot create PostgreSQL default user secret",
					"cluster ID", pgCluster.Status.ID,
				)

				return models.ReconcileRequeue
			}

			pgCluster.Status.DefaultUserSecretName = secret.Name

			logger.Info("PostgreSQL default user secret was created",
				"secret name", secret.Name,
				"cluster ID", pgCluster.Status.ID,
			)
		}
	}

	if pgCluster.Status.ID == "" {
		if pgCluster.Spec.HasRestore() {
			logger.Info(
				"Creating PostgreSQL cluster from backup",
				"original cluster ID", pgCluster.Spec.PgRestoreFrom.ClusterID,
			)

			id, err = r.API.RestorePgCluster(pgCluster.Spec.PgRestoreFrom)
			if err != nil {
				logger.Error(err, "Cannot restore PostgreSQL cluster from backup",
					"original cluster ID", pgCluster.Spec.PgRestoreFrom.ClusterID,
				)

				return models.ReconcileRequeue
			}

			pgCluster.Status.ID = id
			pgCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		} else {
			logger.Info(
				"Creating PostgreSQL cluster",
				"cluster name", pgCluster.Spec.Name,
				"data centres", pgCluster.Spec.DataCentres,
			)

			pgSpec := pgCluster.Spec.ToInstAPI()

			id, err = r.API.CreateCluster(instaclustr.PGSQLEndpoint, pgSpec)
			if err != nil {
				logger.Error(
					err, "Cannot create PostgreSQL cluster",
					"spec", pgCluster.Spec,
				)
				return models.ReconcileRequeue
			}

			pgCluster.Status.ID = id
			pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
		}

		logger.Info(
			"PostgreSQL resource has been created",
			"cluster name", pgCluster.Name,
			"cluster ID", pgCluster.Status.ID,
			"kind", pgCluster.Kind,
			"api version", pgCluster.APIVersion,
			"namespace", pgCluster.Namespace,
		)
	}

	err = r.Status().Patch(ctx, pgCluster, patch)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource status",
			"cluster name", pgCluster.Spec.Name,
			"status", pgCluster.Status,
		)

		return models.ReconcileRequeue
	}

	pgCluster.Annotations[models.DeletionConfirmed] = models.False
	controllerutil.AddFinalizer(pgCluster, models.DeletionFinalizer)

	err = r.Patch(ctx, pgCluster, patch)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource status",
			"cluster name", pgCluster.Spec.Name,
			"status", pgCluster.Status,
		)

		return models.ReconcileRequeue
	}

	err = r.startClusterStatusJob(pgCluster)
	if err != nil {
		logger.Error(err, "Cannot start PostgreSQL cluster status check job",
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.startClusterBackupsJob(pgCluster)
	if err != nil {
		logger.Error(err, "Cannot start PostgreSQL cluster backups check job",
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if pgCluster.Annotations[models.ResourceStateAnnotation] == models.UpdatingEvent {
		return reconcile.Result{Requeue: true}
	}

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) HandleUpdateCluster(
	ctx context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	instPgData, err := r.API.GetPostgreSQL(pgCluster.Status.ID)
	if err != nil {
		logger.Error(
			err, "Cannot get PostgreSQL cluster status from the Instaclustr API",
			"cluster name", pgCluster.Spec.Name,
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	instPg, err := pgCluster.FromInstAPI(instPgData)
	if err != nil {
		logger.Error(
			err, "Cannot convert PostgreSQL cluster status from the Instaclustr API",
			"cluster name", pgCluster.Spec.Name,
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if instPg.Status.CurrentClusterOperationStatus != models.NoOperation {
		logger.Info("PostgreSQL cluster is not ready to update",
			"cluster name", pgCluster.Spec.Name,
			"cluster status", instPg.Status.State,
			"current operation status", instPg.Status.CurrentClusterOperationStatus,
		)

		return models.ReconcileRequeue
	}

	if !pgCluster.Spec.AreDCsEqual(instPg.Spec.DataCentres) {
		err = r.updateDataCentres(pgCluster)
		if err != nil {
			logger.Error(err, "Cannot update Data Centres",
				"cluster name", pgCluster.Spec.Name,
			)

			return models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster data centres were updated",
			"cluster name", pgCluster.Spec.Name,
		)
	}

	instConfigs, err := r.API.GetPostgreSQLConfigs(pgCluster.Status.ID)
	if err != nil {
		logger.Error(err, "Cannot get PostgreSQL cluster configs",
			"cluster name", pgCluster.Spec.Name,
			"clusterID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	for _, instConfig := range instConfigs {
		err = r.reconcileClusterConfigurations(
			pgCluster.Status.ID,
			pgCluster.Spec.ClusterConfigurations,
			instConfig.ConfigurationProperties)
		if err != nil {
			logger.Error(err, "Cannot reconcile PostgreSQL cluster configs",
				"cluster name", pgCluster.Spec.Name,
				"clusterID", pgCluster.Status.ID,
				"configs", pgCluster.Spec.ClusterConfigurations,
				"inst configs", instConfig,
			)

			return models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster configurations were updated",
			"cluster name", pgCluster.Spec.Name,
		)
	}

	err = r.updateDescriptionAndTwoFactorDelete(pgCluster)
	if err != nil {
		logger.Error(err, "Cannot update description and twoFactorDelete",
			"cluster name", pgCluster.Spec.Name,
			"two factor delete", pgCluster.Spec.TwoFactorDelete,
			"description", pgCluster.Spec.Description,
		)

		return models.ReconcileRequeue
	}

	err = r.updateDefaultUserPassword(ctx, pgCluster)
	if err != nil {
		logger.Error(err, "Cannot update PostgreSQL default user password",
			"cluster ID", pgCluster.Status.ID,
			"secret name", pgCluster.Status.DefaultUserSecretName,
		)

		return models.ReconcileRequeue
	}

	pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource metadata",
			"cluster name", pgCluster.Spec.Name,
			"cluster metadata", pgCluster.ObjectMeta,
		)

		return models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was updated",
		"cluster name", pgCluster.Spec.Name,
		"cluster status", pgCluster.Status.State,
	)

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) HandleDeleteCluster(
	ctx context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	_, err := r.API.GetPostgreSQL(pgCluster.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get PostgreSQL cluster status",
			"cluster name", pgCluster.Spec.Name,
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if !errors.Is(err, instaclustr.NotFound) {
		if !clusterresourcesv1alpha1.IsClusterBeingDeleted(pgCluster.DeletionTimestamp, len(pgCluster.Spec.TwoFactorDelete), pgCluster.Annotations[models.DeletionConfirmed]) {
			pgCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			err = r.patchClusterMetadata(ctx, pgCluster, logger)
			if err != nil {
				logger.Error(err, "Cannot patch PostgreSQL cluster metadata",
					"cluster name", pgCluster.Spec.Name,
				)

				return models.ReconcileRequeue
			}

			logger.Info("PostgreSQL cluster deletion is not confirmed",
				"cluster name", pgCluster.Spec.Name,
				"cluster ID", pgCluster.Status.ID,
				"confirmation annotation", models.DeletionConfirmed,
				"annotation value", pgCluster.Annotations[models.DeletionConfirmed],
			)

			return models.ReconcileRequeue
		}

		err = r.API.DeleteCluster(pgCluster.Status.ID, instaclustr.PGSQLEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete PostgreSQL cluster",
				"cluster name", pgCluster.Spec.Name,
				"cluster status", pgCluster.Status.State,
			)

			return models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster is being deleted",
			"cluster name", pgCluster.Spec.Name,
			"cluster ID", pgCluster.Status.ID,
			"status", pgCluster.Status.State,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Deleting PostgreSQL default user secret",
		"cluster ID", pgCluster.Status.ID,
		"secret name", pgCluster.Status.DefaultUserSecretName,
	)

	err = r.deleteSecret(ctx, pgCluster)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Cannot delete PostgreSQL default user secret",
			"cluster ID", pgCluster.Status.ID,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster PostgreSQL default user secret was deleted",
		"cluster ID", pgCluster.Status.ID,
	)

	logger.Info("Deleting cluster backup resources",
		"cluster ID", pgCluster.Status.ID,
	)

	err = r.deleteBackups(ctx, pgCluster.Status.ID, pgCluster.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete PostgreSQL backup resources",
			"cluster ID", pgCluster.Status.ID,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster backup resources were deleted",
		"cluster ID", pgCluster.Status.ID,
	)

	r.Scheduler.RemoveJob(pgCluster.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(pgCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(pgCluster, models.DeletionFinalizer)
	pgCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.patchClusterMetadata(ctx, pgCluster, logger)
	if err != nil {
		logger.Error(
			err, "Cannot patch PostgreSQL resource metadata after finalizer removal",
			"cluster name", pgCluster.Spec.Name,
			"cluster ID", pgCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was deleted",
		"cluster name", pgCluster.Spec.Name,
		"cluster ID", pgCluster.Status.ID,
	)

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) updateDefaultUserPassword(
	ctx context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
) error {
	secret, err := pgCluster.GetUserSecret(ctx, r.Client)
	if err != nil {
		return err
	}

	if secret.Generation == 0 {
		return nil
	}

	password := pgCluster.GetUserPassword(secret)

	isValid := pgCluster.ValidateDefaultUserPassword(password)
	if !isValid {
		return models.ErrNotValidPassword
	}

	err = r.API.UpdatePostgreSQLDefaultUserPassword(pgCluster.Status.ID, password)
	if err != nil {
		return err
	}

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

func (r *PostgreSQLReconciler) startClusterBackupsJob(cluster *clustersv1alpha1.PostgreSQL) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) newWatchStatusJob(pg *clustersv1alpha1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLStatusClusterJob")

	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{Namespace: pg.Namespace, Name: pg.Name}, pg)
		if err != nil {
			l.Error(err, "Cannot get PosgtreSQL custom resource",
				"resource name", pg.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(
			pg.DeletionTimestamp,
			len(pg.Spec.TwoFactorDelete),
			pg.Annotations[models.DeletionConfirmed],
		) {
			l.Info("PostgreSQL cluster is being deleted. Status check job skipped",
				"cluster name", pg.Spec.Name,
				"cluster ID", pg.Status.ID,
			)

			return nil
		}

		instPGData, err := r.API.GetPostgreSQL(pg.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get PostgreSQL cluster status",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
			)

			return err
		}

		instaPG, err := pg.FromInstAPI(instPGData)
		if err != nil {
			l.Error(err, "Cannot convert PostgreSQL cluster status from Instaclustr",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
			)

			return err
		}

		if !areStatusesEqual(&instaPG.Status.ClusterStatus, &pg.Status.ClusterStatus) {
			l.Info("Updating PostgreSQL cluster status",
				"new cluster status", instaPG.Status,
				"old cluster status", pg.Status,
			)

			patch := pg.NewPatch()
			pg.Status.ClusterStatus = instaPG.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL cluster status",
					"cluster name", pg.Spec.Name,
					"cluster ID", pg.Status.ID,
					"instaclustr status", instaPG.Status,
				)
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(pg.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get PostgreSQL cluster maintenance events",
				"cluster name", pg.Spec.Name,
				"cluster ID", pg.Status.ID,
			)

			return err
		}

		if pg.Status.CurrentClusterOperationStatus == models.NoOperation &&
			!pg.Spec.IsEqual(instaPG.Spec) {
			l.Info("Updating PostgreSQL cluster spec",
				"new cluster", instPGData,
				"old cluster", pg,
			)

			patch := pg.NewPatch()
			pg.Spec = instaPG.Spec
			err = r.Patch(context.Background(), pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL cluster spec",
					"cluster name", pg.Spec.Name,
					"cluster ID", pg.Status.ID,
					"instaclustr spec", instPGData,
				)
				return err
			}
		}

		if !pg.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := pg.NewPatch()
			pg.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL cluster maintenance events",
					"cluster name", pg.Spec.Name,
					"cluster ID", pg.Status.ID,
				)

				return err
			}

			l.Info("PostgreSQL cluster maintenance events were updated",
				"cluster ID", pg.Status.ID,
				"events", pg.Status.MaintenanceEvents,
			)
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
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get PostgreSQL cluster backups",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.PgClusterKind)

		k8sBackupList, err := r.listClusterBackups(ctx, cluster.Status.ID, cluster.Namespace)
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
			if k8sBackup.Annotations[models.StartTimestampAnnotation] != "" {
				patch := k8sBackup.NewPatch()
				k8sBackup.Status.Start, err = strconv.Atoi(k8sBackup.Annotations[models.StartTimestampAnnotation])
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
					"backup resource name", k8sBackups[start].Name,
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
				"backup resource name", backupSpec.Name,
			)
		}

		return nil
	}
}

func (r *PostgreSQLReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1alpha1.ClusterBackupList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err := r.Client.List(ctx, backupsList, listOpts...)
	if err != nil {
		return nil, err
	}

	return backupsList, nil
}

func (r *PostgreSQLReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
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
	err = r.DeleteAllOf(ctx, backupType, opts...)
	if err != nil {
		return err
	}

	for _, backup := range backupsList.Items {
		patch := backup.NewPatch()
		controllerutil.RemoveFinalizer(&backup, models.DeletionFinalizer)
		err = r.Patch(ctx, &backup, patch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgreSQLReconciler) deleteSecret(ctx context.Context, pgCluster *clustersv1alpha1.PostgreSQL) error {
	secret, err := pgCluster.GetUserSecret(ctx, r.Client)
	if err != nil {
		return err
	}

	err = r.Client.Delete(ctx, secret)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) updateDataCentres(cluster *clustersv1alpha1.PostgreSQL) error {
	instDCs := cluster.Spec.DCsToInstAPI()
	err := r.API.UpdatePostgreSQLDataCentres(cluster.Status.ID, instDCs)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) reconcileClusterConfigurations(
	clusterID string,
	clusterConfigs map[string]string,
	instConfigs []*models.ConfigurationProperties) error {
	instConfigMap := convertAPIv2ConfigToMap(instConfigs)
	for k8sKey, k8sValue := range clusterConfigs {
		if instValue, exists := instConfigMap[k8sKey]; !exists {
			err := r.API.CreatePostgreSQLConfiguration(clusterID, k8sKey, k8sValue)
			if err != nil {
				return err
			}
		} else if instValue != k8sValue {
			err := r.API.UpdatePostgreSQLConfiguration(clusterID, k8sKey, k8sValue)
			if err != nil {
				return err
			}
		}
	}

	for instKey := range instConfigMap {
		if _, exists := clusterConfigs[instKey]; !exists {
			err := r.API.ResetPostgreSQLConfiguration(clusterID, instKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *PostgreSQLReconciler) patchClusterMetadata(
	ctx context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(pgCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(pgCluster.Finalizers)
	if err != nil {
		return err
	}

	finzlizersPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(ctx, pgCluster, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("PostgreSQL cluster patched",
		"Cluster name", pgCluster.Spec.Name,
		"Finalizers", pgCluster.Finalizers,
		"Annotations", pgCluster.Annotations,
	)
	return nil
}

func (r *PostgreSQLReconciler) updateDescriptionAndTwoFactorDelete(pgCluster *clustersv1alpha1.PostgreSQL) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
	if len(pgCluster.Spec.TwoFactorDelete) != 0 {
		twoFactorDelete = pgCluster.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, pgCluster.Status.ID, pgCluster.Spec.Description, twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) findSecretObject(secret client.Object) []reconcile.Request {
	pg := &clustersv1alpha1.PostgreSQL{}
	pgNamespacedName := types.NamespacedName{
		Namespace: secret.GetNamespace(),
		Name:      secret.GetLabels()[models.ControlledByLabel],
	}
	err := r.Get(context.TODO(), pgNamespacedName, pg)
	if err != nil {
		return []reconcile.Request{}
	}

	if pg.Annotations[models.ResourceStateAnnotation] == models.DeletingEvent {
		return []reconcile.Request{}
	}

	patch := pg.NewPatch()
	pg.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
	err = r.Patch(context.TODO(), pg, patch)
	if err != nil {
		return []reconcile.Request{}
	}

	return []reconcile.Request{{NamespacedName: pgNamespacedName}}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.PostgreSQL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				newObj := event.Object.(*clustersv1alpha1.PostgreSQL)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
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

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
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
		Owns(&k8sCore.Secret{}).
		Watches(
			&source.Kind{Type: &k8sCore.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findSecretObject),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(createEvent event.CreateEvent) bool {
					return createEvent.Object.GetGeneration() == 1
				},
			}),
		).
		Complete(r)
}
