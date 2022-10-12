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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		if errors.IsNotFound(err) {
			r.Scheduler.RemoveJob(req.NamespacedName.String())
		}

		l.Error(err, "unable to fetch Kafka Cluster", "request", req)
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if kafka.Status.ID == "" {
		l.Info("Kafka Cluster ID not found, creating Kafka cluster",
			"Cluster name", kafka.Spec.Name,
			"Data centres", kafka.Spec.DataCentres)

		kafka.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaEndpoint, convertors.KafkaToInstAPI(kafka.Spec))
		if err != nil {
			l.Error(err, "cannot create Kafka cluster",
				"Kafka manifest", kafka.Spec)
			return reconcile.Result{}, err
		}
		l.Info("Kafka resource has been created",
			"Cluster name", kafka.Spec.Name,
			"cluster ID", kafka.Status.ID)

		currentClusterStatus, err := r.API.GetClusterStatus(kafka.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "cannot get Kafka cluster status from the Instaclustr API",
				"kafka cluster spec", kafka.Spec,
				"kafka ID", kafka.Status.ID)
			return reconcile.Result{}, err
		}

		kafka.Status.ClusterStatus = *currentClusterStatus

		err = r.Status().Update(context.Background(), &kafka)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.startClusterStatusJob(&kafka)
		if err != nil {
			l.Error(err, "cannot start cluster status job",
				"Kafka cluster ID", kafka.Status.ID)
			return reconcile.Result{}, err
		}
		l.Info("Cluster status job has been started",
			"kafka cluster ID", kafka.Status.ID)
	}

	return reconcile.Result{}, nil
}

func (r *KafkaReconciler) startClusterStatusJob(kafkaCLuster *clustersv1alpha1.Kafka) error {
	job := r.newWatchStatusJob(kafkaCLuster)

	resourceID := client.ObjectKeyFromObject(kafkaCLuster).String()

	err := r.Scheduler.ScheduleJob(resourceID, scheduler.ClusterStatusInterval, job)
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

		if instaclusterStatus.Status != kafka.Status.Status {
			kafka.Status.Status = instaclusterStatus.Status
			err := r.Status().Update(context.Background(), kafka)
			if err != nil {
				return err
			}
			l.Info("Instacluster Status has been changed",
				"ClusterStatus", kafka.Status.ClusterStatus)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Kafka{}).
		Complete(r)
}
