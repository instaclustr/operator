package instaclustr

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
)

type API interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetClusterStatus(id, clusterEndpoint string) (*v1alpha1.ClusterStatus, error)
	UpdateNodeSize(clusterEndpoint, clusterID string, resizedDataCentre *v1alpha1.ResizedDataCentre, concurrentResizes int, notifySupportContacts bool, nodePurpose string) error
	GetActiveDataCentreResizeOperations(clusterID, dataCentreID string) ([]*modelsv1.DataCentreResizeOperations, error)
	GetClusterConfigurations(clusterEndpoint, clusterID, bundle string) (map[string]string, error)
	UpdateClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName, paramValue string) error
	ResetClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName string) error
	UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1alpha1.TwoFactorDelete) error
	UpdateCluster(id, clusterEndpoint string, InstaDCs any) error
	DeleteCluster(id, clusterEndpoint string) error
	AddDataCentre(id, clusterEndpoint string, dataCentre any) error
	GetAWSPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1alpha1.AWSVPCPeeringStatus, error)
	UpdateAWSPeering(peerID, peeringEndpoint string, peerSpec *clusterresourcesv1alpha1.AWSVPCPeeringSpec) error
	DeleteAWSPeering(peerID, peeringEndpoint string) error
	CreateAWSPeering(url string, AWSSpec *clusterresourcesv1alpha1.AWSVPCPeeringSpec) (*clusterresourcesv1alpha1.AWSVPCPeeringStatus, error)
}
