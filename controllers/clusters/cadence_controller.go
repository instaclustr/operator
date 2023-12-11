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
	"fmt"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// CadenceReconciler reconciles a Cadence object
type CadenceReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CadenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	cadenceCluster := &v1beta1.Cadence{}
	err := r.Client.Get(ctx, req.NamespacedName, cadenceCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Cadence resource is not found",
				"resource name", req.NamespacedName,
			)
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch Cadence resource",
			"resource name", req.NamespacedName,
		)
		return ctrl.Result{}, err
	}

	switch cadenceCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, cadenceCluster, logger)
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, cadenceCluster, logger)
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, cadenceCluster, logger)
	case models.GenericEvent:
		logger.Info("Generic event isn't handled",
			"request", req,
			"event", cadenceCluster.Annotations[models.ResourceStateAnnotation],
		)

		return ctrl.Result{}, nil
	default:
		logger.Info("Unknown event isn't handled",
			"request", req,
			"event", cadenceCluster.Annotations[models.ResourceStateAnnotation],
		)

		return ctrl.Result{}, nil
	}
}

func (r *CadenceReconciler) HandleCreateCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) (ctrl.Result, error) {
	if cadence.Status.ID == "" {
		patch := cadence.NewPatch()

		for _, packagedProvisioning := range cadence.Spec.PackagedProvisioning {
			requeueNeeded, err := r.preparePackagedSolution(ctx, cadence, packagedProvisioning)
			if err != nil {
				logger.Error(err, "Cannot prepare packaged solution for Cadence cluster",
					"cluster name", cadence.Spec.Name,
				)

				r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
					"Cannot prepare packaged solution for Cadence cluster. Reason: %v", err)

				return ctrl.Result{}, err
			}

			if requeueNeeded {
				logger.Info("Waiting for bundled clusters to be created",
					"cadence cluster name", cadence.Spec.Name)

				r.EventRecorder.Event(cadence, models.Normal, "Waiting",
					"Waiting for bundled clusters to be created")

				return models.ReconcileRequeue, nil
			}
		}

		logger.Info(
			"Creating Cadence cluster",
			"cluster name", cadence.Spec.Name,
			"data centres", cadence.Spec.DataCentres,
		)

		cadenceAPISpec, err := cadence.Spec.ToInstAPI(ctx, r.Client)
		if err != nil {
			logger.Error(err, "Cannot convert Cadence cluster manifest to API spec",
				"cluster manifest", cadence.Spec)

			r.EventRecorder.Eventf(cadence, models.Warning, models.ConvertionFailed,
				"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		id, err := r.API.CreateCluster(instaclustr.CadenceEndpoint, cadenceAPISpec)
		if err != nil {
			logger.Error(
				err, "Cannot create Cadence cluster",
				"cadence manifest", cadence.Spec,
			)
			r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		cadence.Status.ID = id
		err = r.Status().Patch(ctx, cadence, patch)
		if err != nil {
			logger.Error(err, "Cannot update Cadence cluster status",
				"cluster name", cadence.Spec.Name,
				"cluster status", cadence.Status,
			)

			r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		if cadence.Spec.Description != "" {
			err = r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, id, cadence.Spec.Description, nil)
			if err != nil {
				logger.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
					"cluster name", cadence.Spec.Name,
					"description", cadence.Spec.Description,
					"twoFactorDelete", cadence.Spec.TwoFactorDelete,
				)

				r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
					"Cluster description and TwoFactoDelete update is failed. Reason: %v", err)
			}
		}

		cadence.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(cadence, models.DeletionFinalizer)

		err = r.Patch(ctx, cadence, patch)
		if err != nil {
			logger.Error(err, "Cannot patch Cadence cluster",
				"cluster name", cadence.Spec.Name, "patch", patch)

			r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		logger.Info(
			"Cadence resource has been created",
			"cluster name", cadence.Name,
			"cluster ID", cadence.Status.ID,
			"kind", cadence.Kind,
			"api version", cadence.APIVersion,
			"namespace", cadence.Namespace,
		)

		r.EventRecorder.Eventf(cadence, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s", id)
	}

	if cadence.Status.State != models.DeletedStatus {
		err := r.startClusterStatusJob(cadence)
		if err != nil {
			logger.Error(err, "Cannot start cluster status job",
				"cadence cluster ID", cadence.Status.ID,
			)

			r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
				"Cluster status check job is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		r.EventRecorder.Event(cadence, models.Normal, models.Created,
			"Cluster status check job is started")
	}

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) HandleUpdateCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) (ctrl.Result, error) {
	iData, err := r.API.GetCadence(cadence.Status.ID)
	if err != nil {
		logger.Error(
			err, "Cannot get Cadence cluster from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	iCadence, err := cadence.FromInstAPI(iData)
	if err != nil {
		logger.Error(
			err, "Cannot convert Cadence cluster from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.ConvertionFailed,
			"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	if iCadence.Status.CurrentClusterOperationStatus != models.NoOperation {
		logger.Info("Cadence cluster is not ready to update",
			"cluster name", iCadence.Spec.Name,
			"cluster state", iCadence.Status.State,
			"current operation status", iCadence.Status.CurrentClusterOperationStatus,
		)

		patch := cadence.NewPatch()
		cadence.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
		cadence.Annotations[models.UpdateQueuedAnnotation] = models.True
		err = r.Patch(ctx, cadence, patch)
		if err != nil {
			logger.Error(err, "Cannot patch Cadence cluster",
				"cluster name", cadence.Spec.Name,
				"patch", patch)

			r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		return models.ReconcileRequeue, nil
	}

	if cadence.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(cadence, iCadence, logger)
	}

	if cadence.Spec.ClusterSettingsNeedUpdate(iCadence.Spec.Cluster) {
		logger.Info("Updating cluster settings",
			"instaclustr description", iCadence.Spec.Description,
			"instaclustr two factor delete", iCadence.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(cadence.Status.ID, cadence.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			logger.Error(err, "Cannot update cluster settings",
				"cluster ID", cadence.Status.ID, "cluster spec", cadence.Spec)
			r.EventRecorder.Eventf(cadence, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	logger.Info("Update request to Instaclustr API has been sent",
		"spec data centres", cadence.Spec.DataCentres,
		"resize settings", cadence.Spec.ResizeSettings,
	)
	err = r.API.UpdateCluster(cadence.Status.ID, instaclustr.CadenceEndpoint, cadence.Spec.NewDCsUpdate())
	if err != nil {
		logger.Error(err, "Cannot update Cadence cluster",
			"cluster ID", cadence.Status.ID,
			"update request", cadence.Spec.NewDCsUpdate(),
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	patch := cadence.NewPatch()
	cadence.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	cadence.Annotations[models.UpdateQueuedAnnotation] = ""
	err = r.Patch(ctx, cadence, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Cadence cluster",
			"cluster name", cadence.Spec.Name,
			"patch", patch)

		r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	logger.Info(
		"Cluster has been updated",
		"cluster name", cadence.Spec.Name,
		"cluster ID", cadence.Status.ID,
		"data centres", cadence.Spec.DataCentres,
	)

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) handleClusterResourcesEvents(
	newObj *v1beta1.Cadence,
	oldObjSpec *v1beta1.CadenceSpec,
) {
	err := HandleResourceEvent(r.Client, models.ClusterbackupRef, oldObjSpec.ClusterResources.ClusterBackups, newObj.Spec.ClusterResources.ClusterBackups, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.ClusterNetworkFirewallRuleRef, oldObjSpec.ClusterResources.ClusterNetworkFirewallRules, newObj.Spec.ClusterResources.ClusterNetworkFirewallRules, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.AWSVPCPeeringRef, oldObjSpec.ClusterResources.AWSVPCPeerings, newObj.Spec.ClusterResources.AWSVPCPeerings, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.AWSSecurityGroupFirewallRuleRef, oldObjSpec.ClusterResources.AWSSecurityGroupFirewallRules, newObj.Spec.ClusterResources.AWSSecurityGroupFirewallRules, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.ExclusionWindowRef, oldObjSpec.ClusterResources.ExclusionWindows, newObj.Spec.ClusterResources.ExclusionWindows, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.GCPVPCPeeringRef, oldObjSpec.ClusterResources.GCPVPCPeerings, newObj.Spec.ClusterResources.GCPVPCPeerings, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
	err = HandleResourceEvent(r.Client, models.AzureVNetPeeringRef, oldObjSpec.ClusterResources.AzureVNetPeerings, newObj.Spec.ClusterResources.AzureVNetPeerings, newObj.Status.ID, newObj.Status.DataCentres)
	if err != nil {
		r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
			CannotHandleUserEvent, err)
	}
}

func (r *CadenceReconciler) DetachClusterresourcesFromCluster(ctx context.Context, l logr.Logger, cadence *v1beta1.Cadence) {
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.ClusterNetworkFirewallRules, models.ClusterNetworkFirewallRuleRef)
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.AWSVPCPeerings, models.AWSVPCPeeringRef)
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.AWSSecurityGroupFirewallRules, models.AWSSecurityGroupFirewallRuleRef)
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.ExclusionWindows, models.ExclusionWindowRef)
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.GCPVPCPeerings, models.GCPVPCPeeringRef)
	r.DetachClusterresources(ctx, l, cadence, cadence.Spec.ClusterResources.AzureVNetPeerings, models.AzureVNetPeeringRef)
}

func (r *CadenceReconciler) DetachClusterresources(ctx context.Context, l logr.Logger, cadence *v1beta1.Cadence, refs []*v1beta1.ClusterResourceRef, kind string) {
	for _, ref := range refs {
		err := HandleDeleteResource(r.Client, ctx, l, kind, ref)
		if err != nil {
			l.Error(err, "Cannot detach clusterresource", "resource kind", kind, "namespace and name", ref)
			r.EventRecorder.Eventf(cadence, models.Warning, models.DeletingEvent,
				"Cannot detach resource. Reason: %v", err)
		}
	}
}

func (r *CadenceReconciler) handleExternalChanges(cadence, iCadence *v1beta1.Cadence, l logr.Logger) (ctrl.Result, error) {
	if !cadence.Spec.AreDCsEqual(iCadence.Spec.DataCentres) {
		l.Info(msgExternalChanges,
			"instaclustr data", iCadence.Spec.DataCentres,
			"k8s resource spec", cadence.Spec.DataCentres)

		msgDiffSpecs, err := createSpecDifferenceMessage(cadence.Spec.DataCentres, iCadence.Spec.DataCentres)
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iCadence.Spec, "k8s resource spec", cadence.Spec)
			return ctrl.Result{}, err
		}
		r.EventRecorder.Eventf(cadence, models.Warning, models.ExternalChanges, msgDiffSpecs)

		return ctrl.Result{}, nil
	}

	patch := cadence.NewPatch()

	cadence.Annotations[models.ExternalChangesAnnotation] = ""

	err := r.Patch(context.Background(), cadence, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", cadence.Spec.Name, "cluster ID", cadence.Status.ID)

		r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	l.Info("External changes have been reconciled", "resource ID", cadence.Status.ID)
	r.EventRecorder.Event(cadence, models.Normal, models.ExternalChanges, "External changes have been reconciled")

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) HandleDeleteCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) (ctrl.Result, error) {
	_, err := r.API.GetCadence(cadence.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(
			err, "Cannot get Cadence cluster status from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	r.DetachClusterresourcesFromCluster(ctx, logger, cadence)

	if !errors.Is(err, instaclustr.NotFound) {
		logger.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID)

		err = r.API.DeleteCluster(cadence.Status.ID, instaclustr.CadenceEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete Cadence cluster",
				"cluster name", cadence.Spec.Name,
				"cluster status", cadence.Status,
			)

			r.EventRecorder.Eventf(cadence, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v", err)

			return ctrl.Result{}, err
		}

		r.EventRecorder.Event(cadence, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if cadence.Spec.TwoFactorDelete != nil {
			patch := cadence.NewPatch()

			cadence.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			cadence.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, cadence, patch)
			if err != nil {
				logger.Error(err, "Cannot patch cluster resource",
					"cluster name", cadence.Spec.Name,
					"cluster state", cadence.Status.State)
				r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return ctrl.Result{}, err
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", cadence.Status.ID)

			r.EventRecorder.Event(cadence, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return ctrl.Result{}, nil
		}
	}

	logger.Info("Cadence cluster is being deleted",
		"cluster name", cadence.Spec.Name,
		"cluster status", cadence.Status)

	for _, packagedProvisioning := range cadence.Spec.PackagedProvisioning {
		err = r.deletePackagedResources(ctx, cadence, packagedProvisioning)
		if err != nil {
			logger.Error(
				err, "Cannot delete Cadence packaged resources",
				"cluster name", cadence.Spec.Name,
				"cluster ID", cadence.Status.ID,
			)

			r.EventRecorder.Eventf(cadence, models.Warning, models.DeletionFailed,
				"Cannot delete Cadence packaged resources. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	r.Scheduler.RemoveJob(cadence.GetJobID(scheduler.StatusChecker))
	patch := cadence.NewPatch()
	controllerutil.RemoveFinalizer(cadence, models.DeletionFinalizer)
	cadence.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent

	err = r.Patch(ctx, cadence, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Cadence cluster",
			"cluster name", cadence.Spec.Name,
			"patch", patch,
		)
		return ctrl.Result{}, err
	}

	err = exposeservice.Delete(r.Client, cadence.Name, cadence.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete Cadence cluster expose service",
			"cluster ID", cadence.Status.ID,
			"cluster name", cadence.Spec.Name,
		)

		return ctrl.Result{}, err
	}

	logger.Info("Cadence cluster was deleted",
		"cluster name", cadence.Spec.Name,
		"cluster ID", cadence.Status.ID,
	)

	r.EventRecorder.Event(cadence, models.Normal, models.Deleted, "Cluster resource is deleted")

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) preparePackagedSolution(
	ctx context.Context,
	cluster *v1beta1.Cadence,
	packagedProvisioning *v1beta1.PackagedProvisioning,
) (bool, error) {
	if len(cluster.Spec.DataCentres) < 1 {
		return false, models.ErrZeroDataCentres
	}

	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, cluster.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return false, err
	}

	cassandraList := &v1beta1.CassandraList{}
	err = r.Client.List(ctx, cassandraList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return false, err
	}

	if len(cassandraList.Items) == 0 {
		appVersions, err := r.API.ListAppVersions(models.CassandraAppKind)
		if err != nil {
			return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
				models.CassandraAppKind, err)
		}

		cassandraVersions := getSortedAppVersions(appVersions, models.CassandraAppType)
		if len(cassandraVersions) == 0 {
			return false, fmt.Errorf("there are no versions for %v kind",
				models.CassandraAppKind)
		}

		cassandraSpec, err := r.newCassandraSpec(cluster, cassandraVersions[len(cassandraVersions)-1].String())
		if err != nil {
			return false, err
		}

		err = r.Client.Create(ctx, cassandraSpec)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	kafkaList := &v1beta1.KafkaList{}
	osList := &v1beta1.OpenSearchList{}
	advancedVisibility := &v1beta1.AdvancedVisibility{
		TargetKafka:      &v1beta1.TargetKafka{},
		TargetOpenSearch: &v1beta1.TargetOpenSearch{},
	}
	var advancedVisibilities []*v1beta1.AdvancedVisibility
	if packagedProvisioning.UseAdvancedVisibility {
		err = r.Client.List(ctx, kafkaList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, err
		}
		if len(kafkaList.Items) == 0 {
			appVersions, err := r.API.ListAppVersions(models.KafkaAppKind)
			if err != nil {
				return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
					models.KafkaAppKind, err)
			}

			kafkaVersions := getSortedAppVersions(appVersions, models.KafkaAppType)
			if len(kafkaVersions) == 0 {
				return false, fmt.Errorf("there are no versions for %v kind",
					models.KafkaAppType)
			}

			kafkaSpec, err := r.newKafkaSpec(cluster, kafkaVersions[len(kafkaVersions)-1].String())
			if err != nil {
				return false, err
			}

			err = r.Client.Create(ctx, kafkaSpec)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		if len(kafkaList.Items[0].Status.DataCentres) == 0 {
			return true, nil
		}

		advancedVisibility.TargetKafka.DependencyCDCID = kafkaList.Items[0].Status.DataCentres[0].ID
		advancedVisibility.TargetKafka.DependencyVPCType = models.VPCPeered

		err = r.Client.List(ctx, osList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, err
		}
		if len(osList.Items) == 0 {
			appVersions, err := r.API.ListAppVersions(models.OpenSearchAppKind)
			if err != nil {
				return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
					models.OpenSearchAppKind, err)
			}

			openSearchVersions := getSortedAppVersions(appVersions, models.OpenSearchAppType)
			if len(openSearchVersions) == 0 {
				return false, fmt.Errorf("there are no versions for %v kind",
					models.OpenSearchAppType)
			}

			// For OpenSearch we cannot use the latest version because is not supported by Cadence. So we use the oldest one.
			osSpec, err := r.newOpenSearchSpec(cluster, openSearchVersions[0].String())
			if err != nil {
				return false, err
			}

			err = r.Client.Create(ctx, osSpec)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		if len(osList.Items[0].Status.DataCentres) == 0 {
			return true, nil
		}

		advancedVisibility.TargetOpenSearch.DependencyCDCID = osList.Items[0].Status.DataCentres[0].ID
		advancedVisibility.TargetOpenSearch.DependencyVPCType = models.VPCPeered
		advancedVisibilities = append(advancedVisibilities, advancedVisibility)
	}

	if len(cassandraList.Items[0].Status.DataCentres) == 0 {
		return true, nil
	}

	cluster.Spec.StandardProvisioning = append(cluster.Spec.StandardProvisioning, &v1beta1.StandardProvisioning{
		AdvancedVisibility: advancedVisibilities,
		TargetCassandra: &v1beta1.TargetCassandra{
			DependencyCDCID:   cassandraList.Items[0].Status.DataCentres[0].ID,
			DependencyVPCType: models.VPCPeered,
		},
	})

	return false, nil
}

func (r *CadenceReconciler) newCassandraSpec(cadence *v1beta1.Cadence, latestCassandraVersion string) (*v1beta1.Cassandra, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.CassandraKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.CassandraChildPrefix + cadence.Name,
		Labels:      map[string]string{models.ControlledByLabel: cadence.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   cadence.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(cadence.Spec.DataCentres) == 0 {
		return nil, models.ErrZeroDataCentres
	}

	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance

	var twoFactorDelete []*v1beta1.TwoFactorDelete
	if len(cadence.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*v1beta1.TwoFactorDelete{
			{
				Email: cadence.Spec.TwoFactorDelete[0].Email,
				Phone: cadence.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}
	var cassNodeSize, network string
	var cassNodesNumber, cassReplicationFactor int
	var cassPrivateIPBroadcastForDiscovery, cassPasswordAndUserAuth bool
	for _, dc := range cadence.Spec.DataCentres {
		for _, pp := range cadence.Spec.PackagedProvisioning {
			cassNodeSize = pp.BundledCassandraSpec.NodeSize
			network = pp.BundledCassandraSpec.Network
			cassNodesNumber = pp.BundledCassandraSpec.NodesNumber
			cassReplicationFactor = pp.BundledCassandraSpec.ReplicationFactor
			cassPrivateIPBroadcastForDiscovery = pp.BundledCassandraSpec.PrivateIPBroadcastForDiscovery
			cassPasswordAndUserAuth = pp.BundledCassandraSpec.PasswordAndUserAuth

			isCassNetworkOverlaps, err := dc.IsNetworkOverlaps(network)
			if err != nil {
				return nil, err
			}
			if isCassNetworkOverlaps {
				return nil, models.ErrNetworkOverlaps
			}
		}
	}

	dcName := models.CassandraChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName

	cassandraDataCentres := []*v1beta1.CassandraDataCentre{
		{
			DataCentre: v1beta1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            cassNodeSize,
				NodesNumber:         cassNodesNumber,
				Network:             network,
			},
			ReplicationFactor:              cassReplicationFactor,
			PrivateIPBroadcastForDiscovery: cassPrivateIPBroadcastForDiscovery,
		},
	}
	spec := v1beta1.CassandraSpec{
		Cluster: v1beta1.Cluster{
			Name:                  models.CassandraChildPrefix + cadence.Name,
			Version:               latestCassandraVersion,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres:         cassandraDataCentres,
		PasswordAndUserAuth: cassPasswordAndUserAuth,
		BundledUseOnly:      true,
	}

	return &v1beta1.Cassandra{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) startClusterStatusJob(cadence *v1beta1.Cadence) error {
	job := r.newWatchStatusJob(cadence)

	err := r.Scheduler.ScheduleJob(cadence.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CadenceReconciler) newWatchStatusJob(cadence *v1beta1.Cadence) scheduler.Job {
	l := log.Log.WithValues("component", "cadenceStatusClusterJob")
	return func() error {
		namespacedName := client.ObjectKeyFromObject(cadence)
		err := r.Get(context.Background(), namespacedName, cadence)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(cadence.GetJobID(scheduler.StatusChecker))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get Cadence custom resource",
				"resource name", cadence.Name,
			)
			return err
		}

		iData, err := r.API.GetCadence(cadence.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if cadence.DeletionTimestamp != nil {
					_, err = r.HandleDeleteCluster(context.Background(), cadence, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), cadence)
			}

			l.Error(err, "Cannot get Cadence cluster from the Instaclustr API",
				"clusterID", cadence.Status.ID,
			)
			return err
		}

		iCadence, err := cadence.FromInstAPI(iData)
		if err != nil {
			l.Error(err, "Cannot convert Cadence cluster from the Instaclustr API",
				"clusterID", cadence.Status.ID,
			)
			return err
		}

		if !areStatusesEqual(&iCadence.Status.ClusterStatus, &cadence.Status.ClusterStatus) ||
			!areSecondaryCadenceTargetsEqual(cadence.Status.TargetSecondaryCadence, iCadence.Status.TargetSecondaryCadence) {
			l.Info("Updating Cadence cluster status",
				"new status", iCadence.Status.ClusterStatus,
				"old status", cadence.Status.ClusterStatus,
			)

			areDCsEqual := areDataCentresEqual(iCadence.Status.ClusterStatus.DataCentres, cadence.Status.ClusterStatus.DataCentres)

			patch := cadence.NewPatch()
			cadence.Status.ClusterStatus = iCadence.Status.ClusterStatus
			cadence.Status.TargetSecondaryCadence = iCadence.Status.TargetSecondaryCadence
			err = r.Status().Patch(context.Background(), cadence, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cadence cluster",
					"cluster name", cadence.Spec.Name,
					"status", cadence.Status.State,
				)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iCadence.Status.ClusterStatus.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					cadence.Name,
					cadence.Namespace,
					cadence.Spec.PrivateNetworkCluster,
					nodes,
					models.CadenceConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iCadence.Status.CurrentClusterOperationStatus == models.NoOperation &&
			cadence.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			cadence.Annotations[models.UpdateQueuedAnnotation] != models.True &&
			!cadence.Spec.AreDCsEqual(iCadence.Spec.DataCentres) {
			l.Info(msgExternalChanges,
				"instaclustr data", iCadence.Spec.DataCentres,
				"k8s resource spec", cadence.Spec.DataCentres)

			patch := cadence.NewPatch()
			cadence.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), cadence, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", cadence.Spec.Name, "cluster state", cadence.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(cadence.Spec.DataCentres, iCadence.Spec.DataCentres)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iCadence.Spec, "k8s resource spec", cadence.Spec)
				return err
			}
			r.EventRecorder.Eventf(cadence, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), cadence)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", cadence.Spec.Name,
				"cluster ID", cadence.Status.ID,
			)
			return err
		}

		if cadence.Status.State == models.RunningStatus && cadence.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := cadence.NewPatch()
			for _, dc := range cadence.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", cadence.Spec.Name,
						"cluster ID", cadence.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), cadence, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", cadence.Spec.Name,
						"cluster ID", cadence.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *CadenceReconciler) newKafkaSpec(cadence *v1beta1.Cadence, latestKafkaVersion string) (*v1beta1.Kafka, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.KafkaKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.KafkaChildPrefix + cadence.Name,
		Labels:      map[string]string{models.ControlledByLabel: cadence.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   cadence.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(cadence.Spec.DataCentres) == 0 {
		return nil, models.ErrZeroDataCentres
	}

	var kafkaTFD []*v1beta1.TwoFactorDelete
	for _, cadenceTFD := range cadence.Spec.TwoFactorDelete {
		twoFactorDelete := &v1beta1.TwoFactorDelete{
			Email: cadenceTFD.Email,
			Phone: cadenceTFD.Phone,
		}
		kafkaTFD = append(kafkaTFD, twoFactorDelete)
	}
	bundledKafkaSpec := cadence.Spec.PackagedProvisioning[0].BundledKafkaSpec

	kafkaNetwork := bundledKafkaSpec.Network
	for _, cadenceDC := range cadence.Spec.DataCentres {
		isKafkaNetworkOverlaps, err := cadenceDC.IsNetworkOverlaps(kafkaNetwork)
		if err != nil {
			return nil, err
		}
		if isKafkaNetworkOverlaps {
			return nil, models.ErrNetworkOverlaps
		}
	}

	kafkaNodeSize := bundledKafkaSpec.NodeSize
	kafkaNodesNumber := bundledKafkaSpec.NodesNumber
	dcName := models.KafkaChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName
	kafkaDataCentres := []*v1beta1.KafkaDataCentre{
		{
			DataCentre: v1beta1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            kafkaNodeSize,
				NodesNumber:         kafkaNodesNumber,
				Network:             kafkaNetwork,
			},
		},
	}

	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance
	clientEncryption := cadence.Spec.DataCentres[0].ClientEncryption
	spec := v1beta1.KafkaSpec{
		Cluster: v1beta1.Cluster{
			Name:                  models.KafkaChildPrefix + cadence.Name,
			Version:               latestKafkaVersion,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       kafkaTFD,
			PCICompliance:         pciCompliance,
		},
		DataCentres:               kafkaDataCentres,
		ReplicationFactor:         bundledKafkaSpec.ReplicationFactor,
		PartitionsNumber:          bundledKafkaSpec.PartitionsNumber,
		AllowDeleteTopics:         true,
		AutoCreateTopics:          true,
		ClientToClusterEncryption: clientEncryption,
		BundledUseOnly:            true,
	}

	return &v1beta1.Kafka{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) newOpenSearchSpec(cadence *v1beta1.Cadence, oldestOpenSearchVersion string) (*v1beta1.OpenSearch, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.OpenSearchKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.OpenSearchChildPrefix + cadence.Name,
		Labels:      map[string]string{models.ControlledByLabel: cadence.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   cadence.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(cadence.Spec.DataCentres) < 1 {
		return nil, models.ErrZeroDataCentres
	}

	bundledOpenSearchSpec := cadence.Spec.PackagedProvisioning[0].BundledOpenSearchSpec

	managerNodes := []*v1beta1.ClusterManagerNodes{{
		NodeSize:         bundledOpenSearchSpec.NodeSize,
		DedicatedManager: false,
	}}

	osReplicationFactor := bundledOpenSearchSpec.ReplicationFactor
	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance

	var twoFactorDelete []*v1beta1.TwoFactorDelete
	if len(cadence.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*v1beta1.TwoFactorDelete{
			{
				Email: cadence.Spec.TwoFactorDelete[0].Email,
				Phone: cadence.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}

	osNetwork := bundledOpenSearchSpec.Network
	isOsNetworkOverlaps, err := cadence.Spec.DataCentres[0].IsNetworkOverlaps(osNetwork)
	if err != nil {
		return nil, err
	}
	if isOsNetworkOverlaps {
		return nil, models.ErrNetworkOverlaps
	}

	dcName := models.OpenSearchChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName

	osDataCentres := []*v1beta1.OpenSearchDataCentre{
		{
			Name:                dcName,
			Region:              dcRegion,
			CloudProvider:       cloudProvider,
			ProviderAccountName: providerAccountName,
			Network:             osNetwork,
			ReplicationFactor:   osReplicationFactor,
		},
	}
	spec := v1beta1.OpenSearchSpec{
		Cluster: v1beta1.Cluster{
			Name:                  models.OpenSearchChildPrefix + cadence.Name,
			Version:               oldestOpenSearchVersion,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres:         osDataCentres,
		ClusterManagerNodes: managerNodes,
		BundledUseOnly:      true,
	}

	return &v1beta1.OpenSearch{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) deletePackagedResources(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	packagedProvisioning *v1beta1.PackagedProvisioning,
) error {
	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, cadence.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return err
	}

	cassandraList := &v1beta1.CassandraList{}
	err = r.Client.List(ctx, cassandraList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}

	if len(cassandraList.Items) != 0 {
		for _, cassandraCluster := range cassandraList.Items {
			err = r.Client.Delete(ctx, &cassandraCluster)
			if err != nil {
				return err
			}
		}
	}

	if packagedProvisioning.UseAdvancedVisibility {
		kafkaList := &v1beta1.KafkaList{}
		err = r.Client.List(ctx, kafkaList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(kafkaList.Items) != 0 {
			for _, kafkaCluster := range kafkaList.Items {
				err = r.Client.Delete(ctx, &kafkaCluster)
				if err != nil {
					return err
				}
			}
		}

		osList := &v1beta1.OpenSearchList{}
		err = r.Client.List(ctx, osList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(osList.Items) != 0 {
			for _, osCluster := range osList.Items {
				err = r.Client.Delete(ctx, &osCluster)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func areSecondaryCadenceTargetsEqual(k8sTargets, iTargets []*v1beta1.TargetCadence) bool {
	for _, iTarget := range iTargets {
		for _, k8sTarget := range k8sTargets {
			return *iTarget == *k8sTarget
		}
	}

	return len(iTargets) == len(k8sTargets)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CadenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay,
				ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&v1beta1.Cadence{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent

				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				oldObj := event.ObjectOld.(*v1beta1.Cadence)
				newObj := event.ObjectNew.(*v1beta1.Cadence)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				r.handleClusterResourcesEvents(newObj, &oldObj.Spec)

				if oldObj.Generation == newObj.Generation {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				event.ObjectNew.SetAnnotations(newObj.Annotations)

				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				annotations := genericEvent.Object.GetAnnotations()
				annotations[models.ResourceStateAnnotation] = models.GenericEvent
				genericEvent.Object.SetAnnotations(annotations)
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Complete(r)
}

func (r *CadenceReconciler) reconcileMaintenanceEvents(ctx context.Context, c *v1beta1.Cadence) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(c.Status.ID)
	if err != nil {
		return err
	}

	if !c.Status.AreMaintenanceEventStatusesEqual(iMEStatuses) {
		patch := c.NewPatch()
		c.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, c, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", c.Status.ID,
			"events", c.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *CadenceReconciler) handleExternalDelete(ctx context.Context, c *v1beta1.Cadence) error {
	l := log.FromContext(ctx)

	patch := c.NewPatch()
	c.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, c, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(c, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.StatusChecker))

	return nil
}
