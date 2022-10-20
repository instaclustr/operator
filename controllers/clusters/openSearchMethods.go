package clusters

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *OpenSearchReconciler) patchClusterMetadata(
	ctx *context.Context,
	openSearchCluster *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(openSearchCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: clustersv1alpha1.ReplaceOperation,
		Path:      clustersv1alpha1.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(openSearchCluster.Finalizers)
	if err != nil {
		return err
	}

	finzlizersPatch := &clustersv1alpha1.PatchRequest{
		Operation: clustersv1alpha1.ReplaceOperation,
		Path:      clustersv1alpha1.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return err
	}

	err = r.Patch(*ctx, openSearchCluster, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("OpenSearch cluster patched",
		"Cluster name", openSearchCluster.Spec.Name,
		"Finalizers", openSearchCluster.Finalizers,
		"Annotations", openSearchCluster.Annotations,
	)
	return nil
}

func (r *OpenSearchReconciler) reconcileDataCentresNodeSize(
	openSearchInstClusterStatus *clustersv1alpha1.ClusterStatus,
	openSearchCluster *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
) error {
	dataCentresToResize := r.checkNodeSizeUpdate(openSearchInstClusterStatus.DataCentres,
		openSearchCluster.Spec.DataCentres)
	if len(dataCentresToResize) > 0 {
		if openSearchInstClusterStatus.Status != modelsv1.RunningStatus {
			return instaclustr.ClusterNotRunning
		}

		err := r.resizeDataCentres(dataCentresToResize, openSearchCluster, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *OpenSearchReconciler) checkNodeSizeUpdate(
	dataCentresFromInst []*clustersv1alpha1.DataCentreStatus,
	openSearchDataCentres []*clustersv1alpha1.OpenSearchDataCentre,
) []*clustersv1alpha1.ResizedDataCentre {
	var resizedDataCentres []*clustersv1alpha1.ResizedDataCentre
	for i, instDataCentre := range dataCentresFromInst {
		if instDataCentre.Nodes[0].Size != openSearchDataCentres[i].NodeSize {
			resizedDataCentres = append(resizedDataCentres, &clustersv1alpha1.ResizedDataCentre{
				CurrentNodeSize: instDataCentre.Nodes[0].Size,
				NewNodeSize:     openSearchDataCentres[i].NodeSize,
				DataCentreID:    instDataCentre.ID,
				Provider:        openSearchDataCentres[i].CloudProvider,
			})
		}
	}

	return resizedDataCentres
}

func (r *OpenSearchReconciler) resizeDataCentres(
	dataCentresToResize []*clustersv1alpha1.ResizedDataCentre,
	openSearchCluster *clustersv1alpha1.OpenSearch,
	logger *logr.Logger,
) error {
	for _, dataCentre := range dataCentresToResize {
		activeResizeOperations, err := r.getDataCentreOperations(openSearchCluster.Status.ID, dataCentre.DataCentreID)
		if err != nil {
			return err
		}
		if len(activeResizeOperations) > 0 {
			return nil
		}

		err = r.API.UpdateNodeSize(
			instaclustr.ClustersEndpointV1,
			openSearchCluster.Status.ID,
			dataCentre,
			openSearchCluster.Spec.ConcurrentResizes,
			openSearchCluster.Spec.NotifySupportContacts,
			"OPENSEARCH_COORDINATOR",
		)
		if err != nil {
			return err
		}

		logger.Info("Data centre resize request was sent",
			"Cluster name", openSearchCluster.Spec.Name,
			"Data centre id", dataCentre.DataCentreID,
			"New node size", dataCentre.NewNodeSize,
		)
	}

	return nil
}

func (r *OpenSearchReconciler) getDataCentreOperations(
	clusterID,
	dataCentreID string,
) ([]*modelsv1.DataCentreResizeOperations, error) {
	activeResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(clusterID, dataCentreID)
	if err != nil {
		return nil, nil
	}

	return activeResizeOperations, nil
}

func (r *OpenSearchReconciler) reconcileClusterConfigurations(
	clusterID,
	clusterStatus string,
	clusterConfigurations map[string]string,
	logger *logr.Logger,
) error {
	instClusterConfigurations, err := r.API.GetClusterConfigurations(instaclustr.ClustersEndpointV1,
		clusterID,
		modelsv1.OpenSearch)
	if err != client.IgnoreNotFound(err) {
		logger.Error(err, "cannot get cluster configurations")
		return err
	}

	if len(instClusterConfigurations) == 0 && len(clusterConfigurations) == 0 {
		return nil
	}

	clusterConfigurationsToUpdate, clusterConfigurationsToReset := r.checkClusterConfigurationsUpdate(instClusterConfigurations,
		clusterConfigurations)

	if len(clusterConfigurationsToUpdate) != 0 {
		if clusterStatus != modelsv1.RunningStatus {
			return instaclustr.ClusterNotRunning
		}
		for name, value := range clusterConfigurationsToUpdate {
			err = r.API.UpdateClusterConfiguration(instaclustr.ClustersEndpointV1,
				clusterID, modelsv1.OpenSearch,
				name,
				value)
			if err != nil {
				logger.Error(err, "cannot update cluster configurations")
				return err
			}
			logger.Info("Cluster configuration updated",
				"Cluster ID", clusterID,
				"Parameter name", name,
				"Parameter value", value,
			)
		}
	}

	if len(clusterConfigurationsToReset) != 0 {
		for _, name := range clusterConfigurationsToReset {
			err = r.API.ResetClusterConfiguration(instaclustr.ClustersEndpointV1, clusterID, modelsv1.OpenSearch, name)
			if err != nil {
				logger.Error(err, "cannot reset cluster configurations")
				return err
			}
			logger.Info("Cluster configuration was reset",
				"ClusterID", clusterID,
				"Parameter name", name,
			)
		}
	}

	return nil
}

func (r *OpenSearchReconciler) checkClusterConfigurationsUpdate(
	instClusterConfigurations map[string]string,
	clusterConfgirurations map[string]string,
) (map[string]string, []string) {
	var clusterConfigurationsToUpdate = make(map[string]string)
	var clusterConfigurationsToReset []string

	if len(instClusterConfigurations) != 0 {
		for name, value := range clusterConfgirurations {
			if instClusterConfigurations[name] != value {
				clusterConfigurationsToUpdate[name] = value
			}
		}

		for name, _ := range instClusterConfigurations {
			if clusterConfgirurations[name] == "" {
				clusterConfigurationsToReset = append(clusterConfigurationsToReset, name)
			}
		}
		return clusterConfigurationsToUpdate, clusterConfigurationsToReset
	}

	clusterConfigurationsToUpdate = clusterConfgirurations

	return clusterConfigurationsToUpdate, clusterConfigurationsToReset
}

func (r *OpenSearchReconciler) updateDescriptionAndTwoFactorDelete(
	openSearchCluster *clustersv1alpha1.OpenSearch) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
	if len(openSearchCluster.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = openSearchCluster.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1,
		openSearchCluster.Status.ID,
		openSearchCluster.Spec.Description,
		twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}
