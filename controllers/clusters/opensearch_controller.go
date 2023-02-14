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
	convertorsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// OpenSearchReconciler reconciles a OpenSearch object
type OpenSearchReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=opensearches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenSearch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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

			return models.ReconcileResult, nil
		}

		logger.Error(err, "Unable to fetch OpenSearch cluster resource",
			"request", req,
		)

		return models.ReconcileResult, nil
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
		return models.ReconcileResult, err
	default:
		logger.Info("OpenSearch resource event isn't handled",
			"cluster manifest", openSearch.Spec,
			"request", req,
			"event", openSearch.Annotations[models.ResourceStateAnnotation],
		)
		return models.ReconcileResult, err
	}
}

func (r *OpenSearchReconciler) HandleCreateCluster(
	ctx context.Context,
	openSearch *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	var id string
	var err error
	patch := openSearch.NewPatch()
	if openSearch.Status.ID == "" {
		if openSearch.Spec.HasRestore() {
			logger.Info(
				"Creating OpenSearch cluster from backup",
				"original cluster ID", openSearch.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreOpenSearchCluster(openSearch.Spec.RestoreFrom)
			if err != nil {
				logger.Error(err, "Cannot restore OpenSearch cluster from backup",
					"original cluster ID", openSearch.Spec.RestoreFrom.ClusterID,
				)

				return models.ReconcileRequeue
			}

			openSearch.Status.ID = id
			err = r.Status().Patch(ctx, openSearch, patch)
			if err != nil {
				logger.Error(err, "Cannot patch OpenSearch cluster status",
					"original cluster ID", openSearch.Spec.RestoreFrom.ClusterID,
					"cluster ID", id,
				)

				return models.ReconcileRequeue
			}
		} else {
			logger.Info(
				"Creating OpenSearch cluster",
				"cluster name", openSearch.Spec.Name,
				"data centres", openSearch.Spec.DataCentres,
			)

			openSearchSpec := convertorsv1.OpenSearchToInstAPI(&openSearch.Spec)

			id, err = r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, openSearchSpec)
			if err != nil {
				logger.Error(err, "Cannot create OpenSearch cluster",
					"cluster name", openSearch.Spec.Name,
					"cluster spec", openSearch.Spec,
				)

				return models.ReconcileRequeue
			}

			openSearch.Status.ID = id
			err = r.Status().Patch(ctx, openSearch, patch)
			if err != nil {
				logger.Error(err, "Cannot patch OpenSearch cluster status",
					"original cluster ID", openSearch.Spec.RestoreFrom.ClusterID,
					"cluster ID", id,
				)

				return models.ReconcileRequeue
			}
		}

		openSearch.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		openSearch.Annotations[models.DeletionConfirmed] = models.False
		controllerutil.AddFinalizer(openSearch, models.DeletionFinalizer)
		err = r.Patch(ctx, openSearch, patch)
		if err != nil {
			logger.Error(err, "Cannot patch OpenSearch cluster spec",
				"cluster ID", openSearch.Status.ID,
				"cluster spec", openSearch.Spec,
			)

			return models.ReconcileRequeue
		}

		if len(openSearch.Spec.TwoFactorDelete) != 0 || openSearch.Spec.Description != "" {
			err = r.updateDescriptionAndTwoFactorDelete(openSearch)
			if err != nil {
				logger.Error(err, "Cannot update description and twoFactorDelete",
					"cluster name", openSearch.Spec.Name,
					"twoFactorDelete", openSearch.Spec.TwoFactorDelete,
					"description", openSearch.Spec.Description,
				)

				return models.ReconcileRequeue
			}
		}
	}

	err = r.startClusterStatusJob(openSearch)
	if err != nil {
		logger.Error(err, "Cannot start OpenSearch cluster status job",
			"cluster ID", openSearch.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.startClusterBackupsJob(openSearch)
	if err != nil {
		logger.Error(err, "Cannot start OpenSearch cluster backups check job",
			"cluster ID", openSearch.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info(
		"OpenSearch resource has been created",
		"cluster name", openSearch.Name,
		"cluster ID", openSearch.Status.ID,
		"kind", openSearch.Kind,
		"api version", openSearch.APIVersion,
		"namespace", openSearch.Namespace,
	)

	return models.ReconcileResult
}

func (r *OpenSearchReconciler) HandleUpdateCluster(
	ctx context.Context,
	openSearch *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	openSearchInstClusterStatus, err := r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(err, "Cannot get OpenSearch cluster status from the Instaclustr API",
			"cluster name", openSearch.Spec.Name,
			"cluster ID", openSearch.Status.ID,
		)

		return models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(openSearch)
	if err != nil {
		logger.Error(err, "Cannot update description and twoFactorDelete",
			"cluster name", openSearch.Spec.Name,
			"twoFactorDelete", openSearch.Spec.TwoFactorDelete,
			"description", openSearch.Spec.Description,
		)

		return models.ReconcileRequeue
	}

	if openSearchInstClusterStatus.Status != models.StatusRUNNING {
		logger.Info("OpenSearch cluster is not ready to update",
			"cluster Name", openSearch.Spec.Name,
			"reason", instaclustr.ClusterNotRunning,
		)

		return models.ReconcileRequeue
	}

	err = r.reconcileDataCentresNodeSize(openSearchInstClusterStatus, openSearch, logger)
	if errors.Is(err, instaclustr.StatusPreconditionFailed) || errors.Is(err, instaclustr.HasActiveResizeOperation) {
		logger.Info("OpenSearch cluster is not ready to resize",
			"cluster name", openSearch.Spec.Name,
			"cluster status", openSearchInstClusterStatus.Status,
			"data centre spec", openSearch.Spec.DataCentres[0],
			"reason", err,
		)

		return models.ReconcileRequeue
	}
	if err != nil {
		logger.Error(err, "Cannot reconcile data centres node size",
			"cluster name", openSearch.Spec.Name,
			"cluster status", openSearchInstClusterStatus.Status,
		)

		return models.ReconcileRequeue
	}

	patch := openSearch.NewPatch()
	openSearch.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, openSearch, patch)
	if err != nil {
		logger.Error(err, "Cannot patch OpenSearch metadata",
			"cluster name", openSearch.Spec.Name,
			"cluster metadata", openSearch.ObjectMeta,
		)

		return models.ReconcileRequeue
	}

	openSearchInstClusterStatus, err = r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "Cannot get OpenSearch cluster status from the Instaclustr API",
			"cluster name", openSearch.Spec.Name,
			"cluster ID", openSearch.Status.ID,
		)

		return models.ReconcileRequeue
	}

	openSearch.Status.ClusterStatus = *openSearchInstClusterStatus
	err = r.Status().Patch(ctx, openSearch, patch)
	if err != nil {
		logger.Error(err, "Cannot update OpenSearch cluster status",
			"cluster name", openSearch.Spec.Name,
			"cluster status", openSearch.Status,
		)

		return models.ReconcileRequeue
	}

	logger.Info("OpenSearch cluster was updated",
		"cluster name", openSearch.Spec.Name,
		"cluster status", openSearch.Status,
	)

	return models.ReconcileResult
}

func (r *OpenSearchReconciler) HandleDeleteCluster(
	ctx context.Context,
	openSearch *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) reconcile.Result {
	patch := openSearch.NewPatch()
	if openSearch.DeletionTimestamp == nil {
		openSearch.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err := r.Patch(ctx, openSearch, patch)
		if err != nil {
			logger.Error(
				err, "Cannot update OpenSearch resource metadata",
				"cluster name", openSearch.Spec.Name,
				"cluster ID", openSearch.Status.ID,
			)
			return models.ReconcileRequeue
		}

		logger.Info("OpenSearch cluster is no longer deleting",
			"cluster name", openSearch.Spec.Name,
		)
		return models.ReconcileResult
	}

	status, err := r.API.GetClusterStatus(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(err, "Cannot get OpenSearch cluster status",
			"cluster name", openSearch.Spec.Name,
			"cluster status", openSearch.Status.Status,
		)

		return models.ReconcileRequeue
	}

	if status != nil {
		if len(openSearch.Spec.TwoFactorDelete) != 0 && openSearch.Annotations[models.DeletionConfirmed] != models.True {
			logger.Info("OpenSearch cluster deletion is not confirmed",
				"cluster name", openSearch.Spec.Name,
				"annotations", openSearch.Annotations,
			)

			openSearch.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			err = r.Patch(ctx, openSearch, patch)
			if err != nil {
				logger.Error(err, "Cannot patch OpenSearch cluster metadata",
					"cluster name", openSearch.Spec.Name,
				)

				return models.ReconcileRequeue
			}
		}

		err = r.API.DeleteCluster(openSearch.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "Cannot delete OpenSearch cluster",
				"cluster name", openSearch.Spec.Name,
				"cluster status", openSearch.Status.Status,
			)

			return models.ReconcileRequeue
		}

		logger.Info("OpenSearch cluster is being deleted",
			"cluster name", openSearch.Spec.Name,
			"cluster status", openSearch.Status.Status,
		)
		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(openSearch.GetJobID(scheduler.BackupsChecker))

	logger.Info("Deleting cluster backup resources",
		"cluster ID", openSearch.Status.ID,
	)

	err = r.deleteBackups(ctx, openSearch.Status.ID, openSearch.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", openSearch.Status.ID,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cluster backup resources were deleted",
		"cluster ID", openSearch.Status.ID,
	)

	r.Scheduler.RemoveJob(openSearch.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(openSearch, models.DeletionFinalizer)
	err = r.Patch(ctx, openSearch, patch)
	if err != nil {
		logger.Error(
			err, "Cannot update OpenSearch resource metadata after finalizer removal",
			"cluster name", openSearch.Spec.Name,
			"cluster ID", openSearch.Status.ID,
		)

		return models.ReconcileRequeue
	}

	logger.Info("OpenSearch cluster was deleted",
		"cluster name", openSearch.Spec.Name,
		"cluster ID", openSearch.Status.ID,
	)

	return models.ReconcileResult
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

func (r *OpenSearchReconciler) newWatchStatusJob(cluster *clustersv1alpha1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("component", "openSearchStatusClusterJob")
	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("OpenSearch cluster is being deleted. Status check job skipped",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instStatus, err := r.API.GetClusterStatus(cluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster status",
				"cluster ID", cluster.Status.ID)

			return err
		}

		if !areStatusesEqual(instStatus, &cluster.Status.ClusterStatus) {
			l.Info("Updating Opensearh cluster status",
				"new status", instStatus,
				"old status", cluster.Status.ClusterStatus,
			)

			patch := cluster.NewPatch()
			instStatus.MaintenanceEvents = cluster.Status.MaintenanceEvents
			cluster.Status.ClusterStatus = *instStatus
			err = r.Status().Patch(context.Background(), cluster, patch)
			if err != nil {
				l.Error(err, "Cannot patch OpenSearch cluster",
					"cluster name", cluster.Spec.Name,
					"status", cluster.Status.Status,
				)

				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster maintenance events",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return err
		}

		if !cluster.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := cluster.NewPatch()
			cluster.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), cluster, patch)
			if err != nil {
				l.Error(err, "Cannot patch OpenSearch cluster maintenance events",
					"cluster name", cluster.Spec.Name,
					"cluster ID", cluster.Status.ID,
				)

				return err
			}

			l.Info("OpenSearch cluster maintenance events were updated",
				"cluster ID", cluster.Status.ID,
				"events", cluster.Status.MaintenanceEvents,
			)
		}

		activeOperations, err := r.API.GetActiveDataCentreResizeOperations(cluster.Status.ID, cluster.Status.CDCID)
		if err != nil {
			l.Error(err, "Cannot get active OpenSearch data centre resize operations",
				"cluster ID", cluster.Status.ID)

			return err
		}

		if len(activeOperations) == 0 {
			instSpec, err := r.API.GetClusterSpec(cluster.Status.ID, instaclustr.ClustersEndpointV1)
			if err != nil {
				l.Error(err, "Cannot get OpenSearch cluster spec",
					"cluster ID", cluster.Status.ID)

				return err
			}

			if !cluster.Spec.IsSpecEqual(instSpec) {
				patch := cluster.NewPatch()
				cluster.Spec.SetSpecFromInst(instSpec)
				cluster.Spec.RestoreFrom = nil
				err = r.Patch(context.TODO(), cluster, patch)
				if err != nil {
					l.Error(err, "Cannot patch OpenSearch cluster spec",
						"cluster name", cluster.Spec.Name,
						"cluster ID", cluster.Status.ID,
						"cluster spec", instSpec,
					)

					return err
				}

				l.Info("OpenSearch cluster spec was updated",
					"cluster ID", cluster.Status.ID,
					"cluster spec", cluster.Spec,
				)
			}
		}

		return nil
	}
}

func (r *OpenSearchReconciler) newWatchBackupsJob(cluster *clustersv1alpha1.OpenSearch) scheduler.Job {
	l := log.Log.WithValues("component", "openSearchBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("OpenSearch cluster is being deleted. Backups check job skipped",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get OpenSearch cluster backups",
				"cluster name", cluster.Spec.Name,
				"clusterID", cluster.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.OsClusterKind)

		k8sBackupList, err := r.listClusterBackups(ctx, cluster.Status.ID, cluster.Namespace)
		if err != nil {
			l.Error(err, "Cannot list OpenSearch cluster backups",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
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

// SetupWithManager sets up the controller with the Manager.
func (r *OpenSearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.OpenSearch{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				newObj := event.Object.(*clustersv1alpha1.OpenSearch)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*clustersv1alpha1.OpenSearch)
				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
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
