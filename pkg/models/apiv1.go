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
	RunningStatus = "RUNNING"
	Disabled      = "DISABLED"
)

type ClusterProviderV1 struct {
	Name                   string            `json:"name"`
	AccountName            string            `json:"accountName,omitempty"`
	CustomVirtualNetworkID string            `json:"customVirtualNetworkId,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
	ResourceGroup          string            `json:"resourceGroup,omitempty"`
	DiskEncryptionKey      string            `json:"diskEncryptionKey,omitempty"`
	ResourceName           string            `json:"resourceName,omitempty"`
}

type RackAllocationV1 struct {
	NumberOfRacks int `json:"numberOfRacks"`
	NodesPerRack  int `json:"nodesPerRack"`
}

type FirewallRule struct {
	Network         string      `json:"network,omitempty"`
	SecurityGroupId string      `json:"securityGroupId,omitempty"`
	Rules           []*RuleType `json:"rules"`
}

type RuleType struct {
	Type string `json:"type"`
}

type TwoFactorDeleteV1 struct {
	DeleteVerifyEmail string `json:"deleteVerifyEmail,omitempty"`
	DeleteVerifyPhone string `json:"deleteVerifyPhone,omitempty"`
}

type PrivateLink struct {
	IAMPrincipalARNs []string `json:"iamPrincipalARNs"`
}

type Bundle struct {
	Bundle  string `json:"bundle"`
	Version string `json:"version"`
}

type DataCentreV1 struct {
	ID                            string          `json:"id,omitempty"`
	Name                          string          `json:"name"`
	CDCName                       string          `json:"cdcName"`
	Provider                      string          `json:"provider"`
	CDCNetwork                    string          `json:"cdcNetwork"`
	ClientEncryption              bool            `json:"clientEncryption"`
	PasswordAuthentication        bool            `json:"passwordAuthentication"`
	UserAuthorization             bool            `json:"userAuthorization"`
	UsePrivateBroadcastRPCAddress bool            `json:"usePrivateBroadcastRPCAddress"`
	PrivateIPOnly                 bool            `json:"privateIPOnly"`
	EncryptionKeyID               string          `json:"encryptionKeyId,omitempty"`
	NodeCount                     int             `json:"nodeCount,omitempty"`
	Nodes                         []*NodeStatusV1 `json:"nodes"`
	PrivateLink                   *PrivateLink    `json:"privateLink,omitempty"`
	CDCStatus                     string          `json:"cdcStatus,omitempty"`
}

type NodeStatusV1 struct {
	ID             string `json:"id"`
	Rack           string `json:"rack"`
	Size           string `json:"size"`
	PublicAddress  string `json:"publicAddress"`
	PrivateAddress string `json:"privateAddress"`
	NodeStatus     string `json:"nodeStatus"`
}

type ResizeRequest struct {
	NewNodeSize           string `json:"newNodeSize"`
	ConcurrentResizes     int    `json:"concurrentResizes"`
	NotifySupportContacts bool   `json:"notifySupportContacts"`
	NodePurpose           string `json:"nodePurpose"`
	ClusterID             string `json:"-"`
	DataCentreID          string `json:"-"`
}

type DataCentreResizeOperations struct {
	CDCID  string `json:"cdc"`
	Status string `json:"completedStatus"`
}

type ClusterConfigurations struct {
	ParameterName  string `json:"parameterName"`
	ParameterValue string `json:"parameterValue"`
}

type ClusterModifyRequest struct {
	TwoFactorDelete *TwoFactorDeleteV1 `json:"twoFactorDelete,omitempty"`
	Description     string             `json:"description,omitempty"`
}

type BundleOptions struct {
	DataNodeSize                 string `json:"dataNodeSize,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
	AlertingPlugin               bool   `json:"alertingPlugin,omitempty"`
	ICUPlugin                    bool   `json:"icuPlugin,omitempty"`
	KNNPlugin                    bool   `json:"knnPlugin,omitempty"`
	NotificationsPlugin          bool   `json:"notificationsPlugin,omitempty"`
	ReportsPlugin                bool   `json:"reportsPlugin,omitempty"`
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
}

type ClusterBackup struct {
	ClusterDataCentres []*BackupDataCentre `json:"clusterDataCentres"`
}

type BackupDataCentre struct {
	Nodes []*BackupNode `json:"nodes"`
}

type BackupNode struct {
	Events []*BackupEvent `json:"events"`
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

type InstaUser struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	InitialPermission string `json:"initial-permissions"`
}

type InstaOpenSearchUser struct {
	*InstaUser `json:",inline"`
	Options    Options `json:"options"`
}

type Options struct {
	IndexPattern string `json:"indexPattern"`
	Role         string `json:"role"`
}
