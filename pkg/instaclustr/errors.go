package instaclustr

import "errors"

var (
	StatusPreconditionFailed  = errors.New("412 - status precondition failed")
	ClusterNotRunning         = errors.New("сluster is not running")
	NotFound                  = errors.New("not found")
	HasActiveResizeOperation  = errors.New("cluster has active resize operation")
	ClusterIsNotReadyToResize = errors.New("cluster is not ready to resize yet")
)
