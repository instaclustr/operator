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

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

// KafkaUserSpec defines the desired state of KafkaUser
type KafkaUserSpec struct {
	Options            *KafkaUserOptions        `json:"options"`
	SecretRef          *v1beta1.SecretReference `json:"secretRef"`
	InitialPermissions string                   `json:"initialPermissions"`
}

type KafkaUserOptions struct {
	OverrideExistingUser bool   `json:"overrideExistingUser,omitempty"`
	SASLSCRAMMechanism   string `json:"saslScramMechanism"`
}

// KafkaUserStatus defines the observed state of KafkaUser
type KafkaUserStatus struct {
	ClustersEvents map[string]string `json:"clustersEvents,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaUser is the Schema for the kafkausers API
type KafkaUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaUserSpec   `json:"spec,omitempty"`
	Status KafkaUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaUserList contains a list of KafkaUser
type KafkaUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaUser `json:"items"`
}

func (ku *KafkaUser) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(ku).String() + "/" + jobName
}

func (ku *KafkaUser) NewPatch() client.Patch {
	old := ku.DeepCopy()
	return client.MergeFrom(old)
}

func (ku *KafkaUser) GetDeletionFinalizer() string {
	return models.DeletionFinalizer + "_" + ku.Namespace + "_" + ku.Name
}

func (ku *KafkaUser) GetDeletionUserFinalizer(clusterID string) string {
	return models.DeletionUserFinalizer + clusterID
}

func (ku *KafkaUser) GetID(clusterID, name string) string {
	return clusterID + "_" + name
}

func init() {
	SchemeBuilder.Register(&KafkaUser{}, &KafkaUserList{})
}

func (ks *KafkaUserSpec) ToInstAPI(clusterID string, username string, password string) *models.KafkaUser {
	return &models.KafkaUser{
		ClusterID:          clusterID,
		InitialPermissions: ks.InitialPermissions,
		Options:            ks.Options.ToInstAPI(),
		Username:           username,
		Password:           password,
	}
}

func (ko *KafkaUserOptions) ToInstAPI() *models.KafkaUserOptions {
	return &models.KafkaUserOptions{
		OverrideExistingUser: ko.OverrideExistingUser,
		SASLSCRAMMechanism:   ko.SASLSCRAMMechanism,
	}

}
