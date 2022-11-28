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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

// ClusterBackupReconciler reconciles a ClusterBackup object
type ClusterBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	API    instaclustr.API
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=clusterbackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ClusterBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	backup := &clusterresourcesv1alpha1.ClusterBackup{}
	err := r.Get(ctx, req.NamespacedName, backup)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("cluster backup resource is not found",
				"Resource name", req.NamespacedName,
			)

			return models.ReconcileResult, nil
		}

		logger.Error(err, "cannot get cluster backup",
			"Backup name", req.NamespacedName,
		)

		return models.ReconcileRequeue, nil
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
			logger.Error(err, "cannot patch cluster backup resource labels",
				"Backup name", backup.Name,
			)

			return models.ReconcileRequeue, nil
		}
	}

	instBackup, err := r.API.GetClusterBackups(instaclustr.ClustersEndpointV1, backup.Spec.ClusterID)
	if err != nil {
		logger.Error(err, "cannot get cluster backups from Instaclustr",
			"Backup name", backup.Name,
			"Cluster ID", backup.Spec.ClusterID,
		)

		return models.ReconcileRequeue, nil
	}

	instBackupEvents := instBackup.GetBackupEvents(models.PgBackupEventType)

	backupsList, err := r.listClusterBackups(ctx, backup.Spec.ClusterID, backup.Namespace)
	if err != nil {
		logger.Error(err, "cannot get cluster backups",
			"Backup name", backup.Name,
			"Cluster ID", backup.Spec.ClusterID,
		)

		return models.ReconcileRequeue, nil
	}

	if len(instBackupEvents) < len(backupsList.Items) {
		err = r.API.TriggerClusterBackup(instaclustr.ClustersEndpointV1, backup.Spec.ClusterID)
		if err != nil {
			logger.Error(err, "cannot trigger cluster backup",
				"Backup name", backup.Name,
				"Cluster ID", backup.Spec.ClusterID,
			)

			return models.ReconcileRequeue, nil
		}

		logger.Info("New cluster backup request was sent",
			"Cluster ID", backup.Spec.ClusterID,
		)
	}

	if backup.Annotations[models.StartAnnotation] != "" &&
		backup.Status.Start == 0 {
		backup.Status.Start, err = strconv.Atoi(backup.Annotations[models.StartAnnotation])
		if err != nil {
			logger.Error(err, "cannot convert backup start timestamp to string",
				"Backup name", backup.Name,
				"Annotations", backup.Annotations,
			)

			return models.ReconcileRequeue, nil
		}

		err = r.Status().Patch(ctx, backup, patch)
		if err != nil {
			logger.Error(err, "cannot patch cluster backup resource status",
				"Backup name", backup.Name,
			)

			return models.ReconcileRequeue, nil
		}
	}

	r.reconcileFinalizers(backup)
	err = r.Patch(ctx, backup, patch)
	if err != nil {
		logger.Error(err, "cannot patch cluster backup resource",
			"Backup name", backup.Name,
		)

		return models.ReconcileRequeue, nil
	}

	logger.Info("Cluster backup resource was reconciled",
		"Backup name", backup.Name,
		"Cluster ID", backup.Spec.ClusterID,
	)

	return models.ReconcileResult, nil
}

func (r *ClusterBackupReconciler) reconcileFinalizers(backup *clusterresourcesv1alpha1.ClusterBackup) {
	for _, finalizer := range backup.Finalizers {
		if finalizer == models.DeletionFinalizer {
			return
		}
	}

	controllerutil.AddFinalizer(backup, models.DeletionFinalizer)
}

func (r *ClusterBackupReconciler) listClusterBackups(ctx context.Context, clusterID, namespace string) (*clusterresourcesv1alpha1.ClusterBackupList, error) {
	backupsList := &clusterresourcesv1alpha1.ClusterBackupList{}
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
		For(&clusterresourcesv1alpha1.ClusterBackup{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				return false
			},
		})).
		Complete(r)
}