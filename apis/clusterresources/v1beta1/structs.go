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

package v1beta1

import (
	"encoding/json"
)

type VPCPeeringSpec struct {
	DataCentreID string   `json:"cdcId,omitempty"`
	PeerSubnets  []string `json:"peerSubnets"`
}

type PeeringStatus struct {
	ID            string `json:"id,omitempty"`
	StatusCode    string `json:"statusCode,omitempty"`
	Name          string `json:"name,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
	CDCID         string `json:"cdcid,omitempty"`
	ResourceState string `json:"resourceState,omitempty"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type FirewallRuleSpec struct {
	ClusterID string `json:"clusterId,omitempty"`
	Type      string `json:"type"`
}

type FirewallRuleStatus struct {
	ID             string `json:"id,omitempty"`
	DeferredReason string `json:"deferredReason,omitempty"`
	Status         string `json:"status,omitempty"`
	ClusterID      string `json:"clusterId,omitempty"`
	ResourceState  string `json:"resourceState,omitempty"`
}

type SecretReference struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
