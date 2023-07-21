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

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// KafkaConnectReconciler reconciles a KafkaConnect object
type KafkaConnectReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=kafkaconnects/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KafkaConnectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	kafkaConnect := &v1beta1.KafkaConnect{}
	err := r.Client.Get(ctx, req.NamespacedName, kafkaConnect)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "KafkaConnect resource is not found", "request", req)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch KafkaConnect", "request", req)
		return models.ReconcileRequeue, err
	}

	switch kafkaConnect.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, kafkaConnect, l), nil
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, kafkaConnect, l), nil
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, kafkaConnect, l), nil
	default:
		l.Info("Event isn't handled", "cluster name", kafkaConnect.Spec.Name,
			"request", req, "event", kafkaConnect.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}
}

func (r *KafkaConnectReconciler) handleCreateCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Creation Event")

	if kc.Status.ID == "" {
		l.Info("Creating Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"data centres", kc.Spec.DataCentres)

		patch := kc.NewPatch()
		var err error

		kc.Status.ID, err = r.API.CreateCluster(instaclustr.KafkaConnectEndpoint, kc.Spec.ToInstAPI())
		if err != nil {
			l.Error(err, "cannot create Kafka Connect in Instaclustr", "Kafka Connect manifest", kc.Spec)
			r.EventRecorder.Eventf(
				kc, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			kc, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s",
			kc.Status.ID,
		)
		err = r.Status().Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "cannot patch Kafka Connect status ", "KC ID", kc.Status.ID)
			r.EventRecorder.Eventf(
				kc, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		kc.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(kc, models.DeletionFinalizer)
		err = r.Patch(ctx, kc, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka Connect", "cluster name", kc.Spec.Name)
			r.EventRecorder.Eventf(
				kc, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
	}

	err := r.startClusterStatusJob(kc)
	if err != nil {
		l.Error(err, "Cannot start cluster status job", "cluster ID", kc.Status.ID)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster has been created",
		"cluster ID", kc.Status.ID)

	r.EventRecorder.Eventf(
		kc, models.Normal, models.Created,
		"Cluster status check job is started",
	)

	err = r.createDefaultSecret(ctx, kc, l)
	if err != nil {
		l.Error(err, "Cannot create default secret for Kafka Connect",
			"cluster name", kc.Spec.Name,
			"clusterID", kc.Status.ID,
		)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.CreationFailed,
			"Default user secret creation on the Instaclustr is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue
	}

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) handleUpdateCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Update Event")

	iData, err := r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil {
		l.Error(err, "Cannot get Kafka Connect from Instaclustr",
			"ClusterID", kc.Status.ID)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	iKC, err := kc.FromInst(iData)
	if err != nil {
		l.Error(err, "Cannot convert Kafka Connect from Instaclustr",
			"ClusterID", kc.Status.ID)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if kc.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(kc, iKC, l)
	}

	if !kc.Spec.IsEqual(iKC.Spec) {
		err = r.API.UpdateKafkaConnect(kc.Status.ID, kc.Spec.NewDCsUpdate())
		if err != nil {
			l.Error(err, "Unable to update Kafka Connect cluster",
				"cluster name", kc.Spec.Name,
				"cluster status", kc.Status,
			)
			r.EventRecorder.Eventf(
				kc, models.Warning, models.UpdateFailed,
				"Cluster update on the Instaclustr API is failed. Reason: %v",
				err,
			)

			patch := kc.NewPatch()
			kc.Annotations[models.UpdateQueuedAnnotation] = models.True
			kc.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			err = r.Patch(ctx, kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", kc.Spec.Name, "cluster ID", kc.Status.ID)

				r.EventRecorder.Eventf(
					kc, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err,
				)
				return models.ReconcileRequeue
			}
			return models.ReconcileRequeue
		}
	}

	patch := kc.NewPatch()
	kc.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	kc.Annotations[models.UpdateQueuedAnnotation] = ""
	err = r.Patch(ctx, kc, patch)
	if err != nil {
		l.Error(err, "Unable to patch Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster status", kc.Status,
		)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster was updated",
		"cluster ID", kc.Status.ID,
	)

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) handleExternalChanges(k, ik *v1beta1.KafkaConnect, l logr.Logger) reconcile.Result {
	if !k.Spec.IsEqual(ik.Spec) {
		l.Info(msgSpecStillNoMatch,
			"specification of k8s resource", k.Spec,
			"data from Instaclustr ", ik.Spec)

		msgDiffSpecs, err := createSpecDifferenceMessage(k.Spec, ik.Spec)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", ik.Spec, "k8s resource spec", k.Spec)
			return models.ExitReconcile
		}
		r.EventRecorder.Eventf(k, models.Warning, models.ExternalChanges, msgDiffSpecs)
		return models.ExitReconcile
	}

	patch := k.NewPatch()

	k.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), k, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", k.Spec.Name, "cluster ID", k.Status.ID)

		r.EventRecorder.Eventf(k, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	l.Info("External changes have been reconciled", "resource ID", k.Status.ID)
	r.EventRecorder.Event(k, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return models.ExitReconcile
}

func (r *KafkaConnectReconciler) handleDeleteCluster(ctx context.Context, kc *v1beta1.KafkaConnect, l logr.Logger) reconcile.Result {
	l = l.WithName("Deletion Event")

	_, err := r.API.GetKafkaConnect(kc.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(err, "Cannot get Kafka Connect cluster",
			"cluster name", kc.Spec.Name,
			"cluster state", kc.Status.ClusterStatus.State)
		r.EventRecorder.Eventf(
			kc, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
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
			return models.ReconcileRequeue
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

				return models.ReconcileRequeue
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", kc.Status.ID)

			r.EventRecorder.Event(kc, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}
	}

	r.Scheduler.RemoveJob(kc.GetJobID(scheduler.StatusChecker))
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
		return models.ReconcileRequeue
	}

	err = exposeservice.Delete(r.Client, kc.Name, kc.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Kafka Connect cluster expose service",
			"cluster ID", kc.Status.ID,
			"cluster name", kc.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	l.Info("Kafka Connect cluster was deleted",
		"cluster name", kc.Spec.Name,
		"cluster ID", kc.Status.ID)

	r.EventRecorder.Eventf(
		kc, models.Normal, models.Deleted,
		"Cluster resource is deleted",
	)

	return models.ExitReconcile
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

	secret := kc.NewDefaultUserSecret(username, password)
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

	return nil
}

func (r *KafkaConnectReconciler) startClusterStatusJob(kc *v1beta1.KafkaConnect) error {
	job := r.newWatchStatusJob(kc)

	err := r.Scheduler.ScheduleJob(kc.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaConnectReconciler) newWatchStatusJob(kc *v1beta1.KafkaConnect) scheduler.Job {
	l := log.Log.WithValues("component", "kafkaConnectStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(kc)
		err := r.Get(context.Background(), namespacedName, kc)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(kc.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get cluster resource",
				"resource name", kc.Name)
			return err
		}

		iData, err := r.API.GetKafkaConnect(kc.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(kc.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", kc.Status.ClusterStatus.ID,
						"cluster name", kc.Spec.Name,
					)

					patch := kc.NewPatch()
					kc.Annotations[models.ClusterDeletionAnnotation] = ""
					kc.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), kc, patch)
					if err != nil {
						l.Error(err, "Cannot patch KafkaConnect cluster resource",
							"cluster ID", kc.Status.ID,
							"cluster name", kc.Spec.Name,
							"resource name", kc.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), kc)
					if err != nil {
						l.Error(err, "Cannot delete KafkaConnect cluster resource",
							"cluster ID", kc.Status.ID,
							"cluster name", kc.Spec.Name,
							"resource name", kc.Name,
						)

						return err
					}

					return nil
				}
			}

			l.Error(err, "Cannot get Kafka Connect from Instaclustr",
				"cluster ID", kc.Status.ID)
			return err
		}

		iKC, err := kc.FromInst(iData)
		if err != nil {
			l.Error(err, "Cannot convert Kafka Connect from Instaclustr",
				"cluster ID", kc.Status.ID)
			return err
		}

		if !areStatusesEqual(&iKC.Status.ClusterStatus, &kc.Status.ClusterStatus) {
			l.Info("Kafka Connect status of k8s is different from Instaclustr. Reconcile statuses..",
				"instaclustr status", iKC.Status,
				"status", kc.Status.ClusterStatus)

			areDCsEqual := areDataCentresEqual(iKC.Status.ClusterStatus.DataCentres, kc.Status.ClusterStatus.DataCentres)

			patch := kc.NewPatch()
			kc.Status.ClusterStatus = iKC.Status.ClusterStatus
			err = r.Status().Patch(context.Background(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka Connect cluster",
					"cluster name", kc.Spec.Name, "cluster state", kc.Status.State)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iKC.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					kc.Name,
					kc.Namespace,
					nodes,
					models.KafkaConnectConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iKC.Status.CurrentClusterOperationStatus == models.NoOperation &&
			kc.Annotations[models.UpdateQueuedAnnotation] != models.True &&
			!kc.Spec.IsEqual(iKC.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iKC.Spec, "k8s resource spec", kc.Spec)

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

		maintEvents, err := r.API.GetMaintenanceEvents(kc.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Kafka Connect cluster maintenance events",
				"cluster name", kc.Spec.Name,
				"cluster ID", kc.Status.ID,
			)

			return err
		}

		if !kc.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := kc.NewPatch()
			kc.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), kc, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka Connect cluster maintenance events",
					"cluster name", kc.Spec.Name,
					"cluster ID", kc.Status.ID,
				)

				return err
			}

			l.Info("Kafka Connect cluster maintenance events were updated",
				"cluster ID", kc.Status.ID,
				"events", kc.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaConnectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
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
