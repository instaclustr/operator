package mock

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

type mockClient struct {
	*http.Client
}

func NewInstAPI() *mockClient {
	return &mockClient{}
}

const (
	StatusID = "statusID"
)

func (c *mockClient) CreateCluster(url string, clusterSpec any) (string, error) {
	return "", nil
}

func (c *mockClient) DoRequest(url string, method string, data []byte) (*http.Response, error) {
	panic("DoRequest: is not implemented")
}

func (c *mockClient) GetOpenSearch(id string) ([]byte, error) {
	panic("GetOpenSearch: is not implemented")
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
	fwRule := &clusterresourcesv1alpha1.FirewallRuleStatus{
		ID:             StatusID,
		Status:         "OK",
		DeferredReason: "NO",
	}
	return fwRule, nil
}

func (c *mockClient) CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1alpha1.FirewallRuleStatus, error) {
	fwRule := &clusterresourcesv1alpha1.FirewallRuleStatus{
		ID:             StatusID,
		Status:         "OK",
		DeferredReason: "NO",
	}
	return fwRule, nil
}

func (c *mockClient) DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error {
	panic("DeleteFirewallRule: is not implemented")
}

func (c *mockClient) GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	panic("GetKafkaUserStatus: is not implemented")
}

func (c *mockClient) CreateKafkaUser(url string, kafkaUser *models.KafkaUser) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	panic("CreateKafkaUser: is not implemented")
}

func (c *mockClient) UpdateKafkaUser(kafkaUserID string, kafkaUserSpec *models.KafkaUser) error {
	panic("UpdateKafkaUser: is not implemented")
}

func (c *mockClient) DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error {
	panic("DeleteKafkaUser: is not implemented")
}

func (c *mockClient) GetTopicStatus(id string) (*kafkamanagementv1alpha1.TopicStatus, error) {
	panic("GetTopicStatus: is not implemented")
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

func (c *mockClient) CreateNodeReload(nr *clusterresourcesv1alpha1.Node) error {
	return nil
}

func (c *mockClient) GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error) {
	return nil, nil
}

func (c *mockClient) CreateKafkaACL(url string, kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec) (*kafkamanagementv1alpha1.KafkaACLStatus, error) {
	kafkaACLStatus := &kafkamanagementv1alpha1.KafkaACLStatus{
		ID: StatusID,
	}
	return kafkaACLStatus, nil
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

func (c *mockClient) CreateExclusionWindow(clusterID string, window clusterresourcesv1alpha1.ExclusionWindowSpec) (string, error) {
	panic("CreateExclusionWindow: is not implemented")
}

func (c *mockClient) GetExclusionWindowsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.ExclusionWindowStatus, error) {
	panic("GetExclusionWindowsStatuses: is not implemented")
}

func (c *mockClient) GetMaintenanceEventsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.MaintenanceEventStatus, error) {
	panic("GetMaintenanceEventsStatuses: is not implemented")
}

func (c *mockClient) GetMaintenanceEvents(clusterID string) ([]*v1alpha1.MaintenanceEvent, error) {
	panic("GetMaintenanceEvents: is not implemented")
}

func (c *mockClient) DeleteExclusionWindow(id string) error {
	panic("DeleteExclusionWindow: is not implemented")
}

func (c *mockClient) UpdateMaintenanceEvent(me clusterresourcesv1alpha1.MaintenanceEventRescheduleSpec) (*clusterresourcesv1alpha1.MaintenanceEventStatus, error) {
	panic("UpdateMaintenanceEvent: is not implemented")
}

func (c *mockClient) RestorePgCluster(restoreData *v1alpha1.PgRestoreFrom) (string, error) {
	panic("RestorePgCluster: is not implemented")
}

func (c *mockClient) RestoreCassandra(restoreData v1alpha1.CassandraRestoreFrom) (string, error) {
	panic("RestoreCassandra: is not implemented")
}

func (c *mockClient) GetCassandra(id string) ([]byte, error) {
	panic("GetCassandra: is not implemented")
}

func (c *mockClient) GetRedis(id string) ([]byte, error) {
	panic("GetRedis: is not implemented")
}

func (c *mockClient) GetKafka(id string) ([]byte, error) {
	panic("GetKafka: is not implemented")
}

func (c *mockClient) GetKafkaConnect(id string) ([]byte, error) {
	panic("GetKafkaConnect: is not implemented")
}

func (c *mockClient) GetZookeeper(id string) ([]byte, error) {
	panic("GetZookeeper: is not implemented")
}

func (c *mockClient) RestoreRedisCluster(restoreData *v1alpha1.RedisRestoreFrom) (string, error) {
	panic("RestoreRedisCluster: is not implemented")
}

func (c *mockClient) RestoreOpenSearchCluster(restoreData *v1alpha1.OpenSearchRestoreFrom) (string, error) {
	panic("RestoreOpenSearchCluster: is not implemented")
}

func (c *mockClient) GetPostgreSQL(id string) ([]byte, error) {
	panic("GetPostgreSQL: is not implemented")
}

func (c *mockClient) UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error {
	panic("UpdatePostgreSQLDataCentres: is not implemented")
}

func (c *mockClient) GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error) {
	panic("GetPostgreSQLConfigs: is not implemented")
}

func (c *mockClient) UpdatePostgreSQLConfiguration(id, name, value string) error {
	panic("UpdatePostgreSQLConfiguration: is not implemented")
}

func (c *mockClient) CreatePostgreSQLConfiguration(id, name, value string) error {
	panic("CreatePostgreSQLConfiguration: is not implemented")
}

func (c *mockClient) ResetPostgreSQLConfiguration(id, name string) error {
	panic("ResetPostgreSQLConfiguration: is not implemented")
}

func (c *mockClient) GetCadence(id string) ([]byte, error) {
	panic("GetCadence: is not implemented")
}

func (c *mockClient) UpdatePostgreSQLDefaultUserPassword(id, password string) error {
	panic("UpdatePostgreSQLDefaultUserPassword: is not implemented")
}

func (c *mockClient) UpdateCassandra(id string, cassandra models.CassandraClusterAPIUpdate) error {
	panic("UpdateCassandra: is not implemented")
}

func (c *mockClient) UpdateKafkaConnect(id string, kc models.KafkaConnectAPIUpdate) error {
	panic("UpdateKafkaConnect: is not implemented")
}
