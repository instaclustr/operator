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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AWSEndpointServicePrincipalSpec defines the desired state of AWSEndpointServicePrincipal
type AWSEndpointServicePrincipalSpec struct {
	// The ID of the cluster data center
	ClusterDataCenterID string      `json:"clusterDataCenterId,omitempty"`
	ClusterRef          *ClusterRef `json:"clusterRef,omitempty"`

	// The Instaclustr ID of the AWS endpoint service
	EndPointServiceID string `json:"endPointServiceId,omitempty"`

	// The IAM Principal ARN
	PrincipalARN string `json:"principalArn"`
}

// AWSEndpointServicePrincipalStatus defines the observed state of AWSEndpointServicePrincipal
type AWSEndpointServicePrincipalStatus struct {
	// The Instaclustr ID of the IAM Principal ARN
	ID    string `json:"id,omitempty"`
	CDCID string `json:"cdcId,omitempty"`

	// The Instaclustr ID of the AWS endpoint service
	EndPointServiceID string `json:"endPointServiceId,omitempty"`

	// State describe current state of the resource
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AWSEndpointServicePrincipal is the Schema for the awsendpointserviceprincipals API
type AWSEndpointServicePrincipal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSEndpointServicePrincipalSpec   `json:"spec,omitempty"`
	Status AWSEndpointServicePrincipalStatus `json:"status,omitempty"`
}

func (p *AWSEndpointServicePrincipal) NewPatch() client.Patch {
	return client.MergeFrom(p.DeepCopy())
}

func (p *AWSEndpointServicePrincipal) GetJobID(job string) string {
	return p.Kind + "/" + p.Namespace + "/" + p.Name + "/" + job
}

//+kubebuilder:object:root=true

// AWSEndpointServicePrincipalList contains a list of AWSEndpointServicePrincipal
type AWSEndpointServicePrincipalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSEndpointServicePrincipal `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSEndpointServicePrincipal{}, &AWSEndpointServicePrincipalList{})
}
