package models

import "time"

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
