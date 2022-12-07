package instaclustr

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	models2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
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
	GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1alpha1.FirewallRuleStatus, error)
	CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1alpha1.FirewallRuleStatus, error)
	DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error
	GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1alpha1.KafkaUserStatus, error)
	CreateKafkaUser(url string, kafkaUser *models2.KafkaUserAPIv2) (*kafkamanagementv1alpha1.KafkaUserStatus, error)
	UpdateKafkaUser(kafkaUserID, kafkaUserEndpoint string, kafkaUserSpec *models2.KafkaUserAPIv2) error
	DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error
	CreateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	DeleteKafkaTopic(url, id string) error
	UpdateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	CreateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error
	DeleteKafkaMirror(url, id string) error
	UpdateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error
	GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error)
	TriggerClusterBackup(url, clusterID string) error
	CreateNodeReload(bundle, nodeID string, nr *modelsv1.NodeReload) error
	GetNodeReloadStatus(bundle, nodeID string) (*modelsv1.NodeReloadStatusAPIv1, error)
	CreateKafkaACL(url string, kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error
	UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error
}
