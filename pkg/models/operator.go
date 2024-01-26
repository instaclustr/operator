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

package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation   = "instaclustr.com/resourceState"
	ClusterDeletionAnnotation = "instaclustr.com/clusterDeletion"
	ExternalChangesAnnotation = "instaclustr.com/externalChanges"
	DeletionFinalizer         = "instaclustr.com/deletionFinalizer"
	StartTimestampAnnotation  = "instaclustr.com/startTimestamp"

	DefaultSecretLabel                = "instaclustr.com/defaultSecret"
	ControlledByLabel                 = "instaclustr.com/controlledBy"
	ClusterIDLabel                    = "instaclustr.com/clusterID"
	ClusterNameLabel                  = "instaclustr.com/clusterName"
	ClustersV1beta1APIVersion         = "clusters.instaclustr.com/v1beta1"
	ClusterresourcesV1beta1APIVersion = "clusterresources.instaclustr.com/v1beta1"
	RedisUserNamespaceLabel           = "instaclustr.com/redisUserNamespace"
	PostgreSQLUserNamespaceLabel      = "instaclustr.com/postgresqlUserNamespace"
	OpenSearchUserNamespaceLabel      = "instaclustr.com/openSearchUserNamespace"

	CassandraKind        = "Cassandra"
	CassandraChildPrefix = "cassandra-"
	CassandraChildDCName = "cassandra-cadence-dc"

	KafkaKind        = "Kafka"
	KafkaChildPrefix = "kafka-"
	KafkaChildDCName = "kafka-cadence-dc"

	OpenSearchKind        = "OpenSearch"
	OpenSearchChildPrefix = "opensearch-"
	OpenSearchChildDCName = "opensearch-cadence-dc"

	K8sAPIVersionV1 = "v1"
	VPCPeered       = "VPC_PEERED"

	True  = "true"
	False = "false"

	Triggered = "triggered"

	ClusterBackupKind                = "ClusterBackup"
	PgClusterKind                    = "PostgreSQL"
	RedisClusterKind                 = "Redis"
	OsClusterKind                    = "OpenSearch"
	CadenceClusterKind               = "Cadence"
	KafkaConnectClusterKind          = "KafkaConnect"
	CassandraClusterKind             = "Cassandra"
	ZookeeperClusterKind             = "Zookeeper"
	ClusterNetworkFirewallRuleKind   = "ClusterNetworkFirewallRule"
	SecretKind                       = "Secret"
	PgBackupEventType                = "postgresql-backup"
	SnapshotUploadEventType          = "snapshot-upload"
	PgBackupPrefix                   = "postgresql-backup-"
	SnapshotUploadPrefix             = "snapshot-upload-"
	ClusterNetworkFirewallRulePrefix = "firewall-rule"
	DefaultUserSecretPrefix          = "default-user-password"
	DefaultUserSecretNameTemplate    = "%s-%s"
	ExposeServiceNameTemplate        = "%s-service"
	PgRestoreValue                   = "postgres"

	CassandraConnectionPort    = 9042
	CadenceConnectionPort      = 7933
	KafkaConnectionPort        = 9092
	KafkaConnectConnectionPort = 8083
	OpenSearchConnectionPort   = 9200
	PgConnectionPort           = 5432
	RedisConnectionPort        = 6379
	ZookeeperConnectionPort    = 2181

	KafkaConnectAppKind = "kafka-connect"
	CadenceAppKind      = "cadence"
	CassandraAppKind    = "cassandra"
	KafkaAppKind        = "kafka"
	OpenSearchAppKind   = "opensearch"
	PgAppKind           = "postgresql"
	RedisAppKind        = "redis"
	ZookeeperAppKind    = "zookeeper"

	KafkaAppType        = "KAFKA"
	RedisAppType        = "REDIS"
	CadenceAppType      = "CADENCE"
	ZookeeperAppType    = "APACHE_ZOOKEEPER"
	OpenSearchAppType   = "OPENSEARCH"
	PgAppType           = "POSTGRESQL"
	KafkaConnectAppType = "KAFKA_CONNECT"
	CassandraAppType    = "APACHE_CASSANDRA"

	DefaultPgUsernameValue = "icpostgresql"
	DefaultPgDbNameValue   = "postgres"
	DefaultPgDbPortValue   = 5432
)

const (
	CreatingEvent        = "creating"
	CreatedEvent         = "created"
	UpdatingEvent        = "updating"
	UpdatedEvent         = "updated"
	DeletingEvent        = "deleting"
	DeletedEvent         = "deleted"
	GenericEvent         = "generic"
	SecretEvent          = "secret"
	ClusterDeletingEvent = "cluster deleting"
)

const (
	Normal           = "Normal"
	Warning          = "Warning"
	Created          = "Created"
	PatchFailed      = "PatchFailed"
	NotFound         = "NotFound"
	CreationFailed   = "CreationFailed"
	FetchFailed      = "FetchFailed"
	ConversionFailed = "ConversionFailed"
	ValidationFailed = "ValidationFailed"
	UpdateFailed     = "UpdateFailed"
	ExternalChanges  = "ExternalChanges"
	DeletionStarted  = "DeletionStarted"
	DeletionFailed   = "DeletionFailed"
	Deleted          = "Deleted"
	ExternalDeleted  = "ExternalDeleted"
)

const (
	ReplaceOperation = "replace"
	AnnotationsPath  = "/metadata/annotations"
	FinalizersPath   = "/metadata/finalizers"

	Username = "username"
	Password = "password"
)

const (
	InProgressME = "in-progress"
	PastME       = "past"
	UpcomingME   = "upcoming"
)

const (
	AWSVPCPeeringStatusCodeDeleted = "deleted"
)

const Requeue60 = time.Second * 60

var (
	ReconcileRequeue   = reconcile.Result{RequeueAfter: Requeue60}
	ImmediatelyRequeue = reconcile.Result{RequeueAfter: 1}
	ExitReconcile      = reconcile.Result{}
)

type Credentials struct {
	Username string
	Password string
}

var ClusterKindsMap = map[string]string{"PostgreSQL": "postgres", "Redis": "redis", "OpenSearch": "opensearch", "Cassandra": "cassandra"}

const (
	CertificateRequestType     = "CERTIFICATE REQUEST"
	RSAPrivateKeyType          = "RSA PRIVATE KEY"
	SignedCertificateSecretKey = "signedCertificate"
	CSRSecretKey               = "csr"
	PrivateKeySecretKey        = "privateKey"
)

const (
	ExternalChangesBaseMessage = "There are external changes on the Instaclustr console. Please reconcile the specification manually."
	SpecPath                   = "spec"
)
