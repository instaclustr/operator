package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation  = "instaclustr.com/resourceState"
	DeletionConfirmed        = "instaclustr.com/deletionConfirmed"
	DeletionFinalizer        = "instaclustr.com/deletionFinalizer"
	UpdatedFieldsAnnotation  = "instaclustr.com/updatedFields"
	StartTimestampAnnotation = "instaclustr.com/startTimestamp"

	ControlledByLabel                  = "instaclustr.com/controlledBy"
	ClusterIDLabel                     = "instaclustr.com/clusterID"
	ClustersV1alpha1APIVersion         = "clusters.instaclustr.com/v1alpha1"
	ClusterresourcesV1alpha1APIVersion = "clusterresources.instaclustr.com/v1alpha1"

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
	V1_3_7     = "1.3.7"
	VPC_PEERED = "VPC_PEERED"

	True     = "true"
	False    = "false"
	Pending  = "pending"
	Canceled = "canceled"

	ClusterBackupKind       = "ClusterBackup"
	PgClusterKind           = "PostgreSQL"
	RedisClusterKind        = "Redis"
	OsClusterKind           = "OpenSearch"
	CassandraClusterKind    = "Cassandra"
	PgBackupEventType       = "postgresql-backup"
	SnapshotUploadEventType = "snapshot-upload"
	PgBackupPrefix          = "postgresql-backup-"
	SnapshotUploadPrefix    = "snapshot-upload-"

	Phone = "Phone"
	Email = "Email"
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

const (
	Requeue60        = time.Second * 60
	Requeue10Minutes = time.Minute * 10
)

var (
	ReconcileRequeue         = reconcile.Result{RequeueAfter: Requeue60}
	ReconcileRequeue10Minute = reconcile.Result{RequeueAfter: Requeue10Minutes}
	ReconcileResult          = reconcile.Result{}
)

const (
	MessagePendingDeletion = `Please confirm cluster deletion via email (or phone), and then ` +
		`set Instaclustr.com/deletionConfirmed annotation to "true". ` +
		`To cancel cluster deletion, set the annotation to "false".`

	MessageDeleteCluster = "Cluster deletion has been confirmed. Deleting a resource."

	MessageCancelClusterDeletion = `Cluster deletion has been canceled. To put the cluster on deletion again, ` +
		`remove "canceled" from Instaclustr.com/deletionConfirmed annotation.`

	MessageUnknownDeleteAnnotation = `Unhandled delete annotation. Cluster deletion is need to be confirmed. ` +
		`If you want to delete cluster, set Instaclustr.com/deletionConfirmed annotation to "true", or` +
		`"false" to cancel cluster deletion.`
)
