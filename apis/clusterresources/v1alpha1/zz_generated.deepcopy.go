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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

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
	out.VPCPeeringStatus = in.VPCPeeringStatus
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
	out.VPCPeeringStatus = in.VPCPeeringStatus
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
	out.VPCPeeringStatus = in.VPCPeeringStatus
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCPeeringStatus) DeepCopyInto(out *VPCPeeringStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCPeeringStatus.
func (in *VPCPeeringStatus) DeepCopy() *VPCPeeringStatus {
	if in == nil {
		return nil
	}
	out := new(VPCPeeringStatus)
	in.DeepCopyInto(out)
	return out
}
