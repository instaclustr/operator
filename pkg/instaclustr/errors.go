/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instaclustr

import "errors"

var (
	StatusPreconditionFailed  = errors.New("412 - status precondition failed")
	ClusterNotRunning         = errors.New("сluster is not running")
	NotFound                  = errors.New("not found")
	HasActiveResizeOperation  = errors.New("cluster has active resize operation")
	ClusterIsNotReadyToResize = errors.New("cluster is not ready to resize yet")
)
