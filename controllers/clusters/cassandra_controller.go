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
	"strings"

	"github.com/go-logr/logr"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	virtcorev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
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

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	IcadminAPI    instaclustr.IcadminAPI
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+virtualmachineinstance.kubevirt.io/node-vm-2-cassandra-cluster:rbac:groups="",resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	cassandra := &v1beta1.Cassandra{}
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
		if cassandra.Spec.OnPremisesSpec != nil {
			return r.handleCreateOnPremisesCluster(ctx, l, cassandra)
		}
		return r.handleCreateCluster(ctx, l, cassandra)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, l, cassandra)
	case models.DeletingEvent:
		if cassandra.Spec.OnPremisesSpec != nil {
			return r.handleDeleteOnPremisesCluster(ctx, l, cassandra)
		}
		return r.handleDeleteCluster(ctx, l, cassandra)
	case models.GenericEvent:
		l.Info("Event isn't handled",
			"cluster name", cassandra.Spec.Name,
			"request", req,
			"event", cassandra.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	default:
		l.Info("Event isn't handled",
			"request", req,
			"event", cassandra.Annotations[models.ResourceStateAnnotation])

		return models.ExitReconcile, nil
	}
}

func (r *CassandraReconciler) handleCreateCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *v1beta1.Cassandra,
) (reconcile.Result, error) {
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

			id, err = r.API.RestoreCluster(cassandra.RestoreInfoToInstAPI(cassandra.Spec.RestoreFrom), models.CassandraAppKind)
			if err != nil {
				l.Error(err, "Cannot restore cluster from backup",
					"original cluster ID", cassandra.Spec.RestoreFrom.ClusterID,
				)

				r.EventRecorder.Eventf(
					cassandra, models.Warning, models.CreationFailed,
					"Cluster restore from backup on the Instaclustr is failed. Reason: %v",
					err,
				)

				return reconcile.Result{}, err
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
				return reconcile.Result{}, err
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
			return reconcile.Result{}, err
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
			return reconcile.Result{}, err
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

	if cassandra.Status.State != models.DeletedStatus {
		err = r.startClusterStatusJob(cassandra)
		if err != nil {
			l.Error(err, "Cannot start cluster status job",
				"cassandra cluster ID", cassandra.Status.ID)

			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.CreationFailed,
				"Cluster status check job is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
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
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			cassandra, models.Normal, models.Created,
			"Cluster backups check job is started",
		)

		if cassandra.Spec.UserRefs != nil {
			err = r.startUsersCreationJob(cassandra)
			if err != nil {
				l.Error(err, "Failed to start user creation job")
				r.EventRecorder.Eventf(cassandra, models.Warning, models.CreationFailed,
					"User creation job is failed. Reason: %v", err)
				return reconcile.Result{}, err
			}

			r.EventRecorder.Event(cassandra, models.Normal, models.Created,
				"Cluster user creation job is started")
		}
	}

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleCreateOnPremisesCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *v1beta1.Cassandra,
) (reconcile.Result, error) {
	l = l.WithName("On-premises Cassandra creation event")
	patch := cassandra.NewPatch()
	if cassandra.Status.ID == "" {
		l.Info(
			"Creating on-premises cluster",
			"cluster name", cassandra.Spec.Name,
			"data centres", cassandra.Spec.DataCentres,
		)

		id, err := r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandra.Spec.ToInstAPI())
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
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			cassandra, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s",
			id,
		)

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
			return reconcile.Result{}, err
		}

		cassandra.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
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
			return reconcile.Result{}, err
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
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			cassandra, models.Normal, models.Created,
			"Cluster status check job is started",
		)
	}

	if len(cassandra.Status.DataCentres) > 0 && cassandra.Status.State != models.RunningStatus {
		err := r.reconcileOnPremResources(ctx, cassandra)
		if err != nil {
			l.Error(
				err, "Cannot create resources for on-premises cluster",
				"cluster spec", cassandra.Spec.OnPremisesSpec,
			)
			r.EventRecorder.Eventf(
				cassandra, models.Warning, models.CreationFailed,
				"Resources creation for on-premises cluster is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		l.Info(
			"On-premises resources have been created",
			"cluster name", cassandra.Spec.Name,
			"on-premises Spec", cassandra.Spec.OnPremisesSpec,
			"cluster ID", cassandra.Status.ID,
		)

	} else {
		l.Info("Waiting for Data Centres provisioning...")
		return models.ReconcileRequeue, nil
	}

	controllerutil.AddFinalizer(cassandra, models.DeletionFinalizer)
	cassandra.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
	err := r.Patch(ctx, cassandra, patch)
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
		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been created",
		"cluster name", cassandra.Spec.Name,
		"cluster ID", cassandra.Status.ID,
		"kind", cassandra.Kind,
		"api Version", cassandra.APIVersion,
		"namespace", cassandra.Namespace,
	)

	err = r.startClusterOnPremisesIPsJob(cassandra)
	if err != nil {
		l.Error(err, "Cannot start cluster on-premises IPs job",
			"cassandra cluster ID", cassandra.Status.ID)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.CreationFailed,
			"Cluster on-premises IPs job is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}
	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleUpdateCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *v1beta1.Cassandra,
) (reconcile.Result, error) {
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
		return reconcile.Result{}, err
	}

	iCassandra, err := cassandra.FromInstAPI(iData)
	if err != nil {
		l.Error(
			err, "Cannot convert cluster from the Instaclustr API",
			"cluster name", cassandra.Spec.Name,
			"cluster ID", cassandra.Status.ID,
		)

		r.EventRecorder.Eventf(
			cassandra, models.Warning, models.ConversionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	if cassandra.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(cassandra, iCassandra, l)
	}

	patch := cassandra.NewPatch()

	if len(cassandra.Spec.TwoFactorDelete) != 0 && len(iCassandra.Spec.TwoFactorDelete) == 0 ||
		cassandra.Spec.Description != iCassandra.Spec.Description {
		l.Info("Updating cluster settings",
			"instaclustr description", iCassandra.Spec.Description,
			"instaclustr two factor delete", iCassandra.Spec.TwoFactorDelete)

		settingsToInstAPI, err := cassandra.Spec.ClusterSettingsUpdateToInstAPI()
		if err != nil {
			l.Error(err, "Cannot convert cluster settings to Instaclustr API",
				"cluster ID", cassandra.Status.ID,
				"cluster spec", cassandra.Spec)

			r.EventRecorder.Eventf(cassandra, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}

		err = r.API.UpdateClusterSettings(cassandra.Status.ID, settingsToInstAPI)
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", cassandra.Status.ID, "cluster spec", cassandra.Spec)

			r.EventRecorder.Eventf(cassandra, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !cassandra.Spec.AreDCsEqual(iCassandra.Spec.DataCentres) {
		l.Info("Update request to Instaclustr API has been sent",
			"spec data centres", cassandra.Spec.DataCentres,
			"resize settings", cassandra.Spec.ResizeSettings,
		)

		err = r.API.UpdateCassandra(cassandra.Status.ID, cassandra.Spec.DCsUpdateToInstAPI())
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

			if errors.Is(err, instaclustr.ClusterIsNotReadyToResize) {
				patch := cassandra.NewPatch()
				cassandra.Annotations[models.UpdateQueuedAnnotation] = models.True
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
					return reconcile.Result{}, err
				}
			}

			return reconcile.Result{}, err
		}
	}

	cassandra.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	cassandra.Annotations[models.UpdateQueuedAnnotation] = ""
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
		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", cassandra.Spec.Name,
		"cluster ID", cassandra.Status.ID,
		"data centres", cassandra.Spec.DataCentres,
	)

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleExternalChanges(cassandra, iCassandra *v1beta1.Cassandra, l logr.Logger) (reconcile.Result, error) {
	if !cassandra.Spec.IsEqual(iCassandra.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", cassandra.Spec,
			"data from Instaclustr ", iCassandra.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(cassandra.Spec, iCassandra.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iCassandra.Spec, "k8s resource spec", cassandra.Spec)
			return models.ExitReconcile, nil
		}

		r.EventRecorder.Eventf(cassandra, models.Warning, models.ExternalChanges, msgDiffSpecs)

		return models.ExitReconcile, nil
	}

	patch := cassandra.NewPatch()

	cassandra.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), cassandra, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", cassandra.Spec.Name, "cluster ID", cassandra.Status.ID)

		r.EventRecorder.Eventf(cassandra, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info("External changes have been reconciled", "resource ID", cassandra.Status.ID)
	r.EventRecorder.Event(cassandra, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleDeleteCluster(
	ctx context.Context,
	l logr.Logger,
	cassandra *v1beta1.Cassandra,
) (reconcile.Result, error) {
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
		return reconcile.Result{}, err
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
			return reconcile.Result{}, err
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

				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", cassandra.Status.ID)

			r.EventRecorder.Event(cassandra, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.StatusChecker))

	l.Info("Deleting cluster backup resources", "cluster ID", cassandra.Status.ID)

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
		return reconcile.Result{}, err
	}

	l.Info("Cluster backup resources were deleted",
		"cluster ID", cassandra.Status.ID,
	)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Deleted,
		"Cluster backup resources are deleted",
	)

	for _, ref := range cassandra.Spec.UserRefs {
		err = r.handleUsersDetach(ctx, l, cassandra, ref)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

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
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, cassandra.Name, cassandra.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Cassandra cluster expose service",
			"cluster ID", cassandra.Status.ID,
			"cluster name", cassandra.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Cluster has been deleted",
		"cluster name", cassandra.Spec.Name,
		"cluster ID", cassandra.Status.ID,
		"kind", cassandra.Kind,
		"api Version", cassandra.APIVersion)

	r.EventRecorder.Eventf(
		cassandra, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *CassandraReconciler) handleDeleteOnPremisesCluster(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
) (reconcile.Result, error) {
	l = l.WithName("On-premises Cassandra deletion event")

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

			return reconcile.Result{}, err
		}
	}

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.StatusChecker))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.OnPremisesIPsChecker))

	err = r.deleteOnPremResources(ctx, c)
	if err != nil {
		l.Error(err, "Cannot delete cluster on-premises resources",
			"cluster ID", c.Status.ID,
		)
		r.EventRecorder.Eventf(
			c, models.Warning, models.DeletionFailed,
			"Cluster on-premises resources deletion is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	l.Info("Cluster on-premises resources are deleted",
		"cluster ID", c.Status.ID,
	)
	r.EventRecorder.Eventf(
		c, models.Normal, models.Deleted,
		"Cluster on-premises resources are deleted deleted",
	)

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

func (r *CassandraReconciler) handleUsersCreate(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.CassandraUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot create a Cassandra user. The resource is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"User is not found, create a new one Cassandra User or provide correct userRef."+
					"Current provided reference: %v", uRef)
			return err
		}

		l.Error(err, "Cannot get Cassandra user", "user", u.Spec)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot get Cassandra user. user reference: %v", uRef)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; exist {
		l.Info("User is already existing on the cluster",
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
		l.Error(err, "Cannot patch the Cassandra User status with the CreatingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
			"Cannot add Cassandra User to the cluster. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for creation", "username", u.Name)

	return nil
}

func (r *CassandraReconciler) handleUsersDelete(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.CassandraUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot delete a Cassandra user, the user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"Cannot delete a Cassandra user, the user %v is not found", req)
			return nil
		}

		l.Error(err, "Cannot get Cassandra user", "user", req)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot get Cassandra user. user reference: %v", req)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; !exist {
		l.Info("User is not existing on the cluster",
			"user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.DeletionFailed,
			"User is not existing on the cluster. User reference: %v", req)

		return nil
	}

	patch := u.NewPatch()

	u.Status.ClustersEvents[c.Status.ID] = models.DeletingEvent

	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the Cassandra User status with the DeletingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot patch the Cassandra User status with the DeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for deletion",
		"User resource", u.Namespace+"/"+u.Name,
		"Cassandra resource", c.Namespace+"/"+c.Name)

	return nil
}

func (r *CassandraReconciler) handleUsersDetach(
	ctx context.Context,
	l logr.Logger,
	c *v1beta1.Cassandra,
	uRef *v1beta1.UserReference,
) error {
	req := types.NamespacedName{
		Namespace: uRef.Namespace,
		Name:      uRef.Name,
	}

	u := &clusterresourcesv1beta1.CassandraUser{}
	err := r.Get(ctx, req, u)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot detach a Cassandra user, the user is not found", "request", req)
			r.EventRecorder.Eventf(c, models.Warning, models.NotFound,
				"Cannot detach a Cassandra user, the user %v is not found", req)
			return nil
		}

		l.Error(err, "Cannot get Cassandra user", "user", req)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot get Cassandra user. user reference: %v", req)
		return err
	}

	if _, exist := u.Status.ClustersEvents[c.Status.ID]; !exist {
		l.Info("User is not existing in the cluster", "user reference", uRef)
		r.EventRecorder.Eventf(c, models.Normal, models.DeletionFailed,
			"User is not existing in the cluster. User reference: %v", uRef)
		return nil
	}

	patch := u.NewPatch()
	u.Status.ClustersEvents[c.Status.ID] = models.ClusterDeletingEvent
	err = r.Status().Patch(ctx, u, patch)
	if err != nil {
		l.Error(err, "Cannot patch the Cassandra user status with the ClusterDeletingEvent",
			"cluster name", c.Spec.Name, "cluster ID", c.Status.ID)
		r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
			"Cannot patch the Cassandra user status with the ClusterDeletingEvent. Reason: %v", err)
		return err
	}

	l.Info("User has been added to the queue for detaching", "username", u.Name)

	return nil
}

func (r *CassandraReconciler) handleUserEvent(
	newObj *v1beta1.Cassandra,
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

		err := r.handleUsersCreate(ctx, l, newObj, newUser)
		if err != nil {
			l.Error(err, "Cannot create Cassandra user in predicate", "user", newUser)
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

		err := r.handleUsersDelete(ctx, l, newObj, oldUser)
		if err != nil {
			l.Error(err, "Cannot delete Cassandra user", "user", oldUser)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot delete user from cluster. Reason: %v", err)
		}
	}
}

func (r *CassandraReconciler) startClusterStatusJob(cassandraCluster *v1beta1.Cassandra) error {
	job := r.newWatchStatusJob(cassandraCluster)

	err := r.Scheduler.ScheduleJob(cassandraCluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
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

func (r *CassandraReconciler) startClusterOnPremisesIPsJob(cluster *v1beta1.Cassandra) error {
	job := r.newWatchOnPremisesIPsJob(cluster)

	err := r.Scheduler.ScheduleJob(cluster.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) newWatchOnPremisesIPsJob(c *v1beta1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraOnPremStatusClusterJob")

	return func() error {

		if c.Spec.OnPremisesSpec != nil && c.Spec.PrivateNetworkCluster {
			gateways, err := r.IcadminAPI.GetGateways(c.Status.DataCentres[0].ID)
			if err != nil {
				l.Error(err, "Cannot get Cassandra SSH-gateway nodes from the Instaclustr API",
					"cluster name", c.Spec.Name,
					"status", c.Status)
				r.EventRecorder.Eventf(
					c, models.Warning, models.FetchFailed,
					"SSH-gateway nodes fetch from the Instaclustr API is failed. Reason: %v",
					err,
				)
				return err
			}

			for _, gateway := range gateways {
				gatewayPods := &k8scorev1.PodList{}
				err = r.List(context.Background(), gatewayPods, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: c.Status.ID,
						models.NodeIDLabel:    gateway.ID,
					}),
					Namespace: c.Namespace,
				})
				if err != nil {
					l.Error(err, "Cannot list SSH-gateway pods",
						"cluster name", c.Spec.Name,
						"clusterID", c.Status.ID,
					)

					r.EventRecorder.Eventf(
						c, models.Warning, models.CreationFailed,
						"Fetching SSH-gateways is failed. Reason: %v",
						err,
					)
					return err
				}

				for _, pod := range gatewayPods.Items {
					if (pod.Status.PodIP != "" && gateway.PrivateAddress == "") ||
						(pod.Status.PodIP != "" && pod.Status.PodIP != gateway.PrivateAddress) {
						err = r.IcadminAPI.SetPrivateGatewayIP(c.Status.DataCentres[0].ID, pod.Status.PodIP)
						if err != nil {
							l.Error(err, "Cannot set Private IP for the SSH-gateway node",
								"cluster name", c.Spec.Name,
								"clusterID", c.Status.ID,
							)

							r.EventRecorder.Eventf(
								c, models.Warning, models.CreationFailed,
								"Setting Private IP for the SSH-gateway node is failed. Reason: %v",
								err,
							)
							return err
						}
					}
				}

				gatewaySVCs := &k8scorev1.ServiceList{}
				err = r.List(context.Background(), gatewaySVCs, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: c.Status.ID,
						models.NodeIDLabel:    gateway.ID,
					}),
					Namespace: c.Namespace,
				})
				if err != nil {
					l.Error(err, "Cannot get services backed by SSH-gateway pods",
						"cluster name", c.Spec.Name,
						"clusterID", c.Status.ID,
					)

					r.EventRecorder.Eventf(
						c, models.Warning, models.CreationFailed,
						"Fetching services backed by SSH-gateway pods is failed. Reason: %v",
						err,
					)
					return err
				}

				for _, svc := range gatewaySVCs.Items {
					if (svc.Status.LoadBalancer.Ingress[0].IP != "" && gateway.PublicAddress == "") ||
						(svc.Status.LoadBalancer.Ingress[0].IP != gateway.PublicAddress) {
						err = r.IcadminAPI.SetPublicGatewayIP(c.Status.DataCentres[0].ID, svc.Status.LoadBalancer.Ingress[0].IP)
						if err != nil {
							l.Error(err, "Cannot set Public IP for the SSH-gateway node",
								"cluster name", c.Spec.Name,
								"clusterID", c.Status.ID,
							)

							r.EventRecorder.Eventf(
								c, models.Warning, models.CreationFailed,
								"Setting Public IP for the SSH-gateway node is failed. Reason: %v",
								err,
							)
							return err
						}

						l.Info("IPs for on-premises cluster ssh-gateway are set",
							"cluster name", c.Spec.Name,
							"clusterID", c.Status.ID,
						)

						r.EventRecorder.Eventf(
							c, models.Normal, models.Created,
							"SSH-gateway IPs are set",
						)
					}
				}
			}
		}

		request := &v1beta1.OnPremiseNode{}
		nodes, err := r.IcadminAPI.GetOnPremisesNodes(c.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cassandra on-premises nodes from the Instaclustr API",
				"cluster name", c.Spec.Name,
				"status", c.Status)
			r.EventRecorder.Eventf(
				c, models.Warning, models.FetchFailed,
				"On-premises nodes fetch from the Instaclustr API is failed. Reason: %v",
				err,
			)
			return err
		}

		for _, node := range nodes {
			nodePods := &k8scorev1.PodList{}
			err = r.List(context.Background(), nodePods, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					models.ClusterIDLabel: c.Status.ID,
					models.NodeIDLabel:    node.ID,
				}),
				Namespace: c.Namespace,
			})
			if err != nil {
				l.Error(err, "Cannot get on-premises cluster pods",
					"cluster name", c.Spec.Name,
					"clusterID", c.Status.ID,
				)

				r.EventRecorder.Eventf(
					c, models.Warning, models.CreationFailed,
					"Fetching on-premises cluster pods is failed. Reason: %v",
					err,
				)
				return err
			}

			nodeSVCs := &k8scorev1.ServiceList{}
			err = r.List(context.Background(), nodeSVCs, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					models.ClusterIDLabel: c.Status.ID,
					models.NodeIDLabel:    node.ID,
				}),
				Namespace: c.Namespace,
			})
			if err != nil {
				l.Error(err, "Cannot get services backed by on-premises cluster pods",
					"cluster name", c.Spec.Name,
					"clusterID", c.Status.ID,
				)

				r.EventRecorder.Eventf(
					c, models.Warning, models.CreationFailed,
					"Fetching services backed by on-premises cluster pods is failed. Reason: %v",
					err,
				)
				return err
			}

			for _, pod := range nodePods.Items {
				if (pod.Status.PodIP != "" && node.PrivateAddress == "") ||
					(pod.Status.PodIP != "" && pod.Status.PodIP != node.PrivateAddress) {
					request.PrivateAddress = pod.Status.PodIP
				}
			}

			for _, svc := range nodeSVCs.Items {
				if (svc.Status.LoadBalancer.Ingress[0].IP != "" && node.PublicAddress == "") ||
					(svc.Status.LoadBalancer.Ingress[0].IP != node.PublicAddress) {
					request.PublicAddress = svc.Status.LoadBalancer.Ingress[0].IP
				}
			}

			if request.PublicAddress != "" || request.PrivateAddress != "" {
				err = r.IcadminAPI.SetNodeIPs(node.ID, request)
				if err != nil {
					l.Error(err, "Cannot set IPs for on-premises cluster nodes",
						"cluster name", c.Spec.Name,
						"clusterID", c.Status.ID,
					)

					r.EventRecorder.Eventf(
						c, models.Warning, models.CreationFailed,
						"Setting IPs for on-premises cluster nodes is failed. Reason: %v",
						err,
					)
					return err
				}

				l.Info("IPs for on-premises cluster node are set",
					"cluster name", c.Spec.Name,
					"clusterID", c.Status.ID,
					"request", request,
				)
				r.EventRecorder.Eventf(
					c, models.Normal, models.Created,
					"Nodes IPs are set",
				)
			}
		}
		return nil
	}
}

func (r *CassandraReconciler) newWatchStatusJob(cassandra *v1beta1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "CassandraStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(cassandra)
		err := r.Get(context.Background(), namespacedName, cassandra)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)

			if cassandra.Spec.OnPremisesSpec != nil {
				r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.StatusChecker))
				r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.OnPremisesIPsChecker))
				return nil
			}

			r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.BackupsChecker))
			r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.UserCreator))
			r.Scheduler.RemoveJob(cassandra.GetJobID(scheduler.StatusChecker))
			return nil
		}

		iData, err := r.API.GetCassandra(cassandra.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(context.Background(), cassandra)
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

			if !areDCsEqual && cassandra.Spec.OnPremisesSpec == nil {
				var nodes []*v1beta1.Node

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
			cassandra.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			cassandra.Annotations[models.UpdateQueuedAnnotation] != models.True &&
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

			msgDiffSpecs, err := createSpecDifferenceMessage(cassandra.Spec, iCassandra.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iCassandra.Spec, "k8s resource spec", cassandra.Spec)
				return err
			}

			r.EventRecorder.Eventf(cassandra, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), cassandra)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", cassandra.Spec.Name,
				"cluster ID", cassandra.Status.ID,
			)
			return err
		}

		if cassandra.Status.State == models.RunningStatus && cassandra.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := cassandra.NewPatch()
			for _, dc := range cassandra.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", cassandra.Spec.Name,
						"cluster ID", cassandra.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), cassandra, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", cassandra.Spec.Name,
						"cluster ID", cassandra.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *CassandraReconciler) newWatchBackupsJob(cluster *v1beta1.Cassandra) scheduler.Job {
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

		iBackups, err := r.API.GetClusterBackups(cluster.Status.ID, models.ClusterKindsMap[cluster.Kind])
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

		for _, ref := range c.Spec.UserRefs {
			err = r.handleUsersCreate(ctx, l, c, ref)
			if err != nil {
				l.Error(err, "Failed to create a user for the cluster", "user ref", ref)
				r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
					"Failed to create a user for the cluster. Reason: %v", err)
				return err
			}
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

	if !c.Status.AreMaintenanceEventStatusesEqual(iMEStatuses) {
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

func (r *CassandraReconciler) reconcileOnPremResources(
	ctx context.Context,
	c *v1beta1.Cassandra,
) error {
	if c.Spec.PrivateNetworkCluster {
		err := r.reconcileSSHGatewayResources(ctx, c)
		if err != nil {
			return err
		}
	}

	err := r.reconcileNodesResources(ctx, c)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) reconcileSSHGatewayResources(
	ctx context.Context,
	c *v1beta1.Cassandra,
) error {
	gateways, err := r.IcadminAPI.GetGateways(c.Status.DataCentres[0].ID)
	if err != nil {
		return err
	}

	for i, gateway := range gateways {
		gatewayDVSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.OSDiskSize)
		if err != nil {
			return err
		}

		gatewayDVName := fmt.Sprintf("%s-%d-%s", models.GatewayDVPrefix, i, strings.ToLower(c.Spec.Name))
		gatewayDV, err := r.createDV(ctx, c, gatewayDVName, gateway.ID, gatewayDVSize, true)
		if err != nil {
			return err
		}

		gatewayCPU := resource.Quantity{}
		gatewayCPU.Set(c.Spec.OnPremisesSpec.SSHGatewayCPU)

		gatewayMemory, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.SSHGatewayMemory)
		if err != nil {
			return err
		}

		gatewayName := fmt.Sprintf("%s-%d-%s", models.GatewayVMPrefix, i, strings.ToLower(c.Spec.Name))

		secretName, err := r.reconcileIgnitionScriptSecret(ctx, c, gatewayName, gateway.ID, gateway.Rack)
		if err != nil {
			return err
		}

		gatewayVM := &virtcorev1.VirtualMachine{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      gatewayName,
		}, gatewayVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			gatewayVM, err = r.newVM(
				ctx,
				c,
				gatewayName,
				gateway.ID,
				gateway.Rack,
				gatewayDV.Name,
				secretName,
				gatewayCPU,
				gatewayMemory)
			if err != nil {
				return err
			}
			err = r.Client.Create(ctx, gatewayVM)
			if err != nil {
				return err
			}
		}

		gatewaySvcName := fmt.Sprintf("%s-%s", models.GatewaySvcPrefix, gatewayName)
		gatewayExposeService := &k8scorev1.Service{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      gatewaySvcName,
		}, gatewayExposeService)

		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			gatewayExposeService = r.newExposeService(c, gatewaySvcName, gatewayName, gateway.ID)
			err = r.Client.Create(ctx, gatewayExposeService)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *CassandraReconciler) reconcileNodesResources(
	ctx context.Context,
	c *v1beta1.Cassandra,
) error {
	nodes, err := r.IcadminAPI.GetOnPremisesNodes(c.Status.ID)
	if err != nil {
		return err
	}

	for i, node := range nodes {
		nodeOSDiskSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.OSDiskSize)
		if err != nil {
			return err
		}

		nodeOSDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeOSDVPrefix, i, strings.ToLower(c.Name))
		nodeOSDV, err := r.createDV(ctx, c, nodeOSDiskDVName, node.ID, nodeOSDiskSize, true)
		if err != nil {
			return err
		}

		nodeDataDiskDVSize, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.DataDiskSize)
		if err != nil {
			return err
		}

		nodeDataDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, strings.ToLower(c.Name))
		nodeDataDV, err := r.createDV(ctx, c, nodeDataDiskDVName, node.ID, nodeDataDiskDVSize, false)
		if err != nil {
			return err
		}

		nodeCPU := resource.Quantity{}
		nodeCPU.Set(c.Spec.OnPremisesSpec.NodeCPU)

		nodeMemory, err := resource.ParseQuantity(c.Spec.OnPremisesSpec.NodeMemory)
		if err != nil {
			return err
		}

		nodeName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, strings.ToLower(c.Name))

		secretName, err := r.reconcileIgnitionScriptSecret(ctx, c, nodeName, node.ID, node.Rack)
		if err != nil {
			return err
		}

		nodeVM := &virtcorev1.VirtualMachine{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      nodeName,
		}, nodeVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			nodeVM, err = r.newVM(
				ctx,
				c,
				nodeName,
				node.ID,
				node.Rack,
				nodeOSDV.Name,
				secretName,
				nodeCPU,
				nodeMemory,
				nodeDataDV.Name)
			if err != nil {
				return err
			}
			err = r.Client.Create(ctx, nodeVM)
			if err != nil {
				return err
			}
		}

		if !c.Spec.PrivateNetworkCluster {
			nodeExposeName := fmt.Sprintf("%s-%s", models.NodeSvcPrefix, nodeName)
			nodeExposeService := &k8scorev1.Service{}
			err = r.Get(ctx, types.NamespacedName{
				Namespace: c.Namespace,
				Name:      nodeExposeName,
			}, nodeExposeService)
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			if k8serrors.IsNotFound(err) {
				nodeExposeService = r.newExposeService(c, nodeExposeName, nodeName, node.ID)
				err = r.Client.Create(ctx, nodeExposeService)
				if err != nil {
					return err
				}
			}
		}

		headlessServiceName := fmt.Sprintf("%s-%s", models.KubevirtSubdomain, c.Spec.Name)
		headlessSVC := &k8scorev1.Service{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      headlessServiceName,
		}, headlessSVC)

		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			ports := []k8scorev1.ServicePort{
				{
					Name: models.InterNode,
					Port: models.Port7000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: models.Port7000,
					},
				},
				{
					Name: models.CQLSH,
					Port: models.Port9042,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: models.Port9042,
					},
				},
			}
			headlessSVC = &k8scorev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       models.ServiceKind,
					APIVersion: models.K8sAPIVersionV1,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      headlessServiceName,
					Namespace: c.Namespace,
					Labels: map[string]string{
						models.ClusterIDLabel: c.Status.ID,
					},
					//Finalizers: []string{models.DeletionFinalizer},
				},
				Spec: k8scorev1.ServiceSpec{
					ClusterIP: "None",
					Ports:     ports,
					Selector: map[string]string{
						models.ClusterIDLabel: c.Status.ID,
					},
				},
			}
			err = r.Client.Create(ctx, headlessSVC)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (r *CassandraReconciler) createDV(
	ctx context.Context,
	c *v1beta1.Cassandra,
	name,
	nodeID string,
	size resource.Quantity,
	isOSDisk bool,
) (*cdiv1beta1.DataVolume, error) {
	dv := &cdiv1beta1.DataVolume{}
	pvc := &k8scorev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      name,
	}, pvc)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if k8serrors.IsNotFound(err) {
		err = r.Get(ctx, types.NamespacedName{
			Namespace: c.Namespace,
			Name:      name,
		}, dv)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if k8serrors.IsNotFound(err) {
			if isOSDisk {
				dv = r.newOSDiskDV(c, name, nodeID, size)
			} else {
				dv = r.newDataDiskDV(c, name, nodeID, size)
			}
			err = r.Client.Create(ctx, dv)
			if err != nil {
				return nil, err
			}
		}
	}
	return dv, nil
}

func (r *CassandraReconciler) newOSDiskDV(
	c *v1beta1.Cassandra,
	name,
	nodeID string,
	size resource.Quantity,
) *cdiv1beta1.DataVolume {
	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				models.ClusterIDLabel: c.Status.ID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: &cdiv1beta1.DataVolumeSource{
				HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
					URL: c.Spec.OnPremisesSpec.OSImageURL,
				},
			},
			PVC: &k8scorev1.PersistentVolumeClaimSpec{
				AccessModes: []k8scorev1.PersistentVolumeAccessMode{
					k8scorev1.ReadWriteOnce,
				},
				Resources: k8scorev1.ResourceRequirements{
					Requests: k8scorev1.ResourceList{
						models.Storage: size,
					},
				},
				StorageClassName: &c.Spec.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

func (r *CassandraReconciler) newDataDiskDV(
	c *v1beta1.Cassandra,
	name,
	nodeID string,
	size resource.Quantity,
) *cdiv1beta1.DataVolume {
	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				models.ClusterIDLabel: c.Status.ID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
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
						models.Storage: size,
					},
				},
				StorageClassName: &c.Spec.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

func (r *CassandraReconciler) newExposeService(
	c *v1beta1.Cassandra,
	name,
	vmName,
	nodeID string,
) *k8scorev1.Service {
	var ports []k8scorev1.ServicePort
	ports = []k8scorev1.ServicePort{{
		Name: models.SSH,
		Port: models.Port22,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: models.Port22,
		},
	},
	}

	if !c.Spec.PrivateNetworkCluster {
		additionalPorts := []k8scorev1.ServicePort{
			{
				Name: models.InterNode,
				Port: models.Port7000,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7000,
				},
			},
			{
				Name: models.CQLSH,
				Port: models.Port9042,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port9042,
				},
			},
			{
				Name: models.JMX,
				Port: models.Port7199,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7199,
				},
			},
		}
		if c.Spec.DataCentres[0].ClientToClusterEncryption {
			sslPort := k8scorev1.ServicePort{
				Name: models.SSL,
				Port: models.Port7001,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7001,
				},
			}
			additionalPorts = append(additionalPorts, sslPort)
		}
		ports = append(ports, additionalPorts...)
	}

	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
			Labels: map[string]string{
				models.ClusterIDLabel: c.Status.ID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				models.KubevirtDomainLabel: vmName,
				models.NodeIDLabel:         nodeID,
			},
			Type: models.LBType,
		},
	}
}

func (r *CassandraReconciler) newVM(
	ctx context.Context,
	c *v1beta1.Cassandra,
	vmName,
	nodeID,
	nodeRack,
	OSDiskDVName,
	ignitionSecretName string,
	cpu,
	memory resource.Quantity,
	storageDVNames ...string,
) (*virtcorev1.VirtualMachine, error) {
	runStrategy := virtcorev1.RunStrategyAlways
	bootOrder1 := uint(1)

	cloudInitSecret := &k8scorev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: c.Spec.OnPremisesSpec.CloudInitScriptNamespacedName.Namespace,
		Name:      c.Spec.OnPremisesSpec.CloudInitScriptNamespacedName.Name,
	}, cloudInitSecret)
	if err != nil {
		return nil, err
	}

	vm := &virtcorev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.VirtualMachineKind,
			APIVersion: models.KubevirtV1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmName,
			Namespace: c.Namespace,
			Labels: map[string]string{
				models.ClusterIDLabel:      c.Status.ID,
				models.NodeIDLabel:         nodeID,
				models.NodeRackLabel:       nodeRack,
				models.KubevirtDomainLabel: vmName,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: virtcorev1.VirtualMachineSpec{
			RunStrategy: &runStrategy,
			Template: &virtcorev1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						models.ClusterIDLabel:      c.Status.ID,
						models.NodeIDLabel:         nodeID,
						models.NodeRackLabel:       nodeRack,
						models.KubevirtDomainLabel: vmName,
					},
				},
				Spec: virtcorev1.VirtualMachineInstanceSpec{
					Hostname:  vmName,
					Subdomain: fmt.Sprintf("%s-%s", models.KubevirtSubdomain, c.Spec.Name),
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
									Name:      models.Boot,
									BootOrder: &bootOrder1,
									IO:        models.Native,
									Cache:     models.None,
									DiskDevice: virtcorev1.DiskDevice{
										Disk: &virtcorev1.DiskTarget{
											Bus: models.Virtio,
										},
									},
								},
								{
									Name:       models.CloudInit,
									DiskDevice: virtcorev1.DiskDevice{},
									Cache:      models.None,
								},
								{
									Name:       models.IgnitionDisk,
									DiskDevice: virtcorev1.DiskDevice{},
									Serial:     models.IgnitionSerial,
									Cache:      models.None,
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
							Name: models.Boot,
							VolumeSource: virtcorev1.VolumeSource{
								PersistentVolumeClaim: &virtcorev1.PersistentVolumeClaimVolumeSource{
									PersistentVolumeClaimVolumeSource: k8scorev1.PersistentVolumeClaimVolumeSource{
										ClaimName: OSDiskDVName,
									},
								},
							},
						},
						{
							Name: models.CloudInit,
							VolumeSource: virtcorev1.VolumeSource{
								CloudInitNoCloud: &virtcorev1.CloudInitNoCloudSource{
									UserDataSecretRef: &k8scorev1.LocalObjectReference{
										Name: c.Spec.OnPremisesSpec.CloudInitScriptNamespacedName.Name,
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
		diskName := fmt.Sprintf("%s-%d-%s", models.DataDisk, i, vm.Name)
		vm.Spec.Template.Spec.Domain.Devices.Disks = append(vm.Spec.Template.Spec.Domain.Devices.Disks, virtcorev1.Disk{
			Name:  diskName,
			IO:    models.Native,
			Cache: models.None,
			DiskDevice: virtcorev1.DiskDevice{
				Disk: &virtcorev1.DiskTarget{
					Bus: models.Virtio,
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

	return vm, nil
}

func (r *CassandraReconciler) reconcileIgnitionScriptSecret(
	ctx context.Context,
	c *v1beta1.Cassandra,
	nodeName,
	nodeID,
	nodeRack string,
) (string, error) {
	ignitionSecret := &k8scorev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      fmt.Sprintf("%s-%s", models.IgnitionScriptSecretPrefix, nodeName),
	}, ignitionSecret)
	if client.IgnoreNotFound(err) != nil {
		return "", err
	}
	if k8serrors.IsNotFound(err) {
		script, err := r.IcadminAPI.GetIgnitionScript(nodeID)
		if err != nil {
			return "", err
		}

		ignitionSecret = &k8scorev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       models.SecretKind,
				APIVersion: models.K8sAPIVersionV1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", models.IgnitionScriptSecretPrefix, nodeName),
				Namespace: c.Namespace,
				Labels: map[string]string{
					models.ControlledByLabel: c.Name,
					models.ClusterIDLabel:    c.Status.ID,
					models.NodeIDLabel:       nodeID,
					models.NodeRackLabel:     nodeRack,
				},
				Finalizers: []string{models.DeletionFinalizer},
			},
			StringData: map[string]string{
				models.Script: script,
			},
		}
		err = r.Create(ctx, ignitionSecret)
		if err != nil {
			return "", err
		}
	}

	return ignitionSecret.Name, nil
}

func (r *CassandraReconciler) deleteOnPremResources(
	ctx context.Context,
	c *v1beta1.Cassandra,
) error {
	vms := &virtcorev1.VirtualMachineList{}
	err := r.List(ctx, vms, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: c.Status.ID,
		}),
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}

	for _, vm := range vms.Items {
		err = r.Delete(ctx, &vm)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(vm.DeepCopy())
		controllerutil.RemoveFinalizer(&vm, models.DeletionFinalizer)
		err = r.Patch(ctx, &vm, patch)
		if err != nil {
			return err
		}
	}

	vmis := &virtcorev1.VirtualMachineInstanceList{}
	err = r.List(ctx, vmis, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: c.Status.ID,
		}),
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}

	for _, vmi := range vmis.Items {
		err = r.Delete(ctx, &vmi)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(vmi.DeepCopy())
		controllerutil.RemoveFinalizer(&vmi, models.DeletionFinalizer)
		err = r.Patch(ctx, &vmi, patch)
		if err != nil {
			return err
		}
	}

	dvs := &cdiv1beta1.DataVolumeList{}
	err = r.List(ctx, dvs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: c.Status.ID,
		}),
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}

	for _, dv := range dvs.Items {
		err = r.Delete(ctx, &dv)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(dv.DeepCopy())
		controllerutil.RemoveFinalizer(&dv, models.DeletionFinalizer)
		err = r.Patch(ctx, &dv, patch)
		if err != nil {
			return err
		}
	}

	svcs := &k8scorev1.ServiceList{}
	err = r.List(ctx, svcs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: c.Status.ID,
		}),
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}

	for _, svc := range svcs.Items {
		err = r.Delete(ctx, &svc)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(svc.DeepCopy())
		controllerutil.RemoveFinalizer(&svc, models.DeletionFinalizer)
		err = r.Patch(ctx, &svc, patch)
		if err != nil {
			return err
		}
	}

	secrets := &k8scorev1.SecretList{}
	err = r.List(ctx, secrets, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: c.Status.ID,
		}),
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		err = r.Delete(ctx, &secret)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(secret.DeepCopy())
		controllerutil.RemoveFinalizer(&secret, models.DeletionFinalizer)
		err = r.Patch(ctx, &secret, patch)
		if err != nil {
			return err
		}
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
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.StatusChecker))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
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

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				oldObj := event.ObjectOld.(*v1beta1.Cassandra)

				r.handleUserEvent(newObj, oldObj.Spec.UserRefs)

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
