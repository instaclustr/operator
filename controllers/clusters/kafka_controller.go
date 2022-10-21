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
	"reflect"

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
			l.Error(err, "Kafka resource is not found",
				"request", req)
			return reconcile.Result{}, nil
		}

		l.Error(err, "unable to fetch Kafka",
			"request", req)
		return reconcile.Result{}, err
	}

	switch kafka.Annotations[clustersv1alpha1.LastEventAnnotation] {
	case clustersv1alpha1.CreateEvent:
		err = r.handleCreateCluster(ctx, &kafka, l)
		if err != nil {
			return ctrl.Result{RequeueAfter: Requeue60}, nil
		}

		return reconcile.Result{}, nil

	//TODO: handle update
	case clustersv1alpha1.UpdateEvent:
		err = r.handleUpdateCluster(ctx, &kafka, l)
		if err != nil {
			return ctrl.Result{RequeueAfter: Requeue60}, nil
		}

		return reconcile.Result{}, nil

	case clustersv1alpha1.DeleteEvent:
		err = r.handleDeleteCluster(ctx, &kafka, l)
		if err != nil {
			if errors.Is(err, instaclustr.ClusterIsBeingDeleted) {
				l.Info("Kafka cluster is deleting",
					"Cluster name", kafka.Spec.Name,
					"Cluster ID", kafka.Status.ID,
					"Kafka cluster status", kafka.Status.ClusterStatus.Status)

				return reconcile.Result{RequeueAfter: Requeue60}, nil
			}

			l.Error(err, "cannot delete Kafka cluster",
				"Cluster name", kafka.Spec.Name,
				"Cluster status", kafka.Status.ClusterStatus.Status,
				"Cluster ID", kafka.Status.ID)
			return ctrl.Result{RequeueAfter: Requeue60}, nil
		}

		return reconcile.Result{}, nil

	case clustersv1alpha1.GenericEvent:
		l.Info("event isn't handled",
			"Cluster name", kafka.Spec.Name, "request", req,
			"event", kafka.Annotations[clustersv1alpha1.LastEventAnnotation])
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *KafkaReconciler) handleCreateCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) error {
	old := kafka.DeepCopy()
	old.Annotations = map[string]string{}
	patch := client.MergeFrom(old)

	controllerutil.AddFinalizer(kafka, clustersv1alpha1.DeletionFinalizer)
	err := r.Patch(ctx, kafka, patch)
	if err != nil {
		return err
	}

	if kafka.Status.ID == "" {
		l.Info("Creating Kafka cluster",
			"Cluster name", kafka.Spec.Name,
			"Data centres", kafka.Spec.DataCentres)

		kafka.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaEndpoint, convertors.KafkaToInstAPI(kafka.Spec))
		if err != nil {
			l.Error(err, "cannot create Kafka cluster",
				"Kafka manifest", kafka.Spec)
			return err
		}
		l.Info("Kafka cluster has been created",
			"cluster ID", kafka.Status.ID)

		err = r.Status().Patch(ctx, kafka, patch)
		if err != nil {
			l.Error(err, "cannot patch Kafka cluster status from the Instaclustr API",
				"kafka cluster spec", kafka.Spec,
				"kafka ID", kafka.Status.ID)
			return err
		}
	}

	err = r.startClusterStatusJob(kafka)
	if err != nil {
		l.Error(err, "cannot start cluster status job",
			"Kafka cluster ID", kafka.Status.ID)
		return err
	}

	return nil
}

func (r *KafkaReconciler) handleUpdateCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) error {

	l.Info("handleUpdateCluster is not implemented")

	return nil
}

func (r *KafkaReconciler) handleDeleteCluster(ctx context.Context, kafka *clustersv1alpha1.Kafka, l logr.Logger) error {
	r.Scheduler.RemoveJob(kafka.GetJobID(scheduler.StatusChecker))

	status, err := r.API.GetClusterStatus(kafka.Status.ID, instaclustr.KafkaEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		return err
	}

	if status != nil {
		err = r.API.DeleteCluster(kafka.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "cannot delete Kafka cluster",
				"Cluster name", kafka.Spec.Name,
				"Cluster status", kafka.Status.Status)
			return err
		}

		return instaclustr.ClusterIsBeingDeleted
	}

	controllerutil.RemoveFinalizer(kafka, clustersv1alpha1.DeletionFinalizer)
	err = r.Update(ctx, kafka)
	if err != nil {
		l.Error(err, "cannot remove finalizer from Kafka",
			"Cluster name", kafka.Spec.Name,
			"Cluster ID", kafka.Status.ID)
		return err
	}

	return nil
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
		instaclusterStatus, err := r.API.GetClusterStatus(kafka.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "cannot get kafka instaclusterStatus", "ClusterID", kafka.Status.ID)
			return err
		}

		if !reflect.DeepEqual(*instaclusterStatus, kafka.Status.ClusterStatus) {
			l.Info("Kafka status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclusterStatus", instaclusterStatus,
				"kafka.Status.ClusterStatus", kafka.Status.ClusterStatus)
			kafka.Status.ClusterStatus = *instaclusterStatus
			err := r.Status().Update(context.Background(), kafka)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Kafka{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{clustersv1alpha1.LastEventAnnotation: clustersv1alpha1.CreateEvent})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{clustersv1alpha1.LastEventAnnotation: clustersv1alpha1.DeleteEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{clustersv1alpha1.LastEventAnnotation: clustersv1alpha1.UpdateEvent})
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				event.Object.SetAnnotations(map[string]string{clustersv1alpha1.LastEventAnnotation: clustersv1alpha1.GenericEvent})
				return true
			},
		})).Complete(r)
}
