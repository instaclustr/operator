package clusters

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

func (r *PostgreSQLReconciler) patchClusterMetadata(
	ctx *context.Context,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(pgCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(pgCluster.Finalizers)
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

	err = r.Patch(*ctx, pgCluster, client.RawPatch(types.JSONPatchType, patchPayload))
	if err != nil {
		return err
	}

	logger.Info("PostgreSQL cluster patched",
		"Cluster name", pgCluster.Spec.Name,
		"Finalizers", pgCluster.Finalizers,
		"Annotations", pgCluster.Annotations,
	)
	return nil
}

func (r *PostgreSQLReconciler) reconcileDataCentresNodeSize(
	pgInstClusterStatus *clustersv1alpha1.ClusterStatus,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) error {
	dataCentresToResize := r.checkNodeSizeUpdate(pgInstClusterStatus.DataCentres, pgCluster.Spec.DataCentres)
	if len(dataCentresToResize) > 0 {
		if pgInstClusterStatus.Status != models.RunningStatus {
			return instaclustr.ClusterNotRunning
		}

		err := r.resizeDataCentres(dataCentresToResize, pgCluster, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgreSQLReconciler) checkNodeSizeUpdate(dataCentresFromInst []*clustersv1alpha1.DataCentreStatus, pgDataCentres []*clustersv1alpha1.PgDataCentre) []*clustersv1alpha1.ResizedDataCentre {
	var resizedDataCentres []*clustersv1alpha1.ResizedDataCentre
	for i, instDataCentre := range dataCentresFromInst {
		if instDataCentre.Nodes[0].Size != pgDataCentres[i].NodeSize {
			resizedDataCentres = append(resizedDataCentres, &clustersv1alpha1.ResizedDataCentre{
				CurrentNodeSize: instDataCentre.Nodes[0].Size,
				NewNodeSize:     pgDataCentres[i].NodeSize,
				DataCentreID:    instDataCentre.ID,
				Provider:        pgDataCentres[i].CloudProvider,
			})
		}
	}

	return resizedDataCentres
}

func (r *PostgreSQLReconciler) resizeDataCentres(
	dataCentresToResize []*clustersv1alpha1.ResizedDataCentre,
	pgCluster *clustersv1alpha1.PostgreSQL,
	logger *logr.Logger,
) error {
	for _, dataCentre := range dataCentresToResize {
		activeResizeOperations, err := r.getDataCentreOperations(pgCluster.Status.ID, dataCentre.DataCentreID)
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
			currentNodeSizeIndex = modelsv1.PgAWSNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.PgAWSNodeTypes[dataCentre.NewNodeSize]
		case modelsv2.AZUREAZ:
			currentNodeSizeIndex = modelsv1.PgAzureNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.PgAzureNodeTypes[dataCentre.NewNodeSize]
		case modelsv2.GCP:
			currentNodeSizeIndex = modelsv1.PgGCPNodeTypes[dataCentre.CurrentNodeSize]
			newNodeSizeIndex = modelsv1.PgGCPNodeTypes[dataCentre.NewNodeSize]
		}

		if currentNodeSizeIndex > newNodeSizeIndex {
			return instaclustr.IncorrectNodeSize
		}

		err = r.API.UpdateNodeSize(
			instaclustr.ClustersEndpointV1,
			pgCluster.Status.ID,
			dataCentre,
			pgCluster.Spec.ConcurrentResizes,
			pgCluster.Spec.NotifySupportContacts,
			modelsv1.PgNodePurpose,
		)
		if err != nil {
			return err
		}

		logger.Info("Data centre resize request was sent",
			"Cluster name", pgCluster.Spec.Name,
			"Data centre id", dataCentre.DataCentreID,
			"New node size", dataCentre.NewNodeSize,
		)
	}

	return nil
}

func (r *PostgreSQLReconciler) getDataCentreOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error) {
	activeResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(clusterID, dataCentreID)
	if err != nil {
		return nil, nil
	}

	return activeResizeOperations, nil
}

func (r *PostgreSQLReconciler) reconcileClusterConfigurations(
	clusterID,
	clusterStatus string,
	clusterConfigurations map[string]string,
	logger *logr.Logger,
) error {
	instClusterConfigurations, err := r.API.GetClusterConfigurations(instaclustr.ClustersEndpointV1, clusterID, modelsv1.PgSQL)
	if err != client.IgnoreNotFound(err) {
		logger.Error(err, "cannot get cluster configurations")
		return err
	}

	if len(instClusterConfigurations) == 0 && len(clusterConfigurations) == 0 {
		return nil
	}

	clusterConfigurationsToUpdate, clusterConfigurationsToReset := r.checkClusterConfigurationsUpdate(instClusterConfigurations, clusterConfigurations)

	if len(clusterConfigurationsToUpdate) != 0 {
		if clusterStatus != models.RunningStatus {
			return instaclustr.ClusterNotRunning
		}
		for name, value := range clusterConfigurationsToUpdate {
			err = r.API.UpdateClusterConfiguration(instaclustr.ClustersEndpointV1, clusterID, modelsv1.PgSQL, name, value)
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
			err = r.API.ResetClusterConfiguration(instaclustr.ClustersEndpointV1, clusterID, modelsv1.PgSQL, name)
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

func (r *PostgreSQLReconciler) checkClusterConfigurationsUpdate(instClusterConfigurations map[string]string, clusterConfgirurations map[string]string) (map[string]string, []string) {
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

func (r *PostgreSQLReconciler) updateDescriptionAndTwoFactorDelete(pgCluster *clustersv1alpha1.PostgreSQL) error {
	var twoFactorDelete *clustersv1alpha1.TwoFactorDelete
	if len(pgCluster.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = pgCluster.Spec.TwoFactorDelete[0]
	}

	err := r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, pgCluster.Status.ID, pgCluster.Spec.Description, twoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}
