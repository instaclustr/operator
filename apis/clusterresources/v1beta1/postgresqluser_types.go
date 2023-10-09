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

// PostgreSQLUserSpec defines the desired state of PostgreSQLUser
type PostgreSQLUserSpec struct {
	SecretRef *SecretReference `json:"secretRef"`
}

// PostgreSQLUserStatus defines the observed state of PostgreSQLUser
type PostgreSQLUserStatus struct {
	// ClustersInfo efficiently stores data about clusters that related to this user.
	// The keys of the map represent the cluster IDs, values are cluster info that consists of default secret namespaced name or event.
	ClustersInfo map[string]ClusterInfo `json:"clustersInfo,omitempty"`
}

type ClusterInfo struct {
	DefaultSecretNamespacedName NamespacedName `json:"defaultSecretNamespacedName"`
	Event                       string         `json:"event,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQLUser is the Schema for the postgresqlusers API
type PostgreSQLUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLUserSpec   `json:"spec,omitempty"`
	Status PostgreSQLUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgreSQLUserList contains a list of PostgreSQLUser
type PostgreSQLUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQLUser `json:"items"`
}

func (r *PostgreSQLUser) NewPatch() client.Patch {
	old := r.DeepCopy()
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&PostgreSQLUser{}, &PostgreSQLUserList{})
}

func (r *PostgreSQLUser) ToInstAPI(username, password string) *models.InstaUser {
	return &models.InstaUser{
		Username:          username,
		Password:          password,
		InitialPermission: "standard",
	}
}

func (r *PostgreSQLUser) GetDeletionFinalizer() string {
	return models.DeletionFinalizer + "_" + r.Namespace + "_" + r.Name
}
