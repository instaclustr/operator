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

// KafkaReconciler reconciles a Kafka object
type KafkaReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kafka object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var kafka clustersv1alpha1.Kafka
	err := r.Client.Get(ctx, req.NamespacedName, &kafka)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka custom resource is not found", "namespaced name ", req.NamespacedName)
			return reconcile.Result{}, nil
		}

		l.Error(err, "Unable to fetch Kafka", "request", req)
		return reconcile.Result{}, err
	}

	switch kafka.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &kafka, l), nil

	case models.UpdatingEvent:
		return r.handleUpdateCluster(&kafka, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &kafka, l), nil

	case models.GenericEvent:
		l.Info("Event isn't handled", "cluster name", kafka.Spec.Name, "request", req,
			"event", kafka.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *KafkaReconciler) handleCreateCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) reconcile.Result {
	l = l.WithName("Creation Event")

	if kafka.Status.ID == "" {
		l.Info("Creating Kafka cluster",
			"cluster name", kafka.Spec.Name,
			"data centres", kafka.Spec.DataCentres)

		patch := kafka.NewPatch()
		var err error

		kafka.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaEndpoint, kafka.Spec.ToInstAPI())
		if err != nil {
			l.Error(err, "Cannot create Kafka cluster", "spec", kafka.Spec)
			return models.ReconcileRequeue
		}
		l.Info("Kafka cluster has been created", "cluster ID", kafka.Status.ID)

		err = r.Status().Patch(ctx, kafka, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka cluster status from the Instaclustr API",
				"spec", kafka.Spec)
			return models.ReconcileRequeue
		}

		kafka.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kafka, models.DeletionFinalizer)

		err = r.Patch(ctx, kafka, patch)
		if err != nil {
			l.Error(err, "Cannot patch kafka", "kafka name", kafka.Spec.Name)
			return models.ReconcileRequeue
		}
	}

	err := r.startClusterStatusJob(kafka)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"kafka cluster ID", kafka.Status.ID)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *KafkaReconciler) handleUpdateCluster(
	k *clustersv1alpha1.Kafka,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Update Event")

	if k.Status.ClusterStatus.Status != StatusRUNNING {
		l.Error(instaclustr.ClusterNotRunning, "Unable to update cluster, cluster still not running",
			"cluster name", k.Spec.Name,
			"cluster status", k.Status.ClusterStatus.Status)
		return models.ReconcileRequeue
	}

	err := r.API.UpdateCluster(
		k.Status.ID,
		instaclustr.KafkaEndpoint,
		k.Spec.ToInstAPIUpdate())
	if err != nil {
		l.Error(err, "Unable to update cluster, got error from Instaclustr",
			"cluster name", k.Spec.Name,
			"cluster status", k.Status.ClusterStatus.Status)
		return models.ReconcileRequeue
	}

	l.Info("Cluster update has been launched",
		"cluster ID", k.Status.ID,
		"data centres", k.Spec.DataCentres)

	return reconcile.Result{}
}

func (r *KafkaReconciler) handleDeleteCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) reconcile.Result {
	l = l.WithName("Deletion Event")

	instaCluster, err := r.API.GetKafka(kafka.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka cluster",
			"cluster name", kafka.Spec.Name,
			"status", kafka.Status.ClusterStatus.Status)
		return models.ReconcileRequeue
	}

	patch := kafka.NewPatch()

	if instaCluster.Status != "" {
		l.Info("Sending cluster deletion to Instaclustr API",
			"cluster name", kafka.Spec.Name, "cluster ID", kafka.Status.ID)

		err = r.API.DeleteCluster(kafka.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Kafka cluster",
				"cluster name", kafka.Spec.Name, "cluster status", kafka.Status.ClusterStatus.Status)
			return models.ReconcileRequeue
		}

		if kafka.Spec.TwoFactorDelete != nil {
			kafka.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			kafka.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err := r.Patch(ctx, kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka cluster",
					"cluster name", kafka.Spec.Name, "status", kafka.Status.Status)
				return models.ReconcileRequeue
			}

			l.Info("Please confirm cluster deletion via email or phone. "+
				"If you have canceled a cluster deletion and want to put the cluster on deletion again, "+
				"remove \"triggered\" from Instaclustr.com/ClusterDeletionAnnotation annotation.",
				"cluster ID", kafka.Status.ID)

			return reconcile.Result{}
		}
	}

	r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(kafka, models.DeletionFinalizer)
	kafka.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kafka, patch)
	if err != nil {
		l.Error(err, "Cannot patch kafka remove recourse",
			"cluster name", kafka.Spec.Name)
		return models.ReconcileRequeue
	}

	l.Info("Kafka cluster was deleted",
		"cluster name", kafka.Spec.Name,
		"cluster ID", kafka.Status.ID)

	return reconcile.Result{}
}

func (r *KafkaReconciler) startClusterStatusJob(kafka *clustersv1alpha1.Kafka) error {
	job := r.newWatchStatusJob(kafka)

	err := r.Scheduler.ScheduleJob(kafka.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaReconciler) newWatchStatusJob(kafka *clustersv1alpha1.Kafka) scheduler.Job {
	l := log.Log.WithValues("component", "kafkaStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(kafka)
		err := r.Get(context.Background(), namespacedName, kafka)
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get kafka custom resource",
				"resource name", kafka.Name)
			return err
		}

		instStatus, err := r.API.GetKafka(kafka.Status.ID)
		if errors.Is(err, instaclustr.NotFound) {
			patch := kafka.NewPatch()
			l.Info("Kafka cluster is not found in Instaclustr. Deleting resource.",
				"cluster ID", kafka.Status.ClusterStatus.ID,
				"cluster name", kafka.Spec.Name,
				"namespaced name", namespacedName)

			controllerutil.RemoveFinalizer(kafka, models.DeletionFinalizer)
			kafka.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
			err = r.Patch(context.Background(), kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch kafka remove recourse",
					"cluster name", kafka.Spec.Name)
				return err
			}

			l.Info("Kafka was deleted",
				"cluster name", kafka.Spec.Name, "cluster ID", kafka.Status.ID)

			r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))

			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get kafka instaclusterStatus", "cluster ID", kafka.Status.ID)
			return err
		}

		if !kafka.Status.IsEqual(instStatus) {
			l.Info("Kafka status of k8s is different from Instaclustr. Reconcile statuses..",
				"instacluster status", instStatus,
				"k8s status", kafka.Status.ClusterStatus)

			patch := kafka.NewPatch()
			kafka.Status.SetFromInst(instStatus)
			err := r.Status().Patch(context.Background(), kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka cluster",
					"cluster name", kafka.Spec.Name, "status", kafka.Status.Status)
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(kafka.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Kafka cluster maintenance events",
				"cluster name", kafka.Spec.Name,
				"cluster ID", kafka.Status.ID,
			)

			return err
		}

		if !kafka.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := kafka.NewPatch()
			kafka.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka cluster maintenance events",
					"cluster name", kafka.Spec.Name,
					"cluster ID", kafka.Status.ID,
				)

				return err
			}

			l.Info("Kafka cluster maintenance events were updated",
				"cluster ID", kafka.Status.ID,
				"events", kafka.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Kafka{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				annots := event.Object.GetAnnotations()
				if annots == nil {
					annots = make(map[string]string)
				}

				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				annots[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatingEvent
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
