package instaclustr

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

type API interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetClusterStatus(id, clusterEndpoint string) (*v1alpha1.ClusterStatus, error)
	UpdateNodeSize(clusterEndpoint string, resizeRequest *models.ResizeRequest) error
	GetActiveDataCentreResizeOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error)
	GetClusterConfigurations(clusterEndpoint, clusterID, bundle string) (map[string]string, error)
	UpdateClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName, paramValue string) error
	ResetClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName string) error
	UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1alpha1.TwoFactorDelete) error
	UpdateCluster(id, clusterEndpoint string, InstaDCs any) error
	DeleteCluster(id, clusterEndpoint string) error
	AddDataCentre(id, clusterEndpoint string, dataCentre any) error
	GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1alpha1.PeeringStatus, error)
	UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error
	DeletePeering(peerID, peeringEndpoint string) error
	CreatePeering(url string, peeringSpec any) (*clusterresourcesv1alpha1.PeeringStatus, error)
	GetFirewallRuleStatus(firewallRuleId string) (*clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus, error)
	CreateFirewallRule(url string, firewallRuleSpec *clusterresourcesv1alpha1.ClusterNetworkFirewallRuleSpec) (*clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus, error)
}
