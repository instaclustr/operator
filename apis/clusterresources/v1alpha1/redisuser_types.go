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

// RedisUserSpec defines the desired state of RedisUser
type RedisUserSpec struct {
	ClusterID               string `json:"clusterId"`
	Username                string `json:"username"`
	PasswordSecretName      string `json:"passwordSecretName"`
	PasswordSecretNamespace string `json:"passwordSecretNamespace"`
	InitialPermissions      string `json:"initialPermissions"`
}

// RedisUserStatus defines the observed state of RedisUser
type RedisUserStatus struct {
	ID string `json:"ID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisUser is the Schema for the redisusers API
type RedisUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisUserSpec   `json:"spec,omitempty"`
	Status RedisUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisUserList contains a list of RedisUser
type RedisUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisUser `json:"items"`
}

func (rs *RedisUserSpec) ToInstAPI(password string) *models.RedisUser {
	return &models.RedisUser{
		ClusterID:          rs.ClusterID,
		Username:           rs.Username,
		Password:           password,
		InitialPermissions: rs.InitialPermissions,
	}
}

func (r *RedisUser) ToInstAPIUpdate(password string) *models.RedisUserUpdate {
	return &models.RedisUserUpdate{
		ID:       r.Status.ID,
		Password: password,
	}
}

func (r *RedisUser) NewPatch() client.Patch {
	old := r.DeepCopy()
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&RedisUser{}, &RedisUserList{})
}
