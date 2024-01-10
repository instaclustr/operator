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

	"github.com/instaclustr/operator/pkg/models"
)

// AzureVNetPeeringSpec defines the desired state of AzureVNetPeering
type AzureVNetPeeringSpec struct {
	PeeringSpec            `json:",inline"`
	PeerResourceGroup      string `json:"peerResourceGroup"`
	PeerSubscriptionID     string `json:"peerSubscriptionId"`
	PeerADObjectID         string `json:"peerAdObjectId,omitempty"`
	PeerVirtualNetworkName string `json:"peerVirtualNetworkName"`
}

// AzureVNetPeeringStatus defines the observed state of AzureVNetPeering
type AzureVNetPeeringStatus struct {
	PeeringStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="StatusCode",type="string",JSONPath=".status.statusCode"

// AzureVNetPeering is the Schema for the azurevnetpeerings API
type AzureVNetPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVNetPeeringSpec   `json:"spec,omitempty"`
	Status AzureVNetPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureVNetPeeringList contains a list of AzureVNetPeering
type AzureVNetPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVNetPeering `json:"items"`
}

func (azure *AzureVNetPeering) GetJobID(jobName string) string {
	return azure.Kind + "/" + client.ObjectKeyFromObject(azure).String() + "/" + jobName
}

func (azure *AzureVNetPeering) NewPatch() client.Patch {
	old := azure.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&AzureVNetPeering{}, &AzureVNetPeeringList{})
}
