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
	"github.com/instaclustr/operator/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AWSEncryptionKeySpec defines the desired state of AWSEncryptionKey
type AWSEncryptionKeySpec struct {
	Alias               string `json:"alias"`
	ARN                 string `json:"arn"`
	ProviderAccountName string `json:"providerAccountName,omitempty"`
}

// AWSEncryptionKeyStatus defines the observed state of AWSEncryptionKey
type AWSEncryptionKeyStatus struct {
	ID    string `json:"id,omitempty"`
	InUse bool   `json:"inUse,omitempty"`
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="InUse",type="string",JSONPath=".status.inUse"

// AWSEncryptionKey is the Schema for the awsencryptionkeys API
type AWSEncryptionKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSEncryptionKeySpec   `json:"spec,omitempty"`
	Status AWSEncryptionKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSEncryptionKeyList contains a list of AWSEncryptionKey
type AWSEncryptionKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSEncryptionKey `json:"items"`
}

func (aws *AWSEncryptionKey) GetJobID(jobName string) string {
	return aws.Kind + "/" + client.ObjectKeyFromObject(aws).String() + "/" + jobName
}

func (aws *AWSEncryptionKey) NewPatch() client.Patch {
	old := aws.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&AWSEncryptionKey{}, &AWSEncryptionKeyList{})
}
