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

package models

const (
	NoOperation         = "NO_OPERATION"
	OperationInProgress = "OPERATION_IN_PROGRESS"

	CreatedStatus = "CREATED"
	DeletedStatus = "DELETED"

	DefaultAccountName = "INSTACLUSTR"

	AWSVPC  = "AWS_VPC"
	GCP     = "GCP"
	AZUREAZ = "AZURE_AZ"
)

type ClusterStatus struct {
	ID                            string `json:"id,omitempty"`
	Status                        string `json:"status,omitempty"`
	CurrentClusterOperationStatus string `json:"currentClusterOperationStatus,omitempty"`
}

type DataCentre struct {
	DataCentreStatus    `json:",inline"`
	Name                string          `json:"name"`
	Network             string          `json:"network"`
	NodeSize            string          `json:"nodeSize,omitempty"`
	NumberOfNodes       int             `json:"numberOfNodes,omitempty"`
	AWSSettings         []*AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings         []*GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings       []*AzureSetting `json:"azureSettings,omitempty"`
	Tags                []*Tag          `json:"tags,omitempty"`
	CloudProvider       string          `json:"cloudProvider"`
	Region              string          `json:"region"`
	ProviderAccountName string          `json:"providerAccountName,omitempty"`
}

type DataCentreStatus struct {
	ID     string  `json:"id,omitempty"`
	Status string  `json:"status,omitempty"`
	Nodes  []*Node `json:"nodes,omitempty"`
}

type CloudProviderSettings struct {
	AWSSettings   []*AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings   []*GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings []*AzureSetting `json:"azureSettings,omitempty"`
}

// HasAWSCloudProviderSettings indicates if the AWSSetting object is not zero.
// We need this because Instaclustr API returns an object of AWSSetting which is not empty,
// but the fields are filled with null values.
func (dc *CloudProviderSettings) HasAWSCloudProviderSettings() bool {
	if dc == nil || len(dc.AWSSettings) == 0 {
		return false
	}

	zero := AWSSetting{}
	for _, s := range dc.AWSSettings {
		if s != nil && *s != zero {
			return true
		}
	}

	return false
}

type AWSSetting struct {
	EBSEncryptionKey       string `json:"ebsEncryptionKey,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
	BackupBucket           string `json:"backupBucket,omitempty"`
}

type GCPSetting struct {
	CustomVirtualNetworkID    string `json:"customVirtualNetworkId,omitempty"`
	DisableSnapshotAutoExpiry bool   `json:"disableSnapshotAutoExpiry,omitempty"`
}

type AzureSetting struct {
	ResourceGroup          string `json:"resourceGroup,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
	StorageNetwork         string `json:"storageNetwork,omitempty"`
}

type Tag struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"nodeSize,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"nodeRoles,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	Rack           string   `json:"rack,omitempty"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber,omitempty"`
	ConfirmationEmail       string `json:"confirmationEmail,omitempty"`
}

type NodeReloadStatus struct {
	NodeID       string `json:"nodeId,omitempty"`
	OperationID  string `json:"operationId,omitempty"`
	TimeCreated  string `json:"timeCreated,omitempty"`
	TimeModified string `json:"timeModified,omitempty"`
	Status       string `json:"status,omitempty"`
	Message      string `json:"message,omitempty"`
}

type ActiveClusters struct {
	AccountID string           `json:"accountId,omitempty"`
	Clusters  []*ActiveCluster `json:"clusters,omitempty"`
}

type ActiveCluster struct {
	Application string `json:"application,omitempty"`
	ID          string `json:"id,omitempty"`
}

type AppVersions struct {
	Application string   `json:"application"`
	Versions    []string `json:"versions"`
}

type PrivateLink struct {
	AdvertisedHostname  string `json:"advertisedHostname"`
	EndPointServiceID   string `json:"endPointServiceId,omitempty"`
	EndPointServiceName string `json:"endPointServiceName,omitempty"`
}

type ClusterSettings struct {
	Description     string           `json:"description"`
	TwoFactorDelete *TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

// ResizeSettings determines how resize requests will be performed for the cluster
type ResizeSettings struct {
	// Setting this property to true will notify the Instaclustr
	// Account's designated support contacts on resize completion
	NotifySupportContacts bool `json:"notifySupportContacts,omitempty"`

	// Number of concurrent nodes to resize during a resize operation
	Concurrency int `json:"concurrency,omitempty"`
}

type ClusterBackup struct {
	ClusterDataCentres []*BackupDataCentre `json:"clusterDataCentres"`
}

type BackupDataCentre struct {
	Nodes []*BackupNode `json:"nodes"`
	Name  string        `json:"name"`
	ID    string        `json:"id"`
}

type BackupNode struct {
	Events         []*BackupEvent `json:"events"`
	PrivateAddress string         `json:"privateAddress,omitempty"`
	PublicAddress  string         `json:"publicAddress"`
	ID             string         `json:"id"`
}

type BackupEvent struct {
	Type     string  `json:"type"`
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
	Start    int     `json:"start"`
	End      int     `json:"end"`
}

type AddonBundle struct {
	Bundle  string         `json:"bundle"`
	Version string         `json:"version"`
	Options *BundleOptions `json:"options"`
}

func (cb *ClusterBackup) GetBackupEvents(clusterKind string) map[int]*BackupEvent {
	var eventType string

	switch clusterKind {
	case PgClusterKind:
		eventType = PgBackupEventType
	default:
		eventType = SnapshotUploadEventType
	}

	instBackupEvents := map[int]*BackupEvent{}
	for _, instDC := range cb.ClusterDataCentres {
		for _, instNode := range instDC.Nodes {
			for _, instEvent := range instNode.Events {
				if instEvent.Type == eventType {
					instBackupEvents[instEvent.Start] = instEvent
				}
			}
		}
	}

	return instBackupEvents
}

type AWSVPCPeering struct {
	ID               string   `json:"id"`
	CDCID            string   `json:"cdcId"`
	DataCentreVPCID  string   `json:"dataCentreVpcId"`
	PeerAWSAccountID string   `json:"peerAwsAccountId"`
	PeerRegion       string   `json:"peerRegion"`
	PeerSubnets      []string `json:"peerSubnets"`
	PeerVpcID        string   `json:"peerVpcId"`
	StatusCode       string   `json:"statusCode"`
}

type AWSEndpointServicePrincipal struct {
	ID                  string `json:"id,omitempty"`
	ClusterDataCenterID string `json:"clusterDataCenterId,omitempty"`
	EndPointServiceID   string `json:"endPointServiceId,omitempty"`
	PrincipalARN        string `json:"principalArn,omitempty"`
}
