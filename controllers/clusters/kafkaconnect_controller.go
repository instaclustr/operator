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
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// KafkaConnectReconciler reconciles a KafkaConnect object
type KafkaConnectReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaConnectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	kafkaConnect := &clustersv1alpha1.KafkaConnect{}
	err := r.Client.Get(ctx, req.NamespacedName, kafkaConnect)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "KafkaConnect resource is not found", "request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch KafkaConnect", "request", req)
		return models.ReconcileRequeue, err
	}

	switch kafkaConnect.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, kafkaConnect, l), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, kafkaConnect, l), nil
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, kafkaConnect, l), nil
	default:
		l.Info("Event isn't handled", "cluster name", kafkaConnect.Spec.Name,
			"request", req, "event", kafkaConnect.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *KafkaConnectReconciler) handleCreateCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Creation Event")

	if kc.Status.ID == "" {
		l.Info("Creating Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"data centres", kc.Spec.DataCentres)

		patch := kc.NewPatch()
		var err error

		kc.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaConnectEndpoint, kc.Spec.ToInstAPI())
		if err != nil {
			l.Error(err, "cannot create Kafka Connect in Instaclustr", "Kafka Connect manifest", kc.Spec)
			return models.ReconcileRequeue
		}

		err = r.Status().Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "cannot patch Kafka Connect status ", "KC ID", kc.Status.ID)
			return models.ReconcileRequeue
		}

		kc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kc, models.DeletionFinalizer)
		err = r.Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka Connect", "cluster name", kc.Spec.Name)
			return models.ReconcileRequeue
		}
	}

	err := r.startClusterStatusJob(kc)
	if err != nil {
		l.Error(err, "Cannot start cluster status job", "cluster ID", kc.Status.ID)
		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster has been created",
		"cluster ID", kc.Status.ID)

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) handleUpdateCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Update Event")

	iData, err := r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get Kafka Connect from Instaclustr",
			"ClusterID", kc.Status.ID)
		return models.ReconcileRequeue
	}

	iKC, err := kc.FromInst(iData)
	if err != nil {
		l.Error(err, "Cannot convert Kafka Connect from Instaclustr",
			"ClusterID", kc.Status.ID)
		return models.ReconcileRequeue
	}

	if !kc.Spec.IsEqual(iKC.Spec) {
		err = r.API.UpdateKafkaConnect(kc.Status.ID, kc.Spec.NewDCsUpdate())
		if err != nil {
			l.Error(err, "Unable to update Kafka Connect cluster",
				"cluster name", kc.Spec.Name,
				"cluster status", kc.Status,
			)
			return models.ReconcileRequeue
		}
	}

	patch := kc.NewPatch()
	kc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Unable to patch Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster status", kc.Status,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster was updated",
		"cluster ID", kc.Status.ID,
	)

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) handleDeleteCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Deletion Event")

	patch := kc.NewPatch()
	err := r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka Connect cluster",
			"cluster name", kc.Spec.Name, "cluster state", kc.Status.State)
		return models.ReconcileRequeue
	}

	_, err = r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster state", kc.Status.ClusterStatus.State)
		return models.ReconcileRequeue
	}

	if !errors.Is(err, instaclustr.NotFound) {
		err = r.API.DeleteCluster(kc.Status.ID, instaclustr.KafkaConnectEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Kafka Connect cluster",
				"cluster name", kc.Spec.Name,
				"cluster state", kc.Status.State)
			return models.ReconcileRequeue
		}

		return models.ReconcileRequeue
	}

	r.Scheduler.RemoveJob(kc.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(kc, models.DeletionFinalizer)
	kc.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from KC",
			"cluster name", kc.Spec.Name)
		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster was deleted",
		"cluster name", kc.Spec.Name,
		"cluster ID", kc.Status.ID)

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) startClusterStatusJob(kc *clustersv1alpha1.KafkaConnect) error {
	job := r.newWatchStatusJob(kc)

	err := r.Scheduler.ScheduleJob(kc.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaConnectReconciler) newWatchStatusJob(kc *clustersv1alpha1.KafkaConnect) scheduler.Job {
	l := log.Log.WithValues("component", "kafkaConnectStatusClusterJob")
	return func() error {
		iData, err := r.API.GetKafkaConnect(kc.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Kafka Connect from Instaclustr",
				"cluster ID", kc.Status.ID)
			return err
		}

		iKC, err := kc.FromInst(iData)
		if err != nil {
			l.Error(err, "Cannot convert Kafka Connect from Instaclustr",
				"cluster ID", kc.Status.ID)
			return err
		}

		if !areStatusesEqual(&iKC.Status.ClusterStatus, &kc.Status.ClusterStatus) {
			l.Info("Kafka Connect status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclustr status", iKC.Status,
				"status", kc.Status.ClusterStatus)

			patch := kc.NewPatch()
			kc.Status.ClusterStatus = iKC.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka Connect cluster",
					"cluster name", kc.Spec.Name, "cluster state", kc.Status.State)
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(kc.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Kafka Connect cluster maintenance events",
				"cluster name", kc.Spec.Name,
				"cluster ID", kc.Status.ID,
			)

			return err
		}

		if !kc.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := kc.NewPatch()
			kc.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka Connect cluster maintenance events",
					"cluster name", kc.Spec.Name,
					"cluster ID", kc.Status.ID,
				)

				return err
			}

			l.Info("Kafka Connect cluster maintenance events were updated",
				"cluster ID", kc.Status.ID,
				"events", kc.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaConnectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.KafkaConnect{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*clustersv1alpha1.KafkaConnect)
				if newObj.Generation == event.ObjectOld.GetGeneration() {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				confirmDeletion(event.ObjectNew)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.GenericEvent})
				return true
			},
		})).Complete(r)
}
