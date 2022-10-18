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
	"encoding/json"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	apiv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	CassandraFinalizerName = "cassandra-insta-finalizer/finalizer"
	StatusRUNNING          = "RUNNING"
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

	cassandraCluster := &clustersv1alpha1.Cassandra{}
	err := r.Client.Get(ctx, req.NamespacedName, cassandraCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Scheduler.RemoveJob(req.NamespacedName.String())
			return reconcile.Result{}, nil
		}
		l.Error(err, "unable to fetch Cassandra cluster")
		return reconcile.Result{}, err
	}

	cassandraAnnotations := cassandraCluster.Annotations
	l.Info("CASSANDRA ANNOTATIONS", "Annot", cassandraAnnotations)
	switch cassandraAnnotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		err = r.HandleCreateCluster(cassandraCluster, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot create Cassandra cluster",
				"Cluster name", cassandraCluster.Spec.Name,
			)
		}
		return reconcile.Result{}, nil
	case models.UpdatingEvent:
		reconcileResult, err := r.HandleUpdateCluster(cassandraCluster, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot update Cassandra cluster",
				"Cluster name", cassandraCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}
		return *reconcileResult, nil
	case models.DeletingEvent:
		err = r.HandleDeleteCluster(cassandraCluster, &l, &ctx)
		if err != nil {
			l.Error(err, "cannot delete Cassandra cluster",
				"Cluster name", cassandraCluster.Spec.Name,
			)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	default:
		l.Info("UNKNOWN EVENT",
			"Cluster spec", cassandraCluster.Spec)
		return reconcile.Result{}, err
	}
}

func (r *CassandraReconciler) HandleCreateCluster(
	cassandraCluster *clustersv1alpha1.Cassandra,
	l *logr.Logger,
	ctx *context.Context,
) error {

	l.Info(
		"Creating Cassandra cluster",
		"Cluster name", cassandraCluster.Spec.Name,
		"Data centres", cassandraCluster.Spec.DataCentres,
	)

	cassandraSpec := apiv2.CassandraToInstAPI(&cassandraCluster.Spec)

	id, err := r.API.CreateCluster(instaclustr.CassandraEndpoint, cassandraSpec)
	if err != nil {
		l.Error(
			err, "cannot create Cassandra cluster",
			"Cassandra cluster spec", cassandraCluster.Spec,
		)
		return err
	}

	cassandraCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent

	err = r.patchClusterMetadata(ctx, cassandraCluster, l)
	if err != nil {
		l.Error(err, "cannot patch Cassandra cluster",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster ID", cassandraCluster.Status.ID,
			"Cluster metadata", cassandraCluster.ObjectMeta,
		)
		return err
	}

	cassandraCluster.Status.ID = id
	err = r.Status().Update(*ctx, cassandraCluster)
	if err != nil {
		l.Error(err, "cannot update Cassandra cluster status",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster ID", cassandraCluster.Status.ID,
			"Kind", cassandraCluster.Kind,
			"API Version", cassandraCluster.APIVersion,
			"Namespace", cassandraCluster.Namespace,
		)
		return err
	}

	err = r.startClusterStatusJob(cassandraCluster)
	if err != nil {
		l.Error(err, "cannot start cluster status job",
			"Cassandra cluster ID", cassandraCluster.Status.ID)
		return err
	}

	l.Info(
		"Cassandra cluster has been created",
		"Cluster name", cassandraCluster.Spec.Name,
		"Cluster ID", cassandraCluster.Status.ID,
		"Kind", cassandraCluster.Kind,
		"API Version", cassandraCluster.APIVersion,
		"Namespace", cassandraCluster.Namespace,
	)
	return nil
}

func (r *CassandraReconciler) HandleUpdateCluster(
	cassandraCluster *clustersv1alpha1.Cassandra,
	l *logr.Logger,
	ctx *context.Context,
) (*reconcile.Result, error) {

	r.Scheduler.RemoveJob(cassandraCluster.GetJobID(scheduler.StatusChecker))
	currentClusterStatus, err := r.API.GetClusterStatus(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
	if err != client.IgnoreNotFound(err) {
		l.Error(
			err, "cannot get Cassandra cluster status from the Instaclustr API",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster ID", cassandraCluster.Status.ID,
		)
		return &reconcile.Result{}, err
	}

	if currentClusterStatus.Status != StatusRUNNING {
		l.Error(instaclustr.ClusterNotRunning, "Cluster is not ready to resize",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster status", cassandraCluster.Status,
		)
		return &reconcile.Result{
			Requeue:      true,
			RequeueAfter: models.Requeue60,
		}, nil
	}

	result := apiv2.CompareCassandraDCs(cassandraCluster.Spec.DataCentres, currentClusterStatus)
	if result != nil {
		err = r.API.UpdateCluster(cassandraCluster.Status.ID,
			instaclustr.CassandraEndpoint,
			result,
		)

		if errors.Is(err, instaclustr.ClusterIsNotReadyToResize) {
			l.Error(err, "cluster is not ready to resize",
				"Cluster name", cassandraCluster.Spec.Name,
				"Cluster status", cassandraCluster.Status,
			)
			return &reconcile.Result{
				Requeue:      true,
				RequeueAfter: models.Requeue60,
			}, nil
		}

		if err != nil {
			l.Error(
				err, "cannot get Cassandra cluster status from the Instaclustr API",
				"Cluster name", cassandraCluster.Spec.Name,
				"Cluster ID", cassandraCluster.Status.ID,
			)
			return &reconcile.Result{}, err
		}

		cassandraCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent

		err = r.patchClusterMetadata(ctx, cassandraCluster, l)
		if err != nil {
			l.Error(err, "cannot patch Cassandra metadata",
				"Cluster name", cassandraCluster.Spec.Name,
				"Cluster metadata", cassandraCluster.ObjectMeta,
			)
			return &reconcile.Result{}, err
		}

		err = r.startClusterStatusJob(cassandraCluster)
		if err != nil {
			l.Error(err, "cannot start cluster status job",
				"Cassandra cluster ID", cassandraCluster.Status.ID)
			return &reconcile.Result{}, err
		}

		l.Info(
			"Cassandra cluster status has been updated",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster ID", cassandraCluster.Status.ID,
			"Kind", cassandraCluster.Kind,
			"API Version", cassandraCluster.APIVersion,
			"Namespace", cassandraCluster.Namespace,
		)
	}
	return &reconcile.Result{}, err
}

func (r *CassandraReconciler) HandleDeleteCluster(
	cassandraCluster *clustersv1alpha1.Cassandra,
	l *logr.Logger,
	ctx *context.Context,
) error {

	r.Scheduler.RemoveJob(cassandraCluster.GetJobID(scheduler.StatusChecker))
	err := r.API.DeleteCluster(cassandraCluster.Status.ID, instaclustr.CassandraEndpoint)
	if err != nil {
		return err
	}

	l.Info("Cassandra cluster has been deleted",
		"Cluster name", cassandraCluster.Spec.Name,
		"Cluster id", cassandraCluster.Status.ID,
	)

	controllerutil.RemoveFinalizer(cassandraCluster, CassandraFinalizerName)
	err = r.Update(*ctx, cassandraCluster)
	if err != nil {
		l.Error(err, "cannot update Cassandra cluster status",
			"Cluster name", cassandraCluster.Spec.Name,
			"Cluster ID", cassandraCluster.Status.ID,
			"Kind", cassandraCluster.Kind,
			"API Version", cassandraCluster.APIVersion,
			"Namespace", cassandraCluster.Namespace,
		)
		return err
	}

	return nil
}

func (r *CassandraReconciler) patchClusterMetadata(
	ctx *context.Context,
	cassandra *clustersv1alpha1.Cassandra,
	logger *logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(cassandra.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(cassandra.Finalizers)
	if err != nil {
		return err
	}

	finzlizersPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(*ctx, cassandra, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("Cassandra cluster patched",
		"Cluster name", cassandra.Spec.Name,
		"Finalizers", cassandra.Finalizers,
		"Annotations", cassandra.Annotations,
	)
	return nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *CassandraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cassandra{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				event.Object.SetFinalizers([]string{CassandraFinalizerName})
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetGeneration() == event.ObjectOld.GetGeneration() {
					return false
				}

				if event.ObjectNew.GetDeletionTimestamp() != nil {
					event.ObjectNew.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.DeletingEvent})
					return true
				}

				event.ObjectNew.SetAnnotations(map[string]string{
					models.ResourceStateAnnotation: models.UpdatingEvent,
				})
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.GenericEvent})
				return true
			},
		})).
		Complete(r)
}
