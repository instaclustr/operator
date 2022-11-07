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
	apiv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

const (
	StatusRUNNING = "RUNNING"
)

// CassandraReconciler reconciles a Cassandra object
type CassandraReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cassandras/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cassandra object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CassandraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var cassandraCluster clustersv1alpha1.Cassandra
	err := r.Client.Get(ctx, req.NamespacedName, &cassandraCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cassandra resource is not found", "request", req)
			return models.ReconcileResult, nil
		}
		l.Error(err, "unable to fetch Cassandra cluster")
		return models.ReconcileResult, err
	}

	switch cassandraCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		reconcileResult := r.handleCreateCluster(&cassandraCluster, l, ctx)
		return *reconcileResult, nil
	case models.UpdatingEvent:
		reconcileResult := r.handleUpdateCluster(&cassandraCluster, l, ctx)
		return *reconcileResult, nil
	case models.DeletingEvent:
		reconcileResult := r.handleDeleteCluster(&cassandraCluster, l, ctx)
		return *reconcileResult, nil
	default:
		l.Info("UNKNOWN EVENT",
			"Cluster spec", cassandraCluster.Spec)
		return models.ReconcileResult, nil
	}
}

func (r *CassandraReconciler) handleCreateCluster(
	cc *clustersv1alpha1.Cassandra,
	l logr.Logger,
	ctx context.Context,
) *reconcile.Result {

	if cc.Status.ID == "" {
		l.Info(
			"Creating Cassandra cluster",
			"Cluster name", cc.Spec.Name,
			"Data centres", cc.Spec.DataCentres,
		)

		cassandraSpec := apiv2.CassandraToInstAPI(&cc.Spec)
		id, err := r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandraSpec)
		if err != nil {
			l.Error(
				err, "cannot create Cassandra cluster",
				"Cassandra cluster spec", cc.Spec,
			)
			return &models.ReconcileResult
		}

		patch := cc.NewPatch()
		cc.Status.ID = id
		err = r.Status().Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "cannot patch Cassandra cluster status",
				"Cluster name", cc.Spec.Name,
				"Cluster ID", cc.Status.ID,
				"Kind", cc.Kind,
				"API Version", cc.APIVersion,
				"Namespace", cc.Namespace,
				"Cluster metadata", cc.ObjectMeta,
			)
			return &models.ReconcileResult
		}

		l.Info(
			"Cassandra cluster has been created",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
			"Kind", cc.Kind,
			"API Version", cc.APIVersion,
			"Namespace", cc.Namespace,
		)

		controllerutil.AddFinalizer(cc, models.DeletionFinalizer)
		cc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		err = r.Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "cannot patch Cassandra cluster",
				"Cluster name", cc.Spec.Name,
				"Cluster ID", cc.Status.ID,
				"Kind", cc.Kind,
				"API Version", cc.APIVersion,
				"Namespace", cc.Namespace,
				"Cluster metadata", cc.ObjectMeta,
			)
			return &models.ReconcileResult
		}

		err = r.startClusterStatusJob(cc)
		if err != nil {
			l.Error(err, "cannot start cluster status job",
				"Cassandra cluster ID", cc.Status.ID)
			return &models.ReconcileResult
		}
	}
	return &models.ReconcileResult
}

func (r *CassandraReconciler) handleUpdateCluster(
	cc *clustersv1alpha1.Cassandra,
	l logr.Logger,
	ctx context.Context,
) *reconcile.Result {
	r.Scheduler.RemoveJob(cc.GetJobID(scheduler.StatusChecker))

	currentClusterStatus, err := r.API.GetClusterStatus(cc.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Cassandra cluster status from the Instaclustr API",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
		)
		return &models.ReconcileResult
	}

	if currentClusterStatus.Status != StatusRUNNING {
		l.Error(instaclustr.ClusterNotRunning, "Cluster is not ready to resize",
			"Cluster name", cc.Spec.Name,
			"Cluster status", cc.Status,
		)
		return &models.ReconcileResult
	}

	result := apiv2.CompareCassandraDCs(cc.Spec.DataCentres, currentClusterStatus)
	if result != nil {
		err = r.API.UpdateCluster(cc.Status.ID,
			instaclustr.CassandraEndpoint,
			result,
		)

		if errors.Is(err, instaclustr.ClusterIsNotReadyToResize) {
			l.Error(err, "cluster is not ready to resize",
				"Cluster name", cc.Spec.Name,
				"Cluster status", cc.Status,
			)
			return &models.ReconcileResult
		}

		if err != nil {
			l.Error(
				err, "cannot update Cassandra cluster status from the Instaclustr API",
				"Cluster name", cc.Spec.Name,
				"Cluster ID", cc.Status.ID,
			)
			return &models.ReconcileResult
		}

		patch := cc.NewPatch()
		cc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.Patch(ctx, cc, patch)
		if err != nil {
			l.Error(err, "cannot patch Cassandra cluster",
				"Cluster name", cc.Spec.Name,
				"Cluster ID", cc.Status.ID,
				"Kind", cc.Kind,
				"API Version", cc.APIVersion,
				"Namespace", cc.Namespace,
				"Cluster metadata", cc.ObjectMeta,
			)
			return &models.ReconcileResult
		}

		err = r.startClusterStatusJob(cc)
		if err != nil {
			l.Error(err, "cannot start cluster status job",
				"Cassandra cluster ID", cc.Status.ID)
			return &models.ReconcileResult
		}

		l.Info(
			"Cassandra cluster status has been updated",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
			"Kind", cc.Kind,
			"API Version", cc.APIVersion,
			"Namespace", cc.Namespace,
		)
	}
	return &models.ReconcileResult
}

func (r *CassandraReconciler) handleDeleteCluster(
	cc *clustersv1alpha1.Cassandra,
	l logr.Logger,
	ctx context.Context,
) *reconcile.Result {

	patch := cc.NewPatch()
	err := r.Patch(ctx, cc, patch)
	if err != nil {
		l.Error(err, "cannot patch Cassandra cluster",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
			"Kind", cc.Kind,
			"API Version", cc.APIVersion,
			"Namespace", cc.Namespace,
			"Cluster metadata", cc.ObjectMeta,
		)
		return &models.ReconcileResult
	}

	status, err := r.API.GetClusterStatus(cc.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "cannot get Cassandra cluster status from the Instaclustr API",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
			"Kind", cc.Kind,
			"API Version", cc.APIVersion,
			"Namespace", cc.Namespace,
		)
		return &models.ReconcileResult
	}

	if status != nil {
		r.Scheduler.RemoveJob(cc.GetJobID(scheduler.StatusChecker))
		err = r.API.DeleteCluster(cc.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "cannot delete Cassandra cluster",
				"Cluster name", cc.Spec.Name,
				"Status", cc.Status.Status,
				"Kind", cc.Kind,
				"API Version", cc.APIVersion,
				"Namespace", cc.Namespace,
			)
			return &models.ReconcileResult
		}

		return &models.ReconcileResult
	}

	controllerutil.RemoveFinalizer(cc, models.DeletionFinalizer)
	cc.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent

	err = r.Patch(ctx, cc, patch)
	if err != nil {
		l.Error(err, "cannot patch Cassandra cluster",
			"Cluster name", cc.Spec.Name,
			"Cluster ID", cc.Status.ID,
			"Kind", cc.Kind,
			"API Version", cc.APIVersion,
			"Namespace", cc.Namespace,
			"Cluster metadata", cc.ObjectMeta,
		)
		return &models.ReconcileResult
	}

	l.Info("Cassandra cluster has been deleted",
		"Cluster name", cc.Spec.Name,
		"Cluster ID", cc.Status.ID,
		"Kind", cc.Kind,
		"API Version", cc.APIVersion,
		"Namespace", cc.Namespace,
	)

	return &models.ReconcileResult
}

func (r *CassandraReconciler) startClusterStatusJob(cassandraCluster *clustersv1alpha1.Cassandra) error {
	job := r.newWatchStatusJob(cassandraCluster)

	err := r.Scheduler.ScheduleJob(cassandraCluster.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CassandraReconciler) newWatchStatusJob(cassandraCluster *clustersv1alpha1.Cassandra) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraStatusClusterJob")
	return func() error {
		instaclusterStatus, err := r.API.GetClusterStatus(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
		if err != nil {
			l.Error(err, "cannot get cassandraCluster instaclusterStatus", "ClusterID", cassandraCluster.Status.ID)
			return err
		}

		if !reflect.DeepEqual(*instaclusterStatus, cassandraCluster.Status.ClusterStatus) {
			l.Info("Cassandra status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclusterStatus", instaclusterStatus,
				"cassandraCluster.Status.ClusterStatus", cassandraCluster.Status.ClusterStatus)
			cassandraCluster.Status.ClusterStatus = *instaclusterStatus
			err := r.Status().Update(context.Background(), cassandraCluster)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func isResourceDeleting(obj client.Object) bool {
	if obj.GetDeletionTimestamp() != nil {
		obj.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
		return true
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cassandra{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				// for operator reboots
				if isResourceDeleting(event.Object) {
					return true
				}

				annotations := event.Object.GetAnnotations()
				annotations[models.ResourceStateAnnotation] = models.CreatingEvent
				event.Object.SetAnnotations(annotations)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if isResourceDeleting(event.ObjectNew) {
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent})
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
