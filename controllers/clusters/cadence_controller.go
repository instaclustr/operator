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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// CadenceReconciler reconciles a Cadence object
type CadenceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	API       instaclustr.API
	Scheduler scheduler.Interface
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;patch

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

	cadenceCluster := &clustersv1alpha1.Cadence{}
	err := r.Client.Get(ctx, req.NamespacedName, cadenceCluster)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("Cadence resource is not found",
				"resource name", req.NamespacedName,
			)
			return models.ReconcileResult, nil
		}

		logger.Error(err, "Unable to fetch Cadence resource",
			"resource name", req.NamespacedName,
		)
		return models.ReconcileResult, nil
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

		return models.ReconcileResult, nil
	default:
		logger.Info("Unknown event isn't handled",
			"request", req,
			"event", cadenceCluster.Annotations[models.ResourceStateAnnotation],
		)

		return models.ReconcileResult, nil
	}
}

func (r *CadenceReconciler) HandleCreateCluster(
	ctx context.Context,
	cadenceCluster *clustersv1alpha1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	if len(cadenceCluster.Spec.DataCentres) < 1 {
		logger.Error(models.ErrZeroDataCentres, "Cadence cluster spec doesn't have data centres",
			"resource name", cadenceCluster.Name,
		)

		return models.ReconcileRequeue
	}

	if cadenceCluster.Status.ID == "" {
		if cadenceCluster.Spec.ProvisioningType == models.PackagedProvisioningType {
			requeueNeeded, err := r.preparePackagedSolution(ctx, cadenceCluster)
			if err != nil {
				logger.Error(err, "Cannot prepare packaged solution for Cadence cluster",
					"cluster name", cadenceCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}

			if requeueNeeded {
				logger.Info("Waiting for bundled clusters to be created",
					"cadence cluster name", cadenceCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}
		}

		logger.Info(
			"Creating Cadence cluster",
			"cluster name", cadenceCluster.Spec.Name,
			"data centres", cadenceCluster.Spec.DataCentres,
		)

		cadenceAPISpec, err := cadenceCluster.Spec.ToInstAPI(ctx, r.Client)
		if err != nil {
			logger.Error(err, "Cannot convert Cadence cluster manifest to API spec",
				"cluster manifest", cadenceCluster.Spec,
			)
			return models.ReconcileRequeue
		}

		id, err := r.API.CreateCluster(instaclustr.CadenceEndpoint, cadenceAPISpec)
		if err != nil {
			logger.Error(
				err, "Cannot create Cadence cluster",
				"cadence manifest", cadenceCluster.Spec,
			)
			return models.ReconcileRequeue
		}

		if cadenceCluster.Spec.Description != "" {
			err = r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, id, cadenceCluster.Spec.Description, nil)
			if err != nil {
				logger.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
					"cluster name", cadenceCluster.Spec.Name,
					"description", cadenceCluster.Spec.Description,
					"twoFactorDelete", cadenceCluster.Spec.TwoFactorDelete,
				)
			}
		}

		cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		cadenceCluster.Annotations[models.DeletionConfirmed] = models.False
		cadenceCluster.Finalizers = append(cadenceCluster.Finalizers, models.DeletionFinalizer)

		patch, err := cadenceCluster.NewClusterMetadataPatch()
		if err != nil {
			logger.Error(err, "Cannot create Cadence cluster metadata patch",
				"cluster name", cadenceCluster.Spec.Name,
				"cluster metadata", cadenceCluster.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		err = r.Client.Patch(ctx, cadenceCluster, patch)
		if err != nil {
			logger.Error(err, "Cannot patch Cadence cluster",
				"cluster name", cadenceCluster.Spec.Name,
				"patch", patch,
			)
			return models.ReconcileRequeue
		}

		statusPatch := cadenceCluster.NewPatch()
		cadenceCluster.Status.ID = id
		err = r.Status().Patch(ctx, cadenceCluster, statusPatch)
		if err != nil {
			logger.Error(err, "Cannot update Cadence cluster status",
				"cluster name", cadenceCluster.Spec.Name,
				"cluster status", cadenceCluster.Status,
			)
			return models.ReconcileRequeue
		}

		logger.Info(
			"Cadence resource has been created",
			"cluster name", cadenceCluster.Name,
			"cluster ID", cadenceCluster.Status.ID,
			"kind", cadenceCluster.Kind,
			"api version", cadenceCluster.APIVersion,
			"namespace", cadenceCluster.Namespace,
		)
	}

	err := r.startClusterStatusJob(cadenceCluster)
	if err != nil {
		logger.Error(err, "Cannot start cluster status job",
			"cadence cluster ID", cadenceCluster.Status.ID,
		)
		return models.ReconcileRequeue
	}

	return models.ReconcileResult
}

func (r *CadenceReconciler) HandleUpdateCluster(
	ctx context.Context,
	cadence *clustersv1alpha1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	cadenceInstClusterStatus, err := r.API.GetCadence(cadence.Status.ID)
	if err != nil {
		logger.Error(
			err, "Cannot get Cadence cluster status from the Instaclustr API",
			"cluster name", cadence.Spec.Name,
			"cluster ID", cadence.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if len(cadenceInstClusterStatus.DataCentres) < 1 {
		logger.Error(models.ErrZeroDataCentres, "Cadence cluster data centres in Instaclustr are empty",
			"cluster name", cadence.Spec.Name,
			"cluster status", cadenceInstClusterStatus,
		)
		return models.ReconcileRequeue
	}

	if cadenceInstClusterStatus.CurrentClusterOperationStatus != models.NoOperation {
		logger.Info("Cadence cluster is not ready to update",
			"cluster name", cadenceInstClusterStatus.Name,
			"cluster status", cadenceInstClusterStatus.Status,
			"current operation status", cadenceInstClusterStatus.CurrentClusterOperationStatus,
		)

		return models.ReconcileRequeue
	}

	err = r.updateDescriptionAndTwoFactorDelete(cadence)
	if err != nil {
		logger.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
			"cluster ID", cadence.Status.ID,
			"description", cadence.Spec.Description,
			"two factor delete", cadence.Spec.TwoFactorDelete,
		)

		return models.ReconcileRequeue
	}

	updateReq := struct {
		DataCentres []*models.CadenceDataCentre `json:"dataCentres"`
	}{
		DataCentres: cadence.Spec.DataCentresToInstAPI(),
	}
	err = r.API.UpdateCluster(cadence.Status.ID, instaclustr.CadenceEndpoint, updateReq)
	if err != nil {
		logger.Error(err, "Cannot update Cadence cluster",
			"cluster ID", cadence.Status.ID,
			"update request", updateReq,
		)

		return models.ReconcileRequeue
	}

	patch := cadence.NewPatch()
	cadence.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Client.Patch(ctx, cadence, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Cadence cluster",
			"cluster name", cadence.Spec.Name,
			"patch", patch,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was updated",
		"cluster name", cadence.Spec.Name,
		"cluster status", cadence.Status,
	)

	return models.ReconcileResult
}

func (r *CadenceReconciler) HandleDeleteCluster(
	ctx context.Context,
	cadenceCluster *clustersv1alpha1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	cadenceInstClusterStatus, err := r.API.GetCadence(cadenceCluster.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(
			err, "Cannot get Cadence cluster status from the Instaclustr API",
			"cluster name", cadenceCluster.Spec.Name,
			"cluster ID", cadenceCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if cadenceInstClusterStatus != nil {
		if len(cadenceCluster.Spec.TwoFactorDelete) != 0 &&
			cadenceCluster.Annotations[models.DeletionConfirmed] != models.True {
			cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			patch, err := cadenceCluster.NewClusterMetadataPatch()
			if err != nil {
				logger.Error(err, "Cannot create Cadence cluster resource metadata patch",
					"cluster name", cadenceCluster.Spec.Name,
				)

				return models.ReconcileRequeue
			}

			err = r.Patch(ctx, cadenceCluster, patch)
			if err != nil {
				logger.Error(err, "Cannot patch Cadence cluster resource metadata",
					"cluster name", cadenceCluster.Spec.Name,
				)

				return models.ReconcileRequeue
			}

			logger.Info("Cadence cluster deletion is not confirmed",
				"cluster name", cadenceCluster.Spec.Name,
				"cluster ID", cadenceCluster.Status.ID,
				"confirmation annotation", models.DeletionConfirmed,
				"annotation value", cadenceCluster.Annotations[models.DeletionConfirmed],
			)

			return models.ReconcileRequeue
		}

		err = r.API.DeleteCluster(cadenceCluster.Status.ID, instaclustr.CadenceEndpoint)
		if err != nil {
			logger.Error(err, "Cannot delete Cadence cluster",
				"cluster name", cadenceCluster.Spec.Name,
				"cluster status", cadenceInstClusterStatus.Status,
			)
			return models.ReconcileRequeue
		}

		logger.Info("Cadence cluster is being deleted",
			"cluster name", cadenceCluster.Spec.Name,
			"cluster status", cadenceInstClusterStatus.Status,
		)

		return models.ReconcileRequeue
	}

	if cadenceCluster.Spec.ProvisioningType == models.PackagedProvisioningType {
		err = r.deletePackagedResources(ctx, cadenceCluster)
		if err != nil {
			logger.Error(
				err, "Cannot delete Cadence packaged resources",
				"cluster name", cadenceCluster.Spec.Name,
				"cluster ID", cadenceCluster.Status.ID,
			)
			return models.ReconcileRequeue
		}
	}

	r.Scheduler.RemoveJob(cadenceCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(cadenceCluster, models.DeletionFinalizer)
	cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	patch, err := cadenceCluster.NewClusterMetadataPatch()
	if err != nil {
		logger.Error(err, "Cannot create Cadence cluster metadata patch",
			"cluster name", cadenceCluster.Spec.Name,
			"cluster metadata", cadenceCluster.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	err = r.Client.Patch(ctx, cadenceCluster, patch)
	if err != nil {
		logger.Error(err, "Cannot patch Cadence cluster",
			"cluster name", cadenceCluster.Spec.Name,
			"patch", patch,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was deleted",
		"cluster name", cadenceCluster.Spec.Name,
		"cluster ID", cadenceCluster.Status.ID,
	)

	return models.ReconcileResult
}

func (r *CadenceReconciler) preparePackagedSolution(ctx context.Context, cluster *clustersv1alpha1.Cadence) (bool, error) {
	if len(cluster.Spec.DataCentres) < 1 {
		return false, models.ErrZeroDataCentres
	}

	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, cluster.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return false, err
	}

	cassandraList := &clustersv1alpha1.CassandraList{}
	err = r.Client.List(ctx, cassandraList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return false, err
	}
	if len(cassandraList.Items) < 1 {
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

	kafkaList := &clustersv1alpha1.KafkaList{}
	osList := &clustersv1alpha1.OpenSearchList{}
	if cluster.Spec.UseAdvancedVisibility {
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

		cluster.Spec.TargetKafkaCDCID = kafkaList.Items[0].Status.DataCentres[0].ID
		cluster.Spec.TargetKafkaVPCType = models.VPCPeered

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

		cluster.Spec.TargetOpenSearchCDCID = osList.Items[0].Status.DataCentres[0].ID
		cluster.Spec.TargetOpenSearchVPCType = models.VPCPeered
	}

	if len(cassandraList.Items[0].Status.DataCentres) == 0 {
		return true, nil
	}

	cluster.Spec.TargetCassandraCDCID = cassandraList.Items[0].Status.DataCentres[0].ID
	cluster.Spec.TargetCassandraVPCType = models.VPCPeered

	return false, nil
}

func (r *CadenceReconciler) newCassandraSpec(cadence *clustersv1alpha1.Cadence) (*clustersv1alpha1.Cassandra, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.CassandraKind,
		APIVersion: models.ClustersV1alpha1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.CassandraChildPrefix + cadence.Name,
		Labels:      map[string]string{models.ControlledByLabel: cadence.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   cadence.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(cadence.Spec.DataCentres) < 1 {
		return nil, models.ErrZeroDataCentres
	}
	cassNodeSize := cadence.Spec.BundledCassandraSpec.NodeSize
	cassNodesNumber := cadence.Spec.BundledCassandraSpec.NodesNumber
	cassReplicationFactor := cadence.Spec.BundledCassandraSpec.ReplicationFactor
	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance

	var twoFactorDelete []*clustersv1alpha1.TwoFactorDelete
	if len(cadence.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*clustersv1alpha1.TwoFactorDelete{
			&clustersv1alpha1.TwoFactorDelete{
				Email: cadence.Spec.TwoFactorDelete[0].Email,
				Phone: cadence.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}

	isCassNetworkOverlaps, err := cadence.Spec.DataCentres[0].IsNetworkOverlaps(cadence.Spec.BundledCassandraSpec.Network)
	if err != nil {
		return nil, err
	}
	if isCassNetworkOverlaps {
		return nil, models.ErrNetworkOverlaps
	}

	dcName := models.CassandraChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	network := cadence.Spec.BundledCassandraSpec.Network
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName
	cassPrivateIPBroadcastForDiscovery := cadence.Spec.BundledCassandraSpec.PrivateIPBroadcastForDiscovery
	cassPasswordAndUserAuth := cadence.Spec.BundledCassandraSpec.PasswordAndUserAuth

	cassandraDataCentres := []*clustersv1alpha1.CassandraDataCentre{
		&clustersv1alpha1.CassandraDataCentre{
			DataCentre: clustersv1alpha1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            cassNodeSize,
				NodesNumber:         int32(cassNodesNumber),
				Network:             network,
			},
			ReplicationFactor:              cassReplicationFactor,
			PrivateIPBroadcastForDiscovery: cassPrivateIPBroadcastForDiscovery,
		},
	}
	spec := clustersv1alpha1.CassandraSpec{
		Cluster: clustersv1alpha1.Cluster{
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

	return &clustersv1alpha1.Cassandra{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) startClusterStatusJob(cadence *clustersv1alpha1.Cadence) error {
	job := r.newWatchStatusJob(cadence)

	err := r.Scheduler.ScheduleJob(cadence.GetJobID(scheduler.StatusChecker), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CadenceReconciler) newWatchStatusJob(cadence *clustersv1alpha1.Cadence) scheduler.Job {
	l := log.Log.WithValues("component", "cadenceStatusClusterJob")
	return func() error {
		err := r.Get(context.Background(), types.NamespacedName{Namespace: cadence.Namespace, Name: cadence.Name}, cadence)
		if err != nil {
			l.Error(err, "Cannot get Cadence custom resource",
				"resource name", cadence.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(
			cadence.DeletionTimestamp,
			len(cadence.Spec.TwoFactorDelete),
			cadence.Annotations[models.DeletionConfirmed],
		) {
			l.Info("Cadence cluster is being deleted. Status check job skipped",
				"cluster name", cadence.Spec.Name,
				"cluster ID", cadence.Status.ID,
			)

			return nil
		}

		instStatus, err := r.API.GetCadence(cadence.Status.ID)
		if err != nil {
			l.Error(err, "Cannot get Cadence cluster status",
				"clusterID", cadence.Status.ID,
			)
			return err
		}

		if !cadence.Status.AreStatusesEqual(instStatus) {
			l.Info("Updating Cadence cluster status",
				"new status", instStatus,
				"old status", cadence.Status.ClusterStatus,
			)

			patch := cadence.NewPatch()
			cadence.Status.SetStatusFromInst(instStatus)
			err = r.Status().Patch(context.Background(), cadence, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cadence cluster",
					"cluster name", cadence.Spec.Name,
					"status", cadence.Status.Status,
				)
				return err
			}
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

func (r *CadenceReconciler) newKafkaSpec(cadence *clustersv1alpha1.Cadence) (*clustersv1alpha1.Kafka, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.KafkaKind,
		APIVersion: models.ClustersV1alpha1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.KafkaChildPrefix + cadence.Name,
		Labels:      map[string]string{models.ControlledByLabel: cadence.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   cadence.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(cadence.Spec.DataCentres) < 1 {
		return nil, models.ErrZeroDataCentres
	}

	var kafkaTFD []*clustersv1alpha1.TwoFactorDelete
	for _, cadenceTFD := range cadence.Spec.TwoFactorDelete {
		twoFactorDelete := &clustersv1alpha1.TwoFactorDelete{
			Email: cadenceTFD.Email,
			Phone: cadenceTFD.Phone,
		}
		kafkaTFD = append(kafkaTFD, twoFactorDelete)
	}

	kafkaNetwork := cadence.Spec.BundledKafkaSpec.Network
	for _, cadenceDC := range cadence.Spec.DataCentres {
		isKafkaNetworkOverlaps, err := cadenceDC.IsNetworkOverlaps(kafkaNetwork)
		if err != nil {
			return nil, err
		}
		if isKafkaNetworkOverlaps {
			return nil, models.ErrNetworkOverlaps
		}
	}

	kafkaNodeSize := cadence.Spec.BundledKafkaSpec.NodeSize
	kafkaNodesNumber := cadence.Spec.BundledKafkaSpec.NodesNumber
	dcName := models.KafkaChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName
	kafkaDataCentres := []*clustersv1alpha1.KafkaDataCentre{{
		DataCentre: clustersv1alpha1.DataCentre{
			Name:                dcName,
			Region:              dcRegion,
			CloudProvider:       cloudProvider,
			ProviderAccountName: providerAccountName,
			NodeSize:            kafkaNodeSize,
			NodesNumber:         int32(kafkaNodesNumber),
			Network:             kafkaNetwork,
		}},
	}

	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance
	clientEncryption := cadence.Spec.DataCentres[0].ClientEncryption
	spec := clustersv1alpha1.KafkaSpec{
		Cluster: clustersv1alpha1.Cluster{
			Name:                  models.KafkaChildPrefix + cadence.Name,
			Version:               models.KafkaV2_8_2,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       kafkaTFD,
			PCICompliance:         pciCompliance,
		},
		DataCentres:               kafkaDataCentres,
		ReplicationFactorNumber:   int32(cadence.Spec.BundledKafkaSpec.ReplicationFactor),
		PartitionsNumber:          int32(cadence.Spec.BundledKafkaSpec.PartitionsNumber),
		AllowDeleteTopics:         true,
		AutoCreateTopics:          true,
		ClientToClusterEncryption: clientEncryption,
	}

	return &clustersv1alpha1.Kafka{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) newOpenSearchSpec(cadence *clustersv1alpha1.Cadence) (*clustersv1alpha1.OpenSearch, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.OpenSearchKind,
		APIVersion: models.ClustersV1alpha1APIVersion,
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

	osNodeSize := cadence.Spec.BundledOpenSearchSpec.NodeSize
	osReplicationFactor := cadence.Spec.BundledOpenSearchSpec.ReplicationFactor
	osNodesNumber := cadence.Spec.BundledOpenSearchSpec.NodesPerReplicationFactor
	slaTier := cadence.Spec.SLATier
	privateClusterNetwork := cadence.Spec.PrivateNetworkCluster
	pciCompliance := cadence.Spec.PCICompliance

	var twoFactorDelete []*clustersv1alpha1.TwoFactorDelete
	if len(cadence.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*clustersv1alpha1.TwoFactorDelete{
			{
				Email: cadence.Spec.TwoFactorDelete[0].Email,
				Phone: cadence.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}

	osNetwork := cadence.Spec.BundledOpenSearchSpec.Network
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

	osDataCentres := []*clustersv1alpha1.OpenSearchDataCentre{
		{
			DataCentre: clustersv1alpha1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            osNodeSize,
				NodesNumber:         int32(osNodesNumber),
				Network:             osNetwork,
			},
			RacksNumber: int32(osReplicationFactor),
		},
	}
	spec := clustersv1alpha1.OpenSearchSpec{
		Cluster: clustersv1alpha1.Cluster{
			Name:                  models.OpenSearchChildPrefix + cadence.Name,
			Version:               models.OpensearchV1_3_7,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres: osDataCentres,
	}

	return &clustersv1alpha1.OpenSearch{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) deletePackagedResources(ctx context.Context, cadence *clustersv1alpha1.Cadence) error {
	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, cadence.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return err
	}

	cassandraList := &clustersv1alpha1.CassandraList{}
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

	if cadence.Spec.UseAdvancedVisibility {
		kafkaList := &clustersv1alpha1.KafkaList{}
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

		osList := &clustersv1alpha1.OpenSearchList{}
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

func (r *CadenceReconciler) updateDescriptionAndTwoFactorDelete(cadence *clustersv1alpha1.Cadence) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
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
		For(&clustersv1alpha1.Cadence{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				annotations := event.Object.GetAnnotations()
				if event.Object.GetDeletionTimestamp() != nil {
					annotations[models.ResourceStateAnnotation] = models.DeletingEvent
					event.Object.SetAnnotations(annotations)
					return true
				}

				annotations[models.ResourceStateAnnotation] = models.CreatingEvent
				event.Object.SetAnnotations(annotations)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				oldObj := event.ObjectOld.(*clustersv1alpha1.Cadence)
				newObj := event.ObjectNew.(*clustersv1alpha1.Cadence)

				if newObj.DeletionTimestamp != nil &&
					(len(newObj.Spec.TwoFactorDelete) == 0 || newObj.Annotations[models.DeletionConfirmed] == models.True) {
					newObj.Annotations[models.ResourceStateAnnotation] = models.DeletingEvent
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
