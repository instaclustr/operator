package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation = "instaclustr.com/resourceState"
	DeletionConfirmed       = "instaclustr.com/deletionConfirmed"
	DeletionFinalizer       = "instaclustr.com/deletionFinalizer"
	UpdatedFieldsAnnotation = "instaclustr.com/updatedFields"

	ControlledByLabel          = "instaclustr.com/controlledBy"
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
)

const (
	CreatingEvent = "creating"
	CreatedEvent  = "created"
	UpdatingEvent = "updating"
	UpdatedEvent  = "updated"
	DeletingEvent = "deleting"
	DeletedEvent  = "deleted"
	GenericEvent  = "generic"
)

const (
	ReplaceOperation = "replace"
	AnnotationsPath  = "/metadata/annotations"
	FinalizersPath   = "/metadata/finalizers"
)

const (
	Requeue60 = time.Second * 60
)

var ReconcileRequeue = reconcile.Result{RequeueAfter: Requeue60}
var ReconcileResult = reconcile.Result{}
