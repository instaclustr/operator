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
	"k8s.io/client-go/tools/record"
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

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=get;list;create;update;patch;deletecollection;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;create;delete;update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pg := &v1beta1.PostgreSQL{}
	err := r.Client.Get(ctx, req.NamespacedName, pg)
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

	switch pg.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, pg, logger), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, pg, logger), nil
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, pg, logger), nil
	case models.SecretEvent:
		return r.handleUpdateDefaultUserPassword(ctx, pg, logger), nil
	case models.GenericEvent:
		logger.Info("PostgreSQL resource generic event isn't handled",
			"cluster name", pg.Spec.Name,
			"request", req,
			"event", pg.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	default:
		logger.Info("PostgreSQL resource event isn't handled",
			"cluster name", pg.Spec.Name,
			"request", req,
			"event", pg.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *PostgreSQLReconciler) handleCreateCluster(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	logger = logger.WithName("PostgreSQL creation event")

	var err error

	patch := pg.NewPatch()
	if pg.Status.ID == "" {
		if pg.Spec.HasRestore() {
			logger.Info(
				"Creating PostgreSQL cluster from backup",
				"original cluster ID", pg.Spec.PgRestoreFrom.ClusterID,
			)

			pg.Status.ID, err = r.API.RestorePgCluster(pg.Spec.PgRestoreFrom)
			if err != nil {
				logger.Error(err, "Cannot restore PostgreSQL cluster from backup",
					"original cluster ID", pg.Spec.PgRestoreFrom.ClusterID,
				)

				r.EventRecorder.Eventf(
					pg, models.Warning, models.CreationFailed,
					"Cluster restoration from backup on Instaclustr cloud is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue
			}

			r.EventRecorder.Eventf(
				pg, models.Normal, models.Created,
				"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
				pg.Spec.PgRestoreFrom.ClusterID,
				pg.Status.ID,
			)
		} else {
			logger.Info(
				"Creating PostgreSQL cluster",
				"cluster name", pg.Spec.Name,
				"data centres", pg.Spec.DataCentres,
			)

			pgSpec := pg.Spec.ToInstAPI()

			pg.Status.ID, err = r.API.CreateCluster(instaclustr.PGSQLEndpoint, pgSpec)
			if err != nil {
				logger.Error(
					err, "Cannot create PostgreSQL cluster",
					"spec", pg.Spec,
				)

				r.EventRecorder.Eventf(
					pg, models.Warning, models.CreationFailed,
					"Cluster creation on the Instaclustr is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue
			}

			r.EventRecorder.Eventf(
				pg, models.Normal, models.Created,
				"Cluster creation request is sent. Cluster ID: %s",
				pg.Status.ID,
			)
		}

		err = r.Status().Patch(ctx, pg, patch)
		if err != nil {
			logger.Error(err, "Cannot patch PostgreSQL resource status",
				"cluster name", pg.Spec.Name,
				"status", pg.Status,
			)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)

			return models.ReconcileRequeue
		}

		logger.Info(
			"PostgreSQL resource has been created",
			"cluster name", pg.Name,
			"cluster ID", pg.Status.ID,
			"kind", pg.Kind,
			"api version", pg.APIVersion,
			"namespace", pg.Namespace,
		)
	}

	controllerutil.AddFinalizer(pg, models.DeletionFinalizer)

	pg.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
	err = r.Patch(ctx, pg, patch)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource",
			"cluster name", pg.Spec.Name,
			"status", pg.Status)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	err = r.startClusterStatusJob(pg)
	if err != nil {
		logger.Error(err, "Cannot start PostgreSQL cluster status check job",
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	err = r.startClusterBackupsJob(pg)
	if err != nil {
		logger.Error(err, "Cannot start PostgreSQL cluster backups check job",
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.CreationFailed,
			"Cluster backups check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Created,
		"Cluster backups check job is started",
	)

	err = r.createDefaultPassword(ctx, pg, logger)
	if err != nil {
		logger.Error(err, "Cannot create default password for PostgreSQL",
			"cluster name", pg.Spec.Name,
			"clusterID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.CreationFailed,
			"Default user secret creation on the Instaclustr is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	if pg.Spec.UserRefs != nil {
		err = r.startUsersCreationJob(pg)
		if err != nil {
			logger.Error(err, "Failed to start user PostreSQL creation job")
			r.EventRecorder.Eventf(pg, models.Warning, models.CreationFailed,
				"User creation job is failed. Reason: %v", err)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Event(pg, models.Normal, models.Created,
			"Cluster user creation job is started")
	}

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) handleUpdateCluster(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	logger = logger.WithName("PostgreSQL update event")

	iData, err := r.API.GetPostgreSQL(pg.Status.ID)
	if err != nil {
		logger.Error(
			err, "Cannot get PostgreSQL cluster status from the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	iPg, err := pg.FromInstAPI(iData)
	if err != nil {
		logger.Error(
			err, "Cannot convert PostgreSQL cluster status from the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if iPg.Status.CurrentClusterOperationStatus != models.NoOperation {
		logger.Info("PostgreSQL cluster is not ready to update",
			"cluster name", pg.Spec.Name,
			"cluster status", iPg.Status.State,
			"current operation status", iPg.Status.CurrentClusterOperationStatus,
		)
		patch := pg.NewPatch()
		pg.Annotations[models.UpdateQueuedAnnotation] = models.True
		err = r.Patch(ctx, pg, patch)
		if err != nil {
			logger.Error(err, "Cannot patch cluster resource",
				"cluster name", pg.Spec.Name, "cluster ID", pg.Status.ID)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		return models.ReconcileRequeue
	}

	if pg.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(pg, iPg, logger)
	}

	if !pg.Spec.AreDCsEqual(iPg.Spec.DataCentres) {
		err = r.updateDataCentres(pg)
		if err != nil {
			logger.Error(err, "Cannot update Data Centres",
				"cluster name", pg.Spec.Name,
			)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v",
				err,
			)

			patch := pg.NewPatch()
			pg.Annotations[models.UpdateQueuedAnnotation] = models.True
			err = r.Patch(ctx, pg, patch)
			if err != nil {
				logger.Error(err, "Cannot patch PostgreSQL metadata",
					"cluster name", pg.Spec.Name,
					"cluster metadata", pg.ObjectMeta,
				)

				r.EventRecorder.Eventf(
					pg, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue
			}
			return models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster data centres were updated",
			"cluster name", pg.Spec.Name,
		)
	}

	iConfigs, err := r.API.GetPostgreSQLConfigs(pg.Status.ID)
	if err != nil {
		logger.Error(err, "Cannot get PostgreSQL cluster configs",
			"cluster name", pg.Spec.Name,
			"clusterID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster configs fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	for _, iConfig := range iConfigs {
		err = r.reconcileClusterConfigurations(
			pg.Status.ID,
			pg.Spec.ClusterConfigurations,
			iConfig.ConfigurationProperties)
		if err != nil {
			logger.Error(err, "Cannot reconcile PostgreSQL cluster configs",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
				"configs", pg.Spec.ClusterConfigurations,
				"inst configs", iConfig,
			)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.UpdateFailed,
				"Cluster configs fetch from the Instaclustr API is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		logger.Info("PostgreSQL cluster configurations were updated",
			"cluster name", pg.Spec.Name,
		)
	}

	err = r.updateDescriptionAndTwoFactorDelete(pg)
	if err != nil {
		logger.Error(err, "Cannot update description and twoFactorDelete",
			"cluster name", pg.Spec.Name,
			"two factor delete", pg.Spec.TwoFactorDelete,
			"description", pg.Spec.Description,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.UpdateFailed,
			"Cluster description and TwoFactoDelete update is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	pg.Annotations[models.UpdateQueuedAnnotation] = ""
	err = r.patchClusterMetadata(ctx, pg, logger)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource metadata",
			"cluster name", pg.Spec.Name,
			"cluster metadata", pg.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was updated",
		"cluster name", pg.Spec.Name,
		"cluster status", pg.Status.State,
	)

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) createUser(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.PostgreSQL,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.PostgreSQLUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot create a PostgreSQL user. The resource is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User is not found, create a new one PostgreSQL User or provide correct userRef."+
					"Current provided reference: %v", uRef)
			return err
		}

		l.Error(err, "Cannot get PostgreSQL user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot get PostgreSQL user. User reference: %v", uRef)
		return err
	}

	secret, err := v1beta1.GetDefaultPgUserSecret(ctx, c.Name, c.Namespace, r.Client)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.EventRecorder.Eventf(
			c, models.Warning, models.FetchFailed,
			"Default user secret fetch is failed. Reason: %v",
			err,
		)

		return err
	}

	defaultSecretNamespacedName := types.NamespacedName{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	if _, exist := u.Status.ClustersInfo[c.Status.ID]; exist {
		l.Info("User is already existing on the cluster",
			"user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.CreationFailed,
			"User is already existing on the cluster. User reference: %v", uRef)

		return nil
	}

	patch := u.NewPatch()

	if u.Status.ClustersInfo == nil {
		u.Status.ClustersInfo = make(map[string]clusterresourcesv1beta1.ClusterInfo)
	}

	u.Status.ClustersInfo[c.Status.ID] = clusterresourcesv1beta1.ClusterInfo{
		DefaultSecretNamespacedName: clusterresourcesv1beta1.NamespacedName{
			Namespace: defaultSecretNamespacedName.Namespace,
			Name:      defaultSecretNamespacedName.Name,
		},
		Event: models.CreatingEvent,
	}

	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the PostgreSQL User status with the CreatingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot add PostgreSQL User to the cluster. Reason: %v", err)
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) handleUserEvent(
	newObj *v1beta1.PostgreSQL,
	oldUsers []*v1beta1.UserReference,
) {
	ctx := context.TODO()
	l := log.FromContext(ctx)

	for _, newUser := range newObj.Spec.UserRefs {
		var exist bool

		for _, oldUser := range oldUsers {
			if *newUser == *oldUser {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		err := r.createUser(ctx, l, newObj, newUser)
		if err != nil {
			l.Error(err, "Cannot create PostgreSQL user in predicate", "user", newUser)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot create user. Reason: %v", err)
		}

		oldUsers = append(oldUsers, newUser)
	}

	for _, oldUser := range oldUsers {
		var exist bool

		for _, newUser := range newObj.Spec.UserRefs {
			if *oldUser == *newUser {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		// TODO: implement user deletion
	}
}

func (r *PostgreSQLReconciler) handleExternalChanges(pg, iPg *v1beta1.PostgreSQL, l logr.Logger) reconcile.Result {
	if !pg.Spec.IsEqual(iPg.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", pg.Spec,
			"data from Instaclustr ", iPg.Spec)
		msgDiffSpecs, err := createSpecDifferenceMessage(pg.Spec, iPg.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iPg.Spec, "k8s resource spec", pg.Spec)
			return models.ExitReconcile
		}
		r.EventRecorder.Eventf(pg, models.Warning, models.ExternalChanges, msgDiffSpecs)

		return models.ExitReconcile
	}

	patch := pg.NewPatch()

	pg.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), pg, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", pg.Spec.Name, "cluster ID", pg.Status.ID)

		r.EventRecorder.Eventf(pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	l.Info("External changes have been reconciled", "resource ID", pg.Status.ID)
	r.EventRecorder.Event(pg, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) handleDeleteCluster(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	logger = logger.WithName("PostgreSQL deletion event")

	_, err := r.API.GetPostgreSQL(pg.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get PostgreSQL cluster status",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if !errors.Is(err, instaclustr.NotFound) {
		logger.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID)

		err = r.API.DeleteCluster(pg.Status.ID, instaclustr.PGSQLEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete PostgreSQL cluster",
				"cluster name", pg.Spec.Name,
				"cluster status", pg.Status.State,
			)
			r.EventRecorder.Eventf(
				pg, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v",
				err,
			)

			return models.ReconcileRequeue
		}

		r.EventRecorder.Event(pg, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if pg.Spec.TwoFactorDelete != nil {
			patch := pg.NewPatch()

			pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			pg.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, pg, patch)
			if err != nil {
				logger.Error(err, "Cannot patch cluster resource",
					"cluster name", pg.Spec.Name,
					"cluster state", pg.Status.State)
				r.EventRecorder.Eventf(pg, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return models.ReconcileRequeue
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", pg.Status.ID)

			r.EventRecorder.Event(pg, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
	}

	logger.Info("PostgreSQL cluster is being deleted. Deleting PostgreSQL default user secret",
		"cluster ID", pg.Status.ID,
	)

	err = r.deleteSecret(ctx, pg)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Cannot delete PostgreSQL default user secret",
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.DeletionFailed,
			"Default user secret deletion is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster PostgreSQL default user secret was deleted",
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Deleted,
		"Default user secret is deleted. Cluster ID: %s",
		pg.Status.ID,
	)

	logger.Info("Deleting cluster backup resources",
		"cluster ID", pg.Status.ID,
	)

	err = r.deleteBackups(ctx, pg.Status.ID, pg.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete PostgreSQL backup resources",
			"cluster ID", pg.Status.ID,
		)
		r.EventRecorder.Eventf(
			pg, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster backup resources were deleted",
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Deleted,
		"Cluster backup resources are deleted",
	)

	r.Scheduler.RemoveJob(pg.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(pg.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(pg, models.DeletionFinalizer)
	pg.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.patchClusterMetadata(ctx, pg, logger)
	if err != nil {
		logger.Error(
			err, "Cannot patch PostgreSQL resource metadata after finalizer removal",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	err = exposeservice.Delete(r.Client, pg.Name, pg.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete PostgreSQL cluster expose service",
			"cluster ID", pg.Status.ID,
			"cluster name", pg.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	logger.Info("PostgreSQL cluster was deleted",
		"cluster name", pg.Spec.Name,
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) handleUpdateDefaultUserPassword(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	logger logr.Logger,
) reconcile.Result {
	logger = logger.WithName("PostgreSQL default user password updating event")

	secret, err := v1beta1.GetDefaultPgUserSecret(ctx, pg.Name, pg.Namespace, r.Client)
	if err != nil {
		logger.Error(err, "Cannot get the default secret for the PostgreSQL cluster",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Fetch default user secret is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	password := string(secret.Data[models.Password])
	isValid := pg.ValidateDefaultUserPassword(password)
	if !isValid {
		logger.Error(err, "Default PostgreSQL user password is not valid. This field must be at least 8 characters long. Must contain characters from at least 3 of the following 4 categories: Uppercase, Lowercase, Numbers, Special Characters",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.ValidationFailed,
			"Validation for default user secret is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	err = r.API.UpdatePostgreSQLDefaultUserPassword(pg.Status.ID, password)
	if err != nil {
		logger.Error(err, "Cannot update default PostgreSQL user password",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.UpdateFailed,
			"Default user password update on the Instaclustr API is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, pg, logger)
	if err != nil {
		logger.Error(err, "Cannot patch PostgreSQL resource metadata",
			"cluster name", pg.Spec.Name,
			"cluster metadata", pg.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	logger.Info("PostgreSQL default user password was updated",
		"cluster name", pg.Spec.Name,
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.UpdatedEvent,
		"Cluster default user password is updated",
	)

	return models.ExitReconcile
}

func (r *PostgreSQLReconciler) startClusterStatusJob(pg *v1beta1.PostgreSQL) error {
	job := r.newWatchStatusJob(pg)

	err := r.Scheduler.ScheduleJob(pg.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) startClusterBackupsJob(pg *v1beta1.PostgreSQL) error {
	job := r.newWatchBackupsJob(pg)

	err := r.Scheduler.ScheduleJob(pg.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) startUsersCreationJob(cluster *v1beta1.PostgreSQL) error {
	job := r.newUsersCreationJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.UserCreator), scheduler.UserCreationInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) newWatchStatusJob(pg *v1beta1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLStatusClusterJob")

	return func() error {
		namespacedName := client.ObjectKeyFromObject(pg)
		err := r.Get(context.Background(), namespacedName, pg)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(pg.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(pg.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get PosgtreSQL custom resource",
				"resource name", pg.Name,
			)
			return err
		}

		instPGData, err := r.API.GetPostgreSQL(pg.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(pg.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", pg.Status.ClusterStatus.ID,
						"cluster name", pg.Spec.Name,
					)

					patch := pg.NewPatch()
					pg.Annotations[models.ClusterDeletionAnnotation] = ""
					pg.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), pg, patch)
					if err != nil {
						l.Error(err, "Cannot patch PostgreSQL cluster resource",
							"cluster ID", pg.Status.ID,
							"cluster name", pg.Spec.Name,
							"resource name", pg.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), pg)
					if err != nil {
						l.Error(err, "Cannot delete PostgreSQL cluster resource",
							"cluster ID", pg.Status.ID,
							"cluster name", pg.Spec.Name,
							"resource name", pg.Name,
						)

						return err
					}

					return nil
				}
			}

			l.Error(err, "Cannot get PostgreSQL cluster status",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
			)

			return err
		}

		iPg, err := pg.FromInstAPI(instPGData)
		if err != nil {
			l.Error(err, "Cannot convert PostgreSQL cluster status from Instaclustr",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
			)

			return err
		}

		if !areStatusesEqual(&iPg.Status.ClusterStatus, &pg.Status.ClusterStatus) {
			l.Info("Updating PostgreSQL cluster status",
				"new cluster status", iPg.Status,
				"old cluster status", pg.Status,
			)

			areDCsEqual := areDataCentresEqual(iPg.Status.ClusterStatus.DataCentres, pg.Status.ClusterStatus.DataCentres)

			patch := pg.NewPatch()
			pg.Status.ClusterStatus = iPg.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL cluster status",
					"cluster name", pg.Spec.Name,
					"cluster ID", pg.Status.ID,
					"instaclustr status", iPg.Status,
				)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iPg.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					pg.Name,
					pg.Namespace,
					nodes,
					models.PgConnectionPort)
				if err != nil {
					return err
				}
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

		if iPg.Status.CurrentClusterOperationStatus == models.NoOperation &&
			pg.Annotations[models.UpdateQueuedAnnotation] != models.True &&
			!pg.Spec.IsEqual(iPg.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iPg.Spec, "k8s resource spec", pg.Spec)

			patch := pg.NewPatch()
			pg.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", pg.Spec.Name, "cluster state", pg.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(pg.Spec, iPg.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iPg.Spec, "k8s resource spec", pg.Spec)
				return err
			}
			r.EventRecorder.Eventf(pg, models.Warning, models.ExternalChanges, msgDiffSpecs)
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

func (r *PostgreSQLReconciler) createDefaultPassword(ctx context.Context, pg *v1beta1.PostgreSQL, l logr.Logger) error {
	iData, err := r.API.GetPostgreSQL(pg.Status.ID)
	if err != nil {
		l.Error(
			err, "Cannot get PostgreSQL cluster status from the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return err
	}

	defaultSecretExists, err := r.DefaultSecretExists(ctx, pg)
	if err != nil {
		return err
	}

	if defaultSecretExists {
		return nil
	}

	defaultUserPassword, err := pg.DefaultPasswordFromInstAPI(iData)
	if err != nil {
		l.Error(err, "Cannot get default user creds for PostgreSQL cluster from the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"clusterID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Default user password fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)

		return err
	}

	secret := pg.NewUserSecret(defaultUserPassword)
	err = r.Client.Create(context.TODO(), secret)
	if err != nil {
		l.Error(err, "Cannot create PostgreSQL default user secret",
			"cluster ID", pg.Status.ID,
			"secret name", secret.Name,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.CreationFailed,
			"Default user secret creation is failed. Reason: %v",
			err,
		)

		return err
	}

	l.Info("PostgreSQL default user secret was created",
		"secret name", secret.Name,
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Created,
		"Default user secret is created. Secret name: %s",
		secret.Name,
	)

	return nil
}

func (r *PostgreSQLReconciler) DefaultSecretExists(ctx context.Context, pg *v1beta1.PostgreSQL) (bool, error) {
	_, err := v1beta1.GetDefaultPgUserSecret(ctx, pg.Name, pg.Namespace, r.Client)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *PostgreSQLReconciler) newWatchBackupsJob(pg *v1beta1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgreSQLBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: pg.Namespace, Name: pg.Name}, pg)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, pg.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get PostgreSQL cluster backups",
				"cluster name", pg.Spec.Name,
				"cluster ID", pg.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.PgClusterKind)

		k8sBackupList, err := r.listClusterBackups(ctx, pg.Status.ID, pg.Namespace)
		if err != nil {
			return err
		}

		k8sBackups := map[int]*clusterresourcesv1beta1.ClusterBackup{}
		unassignedBackups := []*clusterresourcesv1beta1.ClusterBackup{}
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

			backupSpec := pg.NewBackupSpec(start)
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

func (r *PostgreSQLReconciler) newUsersCreationJob(c *v1beta1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("component", "postgresqlUsersCreationJob")

	return func() error {
		ctx := context.Background()

		err := r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      c.Name,
		}, c)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if c.Status.State != models.RunningStatus {
			l.Info("User creation job is scheduled")
			r.EventRecorder.Event(c, models.Normal, models.CreationFailed,
				"User creation job is scheduled, cluster is not in the running state")
			return nil
		}

		for _, ref := range c.Spec.UserRefs {
			err = r.createUser(ctx, l, c, ref)
			if err != nil {
				l.Error(err, "Failed to create a user for the cluster", "user ref", ref)
				r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
					"Failed to create a user for the cluster. Reason: %v", err)
				return err
			}
		}

		l.Info("User creation job successfully finished", "resource name", c.Name)
		r.EventRecorder.Eventf(c, models.Normal, models.Created, "User creation job successfully finished")

		go r.Scheduler.RemoveJob(c.GetJobID(scheduler.UserCreator))

		return nil
	}
}

func (r *PostgreSQLReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1beta1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1beta1.ClusterBackupList{}
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

	backupType := &clusterresourcesv1beta1.ClusterBackup{}
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

func (r *PostgreSQLReconciler) deleteSecret(ctx context.Context, pg *v1beta1.PostgreSQL) error {
	secret, err := v1beta1.GetDefaultPgUserSecret(ctx, pg.Name, pg.Namespace, r.Client)
	if err != nil {
		return err
	}

	err = r.Client.Delete(ctx, secret)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) updateDataCentres(cluster *v1beta1.PostgreSQL) error {
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
	pgCluster *v1beta1.PostgreSQL,
	logger logr.Logger,
) error {
	patchRequest := []*v1beta1.PatchRequest{}

	annotationsPayload, err := json.Marshal(pgCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &v1beta1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(pgCluster.Finalizers)
	if err != nil {
		return err
	}

	finzlizersPatch := &v1beta1.PatchRequest{
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

func (r *PostgreSQLReconciler) updateDescriptionAndTwoFactorDelete(pgCluster *v1beta1.PostgreSQL) error {
	var twoFactorDelete *v1beta1.TwoFactorDelete
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
	s := secret.(*k8sCore.Secret)

	if s.Labels[models.DefaultSecretLabel] != "true" {
		return []reconcile.Request{}
	}

	pg := &v1beta1.PostgreSQL{}
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
	pg.Annotations[models.ResourceStateAnnotation] = models.SecretEvent
	err = r.Patch(context.TODO(), pg, patch)
	if err != nil {
		return []reconcile.Request{}
	}

	return []reconcile.Request{{NamespacedName: pgNamespacedName}}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.PostgreSQL{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				newObj := event.ObjectNew.(*v1beta1.PostgreSQL)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				oldObj := event.ObjectOld.(*v1beta1.PostgreSQL)

				r.handleUserEvent(newObj, oldObj.Spec.UserRefs)

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Owns(&clusterresourcesv1beta1.ClusterBackup{}).
		Owns(&k8sCore.Secret{}).
		Watches(
			&source.Kind{Type: &k8sCore.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findSecretObject),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(createEvent event.CreateEvent) bool {
					return false
				},
			}),
		).
		Complete(r)
}
