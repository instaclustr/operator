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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// OpenSearchReconciler reconciles a OpenSearch object
type OpenSearchReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
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
		return r.HandleUpdateCluster(ctx, openSearch, logger)
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

func (r *OpenSearchReconciler) HandleCreateCluster(
	ctx context.Context,
	o *v1beta1.OpenSearch,
	logger logr.Logger,
) (reconcile.Result, error) {
	logger = logger.WithName("OpenSearch creation event")
	var id string
	var err error
	if o.Status.ID == "" {
		if o.Spec.HasRestore() {
			logger.Info(
				"Creating OpenSearch cluster from backup",
				"original cluster ID", o.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreCluster(o.RestoreInfoToInstAPI(o.Spec.RestoreFrom), models.OpenSearchAppKind)
			if err != nil {
				logger.Error(err, "Cannot restore OpenSearch cluster from backup",
					"original cluster ID", o.Spec.RestoreFrom.ClusterID)

				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
					err)

				return reconcile.Result{}, err
			}

			logger.Info("OpenSearch cluster was created from backup",
				"cluster ID", id,
				"original cluster ID", o.Spec.RestoreFrom.ClusterID)

			r.EventRecorder.Eventf(o, models.Normal, models.Created,
				"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
				o.Spec.RestoreFrom.ClusterID, id)
		} else {
			logger.Info(
				"Creating OpenSearch cluster",
				"cluster name", o.Spec.Name,
				"data centres", o.Spec.DataCentres,
			)

			id, err = r.API.CreateCluster(instaclustr.OpenSearchEndpoint, o.Spec.ToInstAPI())
			if err != nil {
				logger.Error(err, "Cannot create OpenSearch cluster",
					"cluster name", o.Spec.Name,
					"cluster spec", o.Spec)

				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"Cluster creation on the Instaclustr is failed. Reason: %v", err)

				return reconcile.Result{}, err
			}

			logger.Info("OpenSearch cluster was created",
				"cluster ID", id,
				"cluster name", o.Spec.Name)

			r.EventRecorder.Eventf(o, models.Normal, models.Created,
				"Cluster creation request is sent. Cluster ID: %s", id)
		}

		patch := o.NewPatch()
		o.Status.ID = id
		err = r.Status().Patch(ctx, o, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch cluster status",
				"original cluster ID", o.Spec.RestoreFrom.ClusterID,
				"cluster ID", id)

			r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}
		o.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(o, models.DeletionFinalizer)
		err = r.Patch(ctx, o, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch cluster spec",
				"cluster ID", o.Status.ID,
				"spec", o.Spec)

			r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		logger.Info(
			"OpenSearch resource has been created",
			"cluster name", o.Name, "cluster ID", o.Status.ID,
			"api version", o.APIVersion, "namespace", o.Namespace,
		)
	}

	if o.Status.State != models.DeletedStatus {
		err = r.startClusterStatusJob(o)
		if err != nil {
			logger.Error(err, "Cannot start OpenSearch cluster status job",
				"cluster ID", o.Status.ID)

			r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
				"Cluster status check job is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(o, models.Normal, models.Created,
			"Cluster status check job is started")

		err = r.startClusterBackupsJob(o)
		if err != nil {
			logger.Error(err, "Cannot start OpenSearch cluster backups check job",
				"cluster ID", o.Status.ID)

			r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
				"Cluster backups check job is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(o, models.Normal, models.Created,
			"Cluster backups check job is started")

		if o.Spec.UserRefs != nil {
			err = r.startUsersCreationJob(o)
			if err != nil {
				logger.Error(err, "Failed to start user creation job")
				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"User creation job is failed. Reason: %v", err,
				)
				return reconcile.Result{}, err
			}

			r.EventRecorder.Event(o, models.Normal, models.Created,
				"Cluster user creation job is started")
		}
	}

	return models.ExitReconcile, nil
}

func (r *OpenSearchReconciler) HandleUpdateCluster(
	ctx context.Context,
	o *v1beta1.OpenSearch,
	logger logr.Logger,
) (reconcile.Result, error) {
	logger = logger.WithName("OpenSearch update event")

	iData, err := r.API.GetOpenSearch(o.Status.ID)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(
			o, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	iOpenSearch, err := o.FromInstAPI(iData)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	if iOpenSearch.Status.State != models.RunningStatus {
		logger.Info("OpenSearch cluster is not ready to update",
			"cluster Name", o.Spec.Name,
			"reason", instaclustr.ClusterNotRunning,
		)

		patch := o.NewPatch()
		o.Annotations[models.UpdateQueuedAnnotation] = models.True
		err = r.Patch(ctx, o, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch metadata",
				"cluster name", o.Spec.Name,
				"cluster metadata", o.ObjectMeta,
			)

			r.EventRecorder.Eventf(
				o, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	if o.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(o, iOpenSearch, logger)
	}

	if o.Spec.ClusterSettingsNeedUpdate(iOpenSearch.Spec.Cluster) {
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

			patch := o.NewPatch()
			o.Annotations[models.UpdateQueuedAnnotation] = models.True
			err = r.Patch(ctx, o, patch)
			if err != nil {
				logger.Error(err, "Cannot patch OpenSearch metadata",
					"cluster name", o.Spec.Name,
					"cluster metadata", o.ObjectMeta,
				)

				r.EventRecorder.Eventf(
					o, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}

		logger.Info(
			"Cluster has been updated",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID,
			"data centres", o.Spec.DataCentres,
		)
	}

	patch := o.NewPatch()
	o.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	o.Annotations[models.UpdateQueuedAnnotation] = ""
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

func (r *OpenSearchReconciler) handleExternalChanges(o, iO *v1beta1.OpenSearch, l logr.Logger) (reconcile.Result, error) {
	if !o.Spec.IsEqual(iO.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", o.Spec,
			"data from Instaclustr ", iO.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(o.Spec, iO.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iO.Spec, "k8s resource spec", o.Spec)
			return models.ExitReconcile, nil
		}
		r.EventRecorder.Eventf(o, models.Warning, models.ExternalChanges, msgDiffSpecs)

		return models.ExitReconcile, nil
	}

	patch := o.NewPatch()

	o.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), o, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", o.Spec.Name, "cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info("External changes have been reconciled", "resource ID", o.Status.ID)
	r.EventRecorder.Event(o, models.Normal, models.ExternalChanges, "External changes have been reconciled")

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

	for _, ref := range o.Spec.UserRefs {
		err = r.detachUserResource(ctx, logger, o, ref)
		if err != nil {
			return reconcile.Result{}, err
		}
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

func (r *OpenSearchReconciler) startClusterStatusJob(cluster *v1beta1.OpenSearch) error {
	job := r.newWatchStatusJob(cluster)

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

func (r *OpenSearchReconciler) newWatchStatusJob(o *v1beta1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("component", "openSearchStatusClusterJob")
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

		iData, err := r.API.GetOpenSearch(o.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(context.Background(), o)
			}

			l.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
				"cluster ID", o.Status.ID)

			return err
		}

		iO, err := o.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert OpenSearch cluster from the Instaclustr API",
				"cluster ID", o.Status.ID)

			return err
		}

		if !areStatusesEqual(&iO.Status.ClusterStatus, &o.Status.ClusterStatus) {
			l.Info("Updating OpenSearch cluster status",
				"new status", iO.Status.ClusterStatus,
				"old status", o.Status.ClusterStatus,
			)

			areDCsEqual := areDataCentresEqual(iO.Status.ClusterStatus.DataCentres, o.Status.ClusterStatus.DataCentres)

			patch := o.NewPatch()
			o.Status.ClusterStatus = iO.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), o, patch)
			if err != nil {
				l.Error(err, "Cannot patch OpenSearch cluster",
					"cluster name", o.Spec.Name,
					"status", o.Status.State,
				)

				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iO.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					o.Name,
					o.Namespace,
					nodes,
					models.OpenSearchConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iO.Status.CurrentClusterOperationStatus == models.NoOperation &&
			o.Annotations[models.UpdateQueuedAnnotation] != models.True &&
			!o.Spec.IsEqual(iO.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iO.Spec, "k8s resource spec", o.Spec)

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

		for _, ref := range o.Spec.UserRefs {
			err = r.createUser(ctx, logger, o, ref)
			if err != nil {
				logger.Error(err, "Failed to create a user for the cluster",
					"user ref", ref,
				)
				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"Failed to create a user for the cluster. Reason: %v", err,
				)
				return err
			}
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

func (r *OpenSearchReconciler) deleteUser(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.OpenSearch,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.OpenSearchUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "OpenSearch user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User resource is not found, please provide correct userRef."+
					"Current provided reference: %v", uRef)
			return nil
		}

		l.Error(err, "Cannot get OpenSearch user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot get OpenSearch user. user reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; !exist {
		l.Info("User is not existing on the cluster",
			"user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.DeletionFailed,
			"User is not existing on the cluster. User reference: %v", uRef)

		return nil
	}

	patch := u.NewPatch()
	u.Status.ClustersEvents[c.Status.ID] = models.DeletingEvent
	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the OpenSearch user status with the DeletingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot patch the OpenSearch user status with the DeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for deletion", "username", u.Name)

	return nil
}

func (r *OpenSearchReconciler) createUser(
	ctx context.Context,
	logger logr.Logger,
	c *v1beta1.OpenSearch,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.OpenSearchUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Error(err, "OpenSearch user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User is not found, create a new one or provide correct userRef."+
					"Current provided reference: %v", uRef)
			return err
		}

		logger.Error(err, "Cannot get OpenSearch user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot get OpenSearch user. user reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; exist {
		logger.Info("User is already existing on the cluster",
			"user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.CreationFailed,
			"User is already existing on the cluster. User reference: %v", uRef)

		return nil
	}

	patch := u.NewPatch()

	if u.Status.ClustersEvents == nil {
		u.Status.ClustersEvents = make(map[string]string)
	}

	u.Status.ClustersEvents[c.Status.ID] = models.CreatingEvent

	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		logger.Error(err, "Cannot patch the OpenSearch User status with the CreatingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot add OpenSearch User to the cluster. Reason: %v", err)
		return err
	}

	logger.Info("User has been added to the queue for creation", "username", u.Name)

	return nil
}

func (r *OpenSearchReconciler) detachUserResource(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.OpenSearch,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.OpenSearchUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "OpenSearch user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User resource is not found, please provide correct userRef."+
					"Current provided reference: %v", uRef)
			return nil
		}

		l.Error(err, "Cannot get OpenSearch user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot get OpenSearch user. user reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; !exist {
		return nil
	}

	patch := u.NewPatch()
	u.Status.ClustersEvents[c.Status.ID] = models.ClusterDeletingEvent
	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the OpenSearch user status with the ClusterDeletingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot patch the OpenSearch user status with the ClusterDeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("The user has been detached from the cluster")

	return nil
}

func (r *OpenSearchReconciler) handleUserEvent(
	newObj *v1beta1.OpenSearch,
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
			l.Error(err, "Cannot create OpenSearch user in predicate", "user", newUser)
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

		err := r.deleteUser(ctx, l, newObj, oldUser)
		if err != nil {
			l.Error(err, "Cannot delete OpenSearch user", "user", oldUser)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot delete user from cluster. Reason: %v", err)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
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

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if event.ObjectOld.GetGeneration() == newObj.Generation {
					return false
				}

				oldObj := event.ObjectOld.(*v1beta1.OpenSearch)

				r.handleUserEvent(newObj, oldObj.Spec.UserRefs)

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

	if !o.Status.AreMaintenanceEventStatusesEqual(iMEStatuses) {
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
