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
	"github.com/instaclustr/operator/pkg/models"
)

func (r *OpenSearchReconciler) patchClusterMetadata(
	ctx context.Context,
	openSearchCluster *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) error {
	patchRequest := []*clustersv1alpha1.PatchRequest{}

	annotationsPayload, err := json.Marshal(openSearchCluster.Annotations)
	if err != nil {
		return err
	}

	annotationsPatch := &clustersv1alpha1.PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(openSearchCluster.Finalizers)
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

	err = r.Patch(ctx, openSearchCluster, client.RawPatch(types.JSONPatchType, patchPayload))
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
	logger logr.Logger,
) error {
	dataCentreToResize := r.newDataCentreResize(openSearchInstClusterStatus.Options,
		openSearchCluster.Spec.DataCentres[0],
		openSearchInstClusterStatus.DataCentres[0].ID,
	)

	err := r.resizeDataCentres(dataCentreToResize, openSearchCluster, logger)
	if err != nil {
		return err
	}

	return nil
}

func (r *OpenSearchReconciler) newDataCentreResize(
	optionsFromInst *clustersv1alpha1.Options,
	dataCentre *clustersv1alpha1.OpenSearchDataCentre,
	dataCentreID string,
) *clustersv1alpha1.ResizedDataCentre {
	var resizedDataCentre *clustersv1alpha1.ResizedDataCentre
	resizedDataCentre = &clustersv1alpha1.ResizedDataCentre{
		DataCentreID: dataCentreID,
		Provider:     dataCentre.CloudProvider,
	}

	if optionsFromInst.DataNodeSize == "" &&
		optionsFromInst.MasterNodeSize != dataCentre.NodeSize {
		resizedDataCentre.MasterNewNodeSize = dataCentre.NodeSize
	}

	if optionsFromInst.OpenSearchDashboardsNodeSize != "" &&
		optionsFromInst.OpenSearchDashboardsNodeSize != dataCentre.OpenSearchDashboardsNodeSize {
		resizedDataCentre.DashboardNewNodeSize = dataCentre.OpenSearchDashboardsNodeSize
	}

	if optionsFromInst.DataNodeSize != "" {
		if optionsFromInst.DataNodeSize != dataCentre.NodeSize {
			resizedDataCentre.NewNodeSize = dataCentre.NodeSize
		}
		if optionsFromInst.MasterNodeSize != dataCentre.MasterNodeSize {
			resizedDataCentre.MasterNewNodeSize = dataCentre.MasterNodeSize
		}
	}

	return resizedDataCentre
}

func (r *OpenSearchReconciler) resizeDataCentres(
	dataCentreToResize *clustersv1alpha1.ResizedDataCentre,
	openSearchCluster *clustersv1alpha1.OpenSearch,
	logger logr.Logger,
) error {
	activeResizeOperations, err := r.getDataCentreOperations(openSearchCluster.Status.ID, dataCentreToResize.DataCentreID)
	if err != nil {
		return err
	}
	if len(activeResizeOperations) > 0 {
		return instaclustr.HasActiveResizeOperation
	}

	resizeRequest := &models.ResizeRequest{
		NewNodeSize:           openSearchCluster.Spec.DataCentres[0].NodeSize,
		ConcurrentResizes:     openSearchCluster.Spec.ConcurrentResizes,
		NotifySupportContacts: openSearchCluster.Spec.NotifySupportContacts,
		ClusterID:             openSearchCluster.Status.ID,
		DataCentreID:          openSearchCluster.Status.CDCID,
	}

	if dataCentreToResize.NewNodeSize != "" {
		resizeRequest.NodePurpose = modelsv1.OpenSearchDataNodePurpose
		err = r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Data nodes resize request was sent",
			"Cluster name", openSearchCluster.Spec.Name,
			"Data centre id", dataCentreToResize.DataCentreID,
			"New node size", dataCentreToResize.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	if dataCentreToResize.MasterNewNodeSize != "" {
		resizeRequest.NodePurpose = modelsv1.OpenSearchMasterNodePurpose
		err = r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Master nodes resize request was sent",
			"Cluster name", openSearchCluster.Spec.Name,
			"Data centre id", dataCentreToResize.DataCentreID,
			"New node size", dataCentreToResize.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	if dataCentreToResize.DashboardNewNodeSize != "" {
		resizeRequest.NodePurpose = modelsv1.OpenSearchDashBoardsNodePurpose
		err = r.API.UpdateNodeSize(instaclustr.ClustersEndpointV1, resizeRequest)
		if err != nil {
			return err
		}

		logger.Info("Dashboard nodes resize request was sent",
			"Cluster name", openSearchCluster.Spec.Name,
			"Data centre id", dataCentreToResize.DataCentreID,
			"New node size", dataCentreToResize.NewNodeSize,
		)

		return instaclustr.HasActiveResizeOperation
	}

	return nil
}

func (r *OpenSearchReconciler) getDataCentreOperations(
	clusterID,
	dataCentreID string,
) ([]*models.DataCentreResizeOperations, error) {
	activeResizeOperations, err := r.API.GetActiveDataCentreResizeOperations(clusterID, dataCentreID)
	if err != nil {
		return nil, err
	}

	return activeResizeOperations, nil
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
