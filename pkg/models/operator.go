package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation = "instaclustr.com/resourceState"
	DeletionFinalizer       = "instaclustr.com/deletionFinalizer"
	UpdatedFieldsAnnotation = "instaclustr.com/updatedFields"

	ControlledByLabel    = "instaclustr.com/controlledBy"
	CassandraChildPrefix = "cassandra-"

	ClustersV1alpha1APIVersion = "clusters.instaclustr.com/v1alpha1"
	CassandraKind              = "Cassandra"
	CassandraChildDCName       = "cassandra-cadence-dc"
	V3_11_13                   = "3.11.13"
	VPC_PEERED                 = "VPC_PEERED"
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
