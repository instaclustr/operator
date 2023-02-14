package clusters

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

func (r *RedisReconciler) ToInstAPIv1(redisSpec *clustersv1alpha1.RedisSpec) *modelsv1.RedisCluster {
	redisBundles := r.bundlesToInstAPIv1(redisSpec.DataCentres[0], redisSpec.Version, redisSpec.ClientEncryption)

	redisInstProvider := r.providerToInstAPIv1(redisSpec.DataCentres[0])

	redisRackAllocation := r.rackAllocationToInstAPIv1(redisSpec.DataCentres[0])

	var redisTwoFactorDelete *models.TwoFactorDelete
	if len(redisSpec.TwoFactorDelete) != 0 {
		redisTwoFactorDelete = r.twoFactorDeleteToInstAPIv1(redisSpec.TwoFactorDelete)
	}

	return &modelsv1.RedisCluster{
		Cluster: models.Cluster{
			ClusterName:           redisSpec.Name,
			NodeSize:              redisSpec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: redisSpec.PrivateNetworkCluster,
			SLATier:               redisSpec.SLATier,
			Provider:              redisInstProvider,
			TwoFactorDelete:       redisTwoFactorDelete,
			RackAllocation:        redisRackAllocation,
			DataCentre:            redisSpec.DataCentres[0].Region,
			DataCentreCustomName:  redisSpec.DataCentres[0].Name,
			ClusterNetwork:        redisSpec.DataCentres[0].Network,
		},
		Bundles: redisBundles,
	}
}

func (r *RedisReconciler) bundlesToInstAPIv1(
	dataCentre *clustersv1alpha1.RedisDataCentre,
	version string,
	clientEncryption bool,
) []*modelsv1.RedisBundle {
	var redisBundles []*modelsv1.RedisBundle

	redisBundle := &modelsv1.RedisBundle{
		Bundle: models.Bundle{
			Bundle:  modelsv1.Redis,
			Version: version,
		},
		Options: &modelsv1.RedisOptions{
			ClientEncryption: clientEncryption,
			MasterNodes:      dataCentre.MasterNodes,
			ReplicaNodes:     dataCentre.ReplicaNodes,
			PasswordAuth:     dataCentre.PasswordAuth,
		},
	}
	redisBundles = append(redisBundles, redisBundle)

	return redisBundles
}

func (r *RedisReconciler) providerToInstAPIv1(dataCentre *clustersv1alpha1.RedisDataCentre) *models.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dataCentre.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dataCentre.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dataCentre.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dataCentre.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		Tags:                   dataCentre.Tags,
		CustomVirtualNetworkID: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}

func (r *RedisReconciler) rackAllocationToInstAPIv1(dataCentre *clustersv1alpha1.RedisDataCentre) *models.RackAllocation {
	return &models.RackAllocation{
		NodesPerRack:  dataCentre.NodesNumber,
		NumberOfRacks: dataCentre.RacksNumber,
	}
}

func (r *RedisReconciler) twoFactorDeleteToInstAPIv1(twoFactorDelete []*clustersv1alpha1.TwoFactorDelete) *models.TwoFactorDelete {
	return &models.TwoFactorDelete{
		DeleteVerifyEmail: twoFactorDelete[0].Email,
		DeleteVerifyPhone: twoFactorDelete[0].Phone,
	}
}

func (r *RedisReconciler) patchClusterMetadata(
	ctx context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(redisCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(redisCluster.Finalizers)
	if err != nil {
		return err
	}

	finalizersPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finalizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(ctx, redisCluster, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("Redis cluster patched",
		"cluster name", redisCluster.Spec.Name,
		"finalizers", redisCluster.Finalizers,
		"annotations", redisCluster.Annotations,
	)
	return nil
}

func (r *RedisReconciler) reconcileDataCentresNumber(
	clusterStatusFromInst *clustersv1alpha1.ClusterStatus,
	cluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) error {
	dataCentresToAdd := r.checkDataCentresToAdd(clusterStatusFromInst, cluster)
	if len(dataCentresToAdd) > 0 {
		for _, dataCentreToAdd := range dataCentresToAdd {
			instDataCentreToAdd := r.dataCentreToInstAPIv1(dataCentreToAdd, cluster.Spec.Version, cluster.Spec.ClientEncryption)
			err := r.API.AddDataCentre(cluster.Status.ID, instaclustr.ClustersEndpointV1, instDataCentreToAdd)
			if err != nil {
				return err
			}

			logger.Info("Add Redis data centre request was sent",
				"cluster name", cluster.Spec.Name,
				"data centre name", dataCentreToAdd.Name,
			)
		}
	}

	return nil
}

func (r *RedisReconciler) checkDataCentresToAdd(
	clusterStatusFromInst *clustersv1alpha1.ClusterStatus,
	cluster *clustersv1alpha1.Redis,
) []*clustersv1alpha1.RedisDataCentre {
	var dataCentresToAdd []*clustersv1alpha1.RedisDataCentre
	dataCentresNumberDiff := len(cluster.Spec.DataCentres) - len(clusterStatusFromInst.DataCentres)
	if dataCentresNumberDiff > 0 {
		dataCentresToAdd = cluster.Spec.DataCentres[1 : dataCentresNumberDiff+1]
	}

	return dataCentresToAdd
}

func (r *RedisReconciler) dataCentreToInstAPIv1(
	dataCentre *clustersv1alpha1.RedisDataCentre,
	version string,
	clientEncryption bool,
) *modelsv1.RedisDataCentre {
	redisBundles := r.bundlesToInstAPIv1(dataCentre, version, clientEncryption)
	redisProvider := r.providerToInstAPIv1(dataCentre)
	redisRackAllocation := r.rackAllocationToInstAPIv1(dataCentre)

	return &modelsv1.RedisDataCentre{
		DataCentre: models.DataCentre{
			Name:           dataCentre.Name,
			DataCentre:     dataCentre.Region,
			Network:        dataCentre.Network,
			Provider:       redisProvider,
			NodeSize:       dataCentre.NodeSize,
			RackAllocation: redisRackAllocation,
		},
		Bundles: redisBundles,
	}
}

func (r *RedisReconciler) reconcileDataCentresNodeSize(
	instClusterStatus *clustersv1alpha1.ClusterStatus,
	cluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) error {
	dataCentresToResize := r.checkDataCentresToResize(instClusterStatus.DataCentres, cluster.Spec.DataCentres)
	if len(dataCentresToResize) > 0 {
		err := r.resizeDataCentres(dataCentresToResize, cluster, logger)
		if err != nil {
			logger.Error(err, "Cannot resize Redis cluster data centres",
				"cluster name", cluster.Spec.Name,
				"new data centres", cluster.Spec.DataCentres,
				"old data centres", instClusterStatus.DataCentres,
			)

			return err
		}
	}

	return nil
}

func (r *RedisReconciler) resizeDataCentres(
	dataCentresToResize []*clustersv1alpha1.ResizedDataCentre,
	cluster *clustersv1alpha1.Redis,
	logger logr.Logger,
) error {
	for _, dataCentre := range dataCentresToResize {
		activeResizeOperations, err := r.getDataCentreOperations(cluster.Status.ID, dataCentre.DataCentreID)
		if err != nil {
			return err
		}
		if len(activeResizeOperations) > 0 {
			return nil
		}

		var currentNodeSizeIndex int
		var newNodeSizeIndex int
		switch dataCentre.Provider {
		case modelsv2.AWSVPC:
			currentNodeSizeIndex = modelsv1.RedisAWSNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.RedisAWSNodeTypes[dataCentre.NewNodeSize]
		case modelsv2.AZURE, modelsv2.AZUREAZ:
			currentNodeSizeIndex = modelsv1.RedisAzureNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.RedisAzureNodeTypes[dataCentre.NewNodeSize]
		case modelsv2.GCP:
			currentNodeSizeIndex = modelsv1.RedisGCPNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.RedisGCPNodeTypes[dataCentre.NewNodeSize]
		}

		if currentNodeSizeIndex > newNodeSizeIndex {
			return instaclustr.CannotDownsizeNode
		}

		resizeRequest := &models.ResizeRequest{
			NewNodeSize:           dataCentre.NewNodeSize,
			ConcurrentResizes:     cluster.Spec.ConcurrentResizes,
			NotifySupportContacts: cluster.Spec.NotifySupportContacts,
			ClusterID:             cluster.Status.ID,
			DataCentreID:          dataCentre.DataCentreID,
			NodePurpose:           modelsv1.RedisNodePurpose,
		}
		err = r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			if errors.Is(err, instaclustr.StatusPreconditionFailed) {
				logger.Info("Redis cluster is not ready to resize",
					"cluster name", cluster.Spec.Name,
					"cluster status", cluster.Status.Status,
					"reason", err,
				)

				return err
			}

			return err
		}

		logger.Info("Data centre resize request was sent",
			"cluster name", cluster.Spec.Name,
			"data centre ID", dataCentre.DataCentreID,
			"new node size", dataCentre.NewNodeSize,
		)
	}

	return nil
}

func (r *RedisReconciler) getDataCentreOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error) {
	activeResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(clusterID, dataCentreID)
	if err != nil {
		return nil, nil
	}

	return activeResizeOperations, nil
}

func (r *RedisReconciler) checkDataCentresToResize(dataCentresFromInst []*clustersv1alpha1.DataCentreStatus, dataCentres []*clustersv1alpha1.RedisDataCentre) []*clustersv1alpha1.ResizedDataCentre {
	if len(dataCentres) != len(dataCentresFromInst) {
		return nil
	}

	var resizedDataCentres []*clustersv1alpha1.ResizedDataCentre
	for i, instDataCentre := range dataCentresFromInst {
		if instDataCentre.Nodes[0].Size != dataCentres[i].NodeSize {
			resizedDataCentres = append(resizedDataCentres, &clustersv1alpha1.ResizedDataCentre{
				CurrentNodeSize: instDataCentre.Nodes[0].Size,
				NewNodeSize:     dataCentres[i].NodeSize,
				DataCentreID:    instDataCentre.ID,
				Provider:        dataCentres[i].CloudProvider,
			})
		}
	}

	return resizedDataCentres
}

func (r *RedisReconciler) updateDescriptionAndTwoFactorDelete(cluster *clustersv1alpha1.Redis) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
	if len(cluster.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = cluster.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, cluster.Status.ID, cluster.Spec.Description, twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}
