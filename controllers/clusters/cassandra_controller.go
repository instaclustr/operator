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

const (
	StatusRUNNING = "RUNNING"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	cassandra := &clustersv1alpha1.Cassandra{}
	err := r.Client.Get(ctx, req.NamespacedName, cassandra)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Cassandra resource is not found",
				"request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch Cassandra cluster",
			"request", req)
		return models.ReconcileRequeue, err
	}

	switch cassandra.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, l, cassandra), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, l, cassandra), nil
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, l, cassandra), nil
	case models.GenericEvent:
		l.Info("Event isn't handled",
			"cluster name", cassandra.Spec.Name,
			"request", req,
			"event", cassandra.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	default:
		l.Info("Unknown event isn't handled",
			"request", req,
			"event", cassandra.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *CassandraReconciler) handleCreateCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	l = l.WithName("Cassandra creation event")
	var err error
	patch := cassandra.NewPatch()
	if cassandra.Status.ID == "" {
		var id string
		if cassandra.Spec.HasRestore() {
			l.Info(
				"Creating cluster from backup",
				"original cluster ID", cassandra.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreCassandra(*cassandra.Spec.RestoreFrom)
			if err != nil {
				l.Error(err, "Cannot restore cluster from backup",
					"original cluster ID", cassandra.Spec.RestoreFrom.ClusterID,
				)

				r.EventRecorder.Eventf(
					cassandra, models.Warning, models.CreationFailed,
					"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue
			}

			r.EventRecorder.Eventf(
				cassandra, models.Normal, models.Created,
				"Cluster restore request is sent. Original cluster ID: %s, new cluster ID: %s",
				cassandra.Spec.RestoreFrom.ClusterID,
				id,
			)
		} else {
			l.Info(
				"Creating cluster",
				"cluster name", cassandra.Spec.Name,
				"data centres", cassandra.Spec.DataCentres,
			)

			id, err = r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandra.Spec.ToInstAPI())
			if err != nil {
				l.Error(
					err, "Cannot create cluster",
					"cluster spec", cassandra.Spec,
				)
				r.EventRecorder.Eventf(
					cassandra, models.Warning, models.CreationFailed,
					"Cluster creation on the Instaclustr is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue
			}

			r.EventRecorder.Eventf(
				cassandra, models.Normal, models.Created,
				"Cluster creation request is sent. Cluster ID: %s",
				id,
			)
		}

		cassandra.Status.ID = id
		err = r.Status().Patch(ctx, cassandra, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster status",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
				"kind", cassandra.Kind,
				"api Version", cassandra.APIVersion,
				"namespace", cassandra.Namespace,
				"cluster metadata", cassandra.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(cassandra, models.DeletionFinalizer)
		cassandra.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		cassandra.Annotations[models.DeletionConfirmed] = models.False
		err = r.Patch(ctx, cassandra, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
				"kind", cassandra.Kind,
				"api Version", cassandra.APIVersion,
				"namespace", cassandra.Namespace,
				"cluster metadata", cassandra.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Cluster has been created",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
			"kind", cassandra.Kind,
			"api Version", cassandra.APIVersion,
			"namespace", cassandra.Namespace,
		)
	}

	err = r.startClusterStatusJob(cassandra)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"cassandra cluster ID", cassandra.Status.ID)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	err = r.startClusterBackupsJob(cassandra)
	if err != nil {
		l.Error(err, "Cannot start cluster backups check job",
			"cluster ID", cassandra.Status.ID,
		)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.CreationFailed,
			"Cluster backups check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Created,
		"Cluster backups check job is started",
	)

	return models.ExitReconcile
}

func (r *CassandraReconciler) handleUpdateCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	l = l.WithName("Cassandra update event")

	iData, err := r.API.GetCassandra(cassandra.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get cluster from the Instaclustr API",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
		)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	iCassandra, err := cassandra.FromInstAPI(iData)
	if err != nil {
		l.Error(
			err, "Cannot convert cluster from the Instaclustr API",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
		)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if !cassandra.Spec.IsEqual(iCassandra.Spec) {
		err = r.API.UpdateCassandra(cassandra.Status.ID, cassandra.Spec.NewDCsUpdate())
		if err != nil {
			l.Error(err, "Cannot update cluster",
				"cluster ID", cassandra.Status.ID,
				"cluster name", cassandra.Spec.Name,
				"cluster spec", cassandra.Spec,
				"cluster state", cassandra.Status.State,
			)

			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
	}

	patch := cassandra.NewPatch()
	cassandra.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, cassandra, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
			"kind", cassandra.Kind,
			"api Version", cassandra.APIVersion,
			"namespace", cassandra.Namespace,
			"cluster metadata", cassandra.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", cassandra.Spec.Name,
		"cluster ID", cassandra.Status.ID,
		"namespace", cassandra.Namespace,
		"data centres", cassandra.Spec.DataCentres,
	)

	return models.ExitReconcile
}

func (r *CassandraReconciler) handleDeleteCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	l = l.WithName("Cassandra deletion event")

	patch := cassandra.NewPatch()
	err := r.Patch(ctx, cassandra, patch)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
			"kind", cassandra.Kind,
			"api Version", cassandra.APIVersion,
			"namespace", cassandra.Namespace,
			"cluster metadata", cassandra.ObjectMeta,
		)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if len(cassandra.Spec.TwoFactorDelete) != 0 &&
		cassandra.Annotations[models.DeletionConfirmed] != models.True {
		l.Info("Cluster deletion is not confirmed",
			"cluster ID", cassandra.Status.ID,
			"cluster name", cassandra.Spec.Name,
			"deletion confirmation annotation", cassandra.Annotations[models.DeletionConfirmed],
		)

		cassandra.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.Patch(ctx, cassandra, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster resource",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
			)

			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		return models.ExitReconcile
	}

	_, err = r.API.GetCassandra(cassandra.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get cluster from the Instaclustr API",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
			"kind", cassandra.Kind,
			"api Version", cassandra.APIVersion,
			"namespace", cassandra.Namespace,
		)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if !errors.Is(err, instaclustr.NotFound) {
		err = r.API.DeleteCluster(cassandra.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete cluster",
				"cluster name", cassandra.Spec.Name,
				"state", cassandra.Status.State,
				"kind", cassandra.Kind,
				"api Version", cassandra.APIVersion,
				"namespace", cassandra.Namespace,
			)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.DeletionFailed,
				"Cluster deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			cassandra, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent",
		)

		l.Info("Cluster is being deleted",
			"cluster ID", cassandra.Status.ID,
		)

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.StatusChecker))
	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.BackupsChecker))

	l.Info("Deleting cluster backup resources",
		"cluster ID", cassandra.Status.ID,
	)

	err = r.deleteBackups(ctx, cassandra.Status.ID, cassandra.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", cassandra.Status.ID,
		)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.DeletionFailed,
			"Cluster backups deletion is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster backup resources were deleted",
		"cluster ID", cassandra.Status.ID,
	)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Deleted,
		"Cluster backup resources are deleted",
	)

	controllerutil.RemoveFinalizer(cassandra, models.DeletionFinalizer)
	cassandra.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, cassandra, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
			"kind", cassandra.Kind,
			"api Version", cassandra.APIVersion,
			"namespace", cassandra.Namespace,
			"cluster metadata", cassandra.ObjectMeta,
		)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster has been deleted",
		"cluster name", cassandra.Spec.Name,
		"cluster ID", cassandra.Status.ID,
		"kind", cassandra.Kind,
		"api Version", cassandra.APIVersion,
		"namespace", cassandra.Namespace,
	)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile
}

func (r *CassandraReconciler) startClusterStatusJob(cassandraCluster *clustersv1alpha1.Cassandra) error {
	job := r.newWatchStatusJob(cassandraCluster)

	err := r.Scheduler.ScheduleJob(cassandraCluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) startClusterBackupsJob(cluster *clustersv1alpha1.Cassandra) error {
	job := r.newWatchBackupsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.BackupsChecker), scheduler.ClusterBackupsInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) newWatchStatusJob(cassandra *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "CassandraStatusClusterJob")
	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{
			Namespace: cassandra.Namespace,
			Name:      cassandra.Name,
		}, cassandra)
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", cassandra.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cassandra.DeletionTimestamp, len(cassandra.Spec.TwoFactorDelete), cassandra.Annotations[models.DeletionConfirmed]) {
			l.Info("Cluster is being deleted. Status check job skipped",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
			)

			return nil
		}

		iData, err := r.API.GetCassandra(cassandra.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(cassandra.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", cassandra.Status.ClusterStatus.ID,
						"cluster name", cassandra.Spec.Name,
					)

					patch := cassandra.NewPatch()
					cassandra.Annotations[models.DeletionConfirmed] = models.True
					cassandra.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), cassandra, patch)
					if err != nil {
						l.Error(err, "Cannot patch Cassandra cluster resource",
							"cluster ID", cassandra.Status.ID,
							"cluster name", cassandra.Spec.Name,
							"resource name", cassandra.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), cassandra)
					if err != nil {
						l.Error(err, "Cannot delete Cassandra cluster resource",
							"cluster ID", cassandra.Status.ID,
							"cluster name", cassandra.Spec.Name,
							"resource name", cassandra.Name,
						)

						return err
					}

					return nil
				}
			}

			l.Error(err, "Cannot get cluster from the Instaclustr API",
				"clusterID", cassandra.Status.ID)
			return err
		}

		iCassandra, err := cassandra.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert cluster from the Instaclustr API",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
			)
			return err
		}

		if !areStatusesEqual(&iCassandra.Status.ClusterStatus, &cassandra.Status.ClusterStatus) {
			l.Info("Updating cluster status",
				"status from insta", iCassandra.Status.ClusterStatus,
				"status from k8s", cassandra.Status.ClusterStatus)

			patch := cassandra.NewPatch()
			cassandra.Status.ClusterStatus = iCassandra.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), cassandra, patch)
			if err != nil {
				return err
			}
		}

		if iCassandra.Status.CurrentClusterOperationStatus == models.NoOperation &&
			!cassandra.Spec.IsEqual(iCassandra.Spec) {
			patch := cassandra.NewPatch()
			cassandra.Spec = iCassandra.Spec
			err = r.Patch(context.Background(), cassandra, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster ID", cassandra.Status.ID,
					"spec from insta", iCassandra.Spec,
				)

				return err
			}

			l.Info("Cluster spec has been updated",
				"cluster ID", cassandra.Status.ID,
			)
		}

		maintEvents, err := r.API.GetMaintenanceEvents(cassandra.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get cluster maintenance events",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
			)

			return err
		}

		if !cassandra.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := cassandra.NewPatch()
			cassandra.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), cassandra, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster maintenance events",
					"cluster name", cassandra.Spec.Name,
					"cluster ID", cassandra.Status.ID,
				)

				return err
			}

			l.Info("Cluster maintenance events were updated",
				"cluster ID", cassandra.Status.ID,
				"events", cassandra.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

func (r *CassandraReconciler) newWatchBackupsJob(cluster *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraBackupsClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}, cluster)
		if err != nil {
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cluster.DeletionTimestamp, len(cluster.Spec.TwoFactorDelete), cluster.Annotations[models.DeletionConfirmed]) {
			l.Info("Cluster is being deleted. Backups check job skipped",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		iBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get cluster backups",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return err
		}

		iBackupEvents := iBackups.GetBackupEvents(models.CassandraKind)

		k8sBackupList, err := r.listClusterBackups(ctx, cluster.Status.ID, cluster.Namespace)
		if err != nil {
			l.Error(err, "Cannot list cluster backups",
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

func (r *CassandraReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
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

func (r *CassandraReconciler) deleteBackups(ctx context.Context, clusterID, namespace string) error {
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
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cassandra{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				obj := event.Object.(*clustersv1alpha1.Cassandra)
				if obj.DeletionTimestamp != nil &&
					(len(obj.Spec.TwoFactorDelete) == 0 || obj.Annotations[models.DeletionConfirmed] == models.True) {
					obj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*clustersv1alpha1.Cassandra)
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
