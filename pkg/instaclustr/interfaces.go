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

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

type API interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetOpenSearch(id string) ([]byte, error)
	UpdateOpenSearch(id string, o models.OpenSearchInstAPIUpdateRequest) error
	UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1alpha1.TwoFactorDelete) error
	UpdateCluster(id, clusterEndpoint string, instaDCs any) error
	DeleteCluster(id, clusterEndpoint string) error
	GetPeeringStatus(peerID, peeringEndpoint string) (*clusterresourcesv1alpha1.PeeringStatus, error)
	UpdatePeering(peerID, peeringEndpoint string, peerSpec any) error
	DeletePeering(peerID, peeringEndpoint string) error
	CreatePeering(url string, peeringSpec any) (*clusterresourcesv1alpha1.PeeringStatus, error)
	GetFirewallRuleStatus(firewallRuleID string, firewallRuleEndpoint string) (*clusterresourcesv1alpha1.FirewallRuleStatus, error)
	CreateFirewallRule(url string, firewallRuleSpec any) (*clusterresourcesv1alpha1.FirewallRuleStatus, error)
	DeleteFirewallRule(firewallRuleID string, firewallRuleEndpoint string) error
	GetKafkaUserStatus(kafkaUserID, kafkaUserEndpoint string) (*kafkamanagementv1alpha1.KafkaUserStatus, error)
	CreateKafkaUser(url string, kafkaUser *models.KafkaUser) (*kafkamanagementv1alpha1.KafkaUserStatus, error)
	UpdateKafkaUser(kafkaUserID string, kafkaUserSpec *models.KafkaUser) error
	DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error
	GetTopicStatus(id string) ([]byte, error)
	CreateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	DeleteKafkaTopic(url, id string) error
	UpdateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	CreateKafkaMirror(m *kafkamanagementv1alpha1.MirrorSpec) (*kafkamanagementv1alpha1.MirrorStatus, error)
	GetMirrorStatus(id string) (*kafkamanagementv1alpha1.MirrorStatus, error)
	DeleteKafkaMirror(id string) error
	UpdateKafkaMirror(id string, latency int32) error
	GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error)
	TriggerClusterBackup(url, clusterID string) error
	CreateExclusionWindow(clusterID string, window clusterresourcesv1alpha1.ExclusionWindowSpec) (string, error)
	GetExclusionWindowsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.ExclusionWindowStatus, error)
	GetMaintenanceEventsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.MaintenanceEventStatus, error)
	GetMaintenanceEvents(clusterID string) ([]*v1alpha1.MaintenanceEvent, error)
	DeleteExclusionWindow(id string) error
	UpdateMaintenanceEvent(me clusterresourcesv1alpha1.MaintenanceEventRescheduleSpec) (*clusterresourcesv1alpha1.MaintenanceEventStatus, error)
	RestorePgCluster(restoreData *v1alpha1.PgRestoreFrom) (string, error)
	RestoreRedisCluster(restoreData *v1alpha1.RedisRestoreFrom) (string, error)
	RestoreOpenSearchCluster(restoreData *v1alpha1.OpenSearchRestoreFrom) (string, error)
	CreateNodeReload(nr *clusterresourcesv1alpha1.Node) error
	GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error)
	GetRedis(id string) ([]byte, error)
	CreateRedisUser(user *models.RedisUser) (string, error)
	UpdateRedisUser(user *models.RedisUserUpdate) error
	DeleteRedisUser(id string) error
	CreateKafkaACL(url string, kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error
	UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error
	GetCassandra(id string) ([]byte, error)
	UpdateCassandra(id string, cassandra models.CassandraClusterAPIUpdate) error
	GetKafka(id string) ([]byte, error)
	GetKafkaConnect(id string) ([]byte, error)
	UpdateKafkaConnect(id string, kc models.KafkaConnectAPIUpdate) error
	GetZookeeper(id string) ([]byte, error)
	RestoreCassandra(restoreData v1alpha1.CassandraRestoreFrom) (string, error)
	GetPostgreSQL(id string) ([]byte, error)
	UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error
	GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error)
	CreatePostgreSQLConfiguration(id, name, value string) error
	UpdatePostgreSQLConfiguration(id, name, value string) error
	ResetPostgreSQLConfiguration(id, name string) error
	GetCadence(id string) ([]byte, error)
	UpdatePostgreSQLDefaultUserPassword(id, password string) error
	ListClusters() ([]*models.ActiveClusters, error)
	CreateEncryptionKey(encryptionKeySpec any) (*clusterresourcesv1alpha1.AWSEncryptionKeyStatus, error)
	GetEncryptionKeyStatus(encryptionKeyID string, encryptionKeyEndpoint string) (*clusterresourcesv1alpha1.AWSEncryptionKeyStatus, error)
	DeleteEncryptionKey(encryptionKeyID string) error
	CreateUser(userSpec any, clusterID, app string) error
	DeleteUser(username, clusterID, app string) error
}
