package instaclustr

import (
	clusterresourcesv2alpha1 "github.com/instaclustr/operator/apis/clusterresources/v2alpha1"
	"net/http"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
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
	GetCassandraDCs(id, clusterEndpoint string) (*modelsv2.CassandraDCs, error)
	UpdateCassandraCluster(id, clusterEndpoint string, InstaDCs *modelsv2.CassandraDCs) error
	DeleteCluster(id, clusterEndpoint string) error
	AddDataCentre(id, clusterEndpoint string, dataCentre any) error
	CreateGCPPeering(url string, GCPSpec *clusterresourcesv2alpha1.GCPVPCPeeringSpec) (*clusterresourcesv2alpha1.GCPVPCPeeringStatus, error)
	GetGCPPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv2alpha1.GCPVPCPeeringStatus, error)
	DeleteGCPPeering(peerID, peeringEndpoint string) error
}
