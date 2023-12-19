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

	"github.com/instaclustr/operator/pkg/apiextensions"

	"k8s.io/apimachinery/pkg/types"
)

type PeeringSpec struct {
	DataCentreID string      `json:"cdcId,omitempty"`
	PeerSubnets  []string    `json:"peerSubnets"`
	ClusterRef   *ClusterRef `json:"clusterRef,omitempty"`
}

type PeeringStatus struct {
	ID            string `json:"id,omitempty"`
	StatusCode    string `json:"statusCode,omitempty"`
	Name          string `json:"name,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
	CDCID         string `json:"cdcId,omitempty"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type FirewallRuleSpec struct {
	ClusterID  string      `json:"clusterId,omitempty"`
	Type       string      `json:"type"`
	ClusterRef *ClusterRef `json:"clusterRef,omitempty"`
}

type FirewallRuleStatus struct {
	ID             string `json:"id,omitempty"`
	DeferredReason string `json:"deferredReason,omitempty"`
	Status         string `json:"status,omitempty"`
	ClusterID      string `json:"clusterId,omitempty"`
}

type immutablePeeringFields struct {
	DataCentreID string
}

// +kubebuilder:object:generate:=false
type SecretReference = apiextensions.ObjectReference

type ClusterRef struct {
	Name        string `json:"name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	ClusterKind string `json:"clusterKind,omitempty"`
	CDCName     string `json:"cdcName,omitempty"`
}

func (r *ClusterRef) AsNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}
