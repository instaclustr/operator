package v1alpha1

type CloudProvider struct {
	Name                   string `json:"name"`
	AccountName            string `json:"accountName,omitempty"`
	CustomVirtualNetworkId string `json:"customVirtualNetworkId,omitempty"`
	ResourceGroup          string `json:"resourceGroup,omitempty"`
	DiskEncryptionKey      string `json:"diskEncryptionKey,omitempty"`
}

// Bundle is deprecated: delete when not used.
type Bundle struct {
	Bundle  string `json:"bundle"`
	Version string `json:"version"`
}

type GenericDataCentre struct {

	// When use one data centre name of field is dataCentreCustomName for APIv1
	Name string `json:"name,omitempty"`

	// Region. APIv1 : "dataCentre"
	Region string `json:"region"`

	Network  string         `json:"network"`
	Provider *CloudProvider `json:"provider,omitempty"`
	NodeSize string         `json:"nodeSize,omitempty"`

	// APIv2: replicationFactor; APIv1: numberOfRacks
	RacksNumber int32 `json:"racksNumber"`

	// APIv2: numberOfNodes; APIv1: nodesPerRack.
	NodesNumber int32 `json:"nodesNumber"`
}

type DataCentreStatus struct {
	DataCentreID string `json:"dataCentreID"`
	DCStatus     string `json:"DCStatus,omitempty"`
	NodeCount    int32  `json:"nodeCount,omitempty"`
}

type Node struct {
	NodeID         string `json:"nodeID,omitempty"`
	NodeSize       string `json:"nodeSize,omitempty"`
	PublicAddress  string `json:"publicAddress,omitempty"`
	PrivateAddress string `json:"privateAddress,omitempty"`
	NodeStatus     string `json:"nodeStatus,omitempty"`
}

type GenericCluster struct {
	// ClusterName [ 3 .. 32 ] characters.
	// APIv2 : "name", APIv1 : "clusterName".
	ClusterName string `json:"clusterName"`

	Version string `json:"version"`

	// The PCI compliance standards relate to the security of user data and transactional information.
	// Can only be applied clusters provisioned on AWS_VPC, running Cassandra, Kafka, Elasticsearch and Redis.
	// PCI compliance cannot be enabled if the cluster has Spark.
	//
	// APIv1 : "pciCompliantCluster,omitempty"; APIv2 : pciComplianceMode.
	PCICompliance bool `json:"pciCompliance,omitempty"`

	// Required for APIv2, but for APIv1 set "false" as a default.
	PrivateNetworkCluster bool `json:"privateNetworkCluster,omitempty"`

	// Non-production clusters may receive lower priority support and reduced SLAs.
	// Production tier is not available when using Developer class nodes. See SLA Tier for more information.
	// Enum: "PRODUCTION" "NON_PRODUCTION".
	// Required for APIv2, but for APIv1 set "NON_PRODUCTION" as a default.
	SLATier string `json:"slaTier,omitempty"`

	FirewallRules []*FirewallRule `json:"firewallRules,omitempty"`

	// APIv2, unlike AP1, receives an array of TwoFactorDelete (<= 1 items);
	TwoFactorDelete []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`

	Tags []*Tag `json:"tags,omitempty"`

	// running as dependency for another instance.
	// need to be tested before decide what to do with this field
	BundledUseOnlyCluster bool `json:"bundledUseOnlyCluster,omitempty"`
}

type ClusterStatus struct {
	ClusterID                  string `json:"id,omitempty"`
	ClusterCertificateDownload string `json:"clusterCertificateDownload,omitempty"`

	// ClusterStatus shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	ClusterStatus string `json:"status,omitempty"`
}

type FirewallRule struct {
	Network         string     `json:"network,omitempty"`
	SecurityGroupId string     `json:"securityGroupId,omitempty"`
	Rules           []RuleType `json:"rules"`
}

type RuleType struct {
	Type string `json:"type"`
}

type TwoFactorDelete struct {
	// Email address which will be contacted when the cluster is requested to be deleted.
	// APIv1: deleteVerifyEmail; APIv2: confirmationEmail.
	Email string `json:"email"`

	// APIv1: deleteVerifyPhone; APIv2: confirmationPhoneNumber.
	Phone string `json:"phone,omitempty"`
}

type Tag struct {

	// Value of the tag to be added to the Data Centre.
	Value string `json:"value"`

	// Key of the tag to be added to the Data Centre.
	Key string `json:"key"`
}
