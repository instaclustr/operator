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

// OpenSearchUserSpec defines the desired state of OpenSearchUser
type OpenSearchUserSpec struct {
	SecretRef *SecretReference `json:"secretRef"`
}

// OpenSearchUserStatus defines the observed state of OpenSearchUser
type OpenSearchUserStatus struct {
	ClustersEvents map[string]string `json:"clustersEvents,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenSearchUser is the Schema for the opensearchusers API
type OpenSearchUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenSearchUserSpec   `json:"spec,omitempty"`
	Status OpenSearchUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenSearchUserList contains a list of OpenSearchUser
type OpenSearchUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenSearchUser `json:"items"`
}

func (u *OpenSearchUser) ToInstaAPI(username, password string) *models.InstaOpenSearchUser {
	return &models.InstaOpenSearchUser{
		InstaUser: &models.InstaUser{
			Username:          username,
			Password:          password,
			InitialPermission: "standard",
		},
	}
}

func (u *OpenSearchUser) NewPatch() client.Patch {
	old := u.DeepCopy()
	return client.MergeFrom(old)
}

func (u *OpenSearchUser) GetDeletionFinalizer() string {
	return models.DeletionFinalizer + "_" + u.Namespace + "_" + u.Name
}

func (u *OpenSearchUser) GetClusterEvents() map[string]string {
	return u.Status.ClustersEvents
}

func (u *OpenSearchUser) SetClusterEvents(events map[string]string) {
	u.Status.ClustersEvents = events
}

func init() {
	SchemeBuilder.Register(&OpenSearchUser{}, &OpenSearchUserList{})
}
