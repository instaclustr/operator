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
	"strconv"

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

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

// ClusterBackupReconciler reconciles a ClusterBackup object
type ClusterBackupReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ClusterBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	backup := &v1beta1.ClusterBackup{}
	err := r.Get(ctx, req.NamespacedName, backup)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Cluster backup resource is not found",
				"resource name", req.NamespacedName,
			)

			return ctrl.Result{}, nil
		}

		logger.Error(err, "Cannot get cluster backup",
			"backup name", req.NamespacedName,
		)

		return ctrl.Result{}, err
	}

	patch := backup.NewPatch()

	if backup.Labels[models.ClusterIDLabel] != backup.Spec.ClusterID {
		if backup.Labels == nil {
			backup.Labels = map[string]string{models.ClusterIDLabel: backup.Spec.ClusterID}
		} else {
			backup.Labels[models.ClusterIDLabel] = backup.Spec.ClusterID
		}
		err = r.Patch(ctx, backup, patch)
		if err != nil {
			logger.Error(err, "Cannot patch cluster backup resource labels",
				"backup name", backup.Name,
			)

			r.EventRecorder.Eventf(
				backup, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}
	}

	backupsList, err := r.listClusterBackups(ctx, backup.Spec.ClusterID, backup.Namespace)
	if err != nil {
		logger.Error(err, "Cannot get cluster backups",
			"backup name", backup.Name,
			"cluster ID", backup.Spec.ClusterID,
		)

		r.EventRecorder.Eventf(
			backup, models.Warning, models.FetchFailed,
			"Fetch resource from the k8s cluster is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	clusterKind := models.ClusterKindsMap[backup.Spec.ClusterKind]
	if backup.Spec.ClusterKind == models.PgClusterKind {
		clusterKind = models.PgAppKind
	}

	iBackup, err := r.API.GetClusterBackups(backup.Spec.ClusterID, clusterKind)
	if err != nil {
		logger.Error(err, "Cannot get cluster backups from Instaclustr",
			"backup name", backup.Name,
			"cluster ID", backup.Spec.ClusterID,
		)

		r.EventRecorder.Eventf(
			backup, models.Warning, models.FetchFailed,
			"Fetch resource from the Instaclustr API is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	iBackupEvents := iBackup.GetBackupEvents(backup.Spec.ClusterKind)

	if len(iBackupEvents) < len(backupsList.Items) {
		err = r.API.TriggerClusterBackup(backup.Spec.ClusterID, models.ClusterKindsMap[backup.Spec.ClusterKind])
		if err != nil {
			logger.Error(err, "Cannot trigger cluster backup",
				"backup name", backup.Name,
				"cluster ID", backup.Spec.ClusterID,
			)

			r.EventRecorder.Eventf(
				backup, models.Warning, models.CreationFailed,
				"Resource creation on the Instaclustr is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			backup, models.Normal, models.Created,
			"Resource creation request is sent",
		)
		logger.Info("New cluster backup request was sent",
			"cluster ID", backup.Spec.ClusterID,
		)
	}

	if backup.Annotations[models.StartTimestampAnnotation] != "" &&
		backup.Status.Start == 0 {
		backup.Status.Start, err = strconv.Atoi(backup.Annotations[models.StartTimestampAnnotation])
		if err != nil {
			logger.Error(err, "Cannot convert backup start timestamp to int",
				"backup name", backup.Name,
				"annotations", backup.Annotations,
			)

			r.EventRecorder.Eventf(
				backup, models.Warning, models.ConversionFailed,
				"Start timestamp annotation convertion to int is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		err = r.Status().Patch(ctx, backup, patch)
		if err != nil {
			logger.Error(err, "Cannot patch cluster backup resource status",
				"backup name", backup.Name,
			)

			r.EventRecorder.Eventf(
				backup, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)
			return ctrl.Result{}, err
		}

		r.EventRecorder.Eventf(
			backup, models.Normal, models.Created,
			"Start timestamp is set",
		)
	}

	controllerutil.AddFinalizer(backup, models.DeletionFinalizer)
	err = r.Patch(ctx, backup, patch)
	if err != nil {
		logger.Error(err, "Cannot patch cluster backup resource",
			"backup name", backup.Name,
		)

		r.EventRecorder.Eventf(
			backup, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v",
			err,
		)
		return ctrl.Result{}, err
	}

	logger.Info("Cluster backup resource was reconciled",
		"backup name", backup.Name,
		"cluster ID", backup.Spec.ClusterID,
	)

	return ctrl.Result{}, nil
}

func (r *ClusterBackupReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*v1beta1.ClusterBackupList, error) {
	backupsList := &v1beta1.ClusterBackupList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{models.ClusterIDLabel: clusterID},
	}
	err := r.Client.List(ctx, backupsList, listOpts...)
	if err != nil {
		return nil, err
	}

	return backupsList, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay)}).
		For(&v1beta1.ClusterBackup{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return false
			},
		})).
		Complete(r)
}
