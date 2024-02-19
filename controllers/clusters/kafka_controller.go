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
	"fmt"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// KafkaReconciler reconciles a Kafka object
type KafkaReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkas/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var k v1beta1.Kafka
	err := r.Client.Get(ctx, req.NamespacedName, &k)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka custom resource is not found", "namespaced name ", req.NamespacedName)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch Kafka", "request", req)
		return reconcile.Result{}, err
	}

	switch k.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, &k, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, &k, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, &k, l)
	case models.GenericEvent:
		l.Info("Event isn't handled", "cluster name", k.Spec.Name, "request", req,
			"event", k.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}

	return models.ExitReconcile, nil
}

func (r *KafkaReconciler) createCluster(ctx context.Context, k *v1beta1.Kafka, l logr.Logger) error {
	l.Info("Creating cluster",
		"cluster name", k.Spec.Name,
		"data centres", k.Spec.DataCentres)

	b, err := r.API.CreateClusterRaw(instaclustr.KafkaEndpoint, k.Spec.ToInstAPI())
	if err != nil {
		return fmt.Errorf("failed to create kafka cluster, err: %w", err)
	}

	instaModel := models.KafkaCluster{}
	err = json.Unmarshal(b, &instaModel)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json to kafka model, err: %w", err)
	}

	r.EventRecorder.Eventf(
		k, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s",
		instaModel.ID,
	)

	k.Spec.FromInstAPI(&instaModel)
	k.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Update(ctx, k)
	if err != nil {
		return fmt.Errorf("failed to update kafka spec, err: %w", err)
	}

	k.Status.FromInstAPI(&instaModel)
	err = r.Status().Update(ctx, k)
	if err != nil {
		return fmt.Errorf("failed to update kafka status, err: %w", err)
	}

	l.Info("Cluster has been created",
		"cluster ID", k.Status.ID,
	)

	return nil
}

func (r *KafkaReconciler) startJobs(k *v1beta1.Kafka) error {
	err := r.startSyncJob(k)
	if err != nil {
		return fmt.Errorf("failed to start cluster synchronize, err: %w", err)
	}

	r.EventRecorder.Eventf(
		k, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	if k.Spec.UserRefs != nil && k.Status.AvailableUsers == nil {
		err = r.startUsersCreationJob(k)
		if err != nil {
			return fmt.Errorf("failed to start user creation job, err: %w", err)
		}

		r.EventRecorder.Event(k, models.Normal, models.Created,
			"Cluster user creation job is started",
		)
	}

	return nil
}

func (r *KafkaReconciler) handleCreateCluster(ctx context.Context, k *v1beta1.Kafka, l logr.Logger) (reconcile.Result, error) {
	l = l.WithName("Kafka creation Event")

	if k.Status.ID == "" {
		err := r.createCluster(ctx, k, l)
		if err != nil {
			r.EventRecorder.Eventf(k, models.Warning, models.CreationFailed,
				"Failed to create Kafka cluster. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}
	}

	if k.Status.State != models.DeletedStatus {
		patch := k.NewPatch()
		k.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(k, models.DeletionFinalizer)
		err := r.Patch(ctx, k, patch)
		if err != nil {
			r.EventRecorder.Eventf(k, models.Warning, models.CreationFailed,
				"Failed to update resource metadata. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}

		err = r.startJobs(k)
		if err != nil {
			r.EventRecorder.Eventf(k, models.Warning, models.CreationFailed,
				"Failed to start cluster jobs. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}
	}

	return models.ExitReconcile, nil
}

func (r *KafkaReconciler) handleUpdateCluster(
	ctx context.Context,
	k *v1beta1.Kafka,
	req ctrl.Request,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("Kafka update Event")

	instaModel, err := r.API.GetKafka(k.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get cluster from the Instaclustr", "cluster ID", k.Status.ID)
		return reconcile.Result{}, err
	}

	iKafka := v1beta1.Kafka{}
	iKafka.FromInstAPI(instaModel)

	if k.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.KafkaSpec](r.EventRecorder, r.Client, k, &iKafka, l)
	}

	if k.Spec.ClusterSettingsNeedUpdate(&iKafka.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iKafka.Spec.Description,
			"instaclustr two factor delete", iKafka.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(k.Status.ID, k.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", k.Status.ID, "cluster spec", k.Spec)
			r.EventRecorder.Eventf(k, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	err = handleUsersChanges(ctx, r.Client, r, k)
	if err != nil {
		l.Error(err, "Failed to handle users changes")
		r.EventRecorder.Eventf(k, models.Warning, models.PatchFailed,
			"Handling users changes is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	if k.Spec.IsEqual(iKafka.Spec) {
		return models.ExitReconcile, nil
	}

	l.Info("Update request to Instaclustr API has been sent",
		"spec data centres", k.Spec.DataCentres,
		"resize settings", k.Spec.ResizeSettings,
	)

	err = r.API.UpdateCluster(k.Status.ID, instaclustr.KafkaEndpoint, k.Spec.ToInstAPIUpdate())
	if err != nil {
		l.Error(err, "Unable to update cluster on Instaclustr",
			"cluster name", k.Spec.Name, "cluster state", k.Status.State)

		r.EventRecorder.Eventf(k, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	patch := k.NewPatch()
	k.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, k, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", k.Spec.Name, "cluster ID", k.Status.ID)

		r.EventRecorder.Eventf(k, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", k.Spec.Name,
		"cluster ID", k.Status.ID,
		"data centres", k.Spec.DataCentres,
	)

	r.EventRecorder.Event(k, models.Normal, models.UpdatedEvent, "Cluster has been updated")

	return models.ExitReconcile, nil
}

func (r *KafkaReconciler) handleDeleteCluster(ctx context.Context, k *v1beta1.Kafka, l logr.Logger) (reconcile.Result, error) {
	l = l.WithName("Kafka deletion Event")

	_, err := r.API.GetKafka(k.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get cluster from the Instaclustr API",
			"cluster name", k.Spec.Name,
			"cluster state", k.Status.State)
		r.EventRecorder.Eventf(
			k, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	patch := k.NewPatch()
	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", k.Spec.Name,
			"cluster ID", k.Status.ID)

		err = r.API.DeleteCluster(k.Status.ID, instaclustr.KafkaEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete cluster",
				"cluster name", k.Spec.Name,
				"cluster state", k.Status.State)
			r.EventRecorder.Eventf(
				k, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Eventf(
			k, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.",
		)

		if k.Spec.TwoFactorDelete != nil {
			k.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			k.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, k, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", k.Spec.Name,
					"cluster state", k.Status.State)
				r.EventRecorder.Eventf(
					k, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", k.Status.ID)

			r.EventRecorder.Event(k, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	err = detachUsers(ctx, r.Client, r, k)
	if err != nil {
		l.Error(err, "Failed to detach users from the cluster")
		r.EventRecorder.Eventf(k, models.Warning, models.DeletionFailed,
			"Detaching users from the cluster is failed. Reason: %w", err,
		)
		return reconcile.Result{}, err
	}

	r.Scheduler.RemoveJob(k.GetJobID(scheduler.StatusChecker))
	r.Scheduler.RemoveJob(k.GetJobID(scheduler.UserCreator))
	controllerutil.RemoveFinalizer(k, models.DeletionFinalizer)
	k.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, k, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", k.Spec.Name)
		r.EventRecorder.Eventf(
			k, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, k.Name, k.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Kafka cluster expose service",
			"cluster ID", k.Status.ID,
			"cluster name", k.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Cluster was deleted",
		"cluster name", k.Spec.Name,
		"cluster ID", k.Status.ID)

	r.EventRecorder.Eventf(
		k, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

//nolint:unused,deadcode
func (r *KafkaReconciler) startClusterOnPremisesIPsJob(k *v1beta1.Kafka, b *onPremisesBootstrap) error {
	job := newWatchOnPremisesIPsJob(k.Kind, b)

	err := r.Scheduler.ScheduleJob(k.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaReconciler) startSyncJob(kafka *v1beta1.Kafka) error {
	job := r.newSyncJob(kafka)

	err := r.Scheduler.ScheduleJob(kafka.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaReconciler) newSyncJob(k *v1beta1.Kafka) scheduler.Job {
	l := log.Log.WithValues("syncJob", k.GetJobID(scheduler.StatusChecker), "clusterID", k.Status.ID)

	return func() error {
		namespacedName := client.ObjectKeyFromObject(k)
		err := r.Get(context.Background(), namespacedName, k)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(k.GetJobID(scheduler.StatusChecker))
			r.Scheduler.RemoveJob(k.GetJobID(scheduler.UserCreator))
			r.Scheduler.RemoveJob(k.GetJobID(scheduler.BackupsChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", k.Name)
			return err
		}

		instaModel, err := r.API.GetKafka(k.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if k.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), k, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), k)
			}

			l.Error(err, "Cannot get cluster from the Instaclustr", "cluster ID", k.Status.ID)
			return err
		}

		iKafka := v1beta1.Kafka{}
		iKafka.FromInstAPI(instaModel)

		if !k.Status.Equals(&iKafka.Status) {
			l.Info("Updating cluster status")

			areDCsEqual := k.Status.DCsEqual(&iKafka.Status)

			patch := k.NewPatch()
			k.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), k, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", k.Spec.Name, "cluster state", k.Status.State)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iKafka.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					k.Name,
					k.Namespace,
					k.Spec.PrivateNetwork,
					nodes,
					models.KafkaConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := k.Spec.IsEqual(iKafka.Spec)

		if equals && k.Annotations[models.ExternalChangesAnnotation] == models.True {
			err = reconcileExternalChanges(r.Client, r.EventRecorder, k)
			if err != nil {
				return err
			}
		}
		if iKafka.Status.CurrentClusterOperationStatus == models.NoOperation &&
			k.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

			patch := k.NewPatch()
			k.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), k, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", k.Spec.Name, "cluster state", k.Status.State)
				return err
			}

			l.Info("The k8s specification is different from Instaclustr Console. Update operations are blocked.",
				"instaclustr data", iKafka.Spec, "k8s resource spec", k.Spec)

			msgDiffSpecs, err := createSpecDifferenceMessage(k.Spec, iKafka.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iKafka.Spec, "k8s resource spec", k.Spec)
				return err
			}
			r.EventRecorder.Eventf(k, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), k)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", k.Spec.Name,
				"cluster ID", k.Status.ID,
			)
			return err
		}

		if k.Status.State == models.RunningStatus && k.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := k.NewPatch()
			for _, dc := range k.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", k.Spec.Name,
						"cluster ID", k.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), k, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", k.Spec.Name,
						"cluster ID", k.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *KafkaReconciler) startUsersCreationJob(k *v1beta1.Kafka) error {
	job := r.newUsersCreationJob(k)

	err := r.Scheduler.ScheduleJob(k.GetJobID(scheduler.UserCreator), scheduler.UserCreationInterval, job)
	if err != nil {
		return err
	}
	return nil
}

func (r *KafkaReconciler) newUsersCreationJob(k *v1beta1.Kafka) scheduler.Job {
	l := log.Log.WithValues("component", "kafkaUsersCreationJob")
	return func() error {
		ctx := context.Background()

		err := r.Get(ctx, types.NamespacedName{
			Namespace: k.Namespace,
			Name:      k.Name,
		}, k)

		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		if k.Status.State != models.RunningStatus {
			l.Info("User creation job is scheduled")
			r.EventRecorder.Eventf(k, models.Normal, models.CreationFailed,
				"User creation job is scheduled, cluster is not in the running state",
			)

			return nil
		}

		err = handleUsersChanges(ctx, r.Client, r, k)
		if err != nil {
			l.Error(err, "Failed to create users for the cluster")
			r.EventRecorder.Eventf(k, models.Warning, models.CreationFailed,
				"Failed to create users for the cluster. Reason: %v", err)
			return err
		}

		l.Info("User creation job successfully finished")
		r.EventRecorder.Eventf(k, models.Normal, models.Created,
			"User creation job successfully finished",
		)

		r.Scheduler.RemoveJob(k.GetJobID(scheduler.UserCreator))

		return nil
	}
}

func (r *KafkaReconciler) handleExternalDelete(ctx context.Context, k *v1beta1.Kafka) error {
	l := log.FromContext(ctx)

	patch := k.NewPatch()
	k.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, k, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(k, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(k.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(k.GetJobID(scheduler.UserCreator))
	r.Scheduler.RemoveJob(k.GetJobID(scheduler.StatusChecker))

	return nil
}

func (r *KafkaReconciler) NewUserResource() userObject {
	return &clusterresourcesv1beta1.KafkaUser{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: r.RateLimiter}).
		For(&v1beta1.Kafka{}, builder.WithPredicates(predicate.Funcs{
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
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				newObj := event.ObjectNew.(*v1beta1.Kafka)

				if newObj.Status.ID == "" && newObj.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
					return false
				}

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

func (r *KafkaReconciler) reconcileMaintenanceEvents(ctx context.Context, k *v1beta1.Kafka) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(k.Status.ID)
	if err != nil {
		return err
	}

	if !k.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := k.NewPatch()
		k.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, k, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", k.Status.ID,
			"events", k.Status.MaintenanceEvents,
		)
	}

	return nil
}
