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

package instaclustr

import (
	"net/http"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	kafkamanagementv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

type API interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetOpenSearch(id string) ([]byte, error)
	UpdateOpenSearch(id string, o models.OpenSearchInstAPIUpdateRequest) error
	UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1beta1.TwoFactorDelete) error
	UpdateCluster(id, clusterEndpoint string, instaDCs any) error
	DeleteCluster(id, clusterEndpoint string) error
	GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1beta1.PeeringStatus, error)
	UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error
	DeletePeering(peerID, peeringEndpoint string) error
	CreatePeering(url string, peeringSpec any) (*clusterresourcesv1beta1.PeeringStatus, error)
	GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1beta1.FirewallRuleStatus, error)
	CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1beta1.FirewallRuleStatus, error)
	DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error
	GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1beta1.KafkaUserStatus, error)
	CreateKafkaUser(url string, kafkaUser *models.KafkaUser) (*kafkamanagementv1beta1.KafkaUserStatus, error)
	UpdateKafkaUser(kafkaUserID string, kafkaUserSpec *models.KafkaUser) error
	DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error
	CreateKafkaUserCertificate(certRequest *models.CertificateRequest) (*kafkamanagementv1beta1.Certificate, error)
	DeleteKafkaUserCertificate(certificateID string) error
	RenewKafkaUserCertificate(certificateID string) (*kafkamanagementv1beta1.Certificate, error)
	GetTopicStatus(id string) ([]byte, error)
	CreateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error
	DeleteKafkaTopic(url, id string) error
	UpdateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error
	CreateKafkaMirror(m *kafkamanagementv1beta1.MirrorSpec) (*kafkamanagementv1beta1.MirrorStatus, error)
	GetMirrorStatus(id string) (*kafkamanagementv1beta1.MirrorStatus, error)
	DeleteKafkaMirror(id string) error
	UpdateKafkaMirror(id string, latency int32) error
	GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error)
	TriggerClusterBackup(url, clusterID string) error
	CreateExclusionWindow(clusterID string, window *clusterresourcesv1beta1.ExclusionWindowSpec) (string, error)
	GetExclusionWindowsStatus(windowID string) (string, error)
	DeleteExclusionWindow(id string) error
	GetMaintenanceEvents(clusterID, eventType string) ([]*clusterresourcesv1beta1.MaintenanceEventStatus, error)
	FetchMaintenanceEventStatuses(clusterID string) ([]*clusterresourcesv1beta1.ClusteredMaintenanceEventStatus, error)
	RescheduleMaintenanceEvent(me *clusterresourcesv1beta1.MaintenanceEventReschedule) error
	RestorePgCluster(restoreData *v1beta1.PgRestoreFrom) (string, error)
	RestoreRedisCluster(restoreData *v1beta1.RedisRestoreFrom) (string, error)
	RestoreOpenSearchCluster(restoreData *v1beta1.OpenSearchRestoreFrom) (string, error)
	CreateNodeReload(nr *clusterresourcesv1beta1.Node) error
	GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error)
	GetRedis(id string) ([]byte, error)
	UpdateRedis(id string, r *models.RedisDataCentreUpdate) error
	CreateRedisUser(user *models.RedisUser) (string, error)
	UpdateRedisUser(user *models.RedisUserUpdate) error
	DeleteRedisUser(id string) error
	GetRedisUser(id string) (*models.RedisUser, error)
	CreateKafkaACL(url string, kafkaACL *kafkamanagementv1beta1.KafkaACLSpec) (*kafkamanagementv1beta1.KafkaACLStatus, error)
	GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1beta1.KafkaACLStatus, error)
	DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error
	UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error
	GetCassandra(id string) ([]byte, error)
	UpdateCassandra(id string, cassandra models.CassandraClusterAPIUpdate) error
	GetKafka(id string) ([]byte, error)
	GetKafkaConnect(id string) ([]byte, error)
	UpdateKafkaConnect(id string, kc models.KafkaConnectAPIUpdate) error
	GetZookeeper(id string) ([]byte, error)
	RestoreCassandra(restoreData v1beta1.CassandraRestoreFrom) (string, error)
	GetPostgreSQL(id string) ([]byte, error)
	UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error
	GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error)
	CreatePostgreSQLConfiguration(id, name, value string) error
	UpdatePostgreSQLConfiguration(id, name, value string) error
	ResetPostgreSQLConfiguration(id, name string) error
	GetCadence(id string) ([]byte, error)
	UpdatePostgreSQLDefaultUserPassword(id, password string) error
	ListClusters() ([]*models.ActiveClusters, error)
	CreateEncryptionKey(encryptionKeySpec any) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error)
	GetEncryptionKeyStatus(encryptionKeyID string, encryptionKeyEndpoint string) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error)
	DeleteEncryptionKey(encryptionKeyID string) error
	CreateUser(userSpec any, clusterID, app string) error
	DeleteUser(username, clusterID, app string) error
	ListAppVersions(app string) ([]*models.AppVersions, error)
	GetDefaultCredentialsV1(clusterID string) (string, string, error)
	UpdateClusterSettings(clusterID string, settings *models.ClusterSettings) error
	CreateAWSEndpointServicePrincipal(spec any) ([]byte, error)
	DeleteAWSEndpointServicePrincipal(principalID string) error
}
