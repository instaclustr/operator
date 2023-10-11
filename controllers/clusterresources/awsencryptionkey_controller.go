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

package clusterresources

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

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// AWSEncryptionKeyReconciler reconciles a AWSEncryptionKey object
type AWSEncryptionKeyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsencryptionkeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsencryptionkeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=awsencryptionkeys/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AWSEncryptionKey object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *AWSEncryptionKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	encryptionKey := &v1beta1.AWSEncryptionKey{}
	err := r.Client.Get(ctx, req.NamespacedName, encryptionKey)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("AWS encryption key resource is not found",
				"resource name", req.NamespacedName,
			)
			return models.ExitReconcile, nil
		}

		l.Error(err, "Unable to fetch AWS encryption key")
		return models.ReconcileRequeue, err
	}

	switch encryptionKey.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreate(ctx, encryptionKey, &l), nil
	case models.DeletingEvent:
		return r.handleDelete(ctx, encryptionKey, &l), nil
	case models.GenericEvent:
		l.Info("AWS encryption key event isn't handled",
			"alias", encryptionKey.Spec.Alias,
			"arn", encryptionKey.Spec.ARN,
			"provider account name", encryptionKey.Spec.ProviderAccountName,
			"request", req,
			"event", encryptionKey.Annotations[models.ResourceStateAnnotation])
		return models.ExitReconcile, nil
	}

	return models.ExitReconcile, nil
}

func (r *AWSEncryptionKeyReconciler) handleCreate(
	ctx context.Context,
	encryptionKey *v1beta1.AWSEncryptionKey,
	l *logr.Logger,
) reconcile.Result {
	if encryptionKey.Status.ID == "" {
		l.Info(
			"Creating AWS encryption key",
			"alias", encryptionKey.Spec.Alias,
			"arn", encryptionKey.Spec.ARN,
		)

		patch := encryptionKey.NewPatch()

		encryptionKeyStatus, err := r.API.CreateEncryptionKey(&encryptionKey.Spec)
		if err != nil {
			l.Error(
				err, "Cannot create AWS encryption key",
				"spec", encryptionKey.Spec,
			)
			r.EventRecorder.Eventf(
				encryptionKey, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		r.EventRecorder.Eventf(
			encryptionKey, models.Normal, models.Created,
			"Resource creation request is sent",
		)

		encryptionKey.Status = *encryptionKeyStatus
		encryptionKey.Status.State = models.CreatedStatus
		err = r.Status().Patch(ctx, encryptionKey, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWS encryption key status ", "ID", encryptionKey.Status.ID)
			r.EventRecorder.Eventf(
				encryptionKey, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		encryptionKey.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(encryptionKey, models.DeletionFinalizer)
		err = r.Patch(ctx, encryptionKey, patch)
		if err != nil {
			l.Error(err, "Cannot patch AWS encryption key",
				"alias", encryptionKey.Spec.Alias,
				"arn", encryptionKey.Spec.ARN,
			)
			r.EventRecorder.Eventf(
				encryptionKey, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}

		l.Info(
			"AWS encryption key resource has been created",
			"alias", encryptionKey.Spec.Alias,
			"arn", encryptionKey.Spec.ARN,
		)
	}

	err := r.startEncryptionKeyStatusJob(encryptionKey)
	if err != nil {
		l.Error(err, "Cannot start AWS encryption key status checker job",
			"encryption key ID", encryptionKey.Status.ID)
		r.EventRecorder.Eventf(
			encryptionKey, models.Warning, models.CreationFailed,
			"Resource status job creation is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	r.EventRecorder.Eventf(
		encryptionKey, models.Normal, models.Created,
		"Resource status check job is started",
	)

	return models.ExitReconcile
}

func (r *AWSEncryptionKeyReconciler) handleDelete(
	ctx context.Context,
	encryptionKey *v1beta1.AWSEncryptionKey,
	l *logr.Logger,
) reconcile.Result {
	status, err := r.API.GetEncryptionKeyStatus(encryptionKey.Status.ID, instaclustr.AWSEncryptionKeyEndpoint)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get AWS encryption key status from the Instaclustr API",
			"alias", encryptionKey.Spec.Alias,
			"arn", encryptionKey.Spec.ARN,
		)

		r.EventRecorder.Eventf(
			encryptionKey, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	if status != nil {
		err = r.API.DeleteEncryptionKey(encryptionKey.Status.ID)
		if err != nil {
			l.Error(err, "Cannot delete AWS encryption key",
				"encryption key ID", encryptionKey.Status.ID,
				"alias", encryptionKey.Spec.Alias,
				"arn", encryptionKey.Spec.ARN,
			)

			r.EventRecorder.Eventf(
				encryptionKey, models.Warning, models.DeletionFailed,
				"Resource deletion on the Instaclustr is failed. Reason: %v",
				err,
			)
			return models.ReconcileRequeue
		}
		r.EventRecorder.Eventf(
			encryptionKey, models.Normal, models.DeletionStarted,
			"Resource deletion request is sent to the Instaclustr API.",
		)
	}

	r.Scheduler.RemoveJob(encryptionKey.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(encryptionKey, models.DeletionFinalizer)
	encryptionKey.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	patch := encryptionKey.NewPatch()
	err = r.Patch(ctx, encryptionKey, patch)
	if err != nil {
		l.Error(err, "Cannot patch AWS encryption key metadata",
			"alias", encryptionKey.Spec.Alias,
			"arn", encryptionKey.Spec.ARN,
			"status", encryptionKey.Status,
		)

		r.EventRecorder.Eventf(
			encryptionKey, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return models.ReconcileRequeue
	}

	l.Info("AWS encryption key has been deleted",
		"alias", encryptionKey.Spec.Alias,
		"arn", encryptionKey.Spec.ARN,
		"status", encryptionKey.Status,
	)

	r.EventRecorder.Eventf(
		encryptionKey, models.Normal, models.Deleted,
		"Resource is deleted",
	)

	return models.ExitReconcile
}

func (r *AWSEncryptionKeyReconciler) startEncryptionKeyStatusJob(encryptionKey *v1beta1.AWSEncryptionKey) error {
	job := r.newWatchStatusJob(encryptionKey)

	err := r.Scheduler.ScheduleJob(encryptionKey.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *AWSEncryptionKeyReconciler) newWatchStatusJob(encryptionKey *v1beta1.AWSEncryptionKey) scheduler.Job {
	l := log.Log.WithValues("component", "EncryptionKeyStatusJob")
	return func() error {
		ctx := context.Background()

		key := client.ObjectKeyFromObject(encryptionKey)
		err := r.Get(ctx, key, encryptionKey)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
					"namespaced name", key,
				)

				r.Scheduler.RemoveJob(encryptionKey.GetJobID(scheduler.StatusChecker))

				return nil
			}

			return err
		}

		instaEncryptionKeyStatus, err := r.API.GetEncryptionKeyStatus(encryptionKey.Status.ID, instaclustr.AWSEncryptionKeyEndpoint)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				return r.handleExternalDelete(ctx, encryptionKey)
			}

			l.Error(err, "Cannot get AWS encryption key status from Inst API", "encryption key ID", encryptionKey.Status.ID)
			return err
		}

		if !areEncryptionKeyStatusesEqual(instaEncryptionKeyStatus, &encryptionKey.Status) {
			l.Info("AWS encryption key status of k8s is different from Instaclustr. Reconcile statuses..",
				"encryption key status from Inst API", instaEncryptionKeyStatus,
				"encryption key status", encryptionKey.Status)
			patch := encryptionKey.NewPatch()
			encryptionKey.Status = *instaEncryptionKeyStatus
			err := r.Status().Patch(ctx, encryptionKey, patch)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *AWSEncryptionKeyReconciler) handleExternalDelete(ctx context.Context, key *v1beta1.AWSEncryptionKey) error {
	l := log.FromContext(ctx)

	patch := key.NewPatch()
	key.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, key, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(key, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(key.GetJobID(scheduler.StatusChecker))

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSEncryptionKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.AWSEncryptionKey{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if event.Object.GetDeletionTimestamp() != nil {
					event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.AWSEncryptionKey)
				if newObj.Generation == event.ObjectOld.GetGeneration() {
					return false
				}

				if newObj.DeletionTimestamp != nil {
					event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
					return true
				}

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				genericEvent.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.GenericEvent
				return true
			},
		})).
		Complete(r)
}
