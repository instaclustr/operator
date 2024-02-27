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
	"encoding/json"

	clusterresource "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/apiextensions"
	"github.com/instaclustr/operator/pkg/models"
)

type CloudProviderSettings struct {
	CustomVirtualNetworkID    string `json:"customVirtualNetworkId,omitempty"`
	ResourceGroup             string `json:"resourceGroup,omitempty"`
	DiskEncryptionKey         string `json:"diskEncryptionKey,omitempty"`
	BackupBucket              string `json:"backupBucket,omitempty"`
	DisableSnapshotAutoExpiry bool   `json:"disableSnapshotAutoExpiry,omitempty"`
}

type RestoreCDCConfig struct {
	CustomVPCSettings *RestoreCustomVPCSettings `json:"customVpcSettings"`
	RestoreMode       string                    `json:"restoreMode"`
	CDCID             string                    `json:"cdcId"`
}

type RestoreCustomVPCSettings struct {
	VpcID   string `json:"vpcId"`
	Network string `json:"network"`
}

type ClusteredMaintenanceEvent struct {
	InProgress []*clusterresource.MaintenanceEventStatus `json:"inProgress"`
	Past       []*clusterresource.MaintenanceEventStatus `json:"past"`
	Upcoming   []*clusterresource.MaintenanceEventStatus `json:"upcoming"`
}

type OnPremisesSpec struct {
	EnableAutomation   bool       `json:"enableAutomation"`
	StorageClassName   string     `json:"storageClassName"`
	OSDiskSize         string     `json:"osDiskSize"`
	DataDiskSize       string     `json:"dataDiskSize"`
	SSHGatewayCPU      int64      `json:"sshGatewayCPU,omitempty"`
	SSHGatewayMemory   string     `json:"sshGatewayMemory,omitempty"`
	NodeCPU            int64      `json:"nodeCPU"`
	NodeMemory         string     `json:"nodeMemory"`
	OSImageURL         string     `json:"osImageURL"`
	CloudInitScriptRef *Reference `json:"cloudInitScriptRef"`
}

type TwoFactorDelete struct {
	// Email address which will be contacted when the cluster is requested to be deleted.
	Email string `json:"email"`

	Phone string `json:"phone,omitempty"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type PrivateLink struct {
	//+kubebuilder:validation:MinLength:=3
	AdvertisedHostname string `json:"advertisedHostname"`
}

// +kubebuilder:validation:MaxItems:=1
type PrivateLinkSpec []*PrivateLink

func (p PrivateLinkSpec) ToInstAPI() []*models.PrivateLink {
	instaModels := make([]*models.PrivateLink, 0, len(p))
	for _, pl := range p {
		instaModels = append(instaModels, &models.PrivateLink{
			AdvertisedHostname: pl.AdvertisedHostname,
		})
	}

	return instaModels
}

func (p *PrivateLinkSpec) FromInstAPI(o []*models.PrivateLink) {
	*p = make(PrivateLinkSpec, len(o))
	for i, instaModel := range o {
		(*p)[i] = &PrivateLink{
			AdvertisedHostname: instaModel.AdvertisedHostname,
		}
	}
}

type privateLinkStatus struct {
	AdvertisedHostname  string `json:"advertisedHostname"`
	EndPointServiceID   string `json:"endPointServiceId,omitempty"`
	EndPointServiceName string `json:"endPointServiceName,omitempty"`
}

type PrivateLinkStatuses []*privateLinkStatus

func (p1 PrivateLinkStatuses) Equal(p2 PrivateLinkStatuses) bool {
	if len(p1) != len(p2) {
		return false
	}

	for _, link2 := range p2 {
		for _, link1 := range p1 {
			if *link2 != *link1 {
				return false
			}
		}
	}

	return true
}

func (s PrivateLinkStatuses) ToInstAPI() []*models.PrivateLink {
	instaModels := make([]*models.PrivateLink, len(s))
	for i, link := range s {
		instaModels[i] = &models.PrivateLink{
			AdvertisedHostname: link.AdvertisedHostname,
		}
	}

	return instaModels
}

func (p *PrivateLinkStatuses) FromInstAPI(instaModels []*models.PrivateLink) {
	*p = make(PrivateLinkStatuses, len(instaModels))
	for i, instaModel := range instaModels {
		(*p)[i] = &privateLinkStatus{
			AdvertisedHostname:  instaModel.AdvertisedHostname,
			EndPointServiceID:   instaModel.EndPointServiceID,
			EndPointServiceName: instaModel.EndPointServiceName,
		}
	}
}

func privateLinksToInstAPI(p []*PrivateLink) []*models.PrivateLink {
	links := make([]*models.PrivateLink, 0, len(p))
	for _, link := range p {
		links = append(links, &models.PrivateLink{
			AdvertisedHostname: link.AdvertisedHostname,
		})
	}

	return links
}

type immutableCluster struct {
	Name                  string
	Version               string
	PCICompliance         bool
	PrivateNetworkCluster bool
	SLATier               string
}

type immutableDC struct {
	Name                string
	Region              string
	CloudProvider       string
	ProviderAccountName string
	Network             string
}

type ResizeOperation struct {
	// Number of nodes that can be concurrently resized at a given time
	ConcurrentResizes int `json:"concurrentResizes,omitempty"`
	// Replace operations
	ReplaceOperations []*ReplaceOperation `json:"replaceOperations,omitempty"`
	// Timestamp of the creation of the operation
	Created string `json:"created,omitempty"`
	// Timestamp of the completion of the operation
	Completed string `json:"completed,omitempty"`
	// ID of the operation
	ID string `json:"id,omitempty"`
	// New size of the node
	NewNodeSize string `json:"newNodeSize,omitempty"`
	// Timestamp of when Instaclustr Support has been alerted to the resize operation.
	InstaclustrSupportAlerted string `json:"instaclustrSupportAlerted,omitempty"`
	// Purpose of the node
	NodePurpose string `json:"nodePurpose,omitempty"`
	// Status of the operation
	Status string `json:"status,omitempty"`
}

type ResizeSettings struct {
	NotifySupportContacts bool `json:"notifySupportContacts,omitempty"`
	Concurrency           int  `json:"concurrency,omitempty"`
}

func resizeSettingsToInstAPI(rss []*ResizeSettings) []*models.ResizeSettings {
	iRS := make([]*models.ResizeSettings, 0, len(rss))

	for _, rs := range rss {
		iRS = append(iRS, &models.ResizeSettings{
			NotifySupportContacts: rs.NotifySupportContacts,
			Concurrency:           rs.Concurrency,
		})
	}

	return iRS
}

type ReplaceOperation struct {
	// ID of the new node in the replacement operation
	NewNodeID string `json:"newNodeId,omitempty"`
	// Timestamp of the creation of the node replacement operation
	Created string `json:"created,omitempty"`
	// ID of the node replacement operation
	ID string `json:"id,omitempty"`
	// ID of the node being replaced
	NodeID string `json:"nodeId,omitempty"`
	// Status of the node replacement operation
	Status string `json:"status,omitempty"`
}

func (tfd *TwoFactorDelete) ToInstAPI() *models.TwoFactorDelete {
	return &models.TwoFactorDelete{
		ConfirmationPhoneNumber: tfd.Phone,
		ConfirmationEmail:       tfd.Email,
	}
}

func (cps *CloudProviderSettings) AWSToInstAPI() *models.AWSSetting {
	return &models.AWSSetting{
		EBSEncryptionKey:       cps.DiskEncryptionKey,
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
		BackupBucket:           cps.BackupBucket,
	}
}

func (cps *CloudProviderSettings) AzureToInstAPI() *models.AzureSetting {
	return &models.AzureSetting{
		ResourceGroup: cps.ResourceGroup,
	}
}

func (cps *CloudProviderSettings) GCPToInstAPI() *models.GCPSetting {
	return &models.GCPSetting{
		CustomVirtualNetworkID:    cps.CustomVirtualNetworkID,
		DisableSnapshotAutoExpiry: cps.DisableSnapshotAutoExpiry,
	}
}

func areEventStatusesEqual(a, b *clusterresource.ClusteredMaintenanceEventStatus) bool {
	if len(a.Past) != len(b.Past) ||
		len(a.InProgress) != len(b.InProgress) ||
		len(a.Upcoming) != len(b.Upcoming) {
		return false
	}

	for i := range a.Past {
		if a.Past[i].ID != b.Past[i].ID {
			continue
		}
		if !areClusteredMaintenanceEventStatusEqual(a.Past[i], b.Past[i]) {
			return false
		}
	}

	for i := range a.InProgress {
		if a.InProgress[i].ID != b.InProgress[i].ID {
			continue
		}
		if !areClusteredMaintenanceEventStatusEqual(a.InProgress[i], b.InProgress[i]) {
			return false
		}
	}

	for i := range a.Upcoming {
		if a.Upcoming[i].ID != b.Upcoming[i].ID {
			continue
		}
		if !areClusteredMaintenanceEventStatusEqual(a.Upcoming[i], b.Upcoming[i]) {
			return false
		}
	}
	return true
}

func areClusteredMaintenanceEventStatusEqual(a, b *clusterresource.MaintenanceEventStatus) bool {
	return a.Description == b.Description &&
		a.ScheduledStartTime == b.ScheduledStartTime &&
		a.ScheduledEndTime == b.ScheduledEndTime &&
		a.ScheduledStartTimeMax == b.ScheduledStartTimeMax &&
		a.ScheduledStartTimeMin == b.ScheduledStartTimeMin &&
		a.IsFinalized == b.IsFinalized &&
		a.StartTime == b.StartTime &&
		a.EndTime == b.EndTime &&
		a.Outcome == b.Outcome
}

// +kubebuilder:object:generate:=false
type Reference = apiextensions.ObjectReference

type References []*apiextensions.ObjectReference

// Diff returns difference between two References.
// Added stores elements which are presented in new References, but aren't presented in old.
// Deleted stores elements which aren't presented in new References, but are presented in old.
func (old References) Diff(new References) (added, deleted References) {
	// filtering deleted references
	for _, oldRef := range old {
		var exists bool
		for _, newRef := range new {
			if *oldRef == *newRef {
				exists = true
			}
		}

		if !exists {
			deleted = append(deleted, oldRef)
		}
	}

	// filtering added references
	for _, newRef := range new {
		var exists bool
		for _, oldRef := range old {
			if *newRef == *oldRef {
				exists = true
			}
		}

		if !exists {
			added = append(added, newRef)
		}
	}

	return added, deleted
}

// +kubebuilder:validation:MaxItems:=1
type GenericResizeSettings []*ResizeSettings

func (g *GenericResizeSettings) FromInstAPI(instModels []*models.ResizeSettings) {
	*g = make(GenericResizeSettings, len(instModels))
	for i, instModel := range instModels {
		(*g)[i] = &ResizeSettings{
			NotifySupportContacts: instModel.NotifySupportContacts,
			Concurrency:           instModel.Concurrency,
		}
	}
}

func (g *GenericResizeSettings) ToInstAPI() []*models.ResizeSettings {
	instaModels := make([]*models.ResizeSettings, len(*g))
	for i, setting := range *g {
		instaModels[i] = &models.ResizeSettings{
			NotifySupportContacts: setting.NotifySupportContacts,
			Concurrency:           setting.Concurrency,
		}
	}

	return instaModels
}

func (g GenericResizeSettings) Equal(o GenericResizeSettings) bool {
	if len(g) != len(o) {
		return false
	}

	if len(g) > 0 {
		return *g[0] == *o[0]
	}

	return true
}

type AWSSettings struct {
	// ID of a KMS encryption key to encrypt data on nodes.
	// KMS encryption key must be set in Cluster Resources through
	//the Instaclustr Console before provisioning an encrypted Data Centre.
	DiskEncryptionKey string `json:"encryptionKey,omitempty"`

	// VPC ID into which the Data Centre will be provisioned.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`

	// Specify the S3 bucket to use for storing backup data for the cluster data centre.
	// Only available for customers running in their own cloud provider accounts.
	// Currently supported for OpenSearch clusters only.
	BackupBucket string `json:"backupBucket,omitempty"`
}

type GCPSettings struct {
	// Network name or a relative Network or Subnetwork URI.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet.
	//
	// Examples:
	// Network URI: projects/{riyoa-gcp-project-name}/global/networks/{network-name}.
	// Network name: {network-name}, equivalent to projects/{riyoa-gcp-project-name}/global/networks/{network-name}.
	// Same-project subnetwork URI: projects/{riyoa-gcp-project-name}/regions/{region-id}/subnetworks/{subnetwork-name}.
	// Shared VPC subnetwork URI: projects/{riyoa-gcp-host-project-name}/regions/{region-id}/subnetworks/{subnetwork-name}.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`

	// Specify whether the GCS backup bucket should automatically expire data after 7 days or not.
	// Setting this to true will disable automatic expiry and will allow for creation of custom snapshot
	// repositories with customisable retention using the Index Management Plugin.
	// The storage will have to be manually cleared after the cluster is deleted.
	// Only available for customers running in their own cloud provider accounts.
	// Currently supported for OpenSearch clusters only.
	DisableSnapshotAutoExpiry bool `json:"disableSnapshotAutoExpiry,omitempty"`
}

type AzureSettings struct {
	// The name of the Azure Resource Group into which the Data Centre will be provisioned.
	ResourceGroup string `json:"resourceGroup,omitempty"`

	// VNet ID into which the Data Centre will be provisioned.
	// The VNet must have an available address space for the Data Centre's network
	// allocation to be appended to the VNet.
	// Currently supported for PostgreSQL clusters only.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`

	// The private network address block to be used for the storage network.
	// This is only used for certain node sizes, currently limited to those which use Azure NetApp Files:
	// for all other node sizes, this field should not be provided.
	// The network must have a prefix length between /16 and /28, and must be part of a private address range.
	StorageNetwork string `json:"storageNetwork,omitempty"`
}

func nodesEqual(s1, s2 []*Node) bool {
	if len(s1) != len(s2) {
		return false
	}

	m := map[string]*Node{}
	for _, node := range s1 {
		m[node.ID] = node
	}

	for _, s2Node := range s2 {
		s1Node, ok := m[s2Node.ID]
		if !ok {
			return false
		}

		if !s1Node.Equals(s2Node) {
			return false
		}
	}

	return true
}

func nodesFromInstAPI(instaModels []*models.Node) []*Node {
	nodes := make([]*Node, len(instaModels))
	for i, instaModel := range instaModels {
		n := Node{}
		n.FromInstAPI(instaModel)
		nodes[i] = &n
	}

	return nodes
}
