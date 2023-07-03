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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cadence object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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
			return models.ExitReconcile, nil
		}

		logger.Error(err, "Unable to fetch Cadence resource",
			"resource name", req.NamespacedName,
		)
		return models.ReconcileRequeue, nil
	}

	switch cadenceCluster.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.HandleCreateCluster(ctx, cadenceCluster, logger), nil
	case models.UpdatingEvent:
		return r.HandleUpdateCluster(ctx, cadenceCluster, logger), nil
	case models.DeletingEvent:
		return r.HandleDeleteCluster(ctx, cadenceCluster, logger), nil
	case models.GenericEvent:
		logger.Info("Generic event isn't handled",
			"request", req,
			"event", cadenceCluster.Annotations[models.ResourceStateAnnotation],
		)

		return models.ExitReconcile, nil
	default:
		logger.Info("Unknown event isn't handled",
			"request", req,
			"event", cadenceCluster.Annotations[models.ResourceStateAnnotation],
		)

		return models.ExitReconcile, nil
	}
}

func (r *CadenceReconciler) HandleCreateCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) reconcile.Result {
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

				return models.ReconcileRequeue
			}

			if requeueNeeded {
				logger.Info("Waiting for bundled clusters to be created",
					"cadence cluster name", cadence.Spec.Name)

				r.EventRecorder.Event(cadence, models.Normal, "Waiting",
					"Waiting for bundled clusters to be created")

				return models.ReconcileRequeue
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

			return models.ReconcileRequeue
		}

		id, err := r.API.CreateCluster(instaclustr.CadenceEndpoint, cadenceAPISpec)
		if err != nil {
			logger.Error(
				err, "Cannot create Cadence cluster",
				"cadence manifest", cadence.Spec,
			)
			r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr is failed. Reason: %v", err)

			return models.ReconcileRequeue
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

			return models.ReconcileRequeue
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

			return models.ReconcileRequeue
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

	err := r.startClusterStatusJob(cadence)
	if err != nil {
		logger.Error(err, "Cannot start cluster status job",
			"cadence cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.CreationFailed,
			"Cluster status check job is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	r.EventRecorder.Event(cadence, models.Normal, models.Created,
		"Cluster status check job is started")

	return models.ExitReconcile
}

func (r *CadenceReconciler) HandleUpdateCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	iData, err := r.API.GetCadence(cadence.Status.ID)
	if err != nil {
		logger.Error(
			err, "Cannot get Cadence cluster from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
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

		return models.ReconcileRequeue
	}

	if iCadence.Status.CurrentClusterOperationStatus != models.NoOperation {
		logger.Info("Cadence cluster is not ready to update",
			"cluster name", iCadence.Spec.Name,
			"cluster state", iCadence.Status.State,
			"current operation status", iCadence.Status.CurrentClusterOperationStatus,
		)

		return models.ReconcileRequeue
	}

	if cadence.Annotations[models.ExternalChangesAnnotation] == models.True {
		return r.handleExternalChanges(cadence, iCadence, logger)
	}

	err = r.updateDescriptionAndTwoFactorDelete(cadence)
	if err != nil {
		logger.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
			"cluster ID", cadence.Status.ID,
			"description", cadence.Spec.Description,
			"two factor delete", cadence.Spec.TwoFactorDelete,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.UpdateFailed,
			"Cluster description and TwoFactoDelete update is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	err = r.API.UpdateCluster(cadence.Status.ID, instaclustr.CadenceEndpoint, cadence.Spec.NewDCsUpdate())
	if err != nil {
		logger.Error(err, "Cannot update Cadence cluster",
			"cluster ID", cadence.Status.ID,
			"update request", cadence.Spec.NewDCsUpdate(),
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	patch := cadence.NewPatch()
	cadence.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, cadence, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Cadence cluster",
			"cluster name", cadence.Spec.Name,
			"patch", patch)

		r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was updated",
		"cluster name", cadence.Spec.Name,
		"cluster status", cadence.Status,
	)

	return models.ExitReconcile
}

func (r *CadenceReconciler) handleExternalChanges(cadence, iCadence *v1beta1.Cadence, l logr.Logger) reconcile.Result {
	if cadence.Annotations[models.AllowSpecAmendAnnotation] != models.True {
		l.Info("Update is blocked until k8s resource specification is equal with Instaclustr",
			"specification of k8s resource", cadence.Spec,
			"data from Instaclustr ", iCadence.Spec)

		r.EventRecorder.Event(cadence, models.Warning, models.UpdateFailed,
			"There are external changes on the Instaclustr console. Please reconcile the specification manually")

		return models.ExitReconcile
	} else {
		if !cadence.Spec.IsEqual(iCadence.Spec) {
			l.Info(msgSpecStillNoMatch,
				"specification of k8s resource", cadence.Spec,
				"data from Instaclustr ", iCadence.Spec)
			r.EventRecorder.Event(cadence, models.Warning, models.ExternalChanges, msgSpecStillNoMatch)

			return models.ExitReconcile
		}

		patch := cadence.NewPatch()

		cadence.Annotations[models.ExternalChangesAnnotation] = ""
		cadence.Annotations[models.AllowSpecAmendAnnotation] = ""

		err := r.Patch(context.Background(), cadence, patch)
		if err != nil {
			l.Error(err, "Cannot patch cluster resource",
				"cluster name", cadence.Spec.Name, "cluster ID", cadence.Status.ID)

			r.EventRecorder.Eventf(cadence, models.Warning, models.PatchFailed,
				"Cluster resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue
		}

		l.Info("External changes have been reconciled", "resource ID", cadence.Status.ID)
		r.EventRecorder.Event(cadence, models.Normal, models.ExternalChanges, "External changes have been reconciled")

		return models.ExitReconcile
	}
}

func (r *CadenceReconciler) HandleDeleteCluster(
	ctx context.Context,
	cadence *v1beta1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	_, err := r.API.GetCadence(cadence.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(
			err, "Cannot get Cadence cluster status from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		r.EventRecorder.Eventf(cadence, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v", err)

		return models.ReconcileRequeue
	}

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

			return models.ReconcileRequeue
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

				return models.ReconcileRequeue
			}

			logger.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", cadence.Status.ID)

			r.EventRecorder.Event(cadence, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")

			return models.ExitReconcile
		}

		return models.ReconcileRequeue
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

			return models.ReconcileRequeue
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
		return models.ReconcileRequeue
	}

	err = exposeservice.Delete(r.Client, cadence.Name, cadence.Namespace)
	if err != nil {
		logger.Error(err, "Cannot delete Cadence cluster expose service",
			"cluster ID", cadence.Status.ID,
			"cluster name", cadence.Spec.Name,
		)

		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was deleted",
		"cluster name", cadence.Spec.Name,
		"cluster ID", cadence.Status.ID,
	)

	r.EventRecorder.Event(cadence, models.Normal, models.Deleted, "Cluster resource is deleted")

	return models.ExitReconcile
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
		cassandraSpec, err := r.newCassandraSpec(cluster)
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
			kafkaSpec, err := r.newKafkaSpec(cluster)
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
			osSpec, err := r.newOpenSearchSpec(cluster)
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

func (r *CadenceReconciler) newCassandraSpec(cadence *v1beta1.Cadence) (*v1beta1.Cassandra, error) {
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
			Version:               models.CassandraV3_11_13,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres:         cassandraDataCentres,
		PasswordAndUserAuth: cassPasswordAndUserAuth,
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
				activeClusters, err := r.API.ListClusters()
				if err != nil {
					l.Error(err, "Cannot list account active clusters")
					return err
				}

				if !isClusterActive(cadence.Status.ID, activeClusters) {
					l.Info("Cluster is not found in Instaclustr. Deleting resource.",
						"cluster ID", cadence.Status.ClusterStatus.ID,
						"cluster name", cadence.Spec.Name,
					)

					patch := cadence.NewPatch()
					cadence.Annotations[models.ClusterDeletionAnnotation] = ""
					cadence.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					err = r.Patch(context.TODO(), cadence, patch)
					if err != nil {
						l.Error(err, "Cannot patch Cadence cluster resource",
							"cluster ID", cadence.Status.ID,
							"cluster name", cadence.Spec.Name,
							"resource name", cadence.Name,
						)

						return err
					}

					err = r.Delete(context.TODO(), cadence)
					if err != nil {
						l.Error(err, "Cannot delete Cadence cluster resource",
							"cluster ID", cadence.Status.ID,
							"cluster name", cadence.Spec.Name,
							"resource name", cadence.Name,
						)

						return err
					}

					return nil
				}
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

		if !areStatusesEqual(&iCadence.Status.ClusterStatus, &cadence.Status.ClusterStatus) {
			l.Info("Updating Cadence cluster status",
				"new status", iCadence.Status.ClusterStatus,
				"old status", cadence.Status.ClusterStatus,
			)

			areDCsEqual := areDataCentresEqual(iCadence.Status.ClusterStatus.DataCentres, cadence.Status.ClusterStatus.DataCentres)

			patch := cadence.NewPatch()
			cadence.Status.ClusterStatus = iCadence.Status.ClusterStatus
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
					nodes,
					models.CadenceConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if iCadence.Status.CurrentClusterOperationStatus == models.NoOperation &&
			cadence.Annotations[models.ExternalChangesAnnotation] != models.True &&
			!cadence.Spec.IsEqual(iCadence.Spec) {
			l.Info(msgExternalChanges, "instaclustr data", iCadence.Spec, "k8s resource spec", cadence.Spec)

			patch := cadence.NewPatch()
			cadence.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), cadence, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", cadence.Spec.Name, "cluster state", cadence.Status.State)
				return err
			}

			r.EventRecorder.Event(cadence, models.Warning, models.ExternalChanges,
				"There are external changes on the Instaclustr console. Please reconcile the specification manually")
		}

		maintEvents, err := r.API.GetMaintenanceEvents(cadence.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cadence cluster maintenance events",
				"cluster name", cadence.Spec.Name,
				"cluster ID", cadence.Status.ID,
			)

			return err
		}

		if !cadence.Status.AreMaintenanceEventsEqual(maintEvents) {
			patch := cadence.NewPatch()
			cadence.Status.MaintenanceEvents = maintEvents
			err = r.Status().Patch(context.TODO(), cadence, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cadence cluster maintenance events",
					"cluster name", cadence.Spec.Name,
					"cluster ID", cadence.Status.ID,
				)

				return err
			}

			l.Info("Cadence cluster maintenance events were updated",
				"cluster ID", cadence.Status.ID,
				"events", cadence.Status.MaintenanceEvents,
			)
		}

		return nil
	}
}

func (r *CadenceReconciler) newKafkaSpec(cadence *v1beta1.Cadence) (*v1beta1.Kafka, error) {
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
			Version:               models.KafkaV3_1_2,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       kafkaTFD,
			PCICompliance:         pciCompliance,
		},
		DataCentres:               kafkaDataCentres,
		ReplicationFactorNumber:   bundledKafkaSpec.ReplicationFactor,
		PartitionsNumber:          bundledKafkaSpec.PartitionsNumber,
		AllowDeleteTopics:         true,
		AutoCreateTopics:          true,
		ClientToClusterEncryption: clientEncryption,
	}

	return &v1beta1.Kafka{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) newOpenSearchSpec(cadence *v1beta1.Cadence) (*v1beta1.OpenSearch, error) {
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

	osNodeSize := bundledOpenSearchSpec.NodeSize
	osReplicationFactor := bundledOpenSearchSpec.ReplicationFactor
	osNodesNumber := bundledOpenSearchSpec.NodesPerReplicationFactor
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
			DataCentre: v1beta1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            osNodeSize,
				NodesNumber:         osNodesNumber,
				Network:             osNetwork,
			},
			ReplicationFactor: osReplicationFactor,
		},
	}
	spec := v1beta1.OpenSearchSpec{
		Cluster: v1beta1.Cluster{
			Name:                  models.OpenSearchChildPrefix + cadence.Name,
			Version:               models.OpenSearchV1_3_7,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres: osDataCentres,
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

func (r *CadenceReconciler) updateDescriptionAndTwoFactorDelete(cadence *v1beta1.Cadence) error {
	var twoFactorDelete *v1beta1.TwoFactorDelete
	if len(cadence.Spec.TwoFactorDelete) != 0 {
		twoFactorDelete = cadence.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, cadence.Status.ID, cadence.Spec.Description, twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CadenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Cadence{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent

				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				oldObj := event.ObjectOld.(*v1beta1.Cadence)
				newObj := event.ObjectNew.(*v1beta1.Cadence)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

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
		})).
		Complete(r)
}
