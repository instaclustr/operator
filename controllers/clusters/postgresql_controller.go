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
	"fmt"
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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=postgresqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=get;list;create;update;patch;deletecollection;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;watch;list
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	pg := &v1beta1.PostgreSQL{}
	err := r.Client.Get(ctx, req.NamespacedName, pg)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("PostgreSQL custom resource is not found",
				"resource name", req.NamespacedName,
			)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch PostgreSQL cluster",
			"resource name", req.NamespacedName,
		)

		return models.ExitReconcile, err
	}

	switch pg.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, pg, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, pg, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, pg, l)
	case models.SecretEvent:
		return r.handleUpdateDefaultUserPassword(ctx, pg, l)
	case models.GenericEvent:
		l.Info("PostgreSQL resource generic event isn't handled",
			"cluster name", pg.Spec.Name,
			"request", req,
			"event", pg.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	default:
		l.Info("PostgreSQL resource event isn't handled",
			"cluster name", pg.Spec.Name,
			"request", req,
			"event", pg.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *PostgreSQLReconciler) createFromRestore(pg *v1beta1.PostgreSQL, l logr.Logger) (*models.PGCluster, error) {
	l.Info(
		"Creating PostgreSQL cluster from backup",
		"original cluster ID", pg.Spec.PgRestoreFrom.ClusterID,
	)

	id, err := r.API.RestoreCluster(pg.RestoreInfoToInstAPI(pg.Spec.PgRestoreFrom), models.PgRestoreValue)
	if err != nil {
		return nil, fmt.Errorf("failed to restore cluster, err: %v", err)
	}

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Created,
		"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
		pg.Spec.PgRestoreFrom.ClusterID,
		pg.Status.ID,
	)

	instaModel, err := r.API.GetPostgreSQL(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster details, err: %c", err)
	}

	return instaModel, nil
}

func (r *PostgreSQLReconciler) createPostgreSQL(pg *v1beta1.PostgreSQL, l logr.Logger) (*models.PGCluster, error) {
	l.Info(
		"Creating PostgreSQL cluster",
		"cluster name", pg.Spec.Name,
		"data centres", pg.Spec.DataCentres,
	)

	b, err := r.API.CreateClusterRaw(instaclustr.PGSQLEndpoint, pg.Spec.ToInstAPI())
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster, err: %v", err)
	}

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s",
		pg.Status.ID,
	)

	var instaModel models.PGCluster
	err = json.Unmarshal(b, &instaModel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body to models.PGCluster, err: %v", err)
	}

	return &instaModel, nil
}

func (r *PostgreSQLReconciler) createCluster(ctx context.Context, pg *v1beta1.PostgreSQL, l logr.Logger) error {
	var instaModel *models.PGCluster
	var err error

	if pg.Spec.HasRestore() {
		instaModel, err = r.createFromRestore(pg, l)
	} else {
		instaModel, err = r.createPostgreSQL(pg, l)
	}
	if err != nil {
		return err
	}

	patch := pg.NewPatch()

	pg.Spec.FromInstAPI(instaModel)
	pg.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Patch(ctx, pg, patch)
	if err != nil {
		return fmt.Errorf("failed to patch cluster spec, err: %v", err)
	}

	pg.Status.FromInstAPI(instaModel)
	err = r.Status().Patch(ctx, pg, patch)
	if err != nil {
		return fmt.Errorf("failed to patch cluster status, err: %v", err)
	}

	err = r.createDefaultPassword(ctx, pg, l)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLReconciler) handleCreateCluster(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("PostgreSQL creation event")

	if pg.Status.ID == "" {
		err := r.createCluster(ctx, pg, l)
		if err != nil {
			r.EventRecorder.Eventf(pg, models.Warning, models.CreationFailed,
				"Failed to create PostgreSQL cluster. Reason: %w", err,
			)
			return reconcile.Result{}, err
		}
	}

	if pg.Status.State != models.DeletedStatus {
		patch := pg.NewPatch()
		controllerutil.AddFinalizer(pg, models.DeletionFinalizer)
		pg.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err := r.Patch(ctx, pg, patch)
		if err != nil {
			l.Error(err, "Cannot patch PostgreSQL resource",
				"cluster name", pg.Spec.Name,
				"status", pg.Status)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		err = r.startClusterStatusJob(pg)
		if err != nil {
			l.Error(err, "Cannot start PostgreSQL cluster status check job",
				"cluster ID", pg.Status.ID,
			)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.CreationFailed,
				"Cluster status check job is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			pg, models.Normal, models.Created,
			"Cluster status check job is started",
		)

		if pg.Spec.DataCentres[0].CloudProvider == models.ONPREMISES {
			return models.ExitReconcile, nil
		}

		err = r.startClusterBackupsJob(pg)
		if err != nil {
			l.Error(err, "Cannot start PostgreSQL cluster backups check job",
				"cluster ID", pg.Status.ID,
			)

			r.EventRecorder.Eventf(
				pg, models.Warning, models.CreationFailed,
				"Cluster backups check job is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			pg, models.Normal, models.Created,
			"Cluster backups check job is started",
		)
	}

	return models.ExitReconcile, nil
}

func (r *PostgreSQLReconciler) handleUpdateCluster(ctx context.Context, pg *v1beta1.PostgreSQL, req ctrl.Request, l logr.Logger) (reconcile.Result, error) {
	l = l.WithName("PostgreSQL update event")

	instaModel, err := r.API.GetPostgreSQL(pg.Status.ID)
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
		return reconcile.Result{}, err
	}

	iPg := &v1beta1.PostgreSQL{}
	iPg.FromInstAPI(instaModel)

	if pg.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.PgSpec](r.EventRecorder, r.Client, pg, iPg, l)
	}

	if pg.Spec.ClusterSettingsNeedUpdate(&iPg.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iPg.Spec.Description,
			"instaclustr two factor delete", iPg.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(pg.Status.ID, pg.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", pg.Status.ID, "cluster spec", pg.Spec)
			r.EventRecorder.Eventf(pg, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !pg.Spec.DCsEqual(iPg.Spec.DataCentres) {
		l.Info("Update request to Instaclustr API has been sent",
			"spec data centres", pg.Spec.DataCentres,
			"resize settings", pg.Spec.ResizeSettings,
		)

		err = r.updateCluster(pg)
		if err != nil {
			l.Error(err, "Cannot update Data Centres",
				"cluster DC", pg.Spec.DataCentres)

			r.EventRecorder.Eventf(pg, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		l.Info(
			"Cluster has been updated",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
			"data centres", pg.Spec.DataCentres,
		)
	}

	iConfigs, err := r.API.GetPostgreSQLConfigs(pg.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get PostgreSQL cluster configs",
			"cluster name", pg.Spec.Name,
			"clusterID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster configs fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	for _, iConfig := range iConfigs {
		err = r.reconcileClusterConfigurations(
			pg.Status.ID,
			pg.Spec.ClusterConfigurations,
			iConfig.ConfigurationProperties)
		if err != nil {
			l.Error(err, "Cannot reconcile PostgreSQL cluster configs",
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
			return reconcile.Result{}, err
		}

		l.Info("PostgreSQL cluster configurations were updated",
			"cluster name", pg.Spec.Name,
		)
	}

	pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, pg, l)
	if err != nil {
		l.Error(err, "Cannot patch PostgreSQL resource metadata",
			"cluster name", pg.Spec.Name,
			"cluster metadata", pg.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("PostgreSQL cluster was updated",
		"cluster name", pg.Spec.Name,
		"cluster status", pg.Status.State,
	)

	return models.ExitReconcile, nil
}

func (r *PostgreSQLReconciler) handleDeleteCluster(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("PostgreSQL deletion event")

	_, err := r.API.GetPostgreSQL(pg.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get PostgreSQL cluster status",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID)

		err = r.API.DeleteCluster(pg.Status.ID, instaclustr.PGSQLEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete PostgreSQL cluster",
				"cluster name", pg.Spec.Name,
				"cluster status", pg.Status.State,
			)
			r.EventRecorder.Eventf(
				pg, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v",
				err,
			)

			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(pg, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if pg.Spec.TwoFactorDelete != nil {
			patch := pg.NewPatch()

			pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			pg.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, pg, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", pg.Spec.Name,
					"cluster state", pg.Status.State)
				r.EventRecorder.Eventf(pg, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", pg.Status.ID)

			r.EventRecorder.Event(pg, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	l.Info("PostgreSQL cluster is being deleted. Deleting PostgreSQL default user secret",
		"cluster ID", pg.Status.ID,
	)

	l.Info("Deleting cluster backup resources",
		"cluster ID", pg.Status.ID,
	)

	err = r.deleteBackups(ctx, pg.Status.ID, pg.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete PostgreSQL backup resources",
			"cluster ID", pg.Status.ID,
		)
		r.EventRecorder.Eventf(
			pg, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("Cluster backup resources were deleted",
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
	err = r.patchClusterMetadata(ctx, pg, l)
	if err != nil {
		l.Error(
			err, "Cannot patch PostgreSQL resource metadata after finalizer removal",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = r.deleteSecret(ctx, pg)
	if client.IgnoreNotFound(err) != nil {
		l.Error(err, "Cannot delete PostgreSQL default user secret",
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.DeletionFailed,
			"Default user secret deletion is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("Cluster PostgreSQL default user secret was deleted",
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Deleted,
		"Default user secret is deleted. Cluster ID: %s",
		pg.Status.ID,
	)

	err = exposeservice.Delete(r.Client, pg.Name, pg.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete PostgreSQL cluster expose service",
			"cluster ID", pg.Status.ID,
			"cluster name", pg.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("PostgreSQL cluster was deleted",
		"cluster name", pg.Spec.Name,
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *PostgreSQLReconciler) handleUpdateDefaultUserPassword(
	ctx context.Context,
	pg *v1beta1.PostgreSQL,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("PostgreSQL default user password updating event")

	secret, err := v1beta1.GetDefaultPgUserSecret(ctx, pg.Name, pg.Namespace, r.Client)
	if err != nil {
		l.Error(err, "Cannot get the default secret for the PostgreSQL cluster",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.FetchFailed,
			"Fetch default user secret is failed. Reason: %v",
			err,
		)

		return reconcile.Result{}, err
	}

	password := string(secret.Data[models.Password])
	isValid := pg.ValidateDefaultUserPassword(password)
	if !isValid {
		l.Error(err, "Default PostgreSQL user password is not valid. This field must be at least 8 characters long. Must contain characters from at least 3 of the following 4 categories: Uppercase, Lowercase, Numbers, Special Characters",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.ValidationFailed,
			"Validation for default user secret is failed. Reason: %v",
			err,
		)

		return reconcile.Result{}, err
	}

	err = r.API.UpdatePostgreSQLDefaultUserPassword(pg.Status.ID, password)
	if err != nil {
		l.Error(err, "Cannot update default PostgreSQL user password",
			"cluster name", pg.Spec.Name,
			"cluster ID", pg.Status.ID,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.UpdateFailed,
			"Default user password update on the Instaclustr API is failed. Reason: %v",
			err,
		)

		return reconcile.Result{}, err
	}

	pg.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.patchClusterMetadata(ctx, pg, l)
	if err != nil {
		l.Error(err, "Cannot patch PostgreSQL resource metadata",
			"cluster name", pg.Spec.Name,
			"cluster metadata", pg.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("PostgreSQL default user password was updated",
		"cluster name", pg.Spec.Name,
		"cluster ID", pg.Status.ID,
	)

	r.EventRecorder.Eventf(
		pg, models.Normal, models.UpdatedEvent,
		"Cluster default user password is updated",
	)

	return models.ExitReconcile, nil
}

//nolint:unused,deadcode
func (r *PostgreSQLReconciler) startClusterOnPremisesIPsJob(pg *v1beta1.PostgreSQL, b *onPremisesBootstrap) error {
	job := newWatchOnPremisesIPsJob(pg.Kind, b)

	err := r.Scheduler.ScheduleJob(pg.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
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

func (r *PostgreSQLReconciler) newWatchStatusJob(pg *v1beta1.PostgreSQL) scheduler.Job {
	l := log.Log.WithValues("syncJob", pg.GetJobID(scheduler.StatusChecker), "clusterID", pg.Status.ID)

	return func() error {
		namespacedName := client.ObjectKeyFromObject(pg)
		err := r.Get(context.Background(), namespacedName, pg)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(pg.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(pg.GetJobID(scheduler.StatusChecker))
			r.Scheduler.RemoveJob(pg.GetJobID(scheduler.UserCreator))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get PosgtreSQL custom resource",
				"resource name", pg.Name,
			)
			return err
		}

		instaModel, err := r.API.GetPostgreSQL(pg.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if pg.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), pg, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), pg)
			}

			l.Error(err, "Cannot get PostgreSQL cluster status",
				"cluster name", pg.Spec.Name,
				"clusterID", pg.Status.ID,
			)

			return err
		}

		iPg := &v1beta1.PostgreSQL{}
		iPg.FromInstAPI(instaModel)

		if !pg.Status.Equals(&iPg.Status) {
			l.Info("Updating PostgreSQL cluster status")

			areDCsEqual := pg.Status.DCsEqual(iPg.Status.DataCentres)

			patch := pg.NewPatch()
			pg.Status.FromInstAPI(instaModel)
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

				for _, dc := range iPg.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					pg.Name,
					pg.Namespace,
					pg.Spec.PrivateNetwork,
					nodes,
					models.PgConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := pg.Spec.IsEqual(iPg.Spec)

		if equals && pg.Annotations[models.ExternalChangesAnnotation] == models.True {
			patch := pg.NewPatch()
			delete(pg.Annotations, models.ExternalChangesAnnotation)
			err := r.Patch(context.Background(), pg, patch)
			if err != nil {
				return err
			}

			r.EventRecorder.Event(pg, models.Normal, models.ExternalChanges,
				"External changes were automatically reconciled",
			)
		} else if pg.Status.CurrentClusterOperationStatus == models.NoOperation &&
			pg.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

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

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), pg)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", pg.Spec.Name,
				"cluster ID", pg.Status.ID,
			)
			return err
		}

		if pg.Status.State == models.RunningStatus && pg.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := pg.NewPatch()
			for _, dc := range pg.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", pg.Spec.Name,
						"cluster ID", pg.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), pg, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", pg.Spec.Name,
						"cluster ID", pg.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *PostgreSQLReconciler) createDefaultPassword(ctx context.Context, pg *v1beta1.PostgreSQL, l logr.Logger) error {
	patch := pg.NewPatch()
	instaModel, err := r.API.GetPostgreSQL(pg.Status.ID)
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

	secret := pg.NewUserSecret(instaModel.DefaultUserPassword)
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

	pg.Status.DefaultUserSecretRef = &v1beta1.Reference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	err = r.Status().Patch(ctx, pg, patch)
	if err != nil {
		l.Error(err, "Cannot patch PostgreSQL resource",
			"cluster name", pg.Spec.Name,
			"status", pg.Status)

		r.EventRecorder.Eventf(
			pg, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return err
	}

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

		instBackups, err := r.API.GetClusterBackups(pg.Status.ID, models.PgAppKind)
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

func (r *PostgreSQLReconciler) updateCluster(cluster *v1beta1.PostgreSQL) error {
	err := r.API.UpdatePostgreSQL(cluster.Status.ID, cluster.Spec.ToClusterUpdate())
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
	l logr.Logger,
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

	l.Info("PostgreSQL cluster patched",
		"Cluster name", pgCluster.Spec.Name,
		"Finalizers", pgCluster.Finalizers,
		"Annotations", pgCluster.Annotations,
	)
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
		WithOptions(controller.Options{
			RateLimiter: r.RateLimiter}).
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

				if newObj.Status.ID == "" && newObj.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
					return false
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

func (r *PostgreSQLReconciler) reconcileMaintenanceEvents(ctx context.Context, pg *v1beta1.PostgreSQL) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(pg.Status.ID)
	if err != nil {
		return err
	}

	if !pg.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := pg.NewPatch()
		pg.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, pg, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", pg.Status.ID,
			"events", pg.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *PostgreSQLReconciler) handleExternalDelete(ctx context.Context, pg *v1beta1.PostgreSQL) error {
	l := log.FromContext(ctx)

	patch := pg.NewPatch()
	pg.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, pg, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(pg, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(pg.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(pg.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(pg.GetJobID(scheduler.StatusChecker))

	return nil
}
