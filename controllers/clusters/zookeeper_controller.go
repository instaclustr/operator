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

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// ZookeeperReconciler reconciles a Zookeeper object
type ZookeeperReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ZookeeperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	zook := &v1beta1.Zookeeper{}
	err := r.Client.Get(ctx, req.NamespacedName, zook)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Zookeeper resource is not found",
				"request", req)
			return ctrl.Result{}, nil
		}

		l.Error(err, "unable to fetch Zookeeper",
			"request", req)
		return reconcile.Result{}, err
	}

	switch zook.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, zook, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(zook, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, zook, l)
	case models.GenericEvent:
		l.Info("Generic event isn't handled", "cluster name", zook.Spec.Name, "request", req,
			"event", zook.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	default:
		l.Info("Zookeeper resource event isn't handled",
			"cluster name", zook.Spec.Name,
			"request", req,
			"event", zook.Annotations[models.ResourceStateAnnotation],
		)
		return models.ExitReconcile, nil
	}
}

func (r *ZookeeperReconciler) createCluster(ctx context.Context, zook *v1beta1.Zookeeper, l logr.Logger) error {
	id, err := getClusterIDByName(r.API, models.ZookeeperAppType, zook.Spec.Name)
	if err != nil {
		return err
	}

	if id != "" {
		l.Info("Cluster with provided name already exists", "name", zook.Spec.Name, "clusterID", id)
		return fmt.Errorf("cluster %s already exists, please change name property", zook.Spec.Name)
	}

	l.Info("Creating zookeeper cluster",
		"cluster name", zook.Spec.Name,
		"data centres", zook.Spec.DataCentres)

	b, err := r.API.CreateClusterRaw(instaclustr.ZookeeperEndpoint, zook.Spec.ToInstAPI())
	if err != nil {
		return fmt.Errorf("failed to create zookeeper cluster, err: %w", err)
	}

	var instaModel models.ZookeeperCluster
	err = json.Unmarshal(b, &instaModel)
	if err != nil {
		return fmt.Errorf("failed to unmarshal body to models.ZookeeperCluster, err: %w", err)
	}

	patch := zook.NewPatch()

	zook.Spec.FromInstAPI(&instaModel)
	zook.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Patch(ctx, zook, patch)
	if err != nil {
		return fmt.Errorf("failed to patch cluster spec, err: %w", err)
	}

	zook.Status.FromInstAPI(&instaModel)
	err = r.Status().Patch(ctx, zook, patch)
	if err != nil {
		return fmt.Errorf("failed to patch cluster status, err: %w", err)
	}

	l.Info("Zookeeper cluster has been created", "cluster ID", zook.Status.ID)
	r.EventRecorder.Eventf(
		zook, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s",
		zook.Status.ID,
	)

	err = r.createDefaultSecret(ctx, zook, l)
	if err != nil {
		return err
	}

	return nil
}

func (r *ZookeeperReconciler) handleCreateCluster(
	ctx context.Context,
	zook *v1beta1.Zookeeper,
	l logr.Logger,
) (reconcile.Result, error) {
	var err error
	l = l.WithName("Creation Event")

	if zook.Status.ID == "" {
		err := r.createCluster(ctx, zook, l)
		if err != nil {
			r.EventRecorder.Eventf(zook, models.Warning, models.CreationFailed,
				"Failed to create Zookeeper cluster. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}
	}

	if zook.Status.State != models.DeletedStatus {
		patch := zook.NewPatch()
		zook.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(zook, models.DeletionFinalizer)
		err = r.Patch(ctx, zook, patch)
		if err != nil {
			l.Error(err, "Cannot patch zookeeper", "zookeeper name", zook.Spec.Name)
			r.EventRecorder.Eventf(
				zook, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		err = r.startClusterSyncJob(zook)
		if err != nil {
			l.Error(err, "Cannot start cluster status job",
				"zookeeper cluster ID", zook.Status.ID)
			r.EventRecorder.Eventf(
				zook, models.Warning, models.CreationFailed,
				"Cluster status check job creation is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			zook, models.Normal, models.Created,
			"Cluster sync job is started",
		)
	}

	return models.ExitReconcile, nil
}

func (r *ZookeeperReconciler) createDefaultSecret(ctx context.Context, zk *v1beta1.Zookeeper, l logr.Logger) error {
	username, password, err := r.API.GetDefaultCredentialsV1(zk.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get default user creds for Zookeeper cluster from the Instaclustr API",
			"cluster ID", zk.Status.ID,
		)
		r.EventRecorder.Eventf(zk, models.Warning, models.FetchFailed,
			"Default user password fetch from the Instaclustr API is failed. Reason: %v", err,
		)

		return err
	}

	patch := zk.NewPatch()
	secret := newDefaultUserSecret(username, password, zk.Name, zk.Namespace)

	err = controllerutil.SetOwnerReference(zk, secret, r.Scheme)
	if err != nil {
		l.Error(err, "Cannot set secret owner reference with default user credentials",
			"cluster ID", zk.Status.ID,
		)
		r.EventRecorder.Eventf(zk, models.Warning, models.SetOwnerRef,
			"Setting secret owner ref with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	err = r.Create(ctx, secret)
	if err != nil {
		l.Error(err, "Cannot create secret with default user credentials",
			"cluster ID", zk.Status.ID,
		)
		r.EventRecorder.Eventf(zk, models.Warning, models.CreationFailed,
			"Creating secret with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	zk.Status.DefaultUserSecretRef = &v1beta1.Reference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	err = r.Status().Patch(ctx, zk, patch)
	if err != nil {
		l.Error(err, "Cannot patch Zookeeper resource",
			"cluster name", zk.Spec.Name,
			"status", zk.Status)

		r.EventRecorder.Eventf(
			zk, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return err
	}

	return nil
}

func (r *ZookeeperReconciler) handleUpdateCluster(
	zook *v1beta1.Zookeeper,
	req ctrl.Request,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("Update Event")

	instaModel, err := r.API.GetZookeeper(zook.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get cluster from the Instaclustr", "cluster ID", zook.Status.ID)
		return reconcile.Result{}, err
	}

	iZook := &v1beta1.Zookeeper{}
	iZook.FromInstAPI(instaModel)

	if zook.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.ZookeeperSpec](r.EventRecorder, r.Client, zook, iZook, l)
	}

	if zook.Spec.ClusterSettingsNeedUpdate(&iZook.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iZook.Spec.Description,
			"instaclustr two factor delete", iZook.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(zook.Status.ID, zook.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", zook.Status.ID, "cluster spec", zook.Spec)
			r.EventRecorder.Eventf(zook, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	return models.ExitReconcile, nil
}

func (r *ZookeeperReconciler) handleDeleteCluster(
	ctx context.Context,
	zook *v1beta1.Zookeeper,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("Deletion Event")

	_, err := r.API.GetZookeeper(zook.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"status", zook.Status.GenericStatus.State)
		r.EventRecorder.Eventf(
			zook, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	patch := zook.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", zook.Spec.Name,
			"cluster ID", zook.Status.ID)

		err = r.API.DeleteCluster(zook.Status.ID, instaclustr.ZookeeperEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete zookeeper cluster",
				"cluster name", zook.Spec.Name,
				"cluster status", zook.Status.State)
			r.EventRecorder.Eventf(
				zook, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(zook, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if zook.Spec.TwoFactorDelete != nil {
			zook.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			zook.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", zook.Spec.Name,
					"cluster state", zook.Status.State)
				r.EventRecorder.Eventf(
					zook, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", zook.Status.ID)

			r.EventRecorder.Event(zook, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(zook.GetJobID(scheduler.SyncJob))
	controllerutil.RemoveFinalizer(zook, models.DeletionFinalizer)
	zook.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, zook, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from zookeeper",
			"cluster name", zook.Spec.Name)
		r.EventRecorder.Eventf(
			zook, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, zook.Name, zook.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Zookeeper cluster expose service",
			"cluster ID", zook.Status.ID,
			"cluster name", zook.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Zookeeper cluster was deleted",
		"cluster ID", zook.Status.ID,
	)

	r.EventRecorder.Eventf(
		zook, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *ZookeeperReconciler) startClusterSyncJob(Zookeeper *v1beta1.Zookeeper) error {
	job := r.newSyncJob(Zookeeper)

	err := r.Scheduler.ScheduleJob(Zookeeper.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *ZookeeperReconciler) newSyncJob(zook *v1beta1.Zookeeper) scheduler.Job {
	l := log.Log.WithValues("syncJob", zook.GetJobID(scheduler.SyncJob), "clusterID", zook.Status.ID)

	return func() error {
		namespacedName := client.ObjectKeyFromObject(zook)
		err := r.Get(context.Background(), namespacedName, zook)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(zook.GetJobID(scheduler.SyncJob))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", zook.Name)
			return err
		}

		instaModel, err := r.API.GetZookeeper(zook.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if zook.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), zook, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), zook)
			}

			l.Error(err, "Cannot get Zookeeper cluster status from Instaclustr",
				"cluster ID", zook.Status.ID)
			return err
		}

		iZook := &v1beta1.Zookeeper{}
		iZook.FromInstAPI(instaModel)

		if !zook.Status.Equals(&iZook.Status) {
			l.Info("Updating Zookeeper status")

			areDCsEqual := zook.Status.DCsEquals(iZook.Status.DataCentres)

			patch := zook.NewPatch()
			zook.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch Zookeeper cluster",
					"cluster name", zook.Spec.Name,
					"cluster state", zook.Status.State)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iZook.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					zook.Name,
					zook.Namespace,
					zook.Spec.PrivateNetwork,
					nodes,
					models.ZookeeperConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := zook.Spec.IsEqual(iZook.Spec)

		if equals && zook.Annotations[models.ExternalChangesAnnotation] == models.True {
			patch := zook.NewPatch()
			delete(zook.Annotations, models.ExternalChangesAnnotation)
			err := r.Patch(context.Background(), zook, patch)
			if err != nil {
				return err
			}

			r.EventRecorder.Event(zook, models.Normal, models.ExternalChanges,
				"External changes were automatically reconciled",
			)
		} else if zook.Status.CurrentClusterOperationStatus == models.NoOperation &&
			zook.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

			patch := zook.NewPatch()
			zook.Annotations[models.ExternalChangesAnnotation] = models.True
			err = r.Patch(context.Background(), zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", zook.Spec.Name, "cluster state", zook.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(zook.Spec, iZook.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iZook.Spec, "k8s resource spec", zook.Spec)
				return err
			}
			r.EventRecorder.Eventf(zook, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), zook)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", zook.Spec.Name,
				"cluster ID", zook.Status.ID,
			)
			return err
		}

		return nil
	}
}

func (r *ZookeeperReconciler) handleExternalDelete(ctx context.Context, zook *v1beta1.Zookeeper) error {
	l := log.FromContext(ctx)

	patch := zook.NewPatch()
	zook.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, zook, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(zook, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(zook.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(zook.GetJobID(scheduler.SyncJob))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZookeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{RateLimiter: r.RateLimiter}).
		For(&v1beta1.Zookeeper{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.Zookeeper)

				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(newObj); deleting {
					return true
				}

				if newObj.Status.ID == "" && newObj.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
					return false
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if newObj.Generation == event.ObjectOld.GetGeneration() {
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

func (r *ZookeeperReconciler) reconcileMaintenanceEvents(ctx context.Context, z *v1beta1.Zookeeper) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(z.Status.ID)
	if err != nil {
		return err
	}

	if !z.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := z.NewPatch()
		z.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, z, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", z.Status.ID,
			"events", z.Status.MaintenanceEvents,
		)
	}

	return nil
}
