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

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// ZookeeperReconciler reconciles a Zookeeper object
type ZookeeperReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
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
			return models.ExitReconcile, nil
		}

		l.Error(err, "unable to fetch Zookeeper",
			"request", req)
		return models.ReconcileRequeue, err
	}

	switch zook.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, zook, l), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(zook, l), nil
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, zook, l), nil
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

func (r *ZookeeperReconciler) handleCreateCluster(
	ctx context.Context,
	zook *v1beta1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	var err error
	l = l.WithName("Creation Event")

	if zook.Status.ID == "" {
		l.Info("Creating zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"data centres", zook.Spec.DataCentres)

		patch := zook.NewPatch()

		zook.Status.ID, err = r.API.CreateCluster(instaclustr.ZookeeperEndpoint, zook.Spec.ToInstAPI())
		if err != nil {
			l.Error(err, "Cannot create zookeeper cluster", "spec", zook.Spec)
			r.EventRecorder.Eventf(
				zook, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			zook, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s",
			zook.Status.ID,
		)

		err = r.Status().Patch(ctx, zook, patch)
		if err != nil {
			l.Error(err, "Cannot patch zookeeper cluster status from the Instaclustr API",
				"spec", zook.Spec)
			r.EventRecorder.Eventf(
				zook, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

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
			return models.ReconcileRequeue
		}
	}

	err = r.startClusterStatusJob(zook)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"zookeeper cluster ID", zook.Status.ID)
		r.EventRecorder.Eventf(
			zook, models.Warning, models.CreationFailed,
			"Cluster status check job creation is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		zook, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	l.Info("Zookeeper cluster has been created", "cluster ID", zook.Status.ID)

	err = r.createDefaultSecret(ctx, zook, l)
	if err != nil {
		l.Error(err, "Cannot create default secret for Zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"clusterID", zook.Status.ID,
		)
		r.EventRecorder.Eventf(
			zook, models.Warning, models.CreationFailed,
			"Default user secret creation on the Instaclustr is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	return models.ExitReconcile
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

	secret := zk.NewDefaultUserSecret(username, password)
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

	return nil
}

func (r *ZookeeperReconciler) handleUpdateCluster(
	zook *v1beta1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Update Event")

	if zook.Annotations[models.ExternalChangesAnnotation] == models.True {
		r.handleExternalChanges(zook, l)
	} else {
		l.Info("Cluster update is not implemented yet")
	}

	return models.ExitReconcile
}

func (r *ZookeeperReconciler) handleExternalChanges(zook *v1beta1.Zookeeper, l logr.Logger) reconcile.Result {
	iData, err := r.API.GetZookeeper(zook.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get cluster from the Instaclustr", "cluster ID", zook.Status.ID)
		return models.ReconcileRequeue
	}

	iZook, err := zook.FromInstAPI(iData)
	if err != nil {
		l.Error(err, "Cannot convert cluster from the Instaclustr API", "cluster ID", zook.Status.ID)
		return models.ReconcileRequeue
	}

	if !zook.Spec.IsEqual(iZook.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", zook.Spec,
			"data from Instaclustr ", iZook.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(zook.Spec, iZook.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iZook.Spec, "k8s resource spec", zook.Spec)
			return models.ExitReconcile
		}
		r.EventRecorder.Eventf(zook, models.Warning, models.ExternalChanges, msgDiffSpecs)

		return models.ExitReconcile
	}

	patch := zook.NewPatch()

	zook.Annotations[models.ExternalChangesAnnotation] = ""

	err = r.Patch(context.Background(), zook, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", zook.Spec.Name, "cluster ID", zook.Status.ID)

		r.EventRecorder.Eventf(zook, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	l.Info("External changes have been reconciled", "resource ID", zook.Status.ID)
	r.EventRecorder.Event(zook, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return models.ExitReconcile
}

func (r *ZookeeperReconciler) handleDeleteCluster(
	ctx context.Context,
	zook *v1beta1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Deletion Event")

	_, err := r.API.GetZookeeper(zook.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"status", zook.Status.ClusterStatus.State)
		r.EventRecorder.Eventf(
			zook, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
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
			return models.ReconcileRequeue
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
				return models.ReconcileRequeue
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", zook.Status.ID)

			r.EventRecorder.Event(zook, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
	}

	r.Scheduler.RemoveJob(zook.GetJobID(scheduler.StatusChecker))
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
		return models.ReconcileRequeue
	}

	err = exposeservice.Delete(r.Client, zook.Name, zook.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Zookeeper cluster expose service",
			"cluster ID", zook.Status.ID,
			"cluster name", zook.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	l.Info("Zookeeper cluster was deleted",
		"cluster ID", zook.Status.ID,
	)

	r.EventRecorder.Eventf(
		zook, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile
}

func (r *ZookeeperReconciler) startClusterStatusJob(Zookeeper *v1beta1.Zookeeper) error {
	job := r.newWatchStatusJob(Zookeeper)

	err := r.Scheduler.ScheduleJob(Zookeeper.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *ZookeeperReconciler) newWatchStatusJob(zook *v1beta1.Zookeeper) scheduler.Job {
	l := log.Log.WithValues("component", "ZookeeperStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(zook)
		err := r.Get(context.Background(), namespacedName, zook)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(zook.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", zook.Name)
			return err
		}

		iData, err := r.API.GetZookeeper(zook.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleDeleteFromInstaclustrUI(zook, l)
			}

			l.Error(err, "Cannot get Zookeeper cluster status from Instaclustr",
				"cluster ID", zook.Status.ID)
			return err
		}

		iZook, err := zook.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert cluster from the Instaclustr API", "cluster ID", zook.Status.ID)
			return err
		}

		if !areStatusesEqual(&zook.Status.ClusterStatus, &iZook.Status.ClusterStatus) {
			l.Info("Updating Zookeeper status",
				"instaclustr status", iZook.Status,
				"status", zook.Status)

			areDCsEqual := areDataCentresEqual(iZook.Status.ClusterStatus.DataCentres, zook.Status.ClusterStatus.DataCentres)

			patch := zook.NewPatch()
			zook.Status.ClusterStatus = iZook.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch Zookeeper cluster",
					"cluster name", zook.Spec.Name,
					"cluster state", zook.Status.State)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iZook.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					zook.Name,
					zook.Namespace,
					nodes,
					models.ZookeeperConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iZook.Status.CurrentClusterOperationStatus == models.NoOperation &&
			!zook.Spec.IsEqual(iZook.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iZook.Spec, "k8s resource spec", zook.Spec)

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

		//maintEvents, err := r.API.GetMaintenanceEvents(zook.Status.ID)
		//if err != nil {
		//	l.Error(err, "Cannot get Zookeeper cluster maintenance events",
		//		"cluster name", zook.Spec.Name,
		//		"cluster ID", zook.Status.ID,
		//	)
		//
		//	return err
		//}

		//if !zook.Status.AreMaintenanceEventStatusesEqual(maintEvents) {
		//	patch := zook.NewPatch()
		//	zook.Status.MaintenanceEvents = maintEvents
		//	err = r.Status().Patch(context.TODO(), zook, patch)
		//	if err != nil {
		//		l.Error(err, "Cannot patch Zookeeper cluster maintenance events",
		//			"cluster name", zook.Spec.Name,
		//			"cluster ID", zook.Status.ID,
		//		)
		//
		//		return err
		//	}
		//
		//	l.Info("Zookeeper cluster maintenance events were updated",
		//		"cluster ID", zook.Status.ID,
		//		"events", zook.Status.MaintenanceEvents,
		//	)
		//}

		return nil
	}
}

func (r *ZookeeperReconciler) handleDeleteFromInstaclustrUI(zook *v1beta1.Zookeeper, l logr.Logger) error {
	activeClusters, err := r.API.ListClusters()
	if err != nil {
		l.Error(err, "Cannot list account active clusters")
		return err
	}

	if isClusterActive(zook.Status.ID, activeClusters) {
		l.Info("Zookeeper is not found in the Instaclustr but still exist in the Instaclustr list of active cluster",
			"cluster ID", zook.Status.ID,
			"cluster name", zook.Spec.Name,
			"resource name", zook.Name)

		return nil
	}

	l.Info("Cluster is not found in Instaclustr. Deleting resource.",
		"cluster ID", zook.Status.ID,
		"cluster name", zook.Spec.Name)

	patch := zook.NewPatch()
	zook.Annotations[models.ClusterDeletionAnnotation] = ""
	zook.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
	err = r.Patch(context.TODO(), zook, patch)
	if err != nil {
		l.Error(err, "Cannot patch Zookeeper cluster resource",
			"cluster ID", zook.Status.ID,
			"cluster name", zook.Spec.Name,
			"resource name", zook.Name,
		)

		return err
	}

	err = r.Delete(context.TODO(), zook)
	if err != nil {
		l.Error(err, "Cannot delete Zookeeper cluster resource",
			"cluster ID", zook.Status.ID,
			"cluster name", zook.Spec.Name,
			"resource name", zook.Name,
		)

		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZookeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
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
