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

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// KafkaReconciler reconciles a Kafka object
type KafkaReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch Kafka", "request", req)
		return models.ReconcileRequeue, err
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
		return models.ExitReconcile, nil
	}

	return models.ExitReconcile, nil
}

func (r *KafkaReconciler) handleCreateCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) reconcile.Result {
	l = l.WithName("Kafka creation Event")

	var err error
	if kafka.Status.ID == "" {
		l.Info("Creating cluster",
			"cluster name", kafka.Spec.Name,
			"data centres", kafka.Spec.DataCentres)

		patch := kafka.NewPatch()
		kafka.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaEndpoint, kafka.Spec.ToInstAPI())
		if err != nil {
			l.Error(err, "Cannot create cluster",
				"spec", kafka.Spec,
			)
			r.EventRecorder.Eventf(
				kafka, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr cloud is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			kafka, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s",
			kafka.Status.ID,
		)

		err = r.Status().Patch(ctx, kafka, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster status",
				"spec", kafka.Spec,
			)
			r.EventRecorder.Eventf(
				kafka, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		kafka.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kafka, models.DeletionFinalizer)
		err = r.Patch(ctx, kafka, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster resource",
				"name", kafka.Spec.Name,
			)
			r.EventRecorder.Eventf(
				kafka, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
	}

	err = r.startClusterStatusJob(kafka)
	if err != nil {
		l.Error(err, "Cannot start cluster status job",
			"cluster ID", kafka.Status.ID)
		r.EventRecorder.Eventf(
			kafka, models.Warning, models.CreationFailed,
			"Cluster status check job creation is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		kafka, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	l.Info("Cluster has been created",
		"cluster ID", kafka.Status.ID)

	return models.ExitReconcile
}

func (r *KafkaReconciler) handleUpdateCluster(
	k *clustersv1alpha1.Kafka,
	l logr.Logger,
) reconcile.Result {
	l = l.WithName("Kafka update Event")

	if k.Status.ClusterStatus.State != StatusRUNNING {
		l.Error(instaclustr.ClusterNotRunning, "Unable to update cluster, cluster still not running",
			"cluster name", k.Spec.Name,
			"cluster state", k.Status.ClusterStatus.State)
		return models.ReconcileRequeue
	}

	err := r.API.UpdateCluster(
		k.Status.ID,
		instaclustr.KafkaEndpoint,
		k.Spec.ToInstAPIUpdate())
	if err != nil {
		l.Error(err, "Unable to update cluster on Instaclustr",
			"cluster name", k.Spec.Name,
			"cluster state", k.Status.ClusterStatus.State)
		r.EventRecorder.Eventf(
			k, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster update has been launched",
		"cluster ID", k.Status.ID,
		"data centres", k.Spec.DataCentres)

	return models.ExitReconcile
}

func (r *KafkaReconciler) handleDeleteCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) reconcile.Result {
	l = l.WithName("Kafka deletion Event")

	_, err := r.API.GetKafka(kafka.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get cluster from the Instaclustr API",
			"cluster name", kafka.Spec.Name,
			"cluster state", kafka.Status.ClusterStatus.State)
		r.EventRecorder.Eventf(
			kafka, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	patch := kafka.NewPatch()
	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", kafka.Spec.Name,
			"cluster ID", kafka.Status.ID)

		err = r.API.DeleteCluster(kafka.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete cluster",
				"cluster name", kafka.Spec.Name,
				"cluster state", kafka.Status.ClusterStatus.State)
			r.EventRecorder.Eventf(
				kafka, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			kafka, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.",
		)

		if kafka.Spec.TwoFactorDelete != nil {
			kafka.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			kafka.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", kafka.Spec.Name,
					"cluster state", kafka.Status.State)
				r.EventRecorder.Eventf(
					kafka, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue
			}

			l.Info("Please confirm cluster deletion via email or phone. "+
				"If you have canceled a cluster deletion and want to put the cluster on deletion again, "+
				"remove \"triggered\" from Instaclustr.com/ClusterDeletionAnnotation annotation.",
				"cluster ID", kafka.Status.ID)

			return models.ExitReconcile
		}
	}

	r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(kafka, models.DeletionFinalizer)
	kafka.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kafka, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", kafka.Spec.Name)
		r.EventRecorder.Eventf(
			kafka, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Cluster was deleted",
		"cluster name", kafka.Spec.Name,
		"cluster ID", kafka.Status.ID)

	r.EventRecorder.Eventf(
		kafka, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile
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
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", kafka.Name)
			return err
		}

		iData, err := r.API.GetKafka(kafka.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(kafka.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", kafka.Status.ClusterStatus.ID,
						"cluster name", kafka.Spec.Name,
						"namespaced name", namespacedName)

					patch := kafka.NewPatch()
					kafka.Annotations[models.ClusterDeletionAnnotation] = ""
					kafka.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), kafka, patch)
					if err != nil {
						l.Error(err, "Cannot patch Kafka cluster resource",
							"cluster ID", kafka.Status.ID,
							"cluster name", kafka.Spec.Name,
							"resource name", kafka.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), kafka)
					if err != nil {
						l.Error(err, "Cannot delete Kafka cluster resource",
							"cluster ID", kafka.Status.ID,
							"cluster name", kafka.Spec.Name,
							"resource name", kafka.Name,
						)

						return err
					}

					return nil
				}
			}

			l.Error(err, "Cannot get cluster from the Instaclustr API", "cluster ID", kafka.Status.ID)
			return err
		}

		iKafka, err := kafka.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert cluster from the Instaclustr API",
				"cluster ID", kafka.Status.ID,
			)
			return err
		}

		if !areStatusesEqual(&kafka.Status.ClusterStatus, &iKafka.Status.ClusterStatus) {
			l.Info("Updating status",
				"instacluster status", iKafka.Status,
				"k8s status", kafka.Status.ClusterStatus)

			patch := kafka.NewPatch()
			kafka.Status.ClusterStatus = iKafka.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), kafka, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", kafka.Spec.Name, "cluster state", kafka.Status.State)
				return err
			}
		}

		maintEvents, err := r.API.GetMaintenanceEvents(kafka.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get cluster maintenance events",
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
				l.Error(err, "Cannot patch cluster maintenance events",
					"cluster name", kafka.Spec.Name,
					"cluster ID", kafka.Status.ID,
				)

				return err
			}

			l.Info("Cluster maintenance events were updated",
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

				newObj := event.ObjectNew.(*clustersv1alpha1.Kafka)
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
