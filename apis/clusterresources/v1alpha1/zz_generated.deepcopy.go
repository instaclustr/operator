//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEncryptionKey) DeepCopyInto(out *AWSEncryptionKey) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEncryptionKey.
func (in *AWSEncryptionKey) DeepCopy() *AWSEncryptionKey {
	if in == nil {
		return nil
	}
	out := new(AWSEncryptionKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSEncryptionKey) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEncryptionKeyList) DeepCopyInto(out *AWSEncryptionKeyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSEncryptionKey, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEncryptionKeyList.
func (in *AWSEncryptionKeyList) DeepCopy() *AWSEncryptionKeyList {
	if in == nil {
		return nil
	}
	out := new(AWSEncryptionKeyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSEncryptionKeyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEncryptionKeySpec) DeepCopyInto(out *AWSEncryptionKeySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEncryptionKeySpec.
func (in *AWSEncryptionKeySpec) DeepCopy() *AWSEncryptionKeySpec {
	if in == nil {
		return nil
	}
	out := new(AWSEncryptionKeySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSEncryptionKeyStatus) DeepCopyInto(out *AWSEncryptionKeyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSEncryptionKeyStatus.
func (in *AWSEncryptionKeyStatus) DeepCopy() *AWSEncryptionKeyStatus {
	if in == nil {
		return nil
	}
	out := new(AWSEncryptionKeyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSecurityGroupFirewallRule) DeepCopyInto(out *AWSSecurityGroupFirewallRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSecurityGroupFirewallRule.
func (in *AWSSecurityGroupFirewallRule) DeepCopy() *AWSSecurityGroupFirewallRule {
	if in == nil {
		return nil
	}
	out := new(AWSSecurityGroupFirewallRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSecurityGroupFirewallRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSecurityGroupFirewallRuleList) DeepCopyInto(out *AWSSecurityGroupFirewallRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSSecurityGroupFirewallRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSecurityGroupFirewallRuleList.
func (in *AWSSecurityGroupFirewallRuleList) DeepCopy() *AWSSecurityGroupFirewallRuleList {
	if in == nil {
		return nil
	}
	out := new(AWSSecurityGroupFirewallRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSecurityGroupFirewallRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSecurityGroupFirewallRuleSpec) DeepCopyInto(out *AWSSecurityGroupFirewallRuleSpec) {
	*out = *in
	out.FirewallRuleSpec = in.FirewallRuleSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSecurityGroupFirewallRuleSpec.
func (in *AWSSecurityGroupFirewallRuleSpec) DeepCopy() *AWSSecurityGroupFirewallRuleSpec {
	if in == nil {
		return nil
	}
	out := new(AWSSecurityGroupFirewallRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSecurityGroupFirewallRuleStatus) DeepCopyInto(out *AWSSecurityGroupFirewallRuleStatus) {
	*out = *in
	out.FirewallRuleStatus = in.FirewallRuleStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSecurityGroupFirewallRuleStatus.
func (in *AWSSecurityGroupFirewallRuleStatus) DeepCopy() *AWSSecurityGroupFirewallRuleStatus {
	if in == nil {
		return nil
	}
	out := new(AWSSecurityGroupFirewallRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSVPCPeering) DeepCopyInto(out *AWSVPCPeering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSVPCPeering.
func (in *AWSVPCPeering) DeepCopy() *AWSVPCPeering {
	if in == nil {
		return nil
	}
	out := new(AWSVPCPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSVPCPeering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSVPCPeeringList) DeepCopyInto(out *AWSVPCPeeringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSVPCPeering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSVPCPeeringList.
func (in *AWSVPCPeeringList) DeepCopy() *AWSVPCPeeringList {
	if in == nil {
		return nil
	}
	out := new(AWSVPCPeeringList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSVPCPeeringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSVPCPeeringSpec) DeepCopyInto(out *AWSVPCPeeringSpec) {
	*out = *in
	in.VPCPeeringSpec.DeepCopyInto(&out.VPCPeeringSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSVPCPeeringSpec.
func (in *AWSVPCPeeringSpec) DeepCopy() *AWSVPCPeeringSpec {
	if in == nil {
		return nil
	}
	out := new(AWSVPCPeeringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSVPCPeeringStatus) DeepCopyInto(out *AWSVPCPeeringStatus) {
	*out = *in
	out.PeeringStatus = in.PeeringStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSVPCPeeringStatus.
func (in *AWSVPCPeeringStatus) DeepCopy() *AWSVPCPeeringStatus {
	if in == nil {
		return nil
	}
	out := new(AWSVPCPeeringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureVNetPeering) DeepCopyInto(out *AzureVNetPeering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureVNetPeering.
func (in *AzureVNetPeering) DeepCopy() *AzureVNetPeering {
	if in == nil {
		return nil
	}
	out := new(AzureVNetPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AzureVNetPeering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureVNetPeeringList) DeepCopyInto(out *AzureVNetPeeringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AzureVNetPeering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureVNetPeeringList.
func (in *AzureVNetPeeringList) DeepCopy() *AzureVNetPeeringList {
	if in == nil {
		return nil
	}
	out := new(AzureVNetPeeringList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AzureVNetPeeringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureVNetPeeringSpec) DeepCopyInto(out *AzureVNetPeeringSpec) {
	*out = *in
	in.VPCPeeringSpec.DeepCopyInto(&out.VPCPeeringSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureVNetPeeringSpec.
func (in *AzureVNetPeeringSpec) DeepCopy() *AzureVNetPeeringSpec {
	if in == nil {
		return nil
	}
	out := new(AzureVNetPeeringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureVNetPeeringStatus) DeepCopyInto(out *AzureVNetPeeringStatus) {
	*out = *in
	out.PeeringStatus = in.PeeringStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureVNetPeeringStatus.
func (in *AzureVNetPeeringStatus) DeepCopy() *AzureVNetPeeringStatus {
	if in == nil {
		return nil
	}
	out := new(AzureVNetPeeringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraUser) DeepCopyInto(out *CassandraUser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraUser.
func (in *CassandraUser) DeepCopy() *CassandraUser {
	if in == nil {
		return nil
	}
	out := new(CassandraUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CassandraUser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraUserList) DeepCopyInto(out *CassandraUserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CassandraUser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraUserList.
func (in *CassandraUserList) DeepCopy() *CassandraUserList {
	if in == nil {
		return nil
	}
	out := new(CassandraUserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CassandraUserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraUserSpec) DeepCopyInto(out *CassandraUserSpec) {
	*out = *in
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraUserSpec.
func (in *CassandraUserSpec) DeepCopy() *CassandraUserSpec {
	if in == nil {
		return nil
	}
	out := new(CassandraUserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CassandraUserStatus) DeepCopyInto(out *CassandraUserStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CassandraUserStatus.
func (in *CassandraUserStatus) DeepCopy() *CassandraUserStatus {
	if in == nil {
		return nil
	}
	out := new(CassandraUserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBackup) DeepCopyInto(out *ClusterBackup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterBackup.
func (in *ClusterBackup) DeepCopy() *ClusterBackup {
	if in == nil {
		return nil
	}
	out := new(ClusterBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterBackup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBackupList) DeepCopyInto(out *ClusterBackupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterBackup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterBackupList.
func (in *ClusterBackupList) DeepCopy() *ClusterBackupList {
	if in == nil {
		return nil
	}
	out := new(ClusterBackupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterBackupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBackupSpec) DeepCopyInto(out *ClusterBackupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterBackupSpec.
func (in *ClusterBackupSpec) DeepCopy() *ClusterBackupSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterBackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterBackupStatus) DeepCopyInto(out *ClusterBackupStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterBackupStatus.
func (in *ClusterBackupStatus) DeepCopy() *ClusterBackupStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterBackupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkFirewallRule) DeepCopyInto(out *ClusterNetworkFirewallRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkFirewallRule.
func (in *ClusterNetworkFirewallRule) DeepCopy() *ClusterNetworkFirewallRule {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkFirewallRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterNetworkFirewallRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkFirewallRuleList) DeepCopyInto(out *ClusterNetworkFirewallRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterNetworkFirewallRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkFirewallRuleList.
func (in *ClusterNetworkFirewallRuleList) DeepCopy() *ClusterNetworkFirewallRuleList {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkFirewallRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterNetworkFirewallRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkFirewallRuleSpec) DeepCopyInto(out *ClusterNetworkFirewallRuleSpec) {
	*out = *in
	out.FirewallRuleSpec = in.FirewallRuleSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkFirewallRuleSpec.
func (in *ClusterNetworkFirewallRuleSpec) DeepCopy() *ClusterNetworkFirewallRuleSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkFirewallRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkFirewallRuleStatus) DeepCopyInto(out *ClusterNetworkFirewallRuleStatus) {
	*out = *in
	out.FirewallRuleStatus = in.FirewallRuleStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkFirewallRuleStatus.
func (in *ClusterNetworkFirewallRuleStatus) DeepCopy() *ClusterNetworkFirewallRuleStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkFirewallRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExclusionWindowSpec) DeepCopyInto(out *ExclusionWindowSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExclusionWindowSpec.
func (in *ExclusionWindowSpec) DeepCopy() *ExclusionWindowSpec {
	if in == nil {
		return nil
	}
	out := new(ExclusionWindowSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExclusionWindowStatus) DeepCopyInto(out *ExclusionWindowStatus) {
	*out = *in
	out.ExclusionWindowSpec = in.ExclusionWindowSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExclusionWindowStatus.
func (in *ExclusionWindowStatus) DeepCopy() *ExclusionWindowStatus {
	if in == nil {
		return nil
	}
	out := new(ExclusionWindowStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FirewallRuleSpec) DeepCopyInto(out *FirewallRuleSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FirewallRuleSpec.
func (in *FirewallRuleSpec) DeepCopy() *FirewallRuleSpec {
	if in == nil {
		return nil
	}
	out := new(FirewallRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FirewallRuleStatus) DeepCopyInto(out *FirewallRuleStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FirewallRuleStatus.
func (in *FirewallRuleStatus) DeepCopy() *FirewallRuleStatus {
	if in == nil {
		return nil
	}
	out := new(FirewallRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPVPCPeering) DeepCopyInto(out *GCPVPCPeering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPVPCPeering.
func (in *GCPVPCPeering) DeepCopy() *GCPVPCPeering {
	if in == nil {
		return nil
	}
	out := new(GCPVPCPeering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GCPVPCPeering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPVPCPeeringList) DeepCopyInto(out *GCPVPCPeeringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GCPVPCPeering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPVPCPeeringList.
func (in *GCPVPCPeeringList) DeepCopy() *GCPVPCPeeringList {
	if in == nil {
		return nil
	}
	out := new(GCPVPCPeeringList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GCPVPCPeeringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPVPCPeeringSpec) DeepCopyInto(out *GCPVPCPeeringSpec) {
	*out = *in
	in.VPCPeeringSpec.DeepCopyInto(&out.VPCPeeringSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPVPCPeeringSpec.
func (in *GCPVPCPeeringSpec) DeepCopy() *GCPVPCPeeringSpec {
	if in == nil {
		return nil
	}
	out := new(GCPVPCPeeringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPVPCPeeringStatus) DeepCopyInto(out *GCPVPCPeeringStatus) {
	*out = *in
	out.PeeringStatus = in.PeeringStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPVPCPeeringStatus.
func (in *GCPVPCPeeringStatus) DeepCopy() *GCPVPCPeeringStatus {
	if in == nil {
		return nil
	}
	out := new(GCPVPCPeeringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEventRescheduleSpec) DeepCopyInto(out *MaintenanceEventRescheduleSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEventRescheduleSpec.
func (in *MaintenanceEventRescheduleSpec) DeepCopy() *MaintenanceEventRescheduleSpec {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEventRescheduleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEventStatus) DeepCopyInto(out *MaintenanceEventStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEventStatus.
func (in *MaintenanceEventStatus) DeepCopy() *MaintenanceEventStatus {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEventStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEvents) DeepCopyInto(out *MaintenanceEvents) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEvents.
func (in *MaintenanceEvents) DeepCopy() *MaintenanceEvents {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEvents)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MaintenanceEvents) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEventsList) DeepCopyInto(out *MaintenanceEventsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MaintenanceEvents, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEventsList.
func (in *MaintenanceEventsList) DeepCopy() *MaintenanceEventsList {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEventsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MaintenanceEventsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEventsSpec) DeepCopyInto(out *MaintenanceEventsSpec) {
	*out = *in
	if in.ExclusionWindows != nil {
		in, out := &in.ExclusionWindows, &out.ExclusionWindows
		*out = make([]*ExclusionWindowSpec, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ExclusionWindowSpec)
				**out = **in
			}
		}
	}
	if in.MaintenanceEventsReschedules != nil {
		in, out := &in.MaintenanceEventsReschedules, &out.MaintenanceEventsReschedules
		*out = make([]*MaintenanceEventRescheduleSpec, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(MaintenanceEventRescheduleSpec)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEventsSpec.
func (in *MaintenanceEventsSpec) DeepCopy() *MaintenanceEventsSpec {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEventsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaintenanceEventsStatus) DeepCopyInto(out *MaintenanceEventsStatus) {
	*out = *in
	if in.EventsStatuses != nil {
		in, out := &in.EventsStatuses, &out.EventsStatuses
		*out = make([]*MaintenanceEventStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(MaintenanceEventStatus)
				**out = **in
			}
		}
	}
	if in.ExclusionWindowsStatuses != nil {
		in, out := &in.ExclusionWindowsStatuses, &out.ExclusionWindowsStatuses
		*out = make([]*ExclusionWindowStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ExclusionWindowStatus)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaintenanceEventsStatus.
func (in *MaintenanceEventsStatus) DeepCopy() *MaintenanceEventsStatus {
	if in == nil {
		return nil
	}
	out := new(MaintenanceEventsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Node) DeepCopyInto(out *Node) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Node.
func (in *Node) DeepCopy() *Node {
	if in == nil {
		return nil
	}
	out := new(Node)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeReload) DeepCopyInto(out *NodeReload) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeReload.
func (in *NodeReload) DeepCopy() *NodeReload {
	if in == nil {
		return nil
	}
	out := new(NodeReload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeReload) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeReloadList) DeepCopyInto(out *NodeReloadList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeReload, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeReloadList.
func (in *NodeReloadList) DeepCopy() *NodeReloadList {
	if in == nil {
		return nil
	}
	out := new(NodeReloadList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeReloadList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeReloadSpec) DeepCopyInto(out *NodeReloadSpec) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]*Node, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Node)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeReloadSpec.
func (in *NodeReloadSpec) DeepCopy() *NodeReloadSpec {
	if in == nil {
		return nil
	}
	out := new(NodeReloadSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeReloadStatus) DeepCopyInto(out *NodeReloadStatus) {
	*out = *in
	out.NodeInProgress = in.NodeInProgress
	if in.CurrentOperationStatus != nil {
		in, out := &in.CurrentOperationStatus, &out.CurrentOperationStatus
		*out = new(Operation)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeReloadStatus.
func (in *NodeReloadStatus) DeepCopy() *NodeReloadStatus {
	if in == nil {
		return nil
	}
	out := new(NodeReloadStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Operation) DeepCopyInto(out *Operation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Operation.
func (in *Operation) DeepCopy() *Operation {
	if in == nil {
		return nil
	}
	out := new(Operation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PatchRequest) DeepCopyInto(out *PatchRequest) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = make(json.RawMessage, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PatchRequest.
func (in *PatchRequest) DeepCopy() *PatchRequest {
	if in == nil {
		return nil
	}
	out := new(PatchRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PeeringStatus) DeepCopyInto(out *PeeringStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PeeringStatus.
func (in *PeeringStatus) DeepCopy() *PeeringStatus {
	if in == nil {
		return nil
	}
	out := new(PeeringStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisUser) DeepCopyInto(out *RedisUser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisUser.
func (in *RedisUser) DeepCopy() *RedisUser {
	if in == nil {
		return nil
	}
	out := new(RedisUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisUser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisUserList) DeepCopyInto(out *RedisUserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RedisUser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisUserList.
func (in *RedisUserList) DeepCopy() *RedisUserList {
	if in == nil {
		return nil
	}
	out := new(RedisUserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RedisUserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisUserSpec) DeepCopyInto(out *RedisUserSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisUserSpec.
func (in *RedisUserSpec) DeepCopy() *RedisUserSpec {
	if in == nil {
		return nil
	}
	out := new(RedisUserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisUserStatus) DeepCopyInto(out *RedisUserStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisUserStatus.
func (in *RedisUserStatus) DeepCopy() *RedisUserStatus {
	if in == nil {
		return nil
	}
	out := new(RedisUserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretReference) DeepCopyInto(out *SecretReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretReference.
func (in *SecretReference) DeepCopy() *SecretReference {
	if in == nil {
		return nil
	}
	out := new(SecretReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCPeeringSpec) DeepCopyInto(out *VPCPeeringSpec) {
	*out = *in
	if in.PeerSubnets != nil {
		in, out := &in.PeerSubnets, &out.PeerSubnets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCPeeringSpec.
func (in *VPCPeeringSpec) DeepCopy() *VPCPeeringSpec {
	if in == nil {
		return nil
	}
	out := new(VPCPeeringSpec)
	in.DeepCopyInto(out)
	return out
}
