package mock

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	models2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

type mockClient struct {
	*http.Client
}

func NewInstAPI() *mockClient {
	return &mockClient{}
}

const (
	CreatedID = "created"
	StatusID  = "statusID"
)

func (c *mockClient) CreateCluster(url string, clusterSpec any) (string, error) {
	return CreatedID, nil
}

func (c *mockClient) DoRequest(url string, method string, data []byte) (*http.Response, error) {
	panic("DoRequest: is not implemented")
}

func (c *mockClient) GetClusterStatus(id, clusterEndpoint string) (*v1alpha1.ClusterStatus, error) {
	panic("GetClusterStatus: is not implemented")
}
func (c *mockClient) UpdateNodeSize(clusterEndpoint string, resizeRequest *models.ResizeRequest) error {
	panic("UpdateNodeSize: is not implemented")
}
func (c *mockClient) GetActiveDataCentreResizeOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error) {
	panic("GetActiveDataCentreResizeOperations: is not implemented")
}
func (c *mockClient) GetClusterConfigurations(clusterEndpoint, clusterID, bundle string) (map[string]string, error) {
	panic("GetClusterConfigurations: is not implemented")
}
func (c *mockClient) UpdateClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName, paramValue string) error {
	panic("UpdateClusterConfiguration: is not implemented")
}
func (c *mockClient) ResetClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName string) error {
	panic("ResetClusterConfiguration: is not implemented")
}
func (c *mockClient) UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1alpha1.TwoFactorDelete) error {
	panic("UpdateDescriptionAndTwoFactorDelete: is not implemented")
}
func (c *mockClient) UpdateCluster(id, clusterEndpoint string, InstaDCs any) error {
	panic("UpdateCluster: is not implemented")
}
func (c *mockClient) DeleteCluster(id, clusterEndpoint string) error {
	panic("DeleteCluster: is not implemented")
}
func (c *mockClient) AddDataCentre(id, clusterEndpoint string, dataCentre any) error {
	panic("AddDataCentre: is not implemented")
}
func (c *mockClient) GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1alpha1.PeeringStatus, error) {
	panic("GetPeeringStatus: is not implemented")
}
func (c *mockClient) UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error {
	panic("UpdatePeering: is not implemented")
}
func (c *mockClient) DeletePeering(peerID, peeringEndpoint string) error {
	panic("DeletePeering: is not implemented")
}
func (c *mockClient) CreatePeering(url string, peeringSpec any) (*clusterresourcesv1alpha1.PeeringStatus, error) {
	ps := &clusterresourcesv1alpha1.PeeringStatus{
		ID:            StatusID,
		Name:          "name",
		StatusCode:    "statusCode",
		FailureReason: "failureReason",
	}
	return ps, nil
}
func (c *mockClient) GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1alpha1.FirewallRuleStatus, error) {
	panic("GetFirewallRuleStatus: is not implemented")
}
func (c *mockClient) CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1alpha1.FirewallRuleStatus, error) {
	panic("CreateFirewallRule: is not implemented")
}
func (c *mockClient) DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error {
	panic("DeleteFirewallRule: is not implemented")
}
func (c *mockClient) GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	panic("GetKafkaUserStatus: is not implemented")
}
func (c *mockClient) CreateKafkaUser(url string, kafkaUser *models2.KafkaUserAPIv2) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	panic("CreateKafkaUser: is not implemented")
}
func (c *mockClient) UpdateKafkaUser(kafkaUserID, kafkaUserEndpoint string, kafkaUserSpec *models2.KafkaUserAPIv2) error {
	panic("UpdateKafkaUser: is not implemented")
}
func (c *mockClient) DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error {
	panic("DeleteKafkaUser: is not implemented")
}
func (c *mockClient) CreateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error {
	panic("CreateKafkaTopic: is not implemented")
}
func (c *mockClient) DeleteKafkaTopic(url, id string) error {
	panic("DeleteKafkaTopic: is not implemented")
}
func (c *mockClient) UpdateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error {
	panic("UpdateKafkaTopic: is not implemented")
}
func (c *mockClient) CreateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error {
	panic("CreateKafkaMirror: is not implemented")
}
func (c *mockClient) DeleteKafkaMirror(url, id string) error {
	panic("DeleteKafkaMirror: is not implemented")
}
func (c *mockClient) UpdateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error {
	panic("UpdateKafkaMirror: is not implemented")
}
func (c *mockClient) GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error) {
	panic("GetClusterBackups: is not implemented")
}
func (c *mockClient) TriggerClusterBackup(url, clusterID string) error {
	panic("TriggerClusterBackup: is not implemented")
}

func (c *mockClient) CreateNodeReload(bundle, nodeID string, nr *modelsv1.NodeReload) error {
	panic("CreateNodeReload: is not implemented")
}

func (c *mockClient) GetNodeReloadStatus(bundle, nodeID string) (*modelsv1.NodeReloadStatusAPIv1, error) {
	panic("GetNodeReloadStatus: is not implemented")
}

func (c *mockClient) CreateKafkaACL(url string, kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec) (*kafkamanagementv1alpha1.KafkaACLStatus, error) {
	panic("CreateKafkaACL: is not implemented")
}
func (c *mockClient) GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1alpha1.KafkaACLStatus, error) {
	panic("GetKafkaACLStatus: is not implemented")
}
func (c *mockClient) DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error {
	panic("DeleteKafkaACL: is not implemented")
}
func (c *mockClient) UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error {
	panic("UpdateKafkaACL: is not implemented")
}

func (c *mockClient) GetMirrorStatus(id, mirrorEndpoint string) (*kafkamanagementv1alpha1.MirrorStatus, error) {
	panic("GetMirrorStatus: is not implemented")
}
