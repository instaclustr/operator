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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// ZookeeperReconciler reconciles a Zookeeper object
type ZookeeperReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=zookeepers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Zookeeper object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ZookeeperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	zook := &clustersv1alpha1.Zookeeper{}
	err := r.Client.Get(ctx, req.NamespacedName, zook)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Zookeeper resource is not found", "request", req)
			return reconcile.Result{}, nil
		}

		l.Error(err, "unable to fetch Zookeeper", "request", req)
		return reconcile.Result{}, err
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
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ZookeeperReconciler) handleCreateCluster(
	ctx context.Context,
	zook *clustersv1alpha1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Creation Event")

	if zook.Status.ID == "" {
		l.Info("Creating zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"data centres", zook.Spec.DataCentres)

		patch := zook.NewPatch()
		var err error

		zook.Status.ID, err = r.API.CreateCluster(instaclustr.ZookeeperEndpoint, convertors.ZookeeperToInstAPI(zook.Spec))
		if err != nil {
			l.Error(err, "Cannot create zookeeper cluster", "spec", zook.Spec)
			return models.ReconcileRequeue
		}
		l.Info("Zookeeper cluster has been created", "cluster ID", zook.Status.ID)

		err = r.Status().Patch(ctx, zook, patch)
		if err != nil {
			l.Error(err, "Cannot patch zookeeper cluster status from the Instaclustr API",
				"spec", zook.Spec)
			return models.ReconcileRequeue
		}

		zook.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(zook, models.DeletionFinalizer)

		err = r.Patch(ctx, zook, patch)
		if err != nil {
			l.Error(err, "Cannot patch zookeeper", "zookeeper name", zook.Spec.Name)
			return models.ReconcileRequeue
		}
	}

	err := r.startClusterStatusJob(zook)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"zookeeper cluster ID", zook.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *ZookeeperReconciler) handleUpdateCluster(
	zook *clustersv1alpha1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Update Event")

	l.Info("Cluster update is not implemented yet")

	return reconcile.Result{}
}

func (r *ZookeeperReconciler) handleDeleteCluster(
	ctx context.Context,
	zook *clustersv1alpha1.Zookeeper,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Deletion Event")

	patch := zook.NewPatch()
	err := r.Patch(ctx, zook, patch)
	if err != nil {
		l.Error(err, "Cannot patch Zookeeper cluster",
			"cluster name", zook.Spec.Name, "status", zook.Status.Status)
		return models.ReconcileRequeue
	}

	status, err := r.API.GetZookeeper(zook.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get zookeeper cluster",
			"cluster name", zook.Spec.Name,
			"status", zook.Status.ClusterStatus.Status)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteCluster(zook.Status.ID, instaclustr.ZookeeperEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete zookeeper cluster",
				"cluster name", zook.Spec.Name,
				"cluster status", zook.Status.Status)
			return models.ReconcileRequeue
		}

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(zook.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(zook, models.DeletionFinalizer)
	zook.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, zook, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from zookeeper",
			"cluster name", zook.Spec.Name)
		return models.ReconcileRequeue
	}

	l.Info("Zookeeper cluster was deleted",
		"cluster ID", zook.Status.ID,
	)

	return reconcile.Result{}
}

func (r *ZookeeperReconciler) startClusterStatusJob(Zookeeper *clustersv1alpha1.Zookeeper) error {
	job := r.newWatchStatusJob(Zookeeper)

	err := r.Scheduler.ScheduleJob(Zookeeper.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *ZookeeperReconciler) newWatchStatusJob(zook *clustersv1alpha1.Zookeeper) scheduler.Job {
	l := log.Log.WithValues("component", "ZookeeperStatusClusterJob")
	return func() error {
		instStatus, err := r.API.GetZookeeper(zook.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get zookeeper instaclusterStatus", "cluster ID", zook.Status.ID)
			return err
		}

		if !zook.Status.IsEqual(instStatus) {
			l.Info("Zookeeper status of k8s is different from Instaclustr. Reconcile statuses..",
				"instacluster status", instStatus,
				"status", zook.Status.ClusterStatus)

			patch := zook.NewPatch()
			zook.Status.SetFromInst(instStatus)
			err := r.Status().Patch(context.Background(), zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch zookeeper cluster",
					"cluster name", zook.Spec.Name, "Status", zook.Status.Status)
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(zook.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Zookeeper cluster maintenance events",
				"cluster name", zook.Spec.Name,
				"cluster ID", zook.Status.ID,
			)

			return err
		}

		if !zook.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := zook.NewPatch()
			zook.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), zook, patch)
			if err != nil {
				l.Error(err, "Cannot patch Zookeeper cluster maintenance events",
					"cluster name", zook.Spec.Name,
					"cluster ID", zook.Status.ID,
				)

				return err
			}

			l.Info("Zookeeper cluster maintenance events were updated",
				"cluster ID", zook.Status.ID,
				"events", zook.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZookeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Zookeeper{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
				confirmDeletion(event.ObjectNew)
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
