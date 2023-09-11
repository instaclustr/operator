/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CassandraDataCentreV2 struct {

	// Enables commitlog backups and increases the frequency of the default snapshot backups.
	ContinuousBackup bool `json:"continuousBackup"`

	// Number of racks to use when allocating nodes.
	ReplicationFactor int32 `json:"replicationFactor"`

	// List of deleted nodes in the data centre
	DeletedNodes []NodeDetailsV2 `json:"deletedNodes,omitempty"`

	// AWS specific settings for the Data Centre. Cannot be provided with GCP or Azure settings.
	AwsSettings []ProviderAwsSettingsV2 `json:"awsSettings,omitempty"`

	// Total number of nodes in the Data Centre. Must be a multiple of `replicationFactor`.
	NumberOfNodes int32 `json:"numberOfNodes"`

	// Enables broadcast of private IPs for auto-discovery.
	PrivateIpBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`

	// The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between `/12` and `/22` and must be part of a private address space.
	Network string `json:"network"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which  allow you to identify, categorize and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require. Note `tag` is not supported in terraform lifecycle `ignore_changes`.
	Tags []ProviderTagV2 `json:"tags,omitempty"`

	// GCP specific settings for the Data Centre. Cannot be provided with AWS or Azure settings.
	GcpSettings []ProviderGcpSettingsV2 `json:"gcpSettings,omitempty"`

	// Enables Client ⇄ Node Encryption.
	ClientToClusterEncryption bool `json:"clientToClusterEncryption"`

	NodeSize string `json:"nodeSize"`

	// List of non-deleted nodes in the data centre
	Nodes []NodeDetailsV2 `json:"nodes,omitempty"`

	CloudProvider CloudProviderEnumV2 `json:"cloudProvider"`

	// Azure specific settings for the Data Centre. Cannot be provided with AWS or GCP settings.
	AzureSettings []ProviderAzureSettingsV2 `json:"azureSettings,omitempty"`

	// A logical name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// ID of the Cluster Data Centre.
	Id string `json:"id,omitempty"`

	// Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.
	Region string `json:"region"`

	// Create a PrivateLink enabled cluster, see [PrivateLink](https://www.instaclustr.com/support/documentation/useful-information/privatelink/).
	PrivateLink bool `json:"privateLink,omitempty"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the \"Provider Account\" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`

	// Status of the Data Centre.
	Status string `json:"status,omitempty"`
}

// AssertCassandraDataCentreV2Required checks if the required fields are not zero-ed
func AssertCassandraDataCentreV2Required(obj CassandraDataCentreV2) error {
	elements := map[string]interface{}{
		"continuousBackup":               obj.ContinuousBackup,
		"replicationFactor":              obj.ReplicationFactor,
		"numberOfNodes":                  obj.NumberOfNodes,
		"privateIpBroadcastForDiscovery": obj.PrivateIpBroadcastForDiscovery,
		"network":                        obj.Network,
		"clientToClusterEncryption":      obj.ClientToClusterEncryption,
		"nodeSize":                       obj.NodeSize,
		"cloudProvider":                  obj.CloudProvider,
		"name":                           obj.Name,
		"region":                         obj.Region,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.DeletedNodes {
		if err := AssertNodeDetailsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.AwsSettings {
		if err := AssertProviderAwsSettingsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.Tags {
		if err := AssertProviderTagV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.GcpSettings {
		if err := AssertProviderGcpSettingsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.Nodes {
		if err := AssertNodeDetailsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.AzureSettings {
		if err := AssertProviderAzureSettingsV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertCassandraDataCentreV2Constraints checks if the values respects the defined constraints
func AssertCassandraDataCentreV2Constraints(obj CassandraDataCentreV2) error {
	return nil
}
