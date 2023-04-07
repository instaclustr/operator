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
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	virtcorev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
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
	"github.com/instaclustr/operator/pkg/exposeservice"
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
	OnPremisesCfg models.CassandraOnPremisesConfig
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;delete;watch;deletecollection;patch
//+kubebuilder:rbac:groups=*,resources=endpoints,verbs=get;list;watch;create;update;patch;delete

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

	patch := cassandra.NewPatch()
	err = r.Client.Patch(ctx, cassandra, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cassandra resource",
			"name", cassandra.Name)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ExitReconcile, nil
	}

	switch cassandra.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		if cassandra.Spec.OnPremisesSpec != nil {
			return r.handleCreateOnPremises(ctx, l, cassandra), nil
		}
		return r.handleCreateCluster(ctx, l, cassandra), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, l, cassandra), nil
	case models.DeletingEvent:
		if cassandra.Spec.OnPremisesSpec != nil {
			return r.handleDeleteOnPremises(ctx, l, cassandra), nil
		}
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
	if cassandra.Status.ID == "" {
		var id string
		patch := cassandra.NewPatch()

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

	if cassandra.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(cassandra, iCassandra, l)
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

func (r *CassandraReconciler) handleExternalChanges(cassandra, iCassandra *clustersv1alpha1.Cassandra, l logr.Logger) reconcile.Result {
	if cassandra.Annotations[models.AllowSpecAmendAnnotation] != models.True {
		l.Info("Update is blocked until k8s resource specification is equal with Instaclustr",
			"specification of k8s resource", cassandra.Spec,
			"data from Instaclustr ", iCassandra.Spec)

		r.EventRecorder.Event(cassandra, models.Warning, models.UpdateFailed,
			"There are external changes on the Instaclustr console. Please reconcile the specification manually")

		return models.ExitReconcile
	} else {
		if !cassandra.Spec.IsEqual(iCassandra.Spec) {
			l.Info(msgSpecStillNoMatch,
				"specification of k8s resource", cassandra.Spec,
				"data from Instaclustr ", iCassandra.Spec)
			r.EventRecorder.Event(cassandra, models.Warning, models.ExternalChanges, msgSpecStillNoMatch)

			return models.ExitReconcile
		}

		patch := cassandra.NewPatch()

		cassandra.Annotations[models.ExternalChangesAnnotation] = ""
		cassandra.Annotations[models.AllowSpecAmendAnnotation] = ""

		err := r.Patch(context.Background(), cassandra, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster resource",
				"cluster name", cassandra.Spec.Name, "cluster ID", cassandra.Status.ID)

			r.EventRecorder.Eventf(cassandra, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue
		}

		l.Info("External changes have been reconciled", "resource ID", cassandra.Status.ID)
		r.EventRecorder.Event(cassandra, models.Normal, models.ExternalChanges, "External changes have been reconciled")

		return models.ExitReconcile
	}
}

func (r *CassandraReconciler) handleDeleteCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	l = l.WithName("Cassandra deletion event")

	_, err := r.API.GetCassandra(cassandra.Status.ID)
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

	patch := cassandra.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID)

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

		r.EventRecorder.Event(cassandra, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if cassandra.Spec.TwoFactorDelete != nil {
			cassandra.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			cassandra.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, cassandra, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", cassandra.Spec.Name,
					"cluster state", cassandra.Status.State)
				r.EventRecorder.Eventf(cassandra, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", cassandra.Status.ID)

			r.EventRecorder.Event(cassandra, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
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

	err = exposeservice.Delete(r.Client, cassandra.Name, cassandra.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Cassandra cluster expose service",
			"cluster ID", cassandra.Status.ID,
			"cluster name", cassandra.Spec.Name,
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

func (r *CassandraReconciler) handleDeleteOnPremises(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	l = l.WithName("On-premise Cassandra deletion event")

	_, err := r.API.GetCassandraV1(cassandra.Status.ID)
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

		l.Info("Cluster deletion request has been sent",
			"cluster ID", cassandra.Status.ID,
		)

		if cassandra.Spec.TwoFactorDelete != nil {
			patch := cassandra.NewPatch()
			cassandra.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			cassandra.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, cassandra, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", cassandra.Spec.Name,
					"cluster state", cassandra.Status.State)
				r.EventRecorder.Eventf(cassandra, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", cassandra.Status.ID)

			r.EventRecorder.Event(cassandra, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
	}

	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.OnPremisesStatusChecker))
	l.Info("On-premises status checker job has been removed",
		"cluster ID", cassandra.Status.ID,
	)
	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.OnPremisesIPsChecker))
	l.Info("On-premises IPs checker job has been removed",
		"cluster ID", cassandra.Status.ID,
	)
	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.BackupsChecker))
	l.Info("Backups checker job has been removed",
		"cluster ID", cassandra.Status.ID,
	)

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

	l.Info("Deleting cluster on-premises resources",
		"cluster ID", cassandra.Status.ID,
	)

	err = r.deleteOnPremResources(ctx, cassandra.Status.ID, cassandra.Namespace, cassandra.Spec.OnPremisesSpec.DeleteDisksWithVM)
	if err != nil {
		l.Error(err, "Cannot delete cluster on-premises resources",
			"cluster ID", cassandra.Status.ID,
		)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.DeletionFailed,
			"Cluster on-premises resources deletion is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster on-premises resources are deleted",
		"cluster ID", cassandra.Status.ID,
	)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Deleted,
		"Cluster on-premises resources are deleted",
	)

	patch := cassandra.NewPatch()
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
		"Cluster resource has been deleted",
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

func (r *CassandraReconciler) startClusterOnPremisesStatusJob(cluster *clustersv1alpha1.Cassandra) error {
	job := r.newWatchOnPremisesStatusJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.OnPremisesStatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) startClusterOnPremisesIPsJob(cluster *clustersv1alpha1.Cassandra) error {
	job := r.newWatchOnPremisesIPsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) newWatchStatusJob(cassandra *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "CassandraStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(cassandra)
		err := r.Get(context.Background(), namespacedName, cassandra)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", namespacedName)
				r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.BackupsChecker))
				r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.StatusChecker))
				return nil
			}

			l.Error(err, "Cannot get resource in kubernetes",
				"namespaced name", namespacedName)

			return err
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
					cassandra.Annotations[models.ClusterDeletionAnnotation] = ""
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
				"status from Instaclustr", iCassandra.Status.ClusterStatus,
				"status from k8s", cassandra.Status.ClusterStatus)

			areDCsEqual := areDataCentresEqual(iCassandra.Status.ClusterStatus.DataCentres, cassandra.Status.ClusterStatus.DataCentres)

			patch := cassandra.NewPatch()
			cassandra.Status.ClusterStatus = iCassandra.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), cassandra, patch)
			if err != nil {
				return err
			}

			if !areDCsEqual {
				var nodes []*clustersv1alpha1.Node

				for _, dc := range iCassandra.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					cassandra.Name,
					cassandra.Namespace,
					nodes,
					models.CassandraConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iCassandra.Status.CurrentClusterOperationStatus == models.NoOperation &&
			cassandra.Annotations[models.ExternalChangesAnnotation] != models.True &&
			!cassandra.Spec.IsEqual(iCassandra.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iCassandra.Spec, "k8s resource spec", cassandra.Spec)

			patch := cassandra.NewPatch()
			cassandra.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), cassandra, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", cassandra.Spec.Name, "cluster state", cassandra.Status.State)
				return err
			}

			r.EventRecorder.Event(cassandra, models.Warning, models.ExternalChanges,
				"There are external changes on the Instaclustr console. Please reconcile the specification manually")
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
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
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

func (r *CassandraReconciler) newWatchOnPremisesStatusJob(c *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraOnPremStatusClusterJob")

	return func() error {
		ctx := context.Background()
		err := r.Get(ctx, client.ObjectKeyFromObject(c), c)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", client.ObjectKeyFromObject(c))
				r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesStatusChecker))
				r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesIPsChecker))
				r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
				return nil
			}

			l.Error(err, "Cannot get Cassandra cluster from kubernetes",
				"resource key", client.ObjectKeyFromObject(c),
				"cluster name", c.Spec.Name,
			)
			r.EventRecorder.Eventf(
				c, models.Warning, models.FetchFailed,
				"Cassandra cluster fetch from the kubernetes is failed. Reason: %v",
				err,
			)
			return err
		}

		iData, err := r.API.GetCassandraV1(c.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(c.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", c.Status.ClusterStatus.ID,
						"cluster name", c.Spec.Name,
					)

					patch := c.NewPatch()
					c.Annotations[models.ClusterDeletionAnnotation] = ""
					err = r.Patch(context.TODO(), c, patch)
					if err != nil {
						l.Error(err, "Cannot patch Cassandra cluster resource",
							"cluster ID", c.Status.ID,
							"cluster name", c.Spec.Name,
							"resource name", c.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), c)
					if err != nil {
						l.Error(err, "Cannot delete Cassandra cluster resource",
							"cluster ID", c.Status.ID,
							"cluster name", c.Spec.Name,
							"resource name", c.Name,
						)

						return err
					}

					r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesStatusChecker))
					r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesIPsChecker))
					r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))

					return nil
				}
			}

			l.Error(err, "Cannot get Cassandra from the Instaclustr API",
				"cluster name", c.Spec.Name,
				"status", c.Status)
			r.EventRecorder.Eventf(
				c, models.Warning, models.FetchFailed,
				"Cassandra cluster fetch from the Instaclustr API is failed. Reason: %v",
				err,
			)
			return err
		}

		iCass := &clustersv1alpha1.Cassandra{}
		iCass.Status, err = c.Status.FromInstAPIv1(iData)
		if err != nil {
			l.Error(err, "Cannot convert cluster from the Instaclustr API",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)
			return err
		}

		if !areStatusesEqual(&iCass.Status.ClusterStatus, &c.Status.ClusterStatus) {
			l.Info("Updating cluster status",
				"status from Instaclustr", iCass.Status.ClusterStatus,
				"status from k8s", c.Status.ClusterStatus)

			patch := c.NewPatch()
			c.Status.ClusterStatus = iCass.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), c, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *CassandraReconciler) newWatchOnPremisesIPsJob(c *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraOnPremStatusClusterJob")

	return func() error {
		iData, err := r.API.GetCassandraV1(c.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cassandra from the Instaclustr API",
				"cluster name", c.Spec.Name,
				"status", c.Status)
			r.EventRecorder.Eventf(
				c, models.Warning, models.FetchFailed,
				"Cassandra cluster fetch from the Instaclustr API is failed. Reason: %v",
				err,
			)
			return err
		}

		iCass := &clustersv1alpha1.Cassandra{}
		iCass.Status, err = c.Status.FromInstAPIv1(iData)
		if err != nil {
			l.Error(err, "Cannot convert cluster from the Instaclustr API",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)
			return err
		}

		iNodes := iCass.Status.FetchNodes()
		if !r.areIPsSet(iNodes, c.Spec.PrivateNetworkCluster) {
			l.Info("Cassandra nodes IPs are not set",
				"cluster name", c.Spec.Name)
			r.EventRecorder.Event(
				c,
				models.Normal,
				models.IPModifyNeeded,
				models.MsgIPModifyNeeded,
			)
			return nil
		}

		r.EventRecorder.Eventf(
			c, models.Normal, models.Created,
			"Nodes IPs are set",
		)

		r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesIPsChecker))

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

func (r *CassandraReconciler) handleCreateOnPremises(
	ctx context.Context,
	l logr.Logger,
	cassandra *clustersv1alpha1.Cassandra,
) reconcile.Result {
	var err error
	if cassandra.Status.ID == "" {
		l.Info("Cassandra cluster ID is empty. Sending cluster creation request...",
			"cluster name", cassandra.Spec.Name)
		patch := cassandra.NewPatch()

		iCass := cassandra.Spec.OnPremisesToInstAPIv1()
		cassandra.Status.ID, err = r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, iCass)
		if err != nil {
			l.Error(err, "Cannot create Cassandra cluster on the Instaclustr API",
				"cluster name", cassandra.Spec.Name,
				"spec", cassandra.Spec)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.CreationFailed,
				"Cassandra cluster creation on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info("Cassandra cluster creation request is sent",
			"cluster name", cassandra.Spec.Name)

		r.EventRecorder.Eventf(
			cassandra, models.Normal, models.Created,
			"Cassandra cluster creation request is sent",
		)

		err = r.Client.Status().Patch(ctx, cassandra, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cassandra resource status patch",
				"cluster name", cassandra.Spec.Name,
				"status", cassandra.Status)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		controllerutil.AddFinalizer(cassandra, models.DeletionFinalizer)
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

	}

	l.Info("Creating Cassandra on-premises resources...",
		"cluster name", cassandra.Spec.Name)

	err = r.createOnPremisesResources(ctx, cassandra)
	if err != nil {
		l.Error(err, "Cannot create Cassandra on-premises resources",
			"cluster name", cassandra.Spec.Name,
			"status", cassandra.Status)
		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.CreationFailed,
			"On-premises resources creation is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cassandra on-premises resources are created",
		"cluster name", cassandra.Spec.Name)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Created,
		"Cassandra on-premises resources are created",
	)

	patch := cassandra.NewPatch()
	cassandra.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
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

	err = r.startClusterOnPremisesIPsJob(cassandra)
	if err != nil {
		l.Error(err, "Cannot start cluster IPs job",
			"cassandra cluster ID", cassandra.Status.ID)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.CreationFailed,
			"Cluster IPs check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Created,
		"Cluster on-premises IPs check job is started",
	)

	err = r.startClusterOnPremisesStatusJob(cassandra)
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
		"Cluster on-premises status check job is started",
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

func (r *CassandraReconciler) areIPsSet(nodes []*clustersv1alpha1.Node, privateCluster bool) bool {
	for _, node := range nodes {
		if node.PrivateAddress == "" ||
			(node.PublicAddress == "" && !privateCluster) {
			return false
		}
	}

	return true
}

func (r *CassandraReconciler) createOnPremisesResources(ctx context.Context, cassandraCluster *clustersv1alpha1.Cassandra) error {
	cloudInitScriptSecret := &k8scorev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cassandraCluster.Namespace,
		Name:      models.CloudInitSecretName,
	}, cloudInitScriptSecret)
	if err != nil {
		return err
	}

	if cassandraCluster.Spec.PrivateNetworkCluster {
		err = r.createSSHGatewayResources(ctx, cassandraCluster)
		if err != nil {
			return err
		}
	}

	err = r.createNodesResources(ctx, cassandraCluster)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) createSSHGatewayResources(ctx context.Context, c *clustersv1alpha1.Cassandra) error {
	sshIgnitionSecretName := c.Spec.OnPremisesSpec.IgnitionScriptsSecretNames[models.SSH]
	sshIgnitionScriptSecret := &k8scorev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      sshIgnitionSecretName,
	}, sshIgnitionScriptSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.EventRecorder.Event(
				c,
				models.Normal,
				models.IgnitionScriptNeeded,
				models.MsgIgnitionScriptNeeded,
			)
			return err
		}

		return err
	}

	if len(sshIgnitionScriptSecret.Data[models.IgnitionSecretData]) == 0 {
		r.EventRecorder.Event(
			c,
			models.Normal,
			models.IgnitionScriptNeeded,
			models.MsgIgnitionScriptNeeded,
		)
		return models.ErrEmptyIgnitionScript
	}

	sshDVSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.SystemDiskSize)
	if err != nil {
		return err
	}

	sshDVName := fmt.Sprintf("%s-%s", models.SSHDVPrefix, c.Name)
	sshPVC, err := r.providePVC(ctx, c, sshDVName, sshDVSize, true)
	if err != nil {
		return err
	}

	patch := client.MergeFrom(sshPVC.DeepCopy())
	sshPVC.Labels[models.ClusterIDLabel] = c.Status.ID
	err = r.Patch(ctx, sshPVC, patch)
	if err != nil {
		return err
	}

	sshVMCPU := resource.Quantity{}
	sshVMCPU.Set(c.Spec.OnPremisesSpec.SSHGatewayCPU)
	sshVMMemory, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.SSHGatewayMemory)
	if err != nil {
		return err
	}

	sshVMName := fmt.Sprintf("%s-%s", models.SSHVMPrefix, c.Name)
	sshVM := &virtcorev1.VirtualMachine{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      sshVMName,
	}, sshVM)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if k8serrors.IsNotFound(err) {
		sshVM = r.newVM(
			c.Status.ID,
			sshVMName,
			c.Namespace,
			sshDVName,
			sshIgnitionSecretName,
			sshVMCPU,
			sshVMMemory)
		err = r.Client.Create(ctx, sshVM)
		if err != nil {
			return err
		}
	}

	sshExposeName := fmt.Sprintf("%s-%s", models.SSHSvcPrefix, sshVMName)
	sshExposeService := &k8scorev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      sshExposeName,
	}, sshExposeService)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if k8serrors.IsNotFound(err) {
		sshExposeService = r.newExposeService(c.Status.ID, sshExposeName, c.Namespace)
		err = r.Client.Create(ctx, sshExposeService)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *CassandraReconciler) createNodesResources(ctx context.Context, c *clustersv1alpha1.Cassandra) error {
	for i := 1; i <= c.Spec.OnPremisesSpec.NodesNumber; i++ {
		nodeIgnitionSecretName := fmt.Sprintf(models.NodeIgnitionSecretFormat, i)
		nodeIgnitionScriptSecret := &k8scorev1.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      nodeIgnitionSecretName,
		}, nodeIgnitionScriptSecret)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				r.EventRecorder.Event(
					c,
					models.Normal,
					models.IgnitionScriptNeeded,
					models.MsgIgnitionScriptNeeded,
				)
				return err
			}

			return err
		}

		if len(nodeIgnitionScriptSecret.Data[models.IgnitionSecretData]) == 0 {
			r.EventRecorder.Event(
				c,
				models.Normal,
				models.IgnitionScriptNeeded,
				models.MsgIgnitionScriptNeeded,
			)
			return models.ErrEmptyIgnitionScript
		}

		nodeDVSysSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.SystemDiskSize)
		if err != nil {
			return err
		}

		nodeSysDVName := fmt.Sprintf("%s-%d-%s", models.NodeSysDVPrefix, i, c.Name)
		nodeSysPVC, err := r.providePVC(ctx, c, nodeSysDVName, nodeDVSysSize, true)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(nodeSysPVC.DeepCopy())
		nodeSysPVC.Labels[models.ClusterIDLabel] = c.Status.ID
		err = r.Patch(ctx, nodeSysPVC, patch)
		if err != nil {
			return err
		}

		nodeDVStorageSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.DataDiskSize)
		if err != nil {
			return err
		}

		nodeStorageDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, c.Name)
		nodeStoragePVC, err := r.providePVC(ctx, c, nodeStorageDVName, nodeDVStorageSize, false)
		if err != nil {
			return err
		}

		patch = client.MergeFrom(nodeStoragePVC.DeepCopy())
		nodeStoragePVC.Labels[models.ClusterIDLabel] = c.Status.ID
		err = r.Patch(ctx, nodeStoragePVC, patch)
		if err != nil {
			return err
		}

		nodeVMCPU := resource.Quantity{}
		nodeVMCPU.Set(c.Spec.OnPremisesSpec.NodeCPU)
		nodeVMMemory, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.NodeMemory)
		if err != nil {
			return err
		}

		nodeVMName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, c.Name)
		nodeVM := &virtcorev1.VirtualMachine{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      nodeVMName,
		}, nodeVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			nodeVM = r.newVM(
				c.Status.ID,
				nodeVMName,
				c.Namespace,
				nodeSysDVName,
				nodeIgnitionSecretName,
				nodeVMCPU,
				nodeVMMemory,
				nodeStorageDVName)
			err = r.Client.Create(ctx, nodeVM)
			if err != nil {
				return err
			}
		}

		if !c.Spec.PrivateNetworkCluster {
			nodeExposeName := fmt.Sprintf("%s-%s", models.NodeSvcPrefix, nodeVMName)
			nodeExposeService := &k8scorev1.Service{}
			err = r.Get(ctx, types.NamespacedName{
				Namespace: c.Namespace,
				Name:      nodeExposeName,
			}, nodeExposeService)
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			if k8serrors.IsNotFound(err) {
				nodeExposeService = r.newExposeService(c.Status.ID, nodeExposeName, c.Namespace)
				err = r.Client.Create(ctx, nodeExposeService)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Tries to get PVC and DV and creates it if not found
func (r *CassandraReconciler) providePVC(ctx context.Context, c *clustersv1alpha1.Cassandra, name string, size resource.Quantity, isSysDV bool) (*k8scorev1.PersistentVolumeClaim, error) {
	pvc := &k8scorev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      name,
	}, pvc)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if k8serrors.IsNotFound(err) {
		dv := &cdiv1beta1.DataVolume{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      name,
		}, dv)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if k8serrors.IsNotFound(err) {
			if isSysDV {
				dv = r.newSysDV(c, name, size)
			} else {
				dv = r.newStorageDV(c, name, size)
			}
			err = r.Client.Create(ctx, dv)
			if err != nil {
				return nil, err
			}
		}

		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      name,
		}, pvc)
		if err != nil {
			return nil, err
		}
	}
	return pvc, nil
}

func (r *CassandraReconciler) newStorageDV(c *clustersv1alpha1.Cassandra, name string, dataStorage resource.Quantity) *cdiv1beta1.DataVolume {
	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    map[string]string{models.ClusterIDLabel: c.Status.ID},
		},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: &cdiv1beta1.DataVolumeSource{
				Blank: &cdiv1beta1.DataVolumeBlankImage{},
			},
			PVC: &k8scorev1.PersistentVolumeClaimSpec{
				AccessModes: []k8scorev1.PersistentVolumeAccessMode{
					k8scorev1.ReadWriteOnce,
				},
				Resources: k8scorev1.ResourceRequirements{
					Requests: k8scorev1.ResourceList{
						models.Storage: dataStorage,
					},
				},
				StorageClassName: &c.Spec.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

func (r *CassandraReconciler) newSysDV(c *clustersv1alpha1.Cassandra, name string, dataStorage resource.Quantity) *cdiv1beta1.DataVolume {
	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels:    map[string]string{models.ClusterIDLabel: c.Status.ID},
		},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: &cdiv1beta1.DataVolumeSource{
				HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
					URL: r.OnPremisesCfg.DataVolumeImageURL,
				},
			},
			PVC: &k8scorev1.PersistentVolumeClaimSpec{
				AccessModes: []k8scorev1.PersistentVolumeAccessMode{
					k8scorev1.ReadWriteOnce,
				},
				Resources: k8scorev1.ResourceRequirements{
					Requests: k8scorev1.ResourceList{
						models.Storage: dataStorage,
					},
				},
				StorageClassName: &c.Spec.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

func (r *CassandraReconciler) newVM(
	clusterID,
	vmName,
	namespace,
	sysDVName,
	ignitionSecretName string,
	cpu,
	memory resource.Quantity,
	storageDVNames ...string) *virtcorev1.VirtualMachine {
	vm := &virtcorev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.VirtualMachineKind,
			APIVersion: models.KubevirtV1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       vmName,
			Namespace:  namespace,
			Labels:     map[string]string{models.ClusterIDLabel: clusterID},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: virtcorev1.VirtualMachineSpec{
			Running: &models.TrueBool,
			Template: &virtcorev1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						models.KubevirtDomainLabel: vmName,
					},
				},
				Spec: virtcorev1.VirtualMachineInstanceSpec{
					Domain: virtcorev1.DomainSpec{
						Resources: virtcorev1.ResourceRequirements{
							Requests: k8scorev1.ResourceList{
								models.CPU:    cpu,
								models.Memory: memory,
							},
						},
						Devices: virtcorev1.Devices{
							Disks: []virtcorev1.Disk{
								{
									Name:      models.DVDisk,
									BootOrder: &models.BootOrder1,
									DiskDevice: virtcorev1.DiskDevice{
										Disk: &virtcorev1.DiskTarget{
											Bus: models.SATA,
										},
									},
								},
								{
									Name:       models.IgnitionDisk,
									DiskDevice: virtcorev1.DiskDevice{},
									Serial:     models.IgnitionSerial,
								},
							},
							Interfaces: []virtcorev1.Interface{
								{
									Name: models.Default,
									InterfaceBindingMethod: virtcorev1.InterfaceBindingMethod{
										Bridge: &virtcorev1.InterfaceBridge{},
									},
								},
							},
						},
					},
					Volumes: []virtcorev1.Volume{
						{
							Name: models.DVDisk,
							VolumeSource: virtcorev1.VolumeSource{
								PersistentVolumeClaim: &virtcorev1.PersistentVolumeClaimVolumeSource{
									PersistentVolumeClaimVolumeSource: k8scorev1.PersistentVolumeClaimVolumeSource{
										ClaimName: sysDVName,
									},
								},
							},
						},
						{
							Name: models.CloudInitDisk,
							VolumeSource: virtcorev1.VolumeSource{
								CloudInitNoCloud: &virtcorev1.CloudInitNoCloudSource{
									UserDataSecretRef: &k8scorev1.LocalObjectReference{
										Name: models.CloudInitSecretName,
									},
								},
							},
						},
						{
							Name: models.IgnitionDisk,
							VolumeSource: virtcorev1.VolumeSource{
								Secret: &virtcorev1.SecretVolumeSource{
									SecretName: ignitionSecretName,
								},
							},
						},
					},
					Networks: []virtcorev1.Network{
						{
							Name: models.Default,
							NetworkSource: virtcorev1.NetworkSource{
								Pod: &virtcorev1.PodNetwork{},
							},
						},
					},
				},
			},
		},
	}

	for i, dvName := range storageDVNames {
		diskName := fmt.Sprintf("%s-%d", models.DataDisk, i)
		vm.Spec.Template.Spec.Domain.Devices.Disks = append(vm.Spec.Template.Spec.Domain.Devices.Disks, virtcorev1.Disk{
			Name: diskName,
			DiskDevice: virtcorev1.DiskDevice{
				Disk: &virtcorev1.DiskTarget{
					Bus: models.SATA,
				},
			},
			Serial: models.DataDiskSerial,
		})
		vm.Spec.Template.Spec.Volumes = append(vm.Spec.Template.Spec.Volumes, virtcorev1.Volume{
			Name: diskName,
			VolumeSource: virtcorev1.VolumeSource{
				PersistentVolumeClaim: &virtcorev1.PersistentVolumeClaimVolumeSource{
					PersistentVolumeClaimVolumeSource: k8scorev1.PersistentVolumeClaimVolumeSource{
						ClaimName: dvName,
					},
				},
			},
		})
	}

	return vm
}

func (r *CassandraReconciler) newExposeService(clusterID, name, namespace string) *k8scorev1.Service {
	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Labels:     map[string]string{models.ClusterIDLabel: clusterID},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Port: models.Port22,
				},
			},
			Selector: map[string]string{
				models.KubevirtDomainLabel: name,
			},
			Type: models.LBType,
		},
	}
}

func (r *CassandraReconciler) deleteOnPremResources(ctx context.Context, clusterID, namespace string, deleteDisks bool) error {
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	delOpts := []client.DeleteAllOfOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}

	vmType := &virtcorev1.VirtualMachine{}
	err := r.DeleteAllOf(ctx, vmType, delOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	vmList := &virtcorev1.VirtualMachineList{}
	err = r.List(ctx, vmList, listOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	for _, vm := range vmList.Items {
		patch := client.MergeFrom(vm.DeepCopy())
		controllerutil.RemoveFinalizer(&vm, models.DeletionFinalizer)
		err = r.Patch(ctx, &vm, patch)
		if err != nil {
			return err
		}
	}

	dvType := &cdiv1beta1.DataVolume{}
	err = r.DeleteAllOf(ctx, dvType, delOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	dvList := &cdiv1beta1.DataVolumeList{}
	err = r.List(ctx, dvList, listOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	for _, dv := range dvList.Items {
		patch := client.MergeFrom(dv.DeepCopy())
		controllerutil.RemoveFinalizer(&dv, models.DeletionFinalizer)
		err = r.Patch(ctx, &dv, patch)
		if err != nil {
			return err
		}
	}

	svcType := &k8scorev1.Service{}
	err = r.DeleteAllOf(ctx, svcType, delOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	svcList := &k8scorev1.ServiceList{}
	err = r.List(ctx, svcList, listOpts...)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	for _, svc := range svcList.Items {
		patch := client.MergeFrom(svc.DeepCopy())
		controllerutil.RemoveFinalizer(&svc, models.DeletionFinalizer)
		err = r.Patch(ctx, &svc, patch)
		if err != nil {
			return err
		}
	}

	if deleteDisks {
		pvcType := &k8scorev1.PersistentVolumeClaim{}
		err = r.DeleteAllOf(ctx, pvcType, delOpts...)
		if client.IgnoreNotFound(err) != nil {
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

				newObj := event.ObjectNew.(*clustersv1alpha1.Cassandra)

				if newObj.Status.ID == "" || newObj.Annotations[models.ResourceStateAnnotation] == models.CreatingEvent {
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
