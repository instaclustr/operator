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
	CreateClusterRaw(url string, clusterSpec any) ([]byte, error)
	GetOpenSearch(id string) (*models.OpenSearchCluster, error)
	UpdateOpenSearch(id string, o models.OpenSearchInstAPIUpdateRequest) error
	UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1beta1.TwoFactorDelete) error
	UpdateCluster(id, clusterEndpoint string, instaDCs any) error
	DeleteCluster(id, clusterEndpoint string) error
	GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1beta1.PeeringStatus, error)
	GetAWSVPCPeering(peerID string) (*models.AWSVPCPeering, error)
	UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error
	DeletePeering(peerID, peeringEndpoint string) error
	CreateAzureVNetPeering(peeringSpec *clusterresourcesv1beta1.AzureVNetPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error)
	CreateGCPVPCPeering(peeringSpec *clusterresourcesv1beta1.GCPVPCPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error)
	CreateAWSVPCPeering(peeringSpec *clusterresourcesv1beta1.AWSVPCPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error)
	GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1beta1.FirewallRuleStatus, error)
	CreateAWSSecurityGroupFirewallRule(firewallRuleSpec *clusterresourcesv1beta1.AWSSecurityGroupFirewallRuleSpec, clusterID string) (*clusterresourcesv1beta1.FirewallRuleStatus, error)
	CreateClusterNetworkFirewallRule(firewallRuleSpec *clusterresourcesv1beta1.ClusterNetworkFirewallRuleSpec, clusterID string) (*clusterresourcesv1beta1.FirewallRuleStatus, error)
	DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error
	CreateKafkaUser(url string, kafkaUser *models.KafkaUser) (*kafkamanagementv1beta1.KafkaUserStatus, error)
	UpdateKafkaUser(kafkaUserID string, kafkaUserSpec *models.KafkaUser) error
	DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error
	CreateKafkaUserCertificate(certRequest *models.CertificateRequest) (*models.Certificate, error)
	DeleteKafkaUserCertificate(certificateID string) error
	RenewKafkaUserCertificate(certificateID string) (*models.Certificate, error)
	GetTopicStatus(id string) ([]byte, error)
	CreateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error
	DeleteKafkaTopic(url, id string) error
	UpdateKafkaTopic(url string, topic *kafkamanagementv1beta1.Topic) error
	CreateKafkaMirror(m *kafkamanagementv1beta1.MirrorSpec) (*kafkamanagementv1beta1.MirrorStatus, error)
	GetMirrorStatus(id string) (*kafkamanagementv1beta1.MirrorStatus, error)
	DeleteKafkaMirror(id string) error
	UpdateKafkaMirror(id string, latency int32) error
	GetClusterBackups(clusterID, clusterKind string) (*models.ClusterBackup, error)
	TriggerClusterBackup(clusterID, clusterKind string) error
	CreateExclusionWindow(clusterID string, window *clusterresourcesv1beta1.ExclusionWindowSpec) (string, error)
	GetExclusionWindowsStatus(windowID string) (string, error)
	DeleteExclusionWindow(id string) error
	GetMaintenanceEvents(clusterID, eventType string) ([]*clusterresourcesv1beta1.MaintenanceEventStatus, error)
	FetchMaintenanceEventStatuses(clusterID string) ([]*clusterresourcesv1beta1.ClusteredMaintenanceEventStatus, error)
	RescheduleMaintenanceEvent(me *clusterresourcesv1beta1.MaintenanceEventReschedule) error
	CreateNodeReload(nr *clusterresourcesv1beta1.Node) error
	GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error)
	GetRedis(id string) (*models.RedisCluster, error)
	UpdateRedis(id string, r *models.RedisDataCentreUpdate) error
	CreateRedisUser(user *models.RedisUser) (string, error)
	UpdateRedisUser(user *models.RedisUserUpdate) error
	DeleteRedisUser(id string) error
	GetRedisUser(id string) (*models.RedisUser, error)
	CreateKafkaACL(url string, kafkaACL *kafkamanagementv1beta1.KafkaACLSpec) (*kafkamanagementv1beta1.KafkaACLStatus, error)
	GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1beta1.KafkaACLStatus, error)
	DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error
	UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error
	GetCassandra(id string) (*models.CassandraCluster, error)
	UpdateCassandra(id string, cassandra models.CassandraClusterAPIUpdate) error
	GetKafka(id string) (*models.KafkaCluster, error)
	GetKafkaConnect(id string) (*models.KafkaConnectCluster, error)
	UpdateKafkaConnect(id string, kc models.KafkaConnectAPIUpdate) error
	GetZookeeper(id string) ([]byte, error)
	RestoreCluster(restoreData any, clusterKind string) (string, error)
	GetPostgreSQL(id string) (*models.PGCluster, error)
	UpdatePostgreSQL(id string, r *models.PGClusterUpdate) error
	GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error)
	CreatePostgreSQLConfiguration(id, name, value string) error
	UpdatePostgreSQLConfiguration(id, name, value string) error
	ResetPostgreSQLConfiguration(id, name string) error
	GetCadence(id string) (*models.CadenceCluster, error)
	UpdatePostgreSQLDefaultUserPassword(id, password string) error
	ListClusters() ([]*models.ActiveClusters, error)
	CreateEncryptionKey(encryptionKeySpec any) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error)
	GetEncryptionKeyStatus(encryptionKeyID string, encryptionKeyEndpoint string) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error)
	DeleteEncryptionKey(encryptionKeyID string) error
	CreateUser(userSpec any, clusterID, app string) error
	DeleteUser(username, clusterID, app string) error
	FetchUsers(clusterID, app string) ([]string, error)
	ListAppVersions(app string) ([]*models.AppVersions, error)
	GetDefaultCredentialsV1(clusterID string) (string, string, error)
	UpdateClusterSettings(clusterID string, settings *models.ClusterSettings) error
	GetAWSEndpointServicePrincipal(id string) (*models.AWSEndpointServicePrincipal, error)
	CreateOpenSearchEgressRules(rule *clusterresourcesv1beta1.OpenSearchEgressRules) (string, error)
	GetOpenSearchEgressRule(id string) (*clusterresourcesv1beta1.OpenSearchEgressRulesStatus, error)
	DeleteOpenSearchEgressRule(id string) error
	CreateAWSEndpointServicePrincipal(spec clusterresourcesv1beta1.AWSEndpointServicePrincipalSpec, CDCID string) ([]byte, error)
	DeleteAWSEndpointServicePrincipal(principalID string) error
	GetResizeOperationsByClusterDataCentreID(cdcID string) ([]*v1beta1.ResizeOperation, error)
}
