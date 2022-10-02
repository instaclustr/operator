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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RedisBundle struct {
	Bundle  `json:",inline"`
	Options RedisBundleOptions `json:"options,omitempty"`
}

type RedisBundleOptions struct {
	ClientEncryption bool  `json:"clientEncryption,omitempty"`
	PasswordAuth     bool  `json:"passwordAuth,omitempty"`
	MasterNodes      int32 `json:"masterNodes,omitempty"`
	ReplicaNodes     int32 `json:"replicaNodes,omitempty"`
}

type RedisDataCentre struct {
	DataCentre `json:",inline"`
	Bundles    []RedisBundle `json:"bundles,omitempty"`
}

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	Cluster             `json:",inline"`
	Bundles             []RedisBundle     `json:"bundles"`
	PCICompliantCluster bool              `json:"pciCompliantCluster,omitempty"`
	DataCentres         []RedisDataCentre `json:"dataCentres,omitempty"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
