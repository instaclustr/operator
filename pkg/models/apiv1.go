package models

const (
	RunningStatus = "RUNNING"
	Disabled      = "DISABLED"
)

type ClusterV1 struct {
	ID                    string               `json:"id,omitempty"`
	ClusterName           string               `json:"clusterName,omitempty"`
	ClusterStatus         string               `json:"clusterStatus,omitempty"`
	ClusterProvider       []*ClusterProviderV1 `json:"clusterProvider,omitempty"`
	PrivateNetworkCluster bool                 `json:"privateNetworkCluster,omitempty"`
	SLATier               string               `json:"slaTier,omitempty"`
	NodeSize              string               `json:"nodeSize,omitempty"`
	ClusterNetwork        string               `json:"clusterNetwork,omitempty"`
	DataCentre            string               `json:"dataCentre,omitempty"`
	CDCID                 string               `json:"cdcid,omitempty"`
	DataCentreCustomName  string               `json:"dataCentreCustomName,omitempty"`
	RackAllocation        *RackAllocationV1    `json:"rackAllocation,omitempty"`
	FirewallRules         []*FirewallRule      `json:"firewallRules,omitempty"`
	OIDCProvider          string               `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool                 `json:"bundledUseOnlyCluster,omitempty"`
	DataCentres           []*DataCentreV1      `json:"dataCentres,omitempty"`
	BundleOptions         *BundleOptions       `json:"bundleOptions,omitempty"`
	BundleVersion         string               `json:"bundleVersion,omitempty"`
	PCICompliance         string               `json:"pciCompliance,omitempty"`
	AddonBundles          []*AddonBundle       `json:"addonBundles,omitempty"`
}

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
	NumberOfRacks int32 `json:"numberOfRacks"`
	NodesPerRack  int32 `json:"nodesPerRack"`
}

type FirewallRule struct {
	Network         string      `json:"network,omitempty"`
	SecurityGroupId string      `json:"securityGroupId,omitempty"`
	Rules           []*RuleType `json:"rules"`
}

type RuleType struct {
	Type string `json:"type"`
}

type TwoFactorDelete struct {
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
	CDCName                       string          `json:"CDCName"`
	Provider                      string          `json:"provider"`
	CDCNetwork                    string          `json:"cdcNetwork"`
	ClientEncryption              bool            `json:"clientEncryption"`
	PasswordAuthentication        bool            `json:"passwordAuthentication"`
	UserAuthorization             bool            `json:"userAuthorization"`
	UsePrivateBroadcastRPCAddress bool            `json:"usePrivateBroadcastRPCAddress"`
	PrivateIPOnly                 bool            `json:"privateIPOnly"`
	EncryptionKeyID               string          `json:"encryptionKeyId,omitempty"`
	NodeCount                     int32           `json:"nodeCount,omitempty"`
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
	TwoFactorDelete *TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	Description     string           `json:"description,omitempty"`
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

type NodeReloadV1 struct {
	Bundle string `json:"bundle"`
	NodeID string `json:"nodeId"`
}

type NodeReloadStatusAPIv1 struct {
	Operations []*OperationV1 `json:"operations"`
}

type OperationV1 struct {
	TimeCreated  int64  `json:"timeCreated"`
	TimeModified int64  `json:"timeModified"`
	Status       string `json:"status"`
	Message      string `json:"message"`
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
