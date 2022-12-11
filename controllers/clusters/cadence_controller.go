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
				"Resource name", req.NamespacedName,
			)
			return models.ReconcileResult, nil
		}

		logger.Error(err, "unable to fetch Cadence resource",
			"Resource name", req.NamespacedName,
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
		logger.Error(models.ZeroDataCentres, "Cadence cluster spec doesn't have data centres",
			"resource name", cadenceCluster.Name,
		)

		return models.ReconcileRequeue
	}

	if cadenceCluster.Status.ID == "" {
		if cadenceCluster.Spec.ProvisioningType == models.PackagedProvisioningType {
			requeueNeeded, err := r.preparePackagedSolution(ctx, cadenceCluster)
			if err != nil {
				logger.Error(err, "cannot prepare packaged solution for Cadence cluster",
					"Cluster name", cadenceCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}

			if requeueNeeded {
				logger.Info("Waiting for bundled clusters to be created",
					"Cadence cluster name", cadenceCluster.Spec.Name,
				)
				return models.ReconcileRequeue
			}
		}

		logger.Info(
			"Creating Cadence cluster",
			"Cluster name", cadenceCluster.Spec.Name,
			"Data centres", cadenceCluster.Spec.DataCentres,
		)

		cadenceAPISpec, err := cadenceCluster.Spec.ToInstAPIv1(ctx, r.Client)
		if err != nil {
			logger.Error(err, "cannot convert Cadence cluster manifest to API spec",
				"Cluster manifest", cadenceCluster.Spec,
			)
			return models.ReconcileRequeue
		}

		id, err := r.API.CreateCluster(instaclustr.ClustersCreationEndpoint, cadenceAPISpec)
		if err != nil {
			logger.Error(
				err, "cannot create Cadence cluster",
				"Cadence manifest", cadenceCluster.Spec,
			)
			return models.ReconcileRequeue
		}

		if cadenceCluster.Spec.Description != "" {
			err = r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, id, cadenceCluster.Spec.Description, nil)
			if err != nil {
				logger.Error(err, "cannot update Cadence cluster description and TwoFactorDelete",
					"Cluster name", cadenceCluster.Spec.Name,
					"Description", cadenceCluster.Spec.Description,
					"TwoFactorDelete", cadenceCluster.Spec.TwoFactorDelete,
				)
			}
		}

		cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		cadenceCluster.Annotations[models.DeletionConfirmed] = models.False
		cadenceCluster.Finalizers = append(cadenceCluster.Finalizers, models.DeletionFinalizer)

		patch, err := cadenceCluster.NewClusterMetadataPatch()
		if err != nil {
			logger.Error(err, "cannot create Cadence cluster metadata patch",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster metadata", cadenceCluster.ObjectMeta,
			)
			return models.ReconcileRequeue
		}

		err = r.Client.Patch(ctx, cadenceCluster, patch)
		if err != nil {
			logger.Error(err, "cannot patch Cadence cluster",
				"Cluster name", cadenceCluster.Spec.Name,
				"Patch", patch,
			)
			return models.ReconcileRequeue
		}

		statusPatch := cadenceCluster.NewPatch()
		cadenceCluster.Status.ID = id
		err = r.Status().Patch(ctx, cadenceCluster, statusPatch)
		if err != nil {
			logger.Error(err, "cannot update Cadence cluster status",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceCluster.Status,
			)
			return models.ReconcileRequeue
		}

		logger.Info(
			"Cadence resource has been created",
			"Cluster name", cadenceCluster.Name,
			"Cluster ID", cadenceCluster.Status.ID,
			"Kind", cadenceCluster.Kind,
			"Api version", cadenceCluster.APIVersion,
			"Namespace", cadenceCluster.Namespace,
		)
	}

	err := r.startClusterStatusJob(cadenceCluster)
	if err != nil {
		logger.Error(err, "cannot start cluster status job",
			"Cadence cluster ID", cadenceCluster.Status.ID,
		)
		return models.ReconcileRequeue
	}

	return reconcile.Result{}
}

func (r *CadenceReconciler) HandleUpdateCluster(
	ctx context.Context,
	cadenceCluster *clustersv1alpha1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	cadenceInstClusterStatus, err := r.API.GetClusterStatus(cadenceCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get Cadence cluster status from the Instaclustr API",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster ID", cadenceCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if len(cadenceInstClusterStatus.DataCentres) < 1 {
		logger.Error(models.ZeroDataCentres, "Cadence cluster data centres in Instaclustr are empty",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster status", cadenceInstClusterStatus,
		)
		return models.ReconcileRequeue
	}

	cadenceResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(cadenceInstClusterStatus.ID, cadenceInstClusterStatus.CDCID)
	if err != nil {
		logger.Error(
			err, "cannot get Cadence cluster resize operations from the Instaclustr API",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster ID", cadenceCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	updatedFields := &models.CadenceUpdatedFields{}
	annotations := cadenceCluster.Annotations[models.UpdatedFieldsAnnotation]
	err = json.Unmarshal([]byte(annotations), updatedFields)
	if err != nil {
		logger.Error(err, "cannot unmarshal updated fields from annotation",
			"Cluster name", cadenceCluster.Spec.Name,
			"Annotation", cadenceCluster.Annotations[models.UpdatedFieldsAnnotation],
		)

		return models.ReconcileRequeue
	}

	if updatedFields.DescriptionUpdated || updatedFields.TwoFactorDeleteUpdated {
		var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
		if len(cadenceCluster.Spec.TwoFactorDelete) > 0 {
			twoFactorDelete = cadenceCluster.Spec.TwoFactorDelete[0]
		}

		err = r.API.UpdateDescriptionAndTwoFactorDelete(
			instaclustr.ClustersEndpointV1,
			cadenceCluster.Status.ID,
			cadenceCluster.Spec.Description,
			twoFactorDelete,
		)
		if err != nil {
			logger.Error(err, "cannot update Cadence cluster description and twoFactorDelete",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
				"Two factor delete", cadenceCluster.Spec.TwoFactorDelete,
			)

			return models.ReconcileRequeue
		}

		logger.Info("Cadence cluster description and TwoFactorDelete was updated",
			"Cluster name", cadenceCluster.Spec.Name,
			"Description", cadenceCluster.Spec.Description,
			"TwoFactorDelete", twoFactorDelete,
		)
	}

	if updatedFields.NodeSizeUpdated {
		if cadenceInstClusterStatus.Status != StatusRUNNING {
			logger.Info("Cadence cluster is not ready to resize",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
			)
			return models.ReconcileRequeue
		}

		if len(cadenceResizeOperations) > 0 {
			logger.Info("Cadence cluster has active resize operation",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
				"Resizing data centre id", cadenceResizeOperations[0].CDCID,
				"Operation status", cadenceResizeOperations[0].Status,
			)

			return models.ReconcileRequeue
		}

		if len(cadenceCluster.Spec.DataCentres) < 1 {
			logger.Error(models.ZeroDataCentres, "Cadence resource data centres field is empty",
				"Cluster name", cadenceCluster.Spec.Name,
			)
			return models.ReconcileRequeue
		}

		resizeRequest := &models.ResizeRequest{
			NewNodeSize:           cadenceCluster.Spec.DataCentres[0].NodeSize,
			ConcurrentResizes:     cadenceCluster.Spec.ConcurrentResizes,
			NotifySupportContacts: cadenceCluster.Spec.NotifySupportContacts,
			NodePurpose:           models.CadenceNodePurpose,
			ClusterID:             cadenceInstClusterStatus.ID,
			DataCentreID:          cadenceInstClusterStatus.CDCID,
		}

		err = r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if errors.Is(err, instaclustr.StatusPreconditionFailed) {
			logger.Info("Cadence cluster is not ready to resize",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
				"Reason", err,
			)
			return models.ReconcileRequeue
		}
		if err != nil {
			logger.Error(err, "cannot resize Cadence data centre node size",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
				"Current node size", cadenceInstClusterStatus.DataCentres[0].Nodes[0].Size,
				"New node size", cadenceCluster.Spec.DataCentres[0].NodeSize,
				"Resize request", resizeRequest,
			)
			return models.ReconcileRequeue
		}

		logger.Info("Cadence data centre resize request sent",
			"Cluster name", cadenceCluster.Spec.Name,
		)
	}

	cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	patch, err := cadenceCluster.NewClusterMetadataPatch()
	if err != nil {
		logger.Error(err, "cannot create Cadence cluster metadata patch",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster metadata", cadenceCluster.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	err = r.Client.Patch(ctx, cadenceCluster, patch)
	if err != nil {
		logger.Error(err, "cannot patch Cadence cluster",
			"Cluster name", cadenceCluster.Spec.Name,
			"Patch", patch,
		)
		return models.ReconcileRequeue
	}

	cadenceInstClusterStatus, err = r.API.GetClusterStatus(cadenceCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil {
		logger.Error(
			err, "cannot get Cadence cluster status from the Instaclustr API",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster ID", cadenceCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	statusPatch := cadenceCluster.NewPatch()
	cadenceCluster.Status.ClusterStatus = *cadenceInstClusterStatus
	err = r.Status().Patch(ctx, cadenceCluster, statusPatch)
	if err != nil {
		logger.Error(err, "cannot update Cadence cluster status",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster status", cadenceCluster.Status,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was updated",
		"Cluster name", cadenceCluster.Spec.Name,
		"Cluster status", cadenceCluster.Status,
	)

	return models.ReconcileResult
}

func (r *CadenceReconciler) HandleDeleteCluster(
	ctx context.Context,
	cadenceCluster *clustersv1alpha1.Cadence,
	logger logr.Logger,
) reconcile.Result {
	cadenceInstClusterStatus, err := r.API.GetClusterStatus(cadenceCluster.Status.ID, instaclustr.ClustersEndpointV1)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		logger.Error(
			err, "cannot get Cadence cluster status from the Instaclustr API",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster ID", cadenceCluster.Status.ID,
		)

		return models.ReconcileRequeue
	}

	if cadenceInstClusterStatus != nil {
		if len(cadenceCluster.Spec.TwoFactorDelete) != 0 &&
			cadenceCluster.Annotations[models.DeletionConfirmed] != models.True {
			cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
			patch, err := cadenceCluster.NewClusterMetadataPatch()
			if err != nil {
				logger.Error(err, "cannot create Cadence cluster resource metadata patch",
					"Cluster name", cadenceCluster.Spec.Name,
				)

				return models.ReconcileRequeue
			}

			err = r.Patch(ctx, cadenceCluster, patch)
			if err != nil {
				logger.Error(err, "cannot patch Cadence cluster resource metadata",
					"Cluster name", cadenceCluster.Spec.Name,
				)

				return models.ReconcileRequeue
			}

			logger.Info("Cadence cluster deletion is not confirmed",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster ID", cadenceCluster.Status.ID,
				"Confirmation annotation", models.DeletionConfirmed,
				"Annotation value", cadenceCluster.Annotations[models.DeletionConfirmed],
			)

			return models.ReconcileRequeue
		}

		err = r.API.DeleteCluster(cadenceCluster.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			logger.Error(err, "cannot delete Cadence cluster",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster status", cadenceInstClusterStatus.Status,
			)
			return models.ReconcileRequeue
		}

		logger.Info("Cadence cluster is deleting",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster status", cadenceInstClusterStatus.Status,
		)

		return models.ReconcileRequeue
	}

	if cadenceCluster.Spec.ProvisioningType == models.PackagedProvisioningType {
		err = r.deletePackagedResources(ctx, cadenceCluster)
		if err != nil {
			logger.Error(
				err, "cannot delete Cadence packaged resources",
				"Cluster name", cadenceCluster.Spec.Name,
				"Cluster ID", cadenceCluster.Status.ID,
			)
			return models.ReconcileRequeue
		}
	}

	r.Scheduler.RemoveJob(cadenceCluster.GetJobID(scheduler.StatusChecker))
	controllerutil.RemoveFinalizer(cadenceCluster, models.DeletionFinalizer)
	cadenceCluster.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent
	patch, err := cadenceCluster.NewClusterMetadataPatch()
	if err != nil {
		logger.Error(err, "cannot create Cadence cluster metadata patch",
			"Cluster name", cadenceCluster.Spec.Name,
			"Cluster metadata", cadenceCluster.ObjectMeta,
		)
		return models.ReconcileRequeue
	}

	err = r.Client.Patch(ctx, cadenceCluster, patch)
	if err != nil {
		logger.Error(err, "cannot patch Cadence cluster",
			"Cluster name", cadenceCluster.Spec.Name,
			"Patch", patch,
		)
		return models.ReconcileRequeue
	}

	logger.Info("Cadence cluster was deleted",
		"Cluster name", cadenceCluster.Spec.Name,
		"Cluster ID", cadenceCluster.Status.ID,
	)

	return models.ReconcileResult
}

func (r *CadenceReconciler) getDataCentreOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error) {
	activeResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(clusterID, dataCentreID)
	if err != nil {
		return nil, nil
	}

	return activeResizeOperations, nil
}

func (r *CadenceReconciler) preparePackagedSolution(ctx context.Context, cluster *clustersv1alpha1.Cadence) (bool, error) {
	if len(cluster.Spec.DataCentres) < 1 {
		return false, models.ZeroDataCentres
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
		cluster.Spec.TargetKafkaVPCType = models.VPC_PEERED

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
		cluster.Spec.TargetOpenSearchVPCType = models.VPC_PEERED
	}

	if len(cassandraList.Items[0].Status.DataCentres) == 0 {
		return true, nil
	}

	cluster.Spec.TargetCassandraCDCID = cassandraList.Items[0].Status.DataCentres[0].ID
	cluster.Spec.TargetCassandraVPCType = models.VPC_PEERED

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
		return nil, models.ZeroDataCentres
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
		return nil, models.NetworkOverlaps
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
				RacksNumber:         int32(cassReplicationFactor),
				Network:             network,
			},
			PrivateIPBroadcastForDiscovery: cassPrivateIPBroadcastForDiscovery,
		},
	}
	spec := clustersv1alpha1.CassandraSpec{
		Cluster: clustersv1alpha1.Cluster{
			Name:                  models.CassandraChildPrefix + cadence.Name,
			Version:               models.V3_11_13,
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
			l.Error(err, "cannot get Cadence custom resource",
				"Resource name", cadence.Name,
			)
			return err
		}

		if clusterresourcesv1alpha1.IsClusterBeingDeleted(
			cadence.DeletionTimestamp,
			len(cadence.Spec.TwoFactorDelete),
			cadence.Annotations[models.DeletionConfirmed],
		) {
			l.Info("Cadence cluster is being deleted. Status check job skipped",
				"Cluster name", cadence.Spec.Name,
				"Cluster ID", cadence.Status.ID,
			)

			return nil
		}

		instStatus, err := r.API.GetClusterStatus(cadence.Status.ID, instaclustr.ClustersEndpointV1)
		if err != nil {
			l.Error(err, "cannot get Cadence cluster status",
				"ClusterID", cadence.Status.ID,
			)
			return err
		}

		if !isStatusesEqual(instStatus, &cadence.Status.ClusterStatus) {
			l.Info("Updating Cadence cluster status",
				"New status", instStatus,
				"Old status", cadence.Status.ClusterStatus,
			)

			patch := cadence.NewPatch()
			cadence.Status.ClusterStatus = *instStatus
			err = r.Status().Patch(context.Background(), cadence, patch)
			if err != nil {
				l.Error(err, "cannot patch Cadence cluster",
					"Cluster name", cadence.Spec.Name,
					"Status", cadence.Status.Status,
				)
				return err
			}
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
		return nil, models.ZeroDataCentres
	}

	kafkaNodeSize := cadence.Spec.BundledKafkaSpec.NodeSize
	kafkaNodesNumber := cadence.Spec.BundledKafkaSpec.NodesNumber
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

	kafkaNetwork := cadence.Spec.BundledKafkaSpec.Network
	isKafkaNetworkOverlaps, err := cadence.Spec.DataCentres[0].IsNetworkOverlaps(kafkaNetwork)
	if err != nil {
		return nil, err
	}
	if isKafkaNetworkOverlaps {
		return nil, models.NetworkOverlaps
	}

	dcName := models.KafkaChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName

	kafkaDataCentres := []*clustersv1alpha1.KafkaDataCentre{
		&clustersv1alpha1.KafkaDataCentre{
			Name:                dcName,
			Region:              dcRegion,
			CloudProvider:       cloudProvider,
			ProviderAccountName: providerAccountName,
			NodeSize:            kafkaNodeSize,
			NodesNumber:         int32(kafkaNodesNumber),
			Network:             kafkaNetwork,
		},
	}
	spec := clustersv1alpha1.KafkaSpec{
		Cluster: clustersv1alpha1.Cluster{
			Name:                  models.KafkaChildPrefix + cadence.Name,
			Version:               models.V2_7_1,
			SLATier:               slaTier,
			PrivateNetworkCluster: privateClusterNetwork,
			TwoFactorDelete:       twoFactorDelete,
			PCICompliance:         pciCompliance,
		},
		DataCentres:               kafkaDataCentres,
		ReplicationFactorNumber:   int32(cadence.Spec.BundledKafkaSpec.ReplicationFactor),
		PartitionsNumber:          int32(cadence.Spec.BundledKafkaSpec.PartitionsNumber),
		AllowDeleteTopics:         true,
		AutoCreateTopics:          true,
		ClientToClusterEncryption: cadence.Spec.ClientEncryption,
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
		return nil, models.ZeroDataCentres
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
			&clustersv1alpha1.TwoFactorDelete{
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
		return nil, models.NetworkOverlaps
	}

	dcName := models.OpenSearchChildDCName
	dcRegion := cadence.Spec.DataCentres[0].Region
	cloudProvider := cadence.Spec.DataCentres[0].CloudProvider
	providerAccountName := cadence.Spec.DataCentres[0].ProviderAccountName

	osDataCentres := []*clustersv1alpha1.OpenSearchDataCentre{
		&clustersv1alpha1.OpenSearchDataCentre{
			DataCentre: clustersv1alpha1.DataCentre{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				NodeSize:            osNodeSize,
				RacksNumber:         int32(osReplicationFactor),
				NodesNumber:         int32(osNodesNumber),
				Network:             osNetwork,
			},
		},
	}
	spec := clustersv1alpha1.OpenSearchSpec{
		Cluster: clustersv1alpha1.Cluster{
			Name:                  models.OpenSearchChildPrefix + cadence.Name,
			Version:               models.V1_3_5,
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

				updatedFields := newObj.Spec.GetUpdatedFields(&oldObj.Spec)
				updatedFieldsJson, _ := json.Marshal(&updatedFields)

				newObj.Annotations[models.UpdatedFieldsAnnotation] = string(updatedFieldsJson)
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
