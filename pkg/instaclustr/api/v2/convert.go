package v2

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.ClusterStatus, error) {
	var clusterStatusFromInst modelsv2.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.ClusterStatus{
		ID:          clusterStatusFromInst.ID,
		Status:      clusterStatusFromInst.Status,
		DataCentres: dataCentres,
	}

	return clusterStatus, nil
}

func dataCentresFromInstAPI(instaDataCentres []*modelsv2.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPI(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:         dataCentre.ID,
			Status:     dataCentre.Status,
			Nodes:      nodes,
			NodeNumber: dataCentre.NumberOfNodes,
		})
	}

	return dataCentres
}

func nodesFromInstAPI(instaNodes []*modelsv2.NodeStatus) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			ID:             node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			Status:         node.Status,
			Size:           node.NodeSize,
		})
	}
	return nodes
}
