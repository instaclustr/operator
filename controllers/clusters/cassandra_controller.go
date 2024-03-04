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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/controllers/clusterresources"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch;create;update;patch;delete
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	c := &v1beta1.Cassandra{}
	err := r.Client.Get(ctx, req.NamespacedName, c)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Cassandra resource is not found",
				"request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch Cassandra cluster",
			"request", req)
		return reconcile.Result{}, err
	}

	switch c.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, l, c)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, l, req, c)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, l, c)
	case models.GenericEvent:
		l.Info("Event isn't handled",
			"cluster name", c.Spec.Name,
			"request", req,
			"event", c.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	default:
		l.Info("Event isn't handled",
			"request", req,
			"event", c.Annotations[models.ResourceStateAnnotation])

		return models.ExitReconcile, nil
	}
}

func (r *CassandraReconciler) createCassandraFromRestore(c *v1beta1.Cassandra, l logr.Logger) (*models.CassandraCluster, error) {
	l.Info(
		"Creating cluster from backup",
		"original cluster ID", c.Spec.RestoreFrom.ClusterID,
	)

	id, err := r.API.RestoreCluster(c.RestoreInfoToInstAPI(c.Spec.RestoreFrom), models.CassandraAppKind)
	if err != nil {
		l.Error(err, "Cannot restore cluster from backup",
			"original cluster ID", c.Spec.RestoreFrom.ClusterID,
		)

		r.EventRecorder.Eventf(
			c, models.Warning, models.CreationFailed,
			"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
			err,
		)

		return nil, err
	}

	instaModel, err := r.API.GetCassandra(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get cassandra cluster details, err: %w", err)
	}

	r.EventRecorder.Eventf(
		c, models.Normal, models.Created,
		"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
		c.Spec.RestoreFrom.ClusterID,
		instaModel.ID,
	)

	return instaModel, nil
}

func (r *CassandraReconciler) mergeDebezium(c *v1beta1.Cassandra, spec *models.CassandraCluster) error {
	for i, dc := range c.Spec.DataCentres {
		for j, debezium := range dc.Debezium {
			if debezium.ClusterRef != nil {
				cdcID, err := clusterresources.GetDataCentreID(r.Client, context.Background(), debezium.ClusterRef)
				if err != nil {
					return fmt.Errorf("failed to get kafka data centre id, err: %w", err)
				}
				spec.DataCentres[i].Debezium[j].KafkaDataCentreID = cdcID
			}
		}
	}

	return nil
}

func (r *CassandraReconciler) createCassandra(c *v1beta1.Cassandra, l logr.Logger) (*models.CassandraCluster, error) {
	l.Info(
		"Creating cluster",
		"cluster name", c.Spec.Name,
		"data centres", c.Spec.DataCentres,
	)

	iCassandraSpec := c.Spec.ToInstAPI()

	err := r.mergeDebezium(c, iCassandraSpec)
	if err != nil {
		l.Error(err, "Cannot get debezium dependencies for Cassandra cluster")
		return nil, err
	}

	b, err := r.API.CreateClusterRaw(instaclustr.CassandraEndpoint, iCassandraSpec)
	if err != nil {
		l.Error(
			err, "Cannot create cluster",
			"cluster spec", c.Spec,
		)
		r.EventRecorder.Eventf(
			c, models.Warning, models.CreationFailed,
			"Cluster creation on the Instaclustr is failed. Reason: %v",
			err,
		)
		return nil, err
	}

	var instModel models.CassandraCluster
	err = json.Unmarshal(b, &instModel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to model, err: %w", err)
	}

	r.EventRecorder.Eventf(
		c, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s",
		instModel.ID,
	)

	return &instModel, nil
}

func (r *CassandraReconciler) createCluster(ctx context.Context, c *v1beta1.Cassandra, l logr.Logger) error {
	var instModel *models.CassandraCluster
	var err error

	if c.Spec.HasRestore() {
		instModel, err = r.createCassandraFromRestore(c, l)
	} else {
		instModel, err = r.createCassandra(c, l)
	}
	if err != nil {
		return fmt.Errorf("failed to create cluster, err: %w", err)
	}

	c.Spec.FromInstAPI(instModel)
	c.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Update(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to update cassandra spec, err: %w", err)
	}

	c.Status.FromInstAPI(instModel)
	err = r.Status().Update(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to update cassandra status, err: %w", err)
	}

	err = r.createDefaultSecret(ctx, c, l)
	if err != nil {
		return fmt.Errorf("failed to create default cassandra user secret, err: %w", err)
	}

	l.Info(
		"Cluster has been created",
		"cluster name", c.Spec.Name,
		"cluster ID", c.Status.ID,
	)

	return nil
}

func (r *CassandraReconciler) startClusterJobs(c *v1beta1.Cassandra, l logr.Logger) error {
	err := r.startSyncJob(c)
	if err != nil {
		l.Error(err, "Cannot start cluster synchronizer",
			"c cluster ID", c.Status.ID)

		r.EventRecorder.Eventf(
			c, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v",
			err,
		)
		return err
	}

	r.EventRecorder.Eventf(
		c, models.Normal, models.Created,
		"Cluster sync job is started",
	)

	err = r.startClusterBackupsJob(c)
	if err != nil {
		l.Error(err, "Cannot start cluster backups check job",
			"cluster ID", c.Status.ID,
		)

		r.EventRecorder.Eventf(
			c, models.Warning, models.CreationFailed,
			"Cluster backups check job is failed. Reason: %v",
			err,
		)
		return err
	}

	r.EventRecorder.Eventf(
		c, models.Normal, models.Created,
		"Cluster backups check job is started",
	)

	if c.Spec.UserRefs != nil && c.Status.AvailableUsers == nil {
		err = r.startUsersCreationJob(c)
		if err != nil {
			l.Error(err, "Failed to start user creation job")
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"User creation job is failed. Reason: %v", err)
			return err
		}

		r.EventRecorder.Event(c, models.Normal, models.Created,
			"Cluster user creation job is started")
	}

	return nil
}

func (r *CassandraReconciler) handleCreateCluster(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
) (reconcile.Result, error) {
	l = l.WithName("Cassandra creation event")
	if c.Status.ID == "" {
		err := r.createCluster(ctx, c, l)
		if err != nil {
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed, "Failed to create cluster. Reason: %v", err)
			return reconcile.Result{}, err
		}
	}

	if c.Status.State != models.DeletedStatus {
		patch := c.NewPatch()
		c.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(c, models.DeletionFinalizer)
		err := r.Patch(ctx, c, patch)
		if err != nil {
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"Failed to update resource metadata. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}

		err = r.startClusterJobs(c, l)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to start cluster jobs, err: %w", err)
		}

	}

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleUpdateCluster(
	ctx context.Context,
	l logr.Logger,
	req ctrl.Request,
	c *v1beta1.Cassandra,
) (reconcile.Result, error) {
	l = l.WithName("Cassandra update event")

	instaModel, err := r.API.GetCassandra(c.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get cluster from the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID,
		)

		r.EventRecorder.Eventf(
			c, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	iCassandra := v1beta1.Cassandra{}
	iCassandra.FromInstAPI(instaModel)

	if c.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.CassandraSpec](r.EventRecorder, r.Client, c, &iCassandra, l)
	}

	patch := c.NewPatch()

	if c.Spec.ClusterSettingsNeedUpdate(&iCassandra.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iCassandra.Spec.Description,
			"instaclustr two factor delete", iCassandra.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(c.Status.ID, c.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", c.Status.ID, "cluster spec", c.Spec)
			r.EventRecorder.Eventf(c, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !c.Spec.AreDCsEqual(iCassandra.Spec.DataCentres) {
		l.Info("Update request to Instaclustr API has been sent",
			"spec data centres", c.Spec.DataCentres,
			"resize settings", c.Spec.ResizeSettings,
		)

		err = r.API.UpdateCassandra(c.Status.ID, c.Spec.DCsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster",
				"cluster ID", c.Status.ID,
				"cluster spec", c.Spec,
				"cluster state", c.Status.State)

			r.EventRecorder.Eventf(c, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	err = handleUsersChanges(ctx, r.Client, r, c)
	if err != nil {
		l.Error(err, "Failed to handle users changes")
		r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
			"Handling users changes is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	c.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, c, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster ID", c.Status.ID)

		r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", c.Spec.Name,
		"cluster ID", c.Status.ID,
		"data centres", c.Spec.DataCentres,
	)

	r.EventRecorder.Event(c, models.Normal, models.UpdatedEvent, "Cluster has been updated")

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleDeleteCluster(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
) (reconcile.Result, error) {
	l = l.WithName("Cassandra deletion event")

	_, err := r.API.GetCassandra(c.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get cluster from the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID,
			"kind", c.Kind,
			"api Version", c.APIVersion,
			"namespace", c.Namespace,
		)
		r.EventRecorder.Eventf(
			c, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	patch := c.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID)

		err = r.API.DeleteCluster(c.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete cluster",
				"cluster name", c.Spec.Name,
				"state", c.Status.State,
				"kind", c.Kind,
				"api Version", c.APIVersion,
				"namespace", c.Namespace,
			)
			r.EventRecorder.Eventf(
				c, models.Warning, models.DeletionFailed,
				"Cluster deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(c, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if c.Spec.TwoFactorDelete != nil {
			c.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			c.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, c, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", c.Spec.Name,
					"cluster state", c.Status.State)
				r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v", err)

				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", c.Status.ID)

			r.EventRecorder.Event(c, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))

	l.Info("Deleting cluster backup resources", "cluster ID", c.Status.ID)

	err = r.deleteBackups(ctx, c.Status.ID, c.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", c.Status.ID,
		)
		r.EventRecorder.Eventf(
			c, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("Cluster backup resources were deleted",
		"cluster ID", c.Status.ID,
	)

	r.EventRecorder.Eventf(
		c, models.Normal, models.Deleted,
		"Cluster backup resources are deleted",
	)

	err = detachUsers(ctx, r.Client, r, c)
	if err != nil {
		l.Error(err, "Failed to detach users from the cluster")
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Detaching users from the cluster is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	controllerutil.RemoveFinalizer(c, models.DeletionFinalizer)
	c.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, c, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID,
			"kind", c.Kind,
			"api Version", c.APIVersion,
			"namespace", c.Namespace,
			"cluster metadata", c.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			c, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, c.Name, c.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Cassandra cluster expose service",
			"cluster ID", c.Status.ID,
			"cluster name", c.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Cluster has been deleted",
		"cluster name", c.Spec.Name,
		"cluster ID", c.Status.ID,
		"kind", c.Kind,
		"api Version", c.APIVersion)

	r.EventRecorder.Eventf(
		c, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) startSyncJob(c *v1beta1.Cassandra) error {
	job := r.newSyncJob(c)

	err := r.Scheduler.ScheduleJob(c.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) startClusterBackupsJob(cluster *v1beta1.Cassandra) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) startUsersCreationJob(cluster *v1beta1.Cassandra) error {
	job := r.newUsersCreationJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.UserCreator), scheduler.UserCreationInterval, job)
	if err != nil {
		return err
	}

	return nil
}

//nolint:unused,deadcode
func (r *CassandraReconciler) startClusterOnPremisesIPsJob(c *v1beta1.Cassandra, b *onPremisesBootstrap) error {
	job := newWatchOnPremisesIPsJob(c.Kind, b)

	err := r.Scheduler.ScheduleJob(c.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) newSyncJob(c *v1beta1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("syncJob", c.GetJobID(scheduler.SyncJob), "clusterID", c.Status.ID)

	return func() error {
		namespacedName := client.ObjectKeyFromObject(c)
		err := r.Get(context.Background(), namespacedName, c)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(c.GetJobID(scheduler.UserCreator))
			r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))
			return nil
		}

		instaModel, err := r.API.GetCassandra(c.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if c.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), l, c)
					return err
				}

				return r.handleExternalDelete(context.Background(), c)
			}

			l.Error(err, "Cannot get cluster from the Instaclustr API",
				"clusterID", c.Status.ID)
			return err
		}

		iCassandra := v1beta1.Cassandra{}
		iCassandra.FromInstAPI(instaModel)

		if !c.Status.Equals(&iCassandra.Status) {
			l.Info("Updating cluster status")

			dcEqual := c.Status.DataCentresEqual(&iCassandra.Status)

			patch := c.NewPatch()
			c.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), c, patch)
			if err != nil {
				return err
			}

			if !dcEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iCassandra.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					c.Name,
					c.Namespace,
					c.Spec.PrivateNetwork,
					nodes,
					models.CassandraConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := c.Spec.IsEqual(&iCassandra.Spec)

		if equals && c.Annotations[models.ExternalChangesAnnotation] == models.True {
			err = reconcileExternalChanges(r.Client, r.EventRecorder, c)
			if err != nil {
				return err
			}
		} else if c.Status.CurrentClusterOperationStatus == models.NoOperation &&
			c.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

			patch := c.NewPatch()
			c.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), c, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", c.Spec.Name, "cluster state", c.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(c.Spec, iCassandra.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iCassandra.Spec, "k8s resource spec", c.Spec)
				return err
			}

			r.EventRecorder.Eventf(c, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), c)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)
			return err
		}

		if c.Status.State == models.RunningStatus && c.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := c.NewPatch()
			for _, dc := range c.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", c.Spec.Name,
						"cluster ID", c.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), c, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", c.Spec.Name,
						"cluster ID", c.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *CassandraReconciler) newWatchBackupsJob(c *v1beta1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: c.Namespace, Name: c.Name}, c)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		iBackups, err := r.API.GetClusterBackups(c.Status.ID, models.ClusterKindsMap[c.Kind])
		if err != nil {
			l.Error(err, "Cannot get cluster backups",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)

			return err
		}

		iBackupEvents := iBackups.GetBackupEvents(models.CassandraKind)

		k8sBackupList, err := r.listClusterBackups(ctx, c.Status.ID, c.Namespace)
		if err != nil {
			l.Error(err, "Cannot list cluster backups",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)

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

		for start, iBackup := range iBackupEvents {
			if _, exists := k8sBackups[start]; exists {
				if k8sBackups[start].Status.End != 0 {
					continue
				}

				patch := k8sBackups[start].NewPatch()
				k8sBackups[start].Status.UpdateStatus(iBackup)
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
				backupToAssign.Status.Start = iBackup.Start
				backupToAssign.Status.UpdateStatus(iBackup)
				err = r.Status().Patch(context.TODO(), backupToAssign, patch)
				if err != nil {
					return err
				}
				continue
			}

			backupSpec := c.NewBackupSpec(start)
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

func (r *CassandraReconciler) createDefaultSecret(ctx context.Context, c *v1beta1.Cassandra, l logr.Logger) error {
	username, password, err := r.API.GetDefaultCredentialsV1(c.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get default user creds for Cassandra cluster from the Instaclustr API",
			"cluster ID", c.Status.ID,
		)
		r.EventRecorder.Eventf(c, models.Warning, models.FetchFailed,
			"Default user password fetch from the Instaclustr API is failed. Reason: %v", err,
		)

		return err
	}

	patch := c.NewPatch()
	secret := newDefaultUserSecret(username, password, c.Name, c.Namespace)
	err = controllerutil.SetOwnerReference(c, secret, r.Scheme)
	if err != nil {
		l.Error(err, "Cannot set secret owner reference with default user credentials",
			"cluster ID", c.Status.ID,
		)
		r.EventRecorder.Eventf(c, models.Warning, models.SetOwnerRef,
			"Setting secret owner ref with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	err = r.Create(ctx, secret)
	if err != nil {
		l.Error(err, "Cannot create secret with default user credentials",
			"cluster ID", c.Status.ID,
		)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Creating secret with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	l.Info("Default secret was created",
		"secret name", secret.Name,
		"secret namespace", secret.Namespace,
	)

	c.Status.DefaultUserSecretRef = &v1beta1.Reference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	err = r.Status().Patch(ctx, c, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cassandra resource",
			"cluster name", c.Spec.Name,
			"status", c.Status)

		r.EventRecorder.Eventf(
			c, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return err
	}

	return nil
}

func (r *CassandraReconciler) newUsersCreationJob(c *v1beta1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraUsersCreationJob")

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

		err = handleUsersChanges(ctx, r.Client, r, c)
		if err != nil {
			l.Error(err, "Failed to create users for the cluster")
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"Failed to create users for the cluster. Reason: %v", err)
			return err
		}

		l.Info("User creation job successfully finished", "resource name", c.Name)
		r.EventRecorder.Eventf(c, models.Normal, models.Created, "User creation job successfully finished")

		r.Scheduler.RemoveJob(c.GetJobID(scheduler.UserCreator))

		return nil
	}
}

func (r *CassandraReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1beta1.ClusterBackupList, error) {
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

func (r *CassandraReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
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

func (r *CassandraReconciler) reconcileMaintenanceEvents(ctx context.Context, c *v1beta1.Cassandra) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(c.Status.ID)
	if err != nil {
		return err
	}

	if !c.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := c.NewPatch()
		c.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, c, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", c.Status.ID,
			"events", c.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *CassandraReconciler) handleExternalDelete(ctx context.Context, c *v1beta1.Cassandra) error {
	l := log.FromContext(ctx)

	patch := c.NewPatch()
	c.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, c, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(c, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))

	return nil
}

// NewUserResource implements userResourceFactory interface
func (r *CassandraReconciler) NewUserResource() userObject {
	return &clusterresourcesv1beta1.CassandraUser{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{RateLimiter: r.RateLimiter}).
		For(&v1beta1.Cassandra{}, builder.WithPredicates(predicate.Funcs{
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

				newObj := event.ObjectNew.(*v1beta1.Cassandra)

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

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).Complete(r)
}
