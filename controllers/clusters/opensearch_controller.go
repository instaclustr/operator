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
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
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
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// OpenSearchReconciler reconciles a OpenSearch object
type OpenSearchReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *OpenSearchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	openSearch := &v1beta1.OpenSearch{}
	err := r.Client.Get(ctx, req.NamespacedName, openSearch)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("OpenSearch cluster resource is not found",
				"request", req)

			return models.ExitReconcile, nil
		}

		logger.Error(err, "Unable to fetch OpenSearch cluster resource",
			"request", req)

		return reconcile.Result{}, err
	}

	switch openSearch.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, openSearch, logger)
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, openSearch, req, logger)
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, openSearch, logger)
	case models.GenericEvent:
		logger.Info("Opensearch resource generic event",
			"cluster manifest", openSearch.Spec,
			"request", req,
			"event", openSearch.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	default:
		logger.Info("OpenSearch resource event isn't handled",
			"cluster manifest", openSearch.Spec,
			"request", req,
			"event", openSearch.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *OpenSearchReconciler) createOpenSearchFromRestore(o *v1beta1.OpenSearch, logger logr.Logger) (*models.OpenSearchCluster, error) {
	logger.Info(
		"Creating OpenSearch cluster from backup",
		"original cluster ID", o.Spec.RestoreFrom.ClusterID,
	)

	id, err := r.API.RestoreCluster(o.RestoreInfoToInstAPI(o.Spec.RestoreFrom), models.OpenSearchAppKind)
	if err != nil {
		logger.Error(err, "Cannot restore OpenSearch cluster from backup",
			"original cluster ID", o.Spec.RestoreFrom.ClusterID)

		r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
			"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
			err)

		return nil, err
	}

	instaModel, err := r.API.GetOpenSearch(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get opensearch cluster details, err: %w", err)
	}

	logger.Info("OpenSearch cluster was created from backup",
		"cluster ID", instaModel.ID,
		"original cluster ID", o.Spec.RestoreFrom.ClusterID)

	r.EventRecorder.Eventf(o, models.Normal, models.Created,
		"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
		o.Spec.RestoreFrom.ClusterID, instaModel.ID)

	return instaModel, nil
}

func (r *OpenSearchReconciler) createOpenSearch(o *v1beta1.OpenSearch, logger logr.Logger) (*models.OpenSearchCluster, error) {
	logger.Info(
		"Creating OpenSearch cluster",
		"cluster name", o.Spec.Name,
		"data centres", o.Spec.DataCentres,
	)

	b, err := r.API.CreateClusterRaw(instaclustr.OpenSearchEndpoint, o.Spec.ToInstAPI())
	if err != nil {
		logger.Error(err, "Cannot create OpenSearch cluster",
			"cluster name", o.Spec.Name,
			"cluster spec", o.Spec)

		r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
			"Cluster creation on the Instaclustr is failed. Reason: %v", err)

		return nil, err
	}

	var instaModel models.OpenSearchCluster

	err = json.Unmarshal(b, &instaModel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to model")
	}

	logger.Info("OpenSearch cluster was created",
		"cluster ID", instaModel.ID,
		"cluster name", o.Spec.Name)

	r.EventRecorder.Eventf(o, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s", instaModel.ID)

	return &instaModel, nil
}

func (r *OpenSearchReconciler) createCluster(ctx context.Context, o *v1beta1.OpenSearch, logger logr.Logger) error {
	var instaModel *models.OpenSearchCluster
	var err error

	if o.Spec.HasRestore() {
		instaModel, err = r.createOpenSearchFromRestore(o, logger)
	} else {
		instaModel, err = r.createOpenSearch(o, logger)
	}
	if err != nil {
		return err
	}

	o.Spec.FromInstAPI(instaModel)
	o.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Update(ctx, o)
	if err != nil {
		return fmt.Errorf("failed to update cluster spec, err: %w", err)
	}

	o.Status.FromInstAPI(instaModel)
	err = r.Status().Update(ctx, o)
	if err != nil {
		return fmt.Errorf("failed to update cluster status, err : %w", err)
	}

	logger.Info(
		"OpenSearch resource has been created",
		"cluster name", o.Spec.Name, "cluster ID", o.Status.ID,
	)

	return nil
}

func (r *OpenSearchReconciler) startClusterJobs(o *v1beta1.OpenSearch) error {
	err := r.startClusterSyncJob(o)
	if err != nil {
		return fmt.Errorf("failed to start cluster sync job, err: %w", err)
	}

	r.EventRecorder.Event(o, models.Normal, models.Created,
		"Cluster sync job is started")

	err = r.startClusterBackupsJob(o)
	if err != nil {
		return fmt.Errorf("failed to start cluster backups job, err: %w", err)
	}

	r.EventRecorder.Event(o, models.Normal, models.Created,
		"Cluster backups check job is started")

	if o.Spec.UserRefs != nil && o.Status.AvailableUsers == nil {
		err = r.startUsersCreationJob(o)
		if err != nil {
			return fmt.Errorf("failed to start user creation job, err: %w", err)
		}

		r.EventRecorder.Event(o, models.Normal, models.Created,
			"Cluster user creation job is started")
	}

	return nil
}

func (r *OpenSearchReconciler) HandleCreateCluster(
	ctx context.Context,
	o *v1beta1.OpenSearch,
	logger logr.Logger,
) (reconcile.Result, error) {
	logger = logger.WithName("OpenSearch creation event")
	if o.Status.ID == "" {
		err := r.createCluster(ctx, o, logger)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to create cluster, err: %w", err)
		}
	}

	if o.Status.State != models.DeletedStatus {
		patch := o.NewPatch()
		o.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(o, models.DeletionFinalizer)
		err := r.Patch(ctx, o, patch)
		if err != nil {
			r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
				"Failed to update resource metadata. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}

		err = r.startClusterJobs(o)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to run cluster jobs, err: %w", err)
		}
	}

	return models.ExitReconcile, nil
}

func (r *OpenSearchReconciler) HandleUpdateCluster(
	ctx context.Context,
	o *v1beta1.OpenSearch,
	req ctrl.Request,
	logger logr.Logger,
) (reconcile.Result, error) {
	logger = logger.WithName("OpenSearch update event")

	instaModel, err := r.API.GetOpenSearch(o.Status.ID)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(
			o, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	iOpenSearch := &v1beta1.OpenSearch{}
	iOpenSearch.FromInstAPI(instaModel)

	if o.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.OpenSearchSpec](r.EventRecorder, r.Client, o, iOpenSearch, logger)
	}

	if o.Spec.ClusterSettingsNeedUpdate(&iOpenSearch.Spec.GenericClusterSpec) {
		logger.Info("Updating cluster settings",
			"instaclustr description", iOpenSearch.Spec.Description,
			"instaclustr two factor delete", iOpenSearch.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(o.Status.ID, o.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			logger.Error(err, "Cannot update cluster settings",
				"cluster ID", o.Status.ID, "cluster spec", o.Spec)
			r.EventRecorder.Eventf(o, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !o.Spec.IsEqual(iOpenSearch.Spec) {
		logger.Info("Update request to Instaclustr API has been sent",
			"spec data centres", o.Spec.DataCentres,
			"resize settings", o.Spec.ResizeSettings,
		)

		err = r.API.UpdateOpenSearch(o.Status.ID, o.Spec.ToInstAPIUpdate())
		if err != nil {
			logger.Error(err, "Cannot update cluster",
				"cluster ID", o.Status.ID,
				"cluster spec", o.Spec,
				"cluster state", o.Status.State)

			r.EventRecorder.Eventf(o, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		logger.Info(
			"Cluster has been updated",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID,
			"data centres", o.Spec.DataCentres,
		)
	}

	err = handleUsersChanges(ctx, r.Client, r, o)
	if err != nil {
		logger.Error(err, "Failed to handle users changes")
		r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
			"Handling users changes is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	patch := o.NewPatch()
	o.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, o, patch)
	if err != nil {
		logger.Error(err, "Cannot patch OpenSearch metadata",
			"cluster name", o.Spec.Name,
			"cluster metadata", o.ObjectMeta,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	logger.Info("OpenSearch cluster was updated",
		"cluster name", o.Spec.Name,
		"cluster ID", o.Status.ID)

	return models.ExitReconcile, nil
}

func (r *OpenSearchReconciler) HandleDeleteCluster(
	ctx context.Context,
	o *v1beta1.OpenSearch,
	logger logr.Logger,
) (reconcile.Result, error) {
	logger = logger.WithName("OpenSearch deletion event")

	_, err := r.API.GetOpenSearch(o.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get OpenSearch cluster",
			"cluster name", o.Spec.Name,
			"cluster status", o.Status.State)

		r.EventRecorder.Eventf(o, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	patch := o.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		logger.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID)

		err = r.API.DeleteCluster(o.Status.ID, instaclustr.OpenSearchEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete OpenSearch cluster",
				"cluster name", o.Spec.Name,
				"cluster status", o.Status.State)

			r.EventRecorder.Eventf(o, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v", err)

			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(o, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if o.Spec.TwoFactorDelete != nil {
			o.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			o.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, o, patch)
			if err != nil {
				logger.Error(err, "Cannot patch cluster resource",
					"cluster name", o.Spec.Name,
					"cluster state", o.Status.State)
				r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v", err)

				return reconcile.Result{}, err
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", o.Status.ID)

			r.EventRecorder.Event(o, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(o.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(o.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(o.GetJobID(scheduler.StatusChecker))

	logger.Info("Deleting cluster backup resources",
		"cluster ID", o.Status.ID,
	)

	err = r.deleteBackups(ctx, o.Status.ID, o.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", o.Status.ID,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	logger.Info("OpenSearch cluster backup resources were deleted",
		"cluster ID", o.Status.ID,
	)

	err = detachUsers(ctx, r.Client, r, o)
	if err != nil {
		logger.Error(err, "Failed to detach users from the cluster")
		r.EventRecorder.Eventf(o, models.Warning, models.DeletionFailed,
			"Detaching users from the cluster is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	controllerutil.RemoveFinalizer(o, models.DeletionFinalizer)
	err = r.Patch(ctx, o, patch)
	if err != nil {
		logger.Error(
			err, "Cannot update OpenSearch resource metadata after finalizer removal",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, o.Name, o.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete OpenSearch cluster expose service",
			"cluster ID", o.Status.ID,
			"cluster name", o.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	logger.Info("OpenSearch cluster was deleted",
		"cluster name", o.Spec.Name,
		"cluster ID", o.Status.ID,
	)

	r.EventRecorder.Event(o, models.Normal, models.Deleted,
		"Cluster resource is deleted")

	return models.ExitReconcile, nil
}

func (r *OpenSearchReconciler) startClusterSyncJob(cluster *v1beta1.OpenSearch) error {
	job := r.newSyncJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) startClusterBackupsJob(cluster *v1beta1.OpenSearch) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) startUsersCreationJob(cluster *v1beta1.OpenSearch) error {
	job := r.newUsersCreationJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.UserCreator), scheduler.UserCreationInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) newSyncJob(o *v1beta1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("syncJob", o.GetJobID(scheduler.StatusChecker), "clusterID", o.Status.ID)

	return func() error {
		namespacedName := client.ObjectKeyFromObject(o)
		err := r.Get(context.Background(), namespacedName, o)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(o.GetJobID(scheduler.UserCreator))
			r.Scheduler.RemoveJob(o.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(o.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get PosgtreSQL custom resource",
				"resource name", o.Name,
			)
			return err
		}

		instaModel, err := r.API.GetOpenSearch(o.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if o.DeletionTimestamp != nil {
					_, err = r.HandleDeleteCluster(context.Background(), o, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), o)
			}

			l.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
				"cluster ID", o.Status.ID)

			return err
		}

		iO := &v1beta1.OpenSearch{}
		iO.FromInstAPI(instaModel)

		if !iO.Status.Equals(&o.Status) {
			l.Info("Updating OpenSearch cluster status", "old", o.Status, "new", iO.Status)

			dcEqual := o.Status.DataCentreEquals(&iO.Status)

			patch := o.NewPatch()
			o.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), o, patch)
			if err != nil {
				l.Error(err, "Cannot patch OpenSearch cluster",
					"cluster name", o.Spec.Name,
					"status", o.Status.State,
				)

				return err
			}

			if !dcEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iO.Status.DataCentres {
					for _, node := range dc.Nodes {
						n := node.Node
						nodes = append(nodes, &n)
					}
				}

				err = exposeservice.Create(r.Client,
					o.Name,
					o.Namespace,
					o.Spec.PrivateNetwork,
					nodes,
					models.OpenSearchConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := o.Spec.IsEqual(iO.Spec)

		if equals && o.Annotations[models.ExternalChangesAnnotation] == models.True {
			err = reconcileExternalChanges(r.Client, r.EventRecorder, o)
			if err != nil {
				return err
			}
		} else if o.Status.CurrentClusterOperationStatus == models.NoOperation &&
			o.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

			patch := o.NewPatch()
			o.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), o, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", o.Spec.Name, "cluster state", o.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(o.Spec, iO.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iO.Spec, "k8s resource spec", o.Spec)
				return err

			}
			r.EventRecorder.Eventf(o, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), o)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", o.Spec.Name,
				"cluster ID", o.Status.ID,
			)
			return err
		}

		if o.Status.State == models.RunningStatus && o.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := o.NewPatch()
			for _, dc := range o.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", o.Spec.Name,
						"cluster ID", o.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), o, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", o.Spec.Name,
						"cluster ID", o.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *OpenSearchReconciler) newWatchBackupsJob(o *v1beta1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("component", "openSearchBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: o.Namespace, Name: o.Name}, o)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		instBackups, err := r.API.GetClusterBackups(o.Status.ID, models.ClusterKindsMap[o.Kind])
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster backups",
				"cluster name", o.Spec.Name,
				"clusterID", o.Status.ID)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.OsClusterKind)

		k8sBackupList, err := r.listClusterBackups(ctx, o.Status.ID, o.Namespace)
		if err != nil {
			l.Error(err, "Cannot list OpenSearch cluster backups",
				"cluster name", o.Spec.Name,
				"clusterID", o.Status.ID,
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

			backupSpec := o.NewBackupSpec(start)
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

func (r *OpenSearchReconciler) newUsersCreationJob(o *v1beta1.OpenSearch) scheduler.Job {
	logger := log.Log.WithValues("component", "openSearchUsersCreationJob")

	return func() error {
		ctx := context.Background()

		err := r.Get(ctx, types.NamespacedName{
			Namespace: o.Namespace,
			Name:      o.Name,
		}, o)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		if o.Status.State != models.RunningStatus {
			logger.Info("User creation job is scheduled")
			r.EventRecorder.Eventf(o, models.Normal, models.CreationFailed,
				"User creation job is scheduled, cluster is not in the running state",
			)
			return nil
		}

		err = handleUsersChanges(ctx, r.Client, r, o)
		if err != nil {
			logger.Error(err, "Failed to create users for the cluster")
			r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
				"Failed to create users for the cluster. Reason: %v", err)
			return err
		}

		logger.Info("User creation job successfully finished")
		r.EventRecorder.Eventf(o, models.Normal, models.Created,
			"User creation job successfully finished",
		)

		r.Scheduler.RemoveJob(o.GetJobID(scheduler.UserCreator))

		return nil
	}
}

func (r *OpenSearchReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1beta1.ClusterBackupList, error) {
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

func (r *OpenSearchReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
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

func (r *OpenSearchReconciler) NewUserResource() userObject {
	return &clusterresourcesv1beta1.OpenSearchUser{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: r.RateLimiter}).
		For(&v1beta1.OpenSearch{}, builder.WithPredicates(predicate.Funcs{
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

				newObj := event.ObjectNew.(*v1beta1.OpenSearch)

				if newObj.Status.ID == "" && newObj.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
					return false
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if event.ObjectOld.GetGeneration() == newObj.Generation {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
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
		Complete(r)
}

func (r *OpenSearchReconciler) reconcileMaintenanceEvents(ctx context.Context, o *v1beta1.OpenSearch) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(o.Status.ID)
	if err != nil {
		return err
	}

	if !o.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := o.NewPatch()
		o.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, o, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", o.Status.ID,
			"events", o.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *OpenSearchReconciler) handleExternalDelete(ctx context.Context, o *v1beta1.OpenSearch) error {
	l := log.FromContext(ctx)

	patch := o.NewPatch()
	o.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, o, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(o, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(o.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(o.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(o.GetJobID(scheduler.StatusChecker))

	return nil
}
