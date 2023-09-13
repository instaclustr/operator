/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mock

import (
	"net/http"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	clustersv1beta1 "github.com/instaclustr/operator/apis/clusters/v1beta1"
	kafkamanagementv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
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

func (c *mockClient) UpdateOpenSearch(id string, o models.OpenSearchInstAPIUpdateRequest) error {
	panic("UpdateOpenSearch: is not implemented")
}

func (c *mockClient) UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *clustersv1beta1.TwoFactorDelete) error {
	panic("UpdateDescriptionAndTwoFactorDelete: is not implemented")
}

func (c *mockClient) UpdateCluster(id, clusterEndpoint string, InstaDCs any) error {
	panic("UpdateCluster: is not implemented")
}

func (c *mockClient) DeleteCluster(id, clusterEndpoint string) error {
	panic("DeleteCluster: is not implemented")
}

func (c *mockClient) GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1beta1.PeeringStatus, error) {
	panic("GetPeeringStatus: is not implemented")
}

func (c *mockClient) UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error {
	panic("UpdatePeering: is not implemented")
}

func (c *mockClient) DeletePeering(peerID, peeringEndpoint string) error {
	panic("DeletePeering: is not implemented")
}

func (c *mockClient) CreatePeering(url string, peeringSpec any) (*clusterresourcesv1beta1.PeeringStatus, error) {
	ps := &clusterresourcesv1beta1.PeeringStatus{
		ID:            StatusID,
		Name:          "name",
		StatusCode:    "statusCode",
		FailureReason: "failureReason",
	}
	return ps, nil
}

func (c *mockClient) GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1beta1.FirewallRuleStatus, error) {
	fwRule := &clusterresourcesv1beta1.FirewallRuleStatus{
		ID:             StatusID,
		Status:         "OK",
		DeferredReason: "NO",
	}
	return fwRule, nil
}

func (c *mockClient) CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1beta1.FirewallRuleStatus, error) {
	fwRule := &clusterresourcesv1beta1.FirewallRuleStatus{
		ID:             StatusID,
		Status:         "OK",
		DeferredReason: "NO",
	}
	return fwRule, nil
}

func (c *mockClient) DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error {
	panic("DeleteFirewallRule: is not implemented")
}

func (c *mockClient) GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1beta1.KafkaUserStatus, error) {
	panic("GetKafkaUserStatus: is not implemented")
}

func (c *mockClient) CreateKafkaUser(url string, kafkaUser *models.KafkaUser) (*kafkamanagementv1beta1.KafkaUserStatus, error) {
	panic("CreateKafkaUser: is not implemented")
}

func (c *mockClient) CreateKafkaUserCertificate(certRequest *models.CertificateRequest) (*kafkamanagementv1beta1.Certificate, error) {
	panic("CreateKafkaUserCertificate: is not implemented")
}

func (c *mockClient) DeleteKafkaUserCertificate(certificateID string) error {
	panic("DeleteKafkaUserCertificate: is not implemented")
}

func (c *mockClient) RenewKafkaUserCertificate(certificateID string) (*kafkamanagementv1beta1.Certificate, error) {
	panic("RenewKafkaUserCertificate: is not implemented")
}

func (c *mockClient) UpdateKafkaUser(kafkaUserID string, kafkaUserSpec *models.KafkaUser) error {
	panic("UpdateKafkaUser: is not implemented")
}

func (c *mockClient) DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error {
	panic("DeleteKafkaUser: is not implemented")
}

func (c *mockClient) GetTopicStatus(id string) ([]byte, error) {
	panic("GetTopicStatus: is not implemented")
}

func (c *mockClient) CreateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error {
	panic("CreateKafkaTopic: is not implemented")
}

func (c *mockClient) DeleteKafkaTopic(url, id string) error {
	panic("DeleteKafkaTopic: is not implemented")
}

func (c *mockClient) UpdateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error {
	panic("UpdateKafkaTopic: is not implemented")
}

func (c *mockClient) CreateKafkaMirror(m *kafkamanagementv1beta1.MirrorSpec) (*kafkamanagementv1beta1.MirrorStatus, error) {
	panic("CreateKafkaMirror: is not implemented")
}

func (c *mockClient) DeleteKafkaMirror(id string) error {
	panic("DeleteKafkaMirror: is not implemented")
}

func (c *mockClient) UpdateKafkaMirror(id string, latency int32) error {
	panic("UpdateKafkaMirror: is not implemented")
}

func (c *mockClient) GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error) {
	panic("GetClusterBackups: is not implemented")
}

func (c *mockClient) TriggerClusterBackup(url, clusterID string) error {
	panic("TriggerClusterBackup: is not implemented")
}

func (c *mockClient) CreateNodeReload(nr *clusterresourcesv1beta1.Node) error {
	return nil
}

func (c *mockClient) GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error) {
	return nil, nil
}

func (c *mockClient) CreateKafkaACL(url string, kafkaACL *kafkamanagementv1beta1.KafkaACLSpec) (*kafkamanagementv1beta1.KafkaACLStatus, error) {
	kafkaACLStatus := &kafkamanagementv1beta1.KafkaACLStatus{
		ID: StatusID,
	}
	return kafkaACLStatus, nil
}

func (c *mockClient) GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1beta1.KafkaACLStatus, error) {
	panic("GetKafkaACLStatus: is not implemented")
}

func (c *mockClient) DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error {
	panic("DeleteKafkaACL: is not implemented")
}

func (c *mockClient) UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error {
	panic("UpdateKafkaACL: is not implemented")
}

func (c *mockClient) GetMirrorStatus(id string) (*kafkamanagementv1beta1.MirrorStatus, error) {
	panic("GetMirrorStatus: is not implemented")
}

func (c *mockClient) CreateExclusionWindow(clusterID string, window *clusterresourcesv1beta1.ExclusionWindowSpec) (string, error) {
	panic("CreateExclusionWindow: is not implemented")
}

func (c *mockClient) GetExclusionWindowsStatus(windowID string) (string, error) {
	panic("GetExclusionWindowsStatus: is not implemented")
}

func (c *mockClient) DeleteExclusionWindow(id string) error {
	panic("DeleteExclusionWindow: is not implemented")
}

func (c *mockClient) GetMaintenanceEvents(clusterID, eventType string) ([]*clusterresourcesv1beta1.MaintenanceEventStatus, error) {
	panic("GetMaintenanceEvents: is not implemented")
}

func (c *mockClient) FetchMaintenanceEventStatuses(clusterID string) ([]*clusterresourcesv1beta1.ClusteredMaintenanceEventStatus, error) {
	panic("FetchMaintenanceEventStatuses: is not implemented")
}

func (c *mockClient) RescheduleMaintenanceEvent(me *clusterresourcesv1beta1.MaintenanceEventReschedule) error {
	panic("RescheduleMaintenanceEvent: is not implemented")
}

func (c *mockClient) RestorePgCluster(restoreData *clustersv1beta1.PgRestoreFrom) (string, error) {
	panic("RestorePgCluster: is not implemented")
}

func (c *mockClient) RestoreCassandra(restoreData clustersv1beta1.CassandraRestoreFrom) (string, error) {
	panic("RestoreCassandra: is not implemented")
}

func (c *mockClient) GetCassandra(id string) ([]byte, error) {
	panic("GetCassandra: is not implemented")
}

func (c *mockClient) GetRedis(id string) ([]byte, error) {
	panic("GetRedis: is not implemented")
}

func (c *mockClient) UpdateRedis(id string, r *models.RedisDataCentreUpdate) error {
	panic("UpdateRedis: is not implemented")
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

func (c *mockClient) RestoreRedisCluster(restoreData *clustersv1beta1.RedisRestoreFrom) (string, error) {
	panic("RestoreRedisCluster: is not implemented")
}

func (c *mockClient) RestoreOpenSearchCluster(restoreData *clustersv1beta1.OpenSearchRestoreFrom) (string, error) {
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

func (c *mockClient) ListClusters() ([]*models.ActiveClusters, error) {
	panic("ListClusters: is not implemented")
}

func (c *mockClient) CreateRedisUser(user *models.RedisUser) (string, error) {
	panic("CreateRedisUser: is not implemented")
}

func (c *mockClient) UpdateRedisUser(user *models.RedisUserUpdate) error {
	panic("UpdateRedisUser: is not implemented")
}

func (c *mockClient) DeleteRedisUser(id string) error {
	panic("DeleteRedisUser: is not implemented")
}

func (c *mockClient) CreateEncryptionKey(encryptionKeySpec any) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error) {
	encryptionKey := &clusterresourcesv1beta1.AWSEncryptionKeyStatus{
		ID:    StatusID,
		InUse: false,
	}
	return encryptionKey, nil
}

func (c *mockClient) GetEncryptionKeyStatus(encryptionKeyID string, encryptionKeyEndpoint string) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error) {
	encryptionKey := &clusterresourcesv1beta1.AWSEncryptionKeyStatus{
		ID:    StatusID,
		InUse: false,
	}
	return encryptionKey, nil
}

func (c *mockClient) DeleteEncryptionKey(encryptionKeyID string) error {
	return nil
}

func (c *mockClient) ListAppVersions(app string) ([]*models.AppVersions, error) {
	panic("ListAppVersions: is not implemented")
}

func (c *mockClient) CreateUser(userSpec any, clusterID, app string) error {
	panic("CreateUser: is not implemented")
}

func (c *mockClient) DeleteUser(username, clusterID, app string) error {
	panic("DeleteUser: is not implemented")
}

func (c *mockClient) GetDefaultCredentialsV1(clusterID string) (string, string, error) {
	panic("GetDefaultCredentialsV1: is not implemented")
}

func (c *mockClient) UpdateClusterSettings(clusterID string, settings *models.ClusterSettings) error {
	panic("UpdateClusterSettings: is not implemented")
}

func (c *mockClient) CreateAWSEndpointServicePrincipal(spec any) ([]byte, error) {
	panic("CreateAWSEndpointServicePrincipal: is not implemented")
}

func (c *mockClient) DeleteAWSEndpointServicePrincipal(principalID string) error {
	panic("DeleteAWSEndpointServicePrincipal: is not implemented")
}

func (c *mockClient) GetRedisUser(id string) (*models.RedisUser, error) {
	panic("GetRedisUser: is not implemented")
}
