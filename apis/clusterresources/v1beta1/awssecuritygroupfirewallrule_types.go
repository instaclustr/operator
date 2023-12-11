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

// AWSSecurityGroupFirewallRuleSpec defines the desired state of AWSSecurityGroupFirewallRule
type AWSSecurityGroupFirewallRuleSpec struct {
	FirewallRuleSpec `json:",inline"`
	SecurityGroupID  string `json:"securityGroupId"`
}

// AWSSecurityGroupFirewallRuleStatus defines the observed state of AWSSecurityGroupFirewallRule
type AWSSecurityGroupFirewallRuleStatus struct {
	FirewallRuleStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"

// AWSSecurityGroupFirewallRule is the Schema for the awssecuritygroupfirewallrules API
type AWSSecurityGroupFirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSecurityGroupFirewallRuleSpec   `json:"spec,omitempty"`
	Status AWSSecurityGroupFirewallRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSSecurityGroupFirewallRuleList contains a list of AWSSecurityGroupFirewallRule
type AWSSecurityGroupFirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSecurityGroupFirewallRule `json:"items"`
}

func (fr *AWSSecurityGroupFirewallRule) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(fr).String() + "/" + jobName
}

func (fr *AWSSecurityGroupFirewallRule) NewPatch() client.Patch {
	old := fr.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&AWSSecurityGroupFirewallRule{}, &AWSSecurityGroupFirewallRuleList{})
}
