package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation   = "instaclustr.com/resourceState"
	ClusterDeletionAnnotation = "instaclustr.com/clusterDeletion"
	DeletionConfirmed         = "instaclustr.com/deletionConfirmed"
	DeletionFinalizer         = "instaclustr.com/deletionFinalizer"
	UpdatedFieldsAnnotation   = "instaclustr.com/updatedFields"

	ControlledByLabel          = "instaclustr.com/controlledBy"
	ClusterIDLabel             = "instaclustr.com/clusterID"
	ClustersV1alpha1APIVersion = "clusters.instaclustr.com/v1alpha1"

	CassandraKind        = "Cassandra"
	CassandraChildPrefix = "cassandra-"
	CassandraChildDCName = "cassandra-cadence-dc"

	KafkaKind        = "Kafka"
	KafkaChildPrefix = "kafka-"
	KafkaChildDCName = "kafka-cadence-dc"

	OpenSearchKind        = "OpenSearch"
	OpenSearchChildPrefix = "opensearch-"
	OpenSearchChildDCName = "opensearch-cadence-dc"

	V3_11_13   = "3.11.13"
	V2_7_1     = "2.7.1"
	V1_3_5     = "1.3.5"
	VPC_PEERED = "VPC_PEERED"

	True  = "true"
	False = "false"

	Triggered = "triggered"

	ClusterBackupKind                  = "ClusterBackup"
	PgClusterKind                      = "PostgreSQL"
	RedisClusterKind                   = "Redis"
	OsClusterKind                      = "OpenSearch"
	CassandraClusterKind               = "Cassandra"
	PgBackupEventType                  = "postgresql-backup"
	SnapshotUploadEventType            = "snapshot-upload"
	ClusterresourcesV1alpha1APIVersion = "clusterresources.instaclustr.com/v1alpha1"
	StartAnnotation                    = "instaclustr.com/startTimestamp"
	PgBackupPrefix                     = "postgresql-backup-"
	SnapshotUploadPrefix               = "snapshot-upload-"
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
	ReplaceOperation = "replace"
	AnnotationsPath  = "/metadata/annotations"
	FinalizersPath   = "/metadata/finalizers"
)

const Requeue60 = time.Second * 60

var (
	ReconcileRequeue = reconcile.Result{RequeueAfter: Requeue60}
	ReconcileResult  = reconcile.Result{}
)
