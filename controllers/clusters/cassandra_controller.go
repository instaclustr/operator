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
	apiv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

const (
	StatusRUNNING = "RUNNING"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cassandra object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var cassandra clustersv1alpha1.Cassandra
	err := r.Client.Get(ctx, req.NamespacedName, &cassandra)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cassandra resource is not found",
				"request", req)
			return models.ReconcileResult, nil
		}

		l.Error(err, "Unable to fetch Cassandra cluster",
			"request", req)
		return reconcile.Result{}, err
	}

	switch cassandra.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, l, &cassandra), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, l, &cassandra), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, l, &cassandra), nil

	case models.GenericEvent:
		l.Info("Event isn't handled",
			"cluster name", cassandra.Spec.Name,
			"request", req,
			"event", cassandra.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *CassandraReconciler) handleCreateCluster(
	ctx context.Context,
	l logr.Logger,
	cc *clustersv1alpha1.Cassandra,
) reconcile.Result {
	var id string
	var err error
	patch := cc.NewPatch()
	if cc.Status.ID == "" {
		if cc.Spec.HasRestore() {
			l.Info(
				"Creating Cassandra cluster from backup",
				"original cluster ID", cc.Spec.RestoreFrom.ClusterID,
			)

			id, err = r.API.RestoreCassandra(*cc.Spec.RestoreFrom)
			if err != nil {
				l.Error(err, "Cannot restore Cassandra cluster from backup",
					"original cluster ID", cc.Spec.RestoreFrom.ClusterID,
				)

				return models.ReconcileRequeue
			}
		} else {
			l.Info(
				"Creating Cassandra cluster",
				"cluster name", cc.Spec.Name,
				"data centres", cc.Spec.DataCentres,
			)

			cassandraSpec := apiv2.CassandraToInstAPI(&cc.Spec)
			id, err = r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandraSpec)
			if err != nil {
				l.Error(
					err, "Cannot create Cassandra cluster",
					"cassandra cluster spec", cc.Spec,
				)
				return models.ReconcileRequeue
			}
		}

		cc.Status.ID = id
		err = r.Status().Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra cluster status",
				"cluster name", cc.Spec.Name,
				"cluster ID", cc.Status.ID,
				"kind", cc.Kind,
				"api Version", cc.APIVersion,
				"namespace", cc.Namespace,
				"cluster metadata", cc.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(cc, models.DeletionFinalizer)
		cc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		cc.Annotations[models.DeletionConfirmed] = models.False
		err = r.Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra cluster",
				"cluster name", cc.Spec.Name,
				"cluster ID", cc.Status.ID,
				"kind", cc.Kind,
				"api Version", cc.APIVersion,
				"namespace", cc.Namespace,
				"cluster metadata", cc.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"Cassandra cluster has been created",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
			"kind", cc.Kind,
			"api Version", cc.APIVersion,
			"namespace", cc.Namespace,
		)
	}

	err = r.startClusterStatusJob(cc)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"cassandra cluster ID", cc.Status.ID)
		return models.ReconcileRequeue
	}

	err = r.startClusterBackupsJob(cc)
	if err != nil {
		l.Error(err, "Cannot start Cassandra cluster backups check job",
			"cluster ID", cc.Status.ID,
		)

		return models.ReconcileRequeue
	}

	return models.ReconcileResult
}

func (r *CassandraReconciler) handleUpdateCluster(
	ctx context.Context,
	l logr.Logger,
	cc *clustersv1alpha1.Cassandra,
) reconcile.Result {
	currentClusterStatus, err := r.API.GetClusterStatus(cc.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get Cassandra cluster status from the Instaclustr API",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
		)
		return models.ReconcileRequeue
	}

	result := apiv2.CompareCassandraDCs(cc.Spec.DataCentres, currentClusterStatus)
	if result != nil {
		err = r.API.UpdateCluster(cc.Status.ID,
			instaclustr.CassandraEndpoint,
			result,
		)
		if errors.Is(err, instaclustr.ClusterIsNotReadyToResize) {
			l.Error(err, "Cluster is not ready to resize",
				"cluster name", cc.Spec.Name,
				"cluster status", cc.Status,
			)
			return models.ReconcileRequeue
		}
		if err != nil {
			l.Error(
				err, "Cannot update Cassandra cluster status from the Instaclustr API",
				"cluster name", cc.Spec.Name,
				"cluster ID", cc.Status.ID,
			)
			return models.ReconcileRequeue
		}
	}
	patch := cc.NewPatch()
	cc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, cc, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cassandra cluster",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
			"kind", cc.Kind,
			"api Version", cc.APIVersion,
			"namespace", cc.Namespace,
			"cluster metadata", cc.ObjectMeta,
		)
		return models.ReconcileRequeue
	}
	l.Info(
		"Cassandra cluster status has been updated",
		"cluster name", cc.Spec.Name,
		"cluster ID", cc.Status.ID,
		"kind", cc.Kind,
		"api Version", cc.APIVersion,
		"namespace", cc.Namespace,
	)

	return reconcile.Result{}
}

func (r *CassandraReconciler) handleDeleteCluster(
	ctx context.Context,
	l logr.Logger,
	cc *clustersv1alpha1.Cassandra,
) reconcile.Result {
	patch := cc.NewPatch()
	err := r.Patch(ctx, cc, patch)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot patch Cassandra cluster",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
			"kind", cc.Kind,
			"api Version", cc.APIVersion,
			"namespace", cc.Namespace,
			"cluster metadata", cc.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	if len(cc.Spec.TwoFactorDelete) != 0 &&
		cc.Annotations[models.DeletionConfirmed] != models.True {
		l.Info("Cassandra cluster deletion is not confirmed",
			"cluster ID", cc.Status.ID,
			"cluster name", cc.Spec.Name,
		)

		cc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.Status().Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra cluster metadata after finalizer removal",
				"cluster name", cc.Spec.Name,
				"cluster ID", cc.Status.ID,
			)

			return models.ReconcileRequeue
		}

		return models.ReconcileResult
	}

	status, err := r.API.GetClusterStatus(cc.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get Cassandra cluster status from the Instaclustr API",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
			"kind", cc.Kind,
			"api Version", cc.APIVersion,
			"namespace", cc.Namespace,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		r.Scheduler.RemoveJob(cc.GetJobID(scheduler.StatusChecker))
		err = r.API.DeleteCluster(cc.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Cassandra cluster",
				"cluster name", cc.Spec.Name,
				"status", cc.Status.Status,
				"kind", cc.Kind,
				"api Version", cc.APIVersion,
				"namespace", cc.Namespace,
			)
			return models.ReconcileRequeue
		}

		l.Info("Cassandra cluster is being deleted",
			"cluster ID", cc.Status.ID,
		)

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(cc.GetJobID(scheduler.BackupsChecker))

	l.Info("Deleting cluster backup resources",
		"cluster ID", cc.Status.ID,
	)

	err = r.deleteBackups(ctx, cc.Status.ID, cc.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete cluster backup resources",
			"cluster ID", cc.Status.ID,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster backup resources were deleted",
		"cluster ID", cc.Status.ID,
	)

	controllerutil.RemoveFinalizer(cc, models.DeletionFinalizer)
	cc.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent

	err = r.Patch(ctx, cc, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cassandra cluster",
			"cluster name", cc.Spec.Name,
			"cluster ID", cc.Status.ID,
			"kind", cc.Kind,
			"api Version", cc.APIVersion,
			"namespace", cc.Namespace,
			"cluster metadata", cc.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cassandra cluster has been deleted",
		"cluster name", cc.Spec.Name,
		"cluster ID", cc.Status.ID,
		"kind", cc.Kind,
		"api Version", cc.APIVersion,
		"namespace", cc.Namespace,
	)

	return reconcile.Result{}
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

func (r *CassandraReconciler) newWatchStatusJob(cassandraCluster *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraStatusClusterJob")
	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{Namespace: cassandraCluster.Namespace, Name: cassandraCluster.Name}, cassandraCluster)
		if err != nil {
			l.Error(err, "Cannot get Cassandra custom resource",
				"resource name", cassandraCluster.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(cassandraCluster.DeletionTimestamp, len(cassandraCluster.Spec.TwoFactorDelete), cassandraCluster.Annotations[models.DeletionConfirmed]) {
			l.Info("Cassandra cluster is being deleted. Status check job skipped",
				"cluster name", cassandraCluster.Spec.Name,
				"cluster ID", cassandraCluster.Status.ID,
			)

			return nil
		}

		instaclusterStatus, err := r.API.GetClusterStatus(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "Cannot get cassandraCluster instaclusterStatus",
				"clusterID", cassandraCluster.Status.ID)
			return err
		}

		if !areStatusesEqual(instaclusterStatus, &cassandraCluster.Status.ClusterStatus) {
			l.Info("Cassandra status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclusterStatus", instaclusterStatus,
				"cassandraCluster.Status.ClusterStatus", cassandraCluster.Status.ClusterStatus)

			patch := cassandraCluster.NewPatch()
			instaclusterStatus.MaintenanceEvents = cassandraCluster.Status.ClusterStatus.MaintenanceEvents
			cassandraCluster.Status.ClusterStatus = *instaclusterStatus
			err = r.Status().Patch(context.Background(), cassandraCluster, patch)
			if err != nil {
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(cassandraCluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cassandra cluster maintenance events",
				"cluster name", cassandraCluster.Spec.Name,
				"cluster ID", cassandraCluster.Status.ID,
			)

			return err
		}

		if !cassandraCluster.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := cassandraCluster.NewPatch()
			cassandraCluster.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), cassandraCluster, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cassandra cluster maintenance events",
					"cluster name", cassandraCluster.Spec.Name,
					"cluster ID", cassandraCluster.Status.ID,
				)

				return err
			}

			l.Info("Cassandra cluster maintenance events were updated",
				"cluster ID", cassandraCluster.Status.ID,
				"events", cassandraCluster.Status.MaintenanceEvents,
			)
		}

		if instaclusterStatus.CurrentClusterOperationStatus == models.NoOperation {
			instSpec, err := r.API.GetCassandra(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
			if err != nil {
				l.Error(err, "Cannot get Cassandra cluster spec from Instaclustr API",
					"cluster ID", cassandraCluster.Status.ID)

				return err
			}

			if !cassandraCluster.Spec.AreSpecsEqual(instSpec) {
				patch := cassandraCluster.NewPatch()
				cassandraCluster.Spec.RestoreFrom = nil
				cassandraCluster.Spec.SetSpecFromInst(instSpec)
				err = r.Patch(context.Background(), cassandraCluster, patch)
				if err != nil {
					l.Error(err, "Cannot patch Cassandra cluster spec",
						"cluster ID", cassandraCluster.Status.ID,
						"spec from Instaclustr API", instSpec,
					)

					return err
				}

				l.Info("Cassandra cluster spec has been updated",
					"cluster ID", cassandraCluster.Status.ID,
				)
			}
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
			l.Info("Cassandra cluster is being deleted. Backups check job skipped",
				"cluster name", cluster.Spec.Name,
				"cluster ID", cluster.Status.ID,
			)

			return nil
		}

		instBackups, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, cluster.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cassandra cluster backups",
				"cluster name", cluster.Spec.Name,
				"clusterID", cluster.Status.ID,
			)

			return err
		}

		instBackupEvents := instBackups.GetBackupEvents(models.CassandraKind)

		k8sBackupList, err := r.listClusterBackups(ctx, cluster.Status.ID, cluster.Namespace)
		if err != nil {
			l.Error(err, "Cannot list Cassandra cluster backups",
				"cluster name", cluster.Spec.Name,
				"clusterID", cluster.Status.ID,
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
