package instaclustr

import "errors"

var (
	StatusPreconditionFailed = errors.New("412 - status precondition failed")
	ClusterNotRunning        = errors.New("сluster is not running")
	NotFound                 = errors.New("not found")
	IncorrectNodeSize        = errors.New("incorrect node size")
	ClusterIsBeingDeleted    = errors.New("cluster is being deleted")
	HasActiveResizeOperation = errors.New("cluster has active resize operation")
)
