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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
)

// KafkaReconciler reconciles a Kafka object
type KafkaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
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

	var kafkaCluster clustersv1alpha1.Kafka
	err := r.Client.Get(ctx, req.NamespacedName, &kafkaCluster)
	if err != nil {
		l.Error(err, "unable to fetch Kafka Cluster")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if kafkaCluster.Status.ClusterID == "" {
		l.Info(
			"Kafka Cluster ID not found, creating Kafka cluster",
			"Cluster name", kafkaCluster.Spec.Name,
			"Data centres", kafkaCluster.Spec.DataCentres,
		)

		kafkaCluster.Status.ClusterID, err = r.API.CreateCluster(instaclustr.KafkaEndpoint, convertors.KafkaToInstAPI(kafkaCluster.Spec))
		if err != nil {
			l.Error(
				err, "cannot create Kafka cluster",
				"Kafka manifest", kafkaCluster.Spec,
			)
			return reconcile.Result{}, err
		}

		l.Info(
			"Kafka resource has been created",
			"Cluster name", kafkaCluster.Spec.Name,
			"cluster ID", kafkaCluster.Status.ClusterID,
		)

		err = r.Status().Update(context.Background(), &kafkaCluster)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Kafka{}).
		Complete(r)
}
