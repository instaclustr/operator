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

// KafkaACLSpec defines the desired state of KafkaACL
type KafkaACLSpec struct {
	ACLs      []ACL  `json:"acls"`
	UserQuery string `json:"userQuery"`
	ClusterID string `json:"clusterId"`
}

type ACL struct {
	Principal      string `json:"principal"`
	PermissionType string `json:"permissionType"`
	Host           string `json:"host"`
	PatternType    string `json:"patternType"`
	ResourceName   string `json:"resourceName"`
	Operation      string `json:"operation"`
	ResourceType   string `json:"resourceType"`
}

// KafkaACLStatus defines the observed state of KafkaACL
type KafkaACLStatus struct {
	ID string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"

// KafkaACL is the Schema for the kafkaacls API
type KafkaACL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaACLSpec   `json:"spec,omitempty"`
	Status KafkaACLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaACLList contains a list of KafkaACL
type KafkaACLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaACL `json:"items"`
}

func (kacl *KafkaACL) GetJobID(jobName string) string {
	return kacl.Kind + "/" + client.ObjectKeyFromObject(kacl).String() + "/" + jobName
}

func (kacl *KafkaACL) NewPatch() client.Patch {
	old := kacl.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&KafkaACL{}, &KafkaACLList{})
}
