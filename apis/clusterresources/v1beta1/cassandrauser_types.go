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

// CassandraUserSpec defines the desired state of CassandraUser
type CassandraUserSpec struct {
	// SecretRef references to the secret which stores user's credentials
	SecretRef *SecretReference `json:"secretRef"`
}

// CassandraUserStatus defines the observed state of CassandraUser
type CassandraUserStatus struct {
	// ClustersEvents efficiently stores and associates creation or deletion events with their respective cluster IDs.
	// The keys of the map represent the cluster IDs, while the corresponding values represent the events themselves.
	ClustersEvents map[string]string `json:"clustersEvents,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CassandraUser is the Schema for the cassandrausers API
type CassandraUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraUserSpec   `json:"spec,omitempty"`
	Status CassandraUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CassandraUserList contains a list of CassandraUser
type CassandraUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraUser `json:"items"`
}

func (r *CassandraUser) NewPatch() client.Patch {
	old := r.DeepCopy()
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&CassandraUser{}, &CassandraUserList{})
}

func (r *CassandraUser) ToInstAPI(username, password string) *models.InstaUser {
	return &models.InstaUser{
		Username:          username,
		Password:          password,
		InitialPermission: "standard",
	}
}

func (r *CassandraUser) GetDeletionFinalizer() string {
	return models.DeletionFinalizer + "_" + r.Namespace + "_" + r.Name
}

func (r *CassandraUser) GetClusterEvents() map[string]string {
	return r.Status.ClustersEvents
}

func (r *CassandraUser) SetClusterEvents(events map[string]string) {
	r.Status.ClustersEvents = events
}
