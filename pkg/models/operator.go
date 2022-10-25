package models

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ResourceStateAnnotation = "instaclustr.com/resourceState"
	DeletionFinalizer       = "instaclustr.com/deletionFinalizer"
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

var ReconcileRequeue60 = reconcile.Result{RequeueAfter: Requeue60}