package instaclustr

import "errors"

var (
	ClusterIsBeingDeleted     = errors.New("cluster is being deleted")
	NotFound                  = errors.New("not found")
	ClusterIsNotReadyToResize = errors.New("cluster is not ready to resize yet")
	ClusterNotRunning         = errors.New("сluster is not running")
	StatusPreconditionFailed  = errors.New("412 - status precondition failed")
	IncorrectNodeSize         = errors.New("incorrect node size")
)
