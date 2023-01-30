package instaclustr

import (
	"net/http"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
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
	UpdateCluster(id, clusterEndpoint string, instaDCs any) error
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
	CreateKafkaUser(url string, kafkaUser *modelsv2.KafkaUserAPIv2) (*kafkamanagementv1alpha1.KafkaUserStatus, error)
	UpdateKafkaUser(kafkaUserID, kafkaUserEndpoint string, kafkaUserSpec *modelsv2.KafkaUserAPIv2) error
	DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error
	CreateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	DeleteKafkaTopic(url, id string) error
	UpdateKafkaTopic(url string, topic *kafkamanagementv1alpha1.Topic) error
	CreateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error
	GetMirrorStatus(id, mirrorEndpoint string) (*kafkamanagementv1alpha1.MirrorStatus, error)
	DeleteKafkaMirror(url, id string) error
	UpdateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error
	GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error)
	TriggerClusterBackup(url, clusterID string) error
	CreateExclusionWindow(url string, me *clusterresourcesv1alpha1.MaintenanceEventsSpec) (*clusterresourcesv1alpha1.MaintenanceEventsStatus, error)
	GetExclusionWindowStatus(clusterId string, endpoint string) (*clusterresourcesv1alpha1.MaintenanceEventsStatus, error)
	GetMaintenanceEventStatus(eventID string, endpoint string) (*clusterresourcesv1alpha1.MaintenanceEventsStatus, error)
	DeleteExclusionWindow(meStatus *clusterresourcesv1alpha1.MaintenanceEventsStatus, endpoint string) error
	UpdateMaintenanceEvent(me *clusterresourcesv1alpha1.MaintenanceEventsSpec, endpoint string) error
	RestorePgCluster(restoreData *v1alpha1.PgRestoreFrom) (string, error)
	RestoreRedisCluster(restoreData *v1alpha1.RedisRestoreFrom) (string, error)
	RestoreOpenSearchCluster(restoreData *v1alpha1.OpenSearchRestoreFrom) (string, error)
	CreateNodeReload(bundle, nodeID string, nr *modelsv1.NodeReload) error
	GetNodeReloadStatus(bundle, nodeID string) (*modelsv1.NodeReloadStatusAPIv1, error)
	GetClusterSpec(id, clusterEndpoint string) (*models.ClusterSpec, error)
	GetRedisSpec(id, clusterEndpoint string) (*models.RedisCluster, error)
	CreateKafkaACL(url string, kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	GetKafkaACLStatus(kafkaACLID, kafkaACLEndpoint string) (*kafkamanagementv1alpha1.KafkaACLStatus, error)
	DeleteKafkaACL(kafkaACLID, kafkaACLEndpoint string) error
	UpdateKafkaACL(kafkaACLID, kafkaACLEndpoint string, kafkaACLSpec any) error
	GetCassandra(id, clusterEndpoint string) (*modelsv2.CassandraCluster, error)
	RestoreCassandra(restoreData v1alpha1.CassandraRestoreFrom) (string, error)
	GetPostgreSQL(id string) (*models.PGStatus, error)
	UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error
	GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error)
	CreatePostgreSQLConfiguration(id, name, value string) error
	UpdatePostgreSQLConfiguration(id, name, value string) error
	ResetPostgreSQLConfiguration(id, name string) error
	GetCadence(id string) (*models.CadenceAPIv2, error)
	UpdatePostgreSQLDefaultUserPassword(id, password string) error
}
