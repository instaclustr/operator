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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

// AWSVPCPeeringSpec defines the desired state of AWSVPCPeering
type AWSVPCPeeringSpec struct {
	VPCPeeringSpec   `json:",inline"`
	PeerAWSAccountID string `json:"peerAwsAccountId"`
	PeerVPCID        string `json:"peerVpcId"`
	PeerRegion       string `json:"peerRegion,omitempty"`
}

// AWSVPCPeeringStatus defines the observed state of AWSVPCPeering
type AWSVPCPeeringStatus struct {
	PeeringStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSVPCPeering is the Schema for the awsvpcpeerings API
type AWSVPCPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSVPCPeeringSpec   `json:"spec,omitempty"`
	Status AWSVPCPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSVPCPeeringList contains a list of AWSVPCPeering
type AWSVPCPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSVPCPeering `json:"items"`
}

func (aws *AWSVPCPeering) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(aws).String() + "/" + jobName
}

func (aws *AWSVPCPeering) NewPatch() client.Patch {
	old := aws.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&AWSVPCPeering{}, &AWSVPCPeeringList{})
}
