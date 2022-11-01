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
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaConnect object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaConnectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var kafkaConnect clustersv1alpha1.KafkaConnect
	err := r.Client.Get(ctx, req.NamespacedName, &kafkaConnect)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "KafkaConnect resource is not found", "request", req)
			return reconcile.Result{}, nil
		}

		l.Error(err, "unable to fetch KafkaConnect", "request", req)
		return reconcile.Result{}, err
	}

	switch kafkaConnect.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &kafkaConnect, l), nil

	//TODO: handle update
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &kafkaConnect, l), nil

	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &kafkaConnect, l), nil

	default:
		l.Info("event isn't handled", "Cluster name", kafkaConnect.Spec.Name,
			"request", req, "event", kafkaConnect.Annotations[models.ResourceStateAnnotation])
		return reconcile.Result{}, nil
	}
}

func (r *KafkaConnectReconciler) handleCreateCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Creation Event")

	if kc.Status.ID == "" {
		l.Info("Creating Kafka Connect cluster",
			"Cluster name", kc.Spec.Name,
			"Data centres", kc.Spec.DataCentres)

		patch := kc.NewPatch()
		var err error

		kc.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaConnectEndpoint, convertors.KafkaConnectToInstAPI(kc.Spec))
		if err != nil {
			l.Error(err, "cannot create Kafka Connect in Instaclustr", "Kafka Connect manifest", kc.Spec)
			return models.ReconcileRequeue60
		}
		l.Info("Kafka Connect cluster has been created", "cluster ID", kc.Status.ID)
		err = r.Status().Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "cannot patch kafka connect status ", "KC ID", kc.Status.ID)
			return models.ReconcileRequeue60
		}

		kc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kc, models.DeletionFinalizer)

		err = r.Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "cannot patch kafka connect", "KC name", kc.Spec.Name)
			return models.ReconcileRequeue60
		}
	}

	err := r.startClusterStatusJob(kc)
	if err != nil {
		l.Error(err, "cannot start cluster status job", "KC ID", kc.Status.ID)
		return models.ReconcileRequeue60
	}

	return reconcile.Result{}
}

func (r *KafkaConnectReconciler) handleUpdateCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {

	l.Info("handleUpdateCluster is not implemented")

	return reconcile.Result{}
}

func (r *KafkaConnectReconciler) handleDeleteCluster(ctx context.Context, kc *clustersv1alpha1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Deletion Event")

	patch := kc.NewPatch()
	err := r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "cannot patch Kafka Connect cluster",
			"Cluster name", kc.Spec.Name, "Status", kc.Status.Status)
		return models.ReconcileRequeue60
	}

	status, err := r.API.GetClusterStatus(kc.Status.ID, instaclustr.KafkaConnectEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "cannot get Kafka Connect cluster",
			"Cluster name", kc.Spec.Name,
			"Status", kc.Status.ClusterStatus.Status)
		return models.ReconcileRequeue60
	}

	if status != nil {
		err = r.API.DeleteCluster(kc.Status.ID, instaclustr.KafkaConnectEndpoint)
		if err != nil {
			l.Error(err, "cannot delete Kafka Connect cluster",
				"Cluster name", kc.Spec.Name,
				"Status", kc.Status.Status)
			return models.ReconcileRequeue60
		}

		r.Scheduler.RemoveJob(kc.GetJobID(scheduler.StatusChecker))

		return models.ReconcileRequeue60
	}

	controllerutil.RemoveFinalizer(kc, models.DeletionFinalizer)
	kc.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "cannot patch remove finalizer from KC",
			"Cluster name", kc.Spec.Name)
		return models.ReconcileRequeue60
	}

	return reconcile.Result{}
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
		instaclusterStatus, err := r.API.GetClusterStatus(kc.Status.ID, instaclustr.KafkaConnectEndpoint)
		if err != nil {
			l.Error(err, "cannot get KC instaclustrStatus", "ClusterID", kc.Status.ID)
			return err
		}

		if !reflect.DeepEqual(*instaclusterStatus, kc.Status.ClusterStatus) {
			l.Info("Kafka status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclustrStatus", instaclusterStatus,
				"kc.Status.ClusterStatus", kc.Status.ClusterStatus)
			kc.Status.ClusterStatus = *instaclusterStatus
			err := r.Status().Update(context.Background(), kc)
			if err != nil {
				return err
			}
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
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
				confirmDeletion(event.ObjectNew)
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
