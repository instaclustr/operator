package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation   = "instaclustr.com/resourceState"
	ClusterDeletionAnnotation = "instaclustr.com/clusterDeletion"
	ExternalChangesAnnotation = "instaclustr.com/externalChanges"
	AllowSpecAmendAnnotation  = "instaclustr.com/allowSpecAmend"
	DeletionConfirmed         = "instaclustr.com/deletionConfirmed"
	DeletionFinalizer         = "instaclustr.com/deletionFinalizer"
	StartTimestampAnnotation  = "instaclustr.com/startTimestamp"

	ControlledByLabel                  = "instaclustr.com/controlledBy"
	ClusterIDLabel                     = "instaclustr.com/clusterID"
	ClustersV1alpha1APIVersion         = "clusters.instaclustr.com/v1alpha1"
	ClusterresourcesV1alpha1APIVersion = "clusterresources.instaclustr.com/v1alpha1"
	RedisUserNamespaceLabel            = "instaclustr.com/redisUserNamespace"

	CassandraKind        = "Cassandra"
	CassandraChildPrefix = "cassandra-"
	CassandraChildDCName = "cassandra-cadence-dc"

	KafkaKind        = "Kafka"
	KafkaChildPrefix = "kafka-"
	KafkaChildDCName = "kafka-cadence-dc"

	OpenSearchKind        = "OpenSearch"
	OpenSearchChildPrefix = "opensearch-"
	OpenSearchChildDCName = "opensearch-cadence-dc"

	CassandraV3_11_13 = "3.11.13"
	KafkaV3_1_2       = "3.1.2"
	OpensearchV1_3_7  = "opensearch:1.3.7"
	K8sAPIVersionV1   = "v1"
	VPCPeered         = "VPC_PEERED"

	True  = "true"
	False = "false"

	Triggered = "triggered"

	ClusterBackupKind       = "ClusterBackup"
	PgClusterKind           = "PostgreSQL"
	RedisClusterKind        = "Redis"
	OsClusterKind           = "OpenSearch"
	CassandraClusterKind    = "Cassandra"
	SecretKind              = "Secret"
	PgBackupEventType       = "postgresql-backup"
	SnapshotUploadEventType = "snapshot-upload"
	PgBackupPrefix          = "postgresql-backup-"
	SnapshotUploadPrefix    = "snapshot-upload-"
	DefaultUserSecretPrefix = "default-user-password-"
)

const (
	CreatingEvent = "creating"
	CreatedEvent  = "created"
	UpdatingEvent = "updating"
	UpdatedEvent  = "updated"
	DeletingEvent = "deleting"
	DeletedEvent  = "deleted"
	GenericEvent  = "generic"
	SecretEvent   = "secret"
)

const (
	Normal            = "Normal"
	Warning           = "Warning"
	Created           = "Created"
	PatchFailed       = "PatchFailed"
	CreationFailed    = "CreationFailed"
	FetchFailed       = "FetchFailed"
	FetchSecretFailed = "FetchSecretFailed"
	ConvertionFailed  = "ConvertionFailed"
	UpdateFailed      = "UpdateFailed"
	ExternalChanges   = "ExternalChanges"
	DeletionStarted   = "DeletionStarted"
	DeletionFailed    = "DeletionFailed"
	Deleted           = "Deleted"
)

const (
	ReplaceOperation    = "replace"
	AnnotationsPath     = "/metadata/annotations"
	FinalizersPath      = "/metadata/finalizers"
	DefaultUserPassword = "defaultUserPassword"
)

const Requeue60 = time.Second * 60

var (
	ReconcileRequeue = reconcile.Result{RequeueAfter: Requeue60}
	ExitReconcile    = reconcile.Result{}
)
