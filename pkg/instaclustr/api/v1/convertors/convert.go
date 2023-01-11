package convertors

import (
	"encoding/json"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.ClusterStatus, error) {
	var clusterStatusFromInst models.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.ClusterStatus{
		ID:                     clusterStatusFromInst.ID,
		Status:                 clusterStatusFromInst.ClusterStatus,
		DataCentres:            dataCentres,
		CDCID:                  clusterStatusFromInst.CDCID,
		TwoFactorDeleteEnabled: clusterStatusFromInst.TwoFactorDelete,
		Options: &v1alpha1.Options{
			DataNodeSize:                 clusterStatusFromInst.BundleOptions.DataNodeSize,
			MasterNodeSize:               clusterStatusFromInst.BundleOptions.MasterNodeSize,
			OpenSearchDashboardsNodeSize: clusterStatusFromInst.BundleOptions.OpenSearchDashboardsNodeSize,
		},
	}

	return clusterStatus, nil
}

func dataCentresFromInstAPI(instaDataCentres []*models.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus

	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPI(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:              dataCentre.ID,
			Status:          dataCentre.CDCStatus,
			Nodes:           nodes,
			NodeNumber:      dataCentre.NodeCount,
			EncryptionKeyID: dataCentre.EncryptionKeyID,
		})
	}

	return dataCentres
}

func nodesFromInstAPI(instaNodes []*models.NodeStatus) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			ID:             node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			Status:         node.NodeStatus,
			Size:           node.Size,
		})
	}
	return nodes
}

func twoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) *models.TwoFactorDelete {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	return &models.TwoFactorDelete{
		DeleteVerifyEmail: twoFactorDelete[0].Email,
		DeleteVerifyPhone: twoFactorDelete[0].Phone,
	}
}

func NodeReloadStatusFromInstAPI(nodeInProgress clusterresourcesv1alpha1.Node, nrStatus *modelsv1.NodeReloadStatusAPIv1) clusterresourcesv1alpha1.NodeReloadStatus {
	var nrOperations []*clusterresourcesv1alpha1.Operation
	for _, nro := range nrStatus.Operations {
		nrOperation := &clusterresourcesv1alpha1.Operation{
			TimeCreated:  nro.TimeCreated,
			TimeModified: nro.TimeModified,
			Status:       nro.Status,
			Message:      nro.Message,
		}
		nrOperations = append(nrOperations, nrOperation)
	}
	nodeReloadStatus := clusterresourcesv1alpha1.NodeReloadStatus{
		NodeInProgress:         nodeInProgress,
		CurrentOperationStatus: nrOperations,
	}
	return nodeReloadStatus
}
