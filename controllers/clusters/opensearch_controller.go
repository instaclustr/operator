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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
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

	openSearch := &clustersv1alpha1.OpenSearch{}
	err := r.Client.Get(ctx, req.NamespacedName, openSearch)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("OpenSearch cluster resource is not found",
				"request", req,
			)

			return models.ExitReconcile, nil
		}

		logger.Error(err, "Unable to fetch OpenSearch cluster resource",
			"request", req,
		)

		return models.ReconcileRequeue, nil
	}

	switch openSearch.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, openSearch, logger), nil
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, openSearch, logger), nil
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, openSearch, logger), nil
	case models.GenericEvent:
		logger.Info("Opensearch resource generic event",
			"cluster manifest", openSearch.Spec,
			"request", req,
			"event", openSearch.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, err
	default:
		logger.Info("OpenSearch resource event isn't handled",
			"cluster manifest", openSearch.Spec,
			"request", req,
			"event", openSearch.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, err
	}
}

func (r *OpenSearchReconciler) HandleCreateCluster(
	ctx context.Context,
	o *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	var id string
	var err error
	if o.Status.ID == "" {
		if o.Spec.HasRestore() {
			logger.Info(
				"Creating OpenSearch cluster from backup",
				"original cluster ID", o.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreOpenSearchCluster(o.Spec.RestoreFrom)
			if err != nil {
				logger.Error(err, "Cannot restore OpenSearch cluster from backup",
					"original cluster ID", o.Spec.RestoreFrom.ClusterID)

				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
					err)

				return models.ReconcileRequeue
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

			id, err = r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, o.Spec.NewCreateRequestInstAPIv1())
			if err != nil {
				logger.Error(err, "Cannot create OpenSearch cluster",
					"cluster name", o.Spec.Name,
					"cluster spec", o.Spec)

				r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
					"Cluster creation on the Instaclustr is failed. Reason: %v",
					err)

				return models.ReconcileRequeue
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

			return models.ReconcileRequeue
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

			return models.ReconcileRequeue
		}

		if len(o.Spec.TwoFactorDelete) != 0 || o.Spec.Description != "" {
			err = r.updateDescriptionAndTwoFactorDelete(o)
			if err != nil {
				logger.Error(err, "Cannot update description and twoFactorDelete",
					"cluster name", o.Spec.Name,
					"twoFactorDelete", o.Spec.TwoFactorDelete,
					"description", o.Spec.Description,
				)

				return models.ReconcileRequeue
			}
		}
	}

	err = r.startClusterStatusJob(o)
	if err != nil {
		logger.Error(err, "Cannot start OpenSearch cluster status job",
			"cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	r.EventRecorder.Event(o, models.Normal, models.Created,
		"Cluster status check job is started")

	err = r.startClusterBackupsJob(o)
	if err != nil {
		logger.Error(err, "Cannot start OpenSearch cluster backups check job",
			"cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(o, models.Warning, models.CreationFailed,
			"Cluster backups check job is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	r.EventRecorder.Event(o, models.Normal, models.Created,
		"Cluster backups check job is started")

	logger.Info(
		"OpenSearch resource has been created",
		"cluster name", o.Name, "cluster ID", o.Status.ID,
		"api version", o.APIVersion, "namespace", o.Namespace)

	return models.ExitReconcile
}

func (r *OpenSearchReconciler) HandleUpdateCluster(
	ctx context.Context,
	o *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	iData, err := r.API.GetOpenSearch(o.Status.ID)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID)

		r.EventRecorder.Eventf(
			o, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	iOpenSearch, err := o.FromInstAPIv1(iData)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	if iOpenSearch.Status.State != models.RunningStatus {
		logger.Info("OpenSearch cluster is not ready to update",
			"cluster Name", o.Spec.Name,
			"reason", instaclustr.ClusterNotRunning,
		)

		return models.ReconcileRequeue
	}

	if o.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(o, iOpenSearch, logger)
	}

	activeResizes, err := r.API.GetActiveDataCentreResizeOperations(o.Status.ID, o.Status.CDCID)
	if err != nil {
		logger.Error(err, "Cannot get active data centre resizes",
			"cluster name", o.Spec.Name,
			"cluster status", o.Status,
		)
		return models.ReconcileRequeue
	}

	if len(activeResizes) != 0 {
		logger.Info("OpenSearch cluster is not ready to resize",
			"cluster name", o.Spec.Name,
			"cluster status", o.Status,
			"new data centre spec", o.Spec.DataCentres[0],
			"reason", instaclustr.HasActiveResizeOperation,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentreNodeSize(*iOpenSearch, *o, logger)
	if err != nil {
		if errors.Is(err, instaclustr.StatusPreconditionFailed) || errors.Is(err, instaclustr.HasActiveResizeOperation) {
			logger.Info("OpenSearch cluster is not ready to resize",
				"cluster name", o.Spec.Name,
				"cluster status", o.Status,
				"new data centre spec", o.Spec.DataCentres[0],
				"reason", err,
			)

			return models.ReconcileRequeue
		}

		logger.Error(err, "Cannot reconcile data centres node size",
			"cluster name", o.Spec.Name,
			"cluster status", o.Status)

		r.EventRecorder.Eventf(o, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(o)
	if err != nil {
		logger.Error(err, "Cannot update description and twoFactorDelete",
			"cluster name", o.Spec.Name,
			"twoFactorDelete", o.Spec.TwoFactorDelete,
			"description", o.Spec.Description,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.UpdateFailed,
			"Cluster description and TwoFactoDelete update is failed. Reason: %v", err)

		return models.ReconcileRequeue
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

		return models.ReconcileRequeue
	}

	logger.Info("OpenSearch cluster was updated",
		"cluster name", o.Spec.Name,
		"cluster status", o.Status)

	return models.ExitReconcile
}

func (r *OpenSearchReconciler) handleExternalChanges(o, iO *clustersv1alpha1.OpenSearch, l logr.Logger) reconcile.Result {
	if o.Annotations[models.AllowSpecAmendAnnotation] != models.True {
		l.Info("Update is blocked until k8s resource specification is equal with Instaclustr",
			"specification of k8s resource", o.Spec,
			"data from Instaclustr ", iO.Spec)

		r.EventRecorder.Event(o, models.Warning, models.UpdateFailed,
			"There are external changes on the Instaclustr console. Please reconcile the specification manually")

		return models.ExitReconcile
	} else {
		if !o.Spec.IsEqual(iO.Spec) {
			l.Info(msgSpecStillNoMatch,
				"specification of k8s resource", o.Spec,
				"data from Instaclustr ", iO.Spec)
			r.EventRecorder.Event(o, models.Warning, models.ExternalChanges, msgSpecStillNoMatch)

			return models.ExitReconcile
		}

		patch := o.NewPatch()

		o.Annotations[models.ExternalChangesAnnotation] = ""
		o.Annotations[models.AllowSpecAmendAnnotation] = ""

		err := r.Patch(context.Background(), o, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster resource",
				"cluster name", o.Spec.Name, "cluster ID", o.Status.ID)

			r.EventRecorder.Eventf(o, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue
		}

		l.Info("External changes have been reconciled", "resource ID", o.Status.ID)
		r.EventRecorder.Event(o, models.Normal, models.ExternalChanges, "External changes have been reconciled")

		return models.ExitReconcile
	}
}

func (r *OpenSearchReconciler) HandleDeleteCluster(
	ctx context.Context,
	o *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	_, err := r.API.GetOpenSearch(o.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get OpenSearch cluster",
			"cluster name", o.Spec.Name,
			"cluster status", o.Status.State,
		)

		r.EventRecorder.Eventf(o, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	patch := o.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		logger.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", o.Spec.Name,
			"cluster ID", o.Status.ID)

		err = r.API.DeleteCluster(o.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "Cannot delete OpenSearch cluster",
				"cluster name", o.Spec.Name,
				"cluster status", o.Status.State,
			)

			r.EventRecorder.Eventf(o, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v", err)

			return models.ReconcileRequeue
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

				return models.ReconcileRequeue
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", o.Status.ID)

			r.EventRecorder.Event(o, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
	}

	r.Scheduler.RemoveJob(o.GetJobID(scheduler.BackupsChecker))

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

		return models.ReconcileRequeue
	}

	logger.Info("OpenSearch cluster backup resources were deleted",
		"cluster ID", o.Status.ID,
	)

	r.Scheduler.RemoveJob(o.GetJobID(scheduler.StatusChecker))
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

		return models.ReconcileRequeue
	}

	logger.Info("OpenSearch cluster was deleted",
		"cluster name", o.Spec.Name,
		"cluster ID", o.Status.ID,
	)

	r.EventRecorder.Event(o, models.Normal, models.Deleted,
		"Cluster resource is deleted")

	return models.ExitReconcile
}

func (r *OpenSearchReconciler) startClusterStatusJob(cluster *clustersv1alpha1.OpenSearch) error {
	job := r.newWatchStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) startClusterBackupsJob(cluster *clustersv1alpha1.OpenSearch) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) newWatchStatusJob(o *clustersv1alpha1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("component", "openSearchStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(o)
		err := r.Get(context.Background(), namespacedName, o)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
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
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(o.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", o.Status.ClusterStatus.ID,
						"cluster name", o.Spec.Name,
					)

					patch := o.NewPatch()
					o.Annotations[models.ClusterDeletionAnnotation] = ""
					o.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), o, patch)
					if err != nil {
						l.Error(err, "Cannot patch OpenSearch cluster resource",
							"cluster ID", o.Status.ID,
							"cluster name", o.Spec.Name,
							"resource name", o.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), o)
					if err != nil {
						l.Error(err, "Cannot delete OpenSearch cluster resource",
							"cluster ID", o.Status.ID,
							"cluster name", o.Spec.Name,
							"resource name", o.Name,
						)

						return err
					}

					return nil
				}
			}

			l.Error(err, "Cannot get OpenSearch cluster from the Instaclustr API",
				"cluster ID", o.Status.ID)

			return err
		}

		iO, err := o.FromInstAPIv1(iData)
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
		}

		maintEvents, err := r.API.GetMaintenanceEvents(o.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster maintenance events",
				"cluster name", o.Spec.Name,
				"cluster ID", o.Status.ID,
			)

			return err
		}

		if iO.Status.CurrentClusterOperationStatus == models.NoOperation &&
			o.Annotations[models.ExternalChangesAnnotation] != models.True &&
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

			r.EventRecorder.Event(o, models.Warning, models.ExternalChanges,
				"There are external changes on the Instaclustr console. Please reconcile the specification manually")
		}

		if !o.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := o.NewPatch()
			o.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), o, patch)
			if err != nil {
				l.Error(err, "Cannot patch OpenSearch cluster maintenance events",
					"cluster name", o.Spec.Name,
					"cluster ID", o.Status.ID,
				)

				return err
			}

			l.Info("OpenSearch cluster maintenance events were updated",
				"cluster ID", o.Status.ID,
				"events", o.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

func (r *OpenSearchReconciler) newWatchBackupsJob(o *clustersv1alpha1.OpenSearch) scheduler.Job {
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

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, o.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster backups",
				"cluster name", o.Spec.Name,
				"clusterID", o.Status.ID,
			)

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

func (r *OpenSearchReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
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

func (r *OpenSearchReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
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

func (r *OpenSearchReconciler) reconcileDataCentreNodeSize(iO, o clustersv1alpha1.OpenSearch, logger logr.Logger) error {
	if len(o.Spec.DataCentres) == 0 ||
		len(iO.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	iDC := iO.Spec.DataCentres[0]
	dc := o.Spec.DataCentres[0]

	resizeRequest := &models.ResizeRequest{
		ConcurrentResizes:     o.Spec.ConcurrentResizes,
		NotifySupportContacts: o.Spec.NotifySupportContacts,
		ClusterID:             o.Status.ID,
		DataCentreID:          o.Status.CDCID,
	}

	if iDC.NodeSize != dc.NodeSize {
		resizeRequest.NewNodeSize = dc.NodeSize
		resizeRequest.NodePurpose = models.OpenSearchDataNodePurposeV1
		err := r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Data nodes resize request was sent",
			"cluster name", o.Spec.Name,
			"data centre id", resizeRequest.DataCentreID,
			"new node size", resizeRequest.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	if iDC.MasterNodeSize != dc.MasterNodeSize {
		resizeRequest.NewNodeSize = dc.MasterNodeSize
		resizeRequest.NodePurpose = models.OpenSearchMasterNodePurposeV1
		err := r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Master nodes resize request was sent",
			"cluster name", o.Spec.Name,
			"data centre id", resizeRequest.DataCentreID,
			"new node size", resizeRequest.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	if iDC.OpenSearchDashboardsNodeSize != dc.OpenSearchDashboardsNodeSize {
		resizeRequest.NodePurpose = models.OpenSearchDashBoardsNodePurposeV1
		err := r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Dashboard nodes resize request was sent",
			"cluster name", o.Spec.Name,
			"data centre id", resizeRequest.DataCentreID,
			"new node size", resizeRequest.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	return nil
}

func (r *OpenSearchReconciler) updateDescriptionAndTwoFactorDelete(
	openSearchCluster *clustersv1alpha1.OpenSearch) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
	if len(openSearchCluster.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = openSearchCluster.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1,
		openSearchCluster.Status.ID,
		openSearchCluster.Spec.Description,
		twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.OpenSearch{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				newObj := event.ObjectNew.(*clustersv1alpha1.OpenSearch)

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
		})).
		Complete(r)
}
