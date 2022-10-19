package clusters

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

func (r *RedisReconciler) ToInstAPIv1(redisSpec *clustersv1alpha1.RedisSpec) *modelsv1.RedisCluster {
	redisBundles := r.bundlesToInstAPIv1(redisSpec.DataCentres[0], redisSpec.Version)

	redisInstProvider := r.providerToInstAPI(redisSpec.DataCentres[0])

	var redisTwoFactorDelete *modelsv1.TwoFactorDelete
	if len(redisSpec.TwoFactorDelete) != 0 {
		redisTwoFactorDelete = r.twoFactorDeleteToInstAPI(redisSpec.TwoFactorDelete)
	}

	redisInstCluster := &modelsv1.RedisCluster{
		Cluster: modelsv1.Cluster{
			ClusterName:           redisSpec.Name,
			NodeSize:              redisSpec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: redisSpec.PrivateNetworkCluster,
			SLATier:               redisSpec.SLATier,
			Provider:              redisInstProvider,
			TwoFactorDelete:       redisTwoFactorDelete,
		},
		Bundles: redisBundles,
	}

	redisRackAllocation := &modelsv1.RackAllocation{
		NodesPerRack:  redisSpec.DataCentres[0].NodesNumber,
		NumberOfRacks: redisSpec.DataCentres[0].RacksNumber,
	}

	redisInstCluster.DataCentre = redisSpec.DataCentres[0].Region
	redisInstCluster.DataCentreCustomName = redisSpec.DataCentres[0].Name
	redisInstCluster.NodeSize = redisSpec.DataCentres[0].NodeSize
	redisInstCluster.ClusterNetwork = redisSpec.DataCentres[0].Network
	redisInstCluster.RackAllocation = redisRackAllocation

	return redisInstCluster

	// Can be used in APIv2 if it supports multiple DC for Redis
	//
	//var redisInstDCs []*modelsv1.RedisDataCentre
	//for _, dataCentre := range redisSpec.DataCentres {
	//	redisBundles = r.bundlesToInstAPIv1(dataCentre, redisSpec.Version)
	//
	//	redisInstProvider = r.providerToInstAPI(dataCentre)
	//
	//	redisRackAlloc := &modelsv1.RackAllocation{
	//		NodesPerRack:  dataCentre.NodesNumber,
	//		NumberOfRacks: dataCentre.RacksNumber,
	//	}
	//
	//	redisInstDC := &modelsv1.RedisDataCentre{
	//		DataCentre: modelsv1.DataCentre{
	//			Name:           dataCentre.Name,
	//			DataCentre:     dataCentre.Region,
	//			Network:        dataCentre.Network,
	//			Provider:       redisInstProvider,
	//			NodeSize:       dataCentre.NodeSize,
	//			RackAllocation: redisRackAlloc,
	//		},
	//		Bundles: redisBundles,
	//	}
	//
	//	redisInstDCs = append(redisInstDCs, redisInstDC)
	//}
	//
	//redisInstCluster.DataCentres = redisInstDCs
	//
	//return redisInstCluster
}

func (r *RedisReconciler) bundlesToInstAPIv1(dataCentre *clustersv1alpha1.RedisDataCentre, version string) []*modelsv1.RedisBundle {
	var redisBundles []*modelsv1.RedisBundle

	redisBundle := &modelsv1.RedisBundle{
		Bundle: modelsv1.Bundle{
			Bundle:  modelsv1.Redis,
			Version: version,
		},
		Options: &modelsv1.RedisOptions{
			ClientEncryption: dataCentre.ClientEncryption,
			MasterNodes:      dataCentre.MasterNodes,
			ReplicaNodes:     dataCentre.ReplicaNodes,
			PasswordAuth:     dataCentre.PasswordAuth,
		},
	}
	redisBundles = append(redisBundles, redisBundle)

	return redisBundles
}

func (r *RedisReconciler) providerToInstAPI(dataCentre *clustersv1alpha1.RedisDataCentre) *modelsv1.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dataCentre.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dataCentre.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dataCentre.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dataCentre.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &modelsv1.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		Tags:                   dataCentre.Tags,
		CustomVirtualNetworkId: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}

func (r *RedisReconciler) twoFactorDeleteToInstAPI(twoFactorDelete []*clustersv1alpha1.TwoFactorDelete) *modelsv1.TwoFactorDelete {
	return &modelsv1.TwoFactorDelete{
		DeleteVerifyEmail: twoFactorDelete[0].Email,
		DeleteVerifyPhone: twoFactorDelete[0].Phone,
	}
}

func (r *RedisReconciler) patchClusterMetadata(
	ctx *context.Context,
	redisCluster *clustersv1alpha1.Redis,
	logger *logr.Logger,
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

	finzlizersPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(*ctx, redisCluster, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("Redis cluster patched",
		"Cluster name", redisCluster.Spec.Name,
		"Finalizers", redisCluster.Finalizers,
		"Annotations", redisCluster.Annotations,
	)
	return nil
}
