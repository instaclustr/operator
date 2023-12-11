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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

// ClusterBackupSpec defines the desired state of ClusterBackup
type ClusterBackupSpec struct {
	ClusterKind string `json:"clusterKind"`
}

// ClusterBackupStatus defines the observed state of ClusterBackup
type ClusterBackupStatus struct {
	OperationStatus string `json:"operationStatus,omitempty"`
	Progress        string `json:"progress,omitempty"`
	Start           int    `json:"start,omitempty"`
	End             int    `json:"end,omitempty"`
	ClusterID       string `json:"clusterId,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterBackup is the Schema for the clusterbackups API
type ClusterBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterBackupSpec   `json:"spec,omitempty"`
	Status ClusterBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterBackupList contains a list of ClusterBackup
type ClusterBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBackup `json:"items"`
}

func (cb *ClusterBackup) NewPatch() client.Patch {
	old := cb.DeepCopy()
	return client.MergeFrom(old)
}

func (cbs *ClusterBackupStatus) UpdateStatus(instBackup *models.BackupEvent) {
	cbs.OperationStatus = instBackup.State
	cbs.End = instBackup.End
	cbs.Progress = fmt.Sprintf("%f", instBackup.Progress)
}

func (cb *ClusterBackup) AttachToCluster(id string) {
	cb.Status.ClusterID = id
}

func (cb *ClusterBackup) DetachFromCluster() {

}

func init() {
	SchemeBuilder.Register(&ClusterBackup{}, &ClusterBackupList{})
}
