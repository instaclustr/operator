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
	"github.com/instaclustr/operator/controllers/clusterresources"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// KafkaConnectReconciler reconciles a KafkaConnect object
type KafkaConnectReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/finalizers,verbs=update
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
func (r *KafkaConnectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	kc := &v1beta1.KafkaConnect{}
	err := r.Client.Get(ctx, req.NamespacedName, kc)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "KafkaConnect resource is not found", "request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch KafkaConnect", "request", req)
		return reconcile.Result{}, err
	}

	switch kc.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, kc, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, kc, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, kc, l)
	default:
		l.Info("Event isn't handled", "cluster name", kc.Spec.Name,
			"request", req, "event", kc.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *KafkaConnectReconciler) mergeManagedClusterFromRef(ctx context.Context, kc *v1beta1.KafkaConnect) error {
	managedCluster := kc.Spec.GetManagedCluster()
	if managedCluster == nil || managedCluster.ClusterRef == nil {
		return nil
	}

	targetClusterID, err := clusterresources.GetClusterID(r.Client, ctx, managedCluster.ClusterRef)
	if err != nil {
		return fmt.Errorf("failed to get managed cluster id by ref %s for kind %s, err: %w",
			managedCluster.ClusterRef.AsNamespacedName().String(),
			managedCluster.ClusterRef.ClusterKind,
			err)
	}

	managedCluster.TargetKafkaClusterID = targetClusterID

	return nil
}

func (r *KafkaConnectReconciler) createKafkaConnect(ctx context.Context, kc *v1beta1.KafkaConnect) (*models.KafkaConnectCluster, error) {
	err := r.mergeManagedClusterFromRef(ctx, kc)
	if err != nil {
		return nil, err
	}

	b, err := r.API.CreateClusterRaw(instaclustr.KafkaConnectEndpoint, kc.Spec.ToInstAPI())
	if err != nil {
		return nil, fmt.Errorf("failed to create KafkaConnect cluster, err: %w", err)
	}

	var instaModel models.KafkaConnectCluster
	err = json.Unmarshal(b, &instaModel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body to KafkaConnect model, err: %w", err)
	}

	return &instaModel, nil
}

func (r *KafkaConnectReconciler) createCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) error {
	if !kc.Spec.Inherits() {
		id, err := getClusterIDByName(r.API, models.KafkaConnectAppType, kc.Spec.Name)
		if err != nil {
			return err
		}

		if id != "" {
			l.Info("Cluster with provided name already exists", "name", kc.Spec.Name, "clusterID", id)
			return fmt.Errorf("cluster %s already exists, please change name property", kc.Spec.Name)
		}
	}

	var instaModel *models.KafkaConnectCluster
	var err error

	switch {
	case kc.Spec.Inherits():
		l.Info("Inheriting from the cluster", "clusterID", kc.Spec.InheritsFrom)
		instaModel, err = r.API.GetKafkaConnect(kc.Spec.InheritsFrom)
	default:
		instaModel, err = r.createKafkaConnect(ctx, kc)
	}
	if err != nil {
		return err
	}

	kc.Spec.FromInstAPI(instaModel)
	kc.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Update(ctx, kc)
	if err != nil {
		return fmt.Errorf("failed to update resource spec, err: %w", err)
	}

	kc.Status.FromInstAPI(instaModel)
	err = r.Status().Update(ctx, kc)
	if err != nil {
		return fmt.Errorf("failed to update resource status, err: %w", err)
	}

	l.Info("KafkaConnect cluster has been created",
		"clusterID", kc.Status.ID,
	)
	r.EventRecorder.Eventf(
		kc, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s",
		kc.Status.ID,
	)

	err = r.createDefaultSecret(ctx, kc, l)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaConnectReconciler) handleCreateCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) (reconcile.Result, error) {
	l = l.WithName("Creation Event")

	if kc.Status.ID == "" {
		err := r.createCluster(ctx, kc, l)
		if err != nil {
			r.EventRecorder.Eventf(kc, models.Warning, models.CreationFailed,
				"Failed to create cluster resource. Reason: %v", err,
			)
			return reconcile.Result{}, err
		}
	}

	if kc.Status.State != models.DeletedStatus {
		patch := kc.NewPatch()
		kc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kc, models.DeletionFinalizer)
		err := r.Patch(ctx, kc, patch)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.startSyncJob(kc)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to start sync job, err: %w", err)
		}

		r.EventRecorder.Eventf(
			kc, models.Normal, models.Created,
			"Cluster sync job is started",
		)
	}

	return models.ExitReconcile, nil
}

func (r *KafkaConnectReconciler) handleUpdateCluster(
	ctx context.Context,
	kc *v1beta1.KafkaConnect,
	req ctrl.Request,
	l logr.Logger,
) (reconcile.Result, error) {
	l = l.WithName("Update Event")

	instaModel, err := r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get Kafka Connect from Instaclustr",
			"ClusterID", kc.Status.ID)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	iKC := &v1beta1.KafkaConnect{}
	iKC.FromInstAPI(instaModel)

	if kc.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.KafkaConnectSpec](r.EventRecorder, r.Client, kc, iKC, l)
	}

	if kc.Spec.ClusterSettingsNeedUpdate(&iKC.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iKC.Spec.Description,
			"instaclustr two factor delete", iKC.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(kc.Status.ID, kc.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", kc.Status.ID, "cluster spec", kc.Spec)
			r.EventRecorder.Eventf(kc, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	if !kc.Spec.Equals(&iKC.Spec) {
		l.Info("Update request to Instaclustr API has been sent",
			"spec data centres", kc.Spec.DataCentres)

		err = r.API.UpdateKafkaConnect(kc.Status.ID, kc.Spec.NewDCsUpdate())
		if err != nil {
			l.Error(err, "Unable to update Kafka Connect cluster",
				"cluster name", kc.Spec.Name,
				"cluster status", kc.Status)

			r.EventRecorder.Eventf(kc, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v", err)

			return reconcile.Result{}, err
		}
	}

	patch := kc.NewPatch()
	kc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Unable to patch Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster status", kc.Status)

		r.EventRecorder.Eventf(kc, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", kc.Spec.Name,
		"cluster ID", kc.Status.ID,
		"data centres", kc.Spec.DataCentres,
	)

	return models.ExitReconcile, nil
}

func (r *KafkaConnectReconciler) handleDeleteCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) (reconcile.Result, error) {
	l = l.WithName("Deletion Event")

	_, err := r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster state", kc.Status.State)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	patch := kc.NewPatch()

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", kc.Spec.Name,
			"cluster ID", kc.Status.ID)

		err = r.API.DeleteCluster(kc.Status.ID, instaclustr.KafkaConnectEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Kafka Connect cluster",
				"cluster name", kc.Spec.Name,
				"cluster state", kc.Status.State)
			r.EventRecorder.Eventf(
				kc, models.Warning, models.DeletionFailed,
				"Cluster deletion on the Instaclustr API is failed. Reason: %v",
				err,
			)
			return reconcile.Result{}, err
		}

		r.EventRecorder.Event(kc, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if kc.Spec.TwoFactorDelete != nil {
			kc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			kc.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", kc.Spec.Name,
					"cluster state", kc.Status.State)
				r.EventRecorder.Eventf(kc, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return reconcile.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", kc.Status.ID)

			r.EventRecorder.Event(kc, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile, nil
		}
	}

	r.Scheduler.RemoveJob(kc.GetJobID(scheduler.SyncJob))
	controllerutil.RemoveFinalizer(kc, models.DeletionFinalizer)
	kc.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Cannot patch remove finalizer from KC",
			"cluster name", kc.Spec.Name)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return reconcile.Result{}, err
	}

	err = exposeservice.Delete(r.Client, kc.Name, kc.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Kafka Connect cluster expose service",
			"cluster ID", kc.Status.ID,
			"cluster name", kc.Spec.Name,
		)

		return reconcile.Result{}, err
	}

	l.Info("Kafka Connect cluster was deleted",
		"cluster name", kc.Spec.Name,
		"cluster ID", kc.Status.ID)

	r.EventRecorder.Eventf(
		kc, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile, nil
}

func (r *KafkaConnectReconciler) createDefaultSecret(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) error {
	username, password, err := r.API.GetDefaultCredentialsV1(kc.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get default user creds for Kafka Connect cluster from the Instaclustr API",
			"cluster ID", kc.Status.ID,
		)
		r.EventRecorder.Eventf(kc, models.Warning, models.FetchFailed,
			"Default user password fetch from the Instaclustr API is failed. Reason: %v", err,
		)

		return err
	}

	patch := kc.NewPatch()
	secret := newDefaultUserSecret(username, password, kc.Name, kc.Namespace)
	err = controllerutil.SetOwnerReference(kc, secret, r.Scheme)
	if err != nil {
		l.Error(err, "Cannot set secret owner reference with default user credentials",
			"cluster ID", kc.Status.ID,
		)
		r.EventRecorder.Eventf(kc, models.Warning, models.SetOwnerRef,
			"Setting secret owner ref with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	err = r.Create(ctx, secret)
	if err != nil {
		l.Error(err, "Cannot create secret with default user credentials",
			"cluster ID", kc.Status.ID,
		)
		r.EventRecorder.Eventf(kc, models.Warning, models.CreationFailed,
			"Creating secret with default user credentials is failed. Reason: %v", err,
		)

		return err
	}

	l.Info("Default secret was created",
		"secret name", secret.Name,
		"secret namespace", secret.Namespace,
	)

	kc.Status.DefaultUserSecretRef = &v1beta1.Reference{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}

	err = r.Status().Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Cannot patch Kafka Connect resource",
			"cluster name", kc.Spec.Name,
			"status", kc.Status)

		r.EventRecorder.Eventf(
			kc, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return err
	}

	return nil
}

//nolint:unused,deadcode
func (r *KafkaConnectReconciler) startClusterOnPremisesIPsJob(k *v1beta1.KafkaConnect, b *onPremisesBootstrap) error {
	job := newWatchOnPremisesIPsJob(k.Kind, b)

	err := r.Scheduler.ScheduleJob(k.GetJobID(scheduler.OnPremisesIPsChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaConnectReconciler) startSyncJob(kc *v1beta1.KafkaConnect) error {
	job := r.newSyncJob(kc)

	err := r.Scheduler.ScheduleJob(kc.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaConnectReconciler) newSyncJob(kc *v1beta1.KafkaConnect) scheduler.Job {
	l := log.Log.WithValues("component", "kafkaConnectStatusClusterJob")

	return func() error {
		namespacedName := client.ObjectKeyFromObject(kc)
		err := r.Get(context.Background(), namespacedName, kc)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(kc.GetJobID(scheduler.SyncJob))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", kc.Name)
			return err
		}

		instaModel, err := r.API.GetKafkaConnect(kc.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if kc.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), kc, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), kc)
			}

			l.Error(err, "Cannot get Kafka Connect from Instaclustr",
				"cluster ID", kc.Status.ID)
			return err
		}

		iKC := &v1beta1.KafkaConnect{}
		iKC.FromInstAPI(instaModel)

		if !kc.Status.Equals(&iKC.Status) {
			l.Info("Kafka Connect status of k8s is different from Instaclustr. Reconcile statuses..")

			areDCsEqual := kc.Status.DCsEqual(iKC.Status.DataCentres)

			patch := kc.NewPatch()
			kc.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka Connect cluster",
					"cluster name", kc.Spec.Name, "cluster state", kc.Status.State)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iKC.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					kc.Name,
					kc.Namespace,
					kc.Spec.PrivateNetwork,
					nodes,
					models.KafkaConnectConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		equals := kc.Spec.Equals(&iKC.Spec)

		if equals && kc.Annotations[models.ExternalChangesAnnotation] == models.True {
			err = reconcileExternalChanges(r.Client, r.EventRecorder, kc)
			if err != nil {
				return err
			}
		} else if kc.Status.CurrentClusterOperationStatus == models.NoOperation &&
			kc.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!equals {

			patch := kc.NewPatch()
			kc.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", kc.Spec.Name, "cluster state", kc.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(kc.Spec, iKC.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iKC.Spec, "k8s resource spec", kc.Spec)
				return err
			}
			r.EventRecorder.Eventf(kc, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), kc)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", kc.Spec.Name,
				"cluster ID", kc.Status.ID,
			)
			return err
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaConnectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{RateLimiter: r.RateLimiter}).
		For(&v1beta1.KafkaConnect{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				event.Object.SetAnnotations(map[string]string{models.ResourceStateAnnotation: models.CreatingEvent})
				confirmDeletion(event.Object)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				newObj := event.ObjectNew.(*v1beta1.KafkaConnect)
				if newObj.Generation == event.ObjectOld.GetGeneration() {
					return false
				}

				if newObj.Status.ID == "" && newObj.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
					return false
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
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

func (r *KafkaConnectReconciler) reconcileMaintenanceEvents(ctx context.Context, kc *v1beta1.KafkaConnect) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(kc.Status.ID)
	if err != nil {
		return err
	}

	if !kc.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := kc.NewPatch()
		kc.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, kc, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", kc.Status.ID,
			"events", kc.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *KafkaConnectReconciler) handleExternalDelete(ctx context.Context, kc *v1beta1.KafkaConnect) error {
	l := log.FromContext(ctx)

	patch := kc.NewPatch()
	kc.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, kc, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(kc, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(kc.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(kc.GetJobID(scheduler.SyncJob))

	return nil
}
