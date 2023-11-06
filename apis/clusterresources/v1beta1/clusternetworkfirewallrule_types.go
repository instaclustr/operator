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

// ClusterNetworkFirewallRuleSpec defines the desired state of ClusterNetworkFirewallRule
type ClusterNetworkFirewallRuleSpec struct {
	FirewallRuleSpec `json:",inline"`
	Network          string `json:"network"`
}

// ClusterNetworkFirewallRuleStatus defines the observed state of ClusterNetworkFirewallRule
type ClusterNetworkFirewallRuleStatus struct {
	FirewallRuleStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterNetworkFirewallRule is the Schema for the clusternetworkfirewallrules API
type ClusterNetworkFirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterNetworkFirewallRuleSpec   `json:"spec,omitempty"`
	Status ClusterNetworkFirewallRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterNetworkFirewallRuleList contains a list of ClusterNetworkFirewallRule
type ClusterNetworkFirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterNetworkFirewallRule `json:"items"`
}

func (fr *ClusterNetworkFirewallRule) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(fr).String() + "/" + jobName
}

func (fr *ClusterNetworkFirewallRule) NewPatch() client.Patch {
	old := fr.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (fr *ClusterNetworkFirewallRule) AttachToCluster(id string) {
	fr.Status.ClusterID = id
	fr.Status.ResourceState = models.CreatingEvent
}

func (fr *ClusterNetworkFirewallRule) DetachFromCluster() {
	fr.Status.ResourceState = models.DeletingEvent
}

func init() {
	SchemeBuilder.Register(&ClusterNetworkFirewallRule{}, &ClusterNetworkFirewallRuleList{})
}
