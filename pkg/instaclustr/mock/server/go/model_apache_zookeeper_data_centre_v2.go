/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ApacheZookeeperDataCentreV2 -
type ApacheZookeeperDataCentreV2 struct {

	// AWS specific settings for the Data Centre. Cannot be provided with GCP or Azure settings.
	AwsSettings []ProviderAwsSettingsV2 `json:"awsSettings,omitempty"`

	// Total number of Zookeeper nodes in the Data Centre.
	NumberOfNodes int32 `json:"numberOfNodes"`

	// The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between `/12` and `/22` and must be part of a private address space.
	Network string `json:"network"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which  allow you to identify, categorize and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.
	Tags []ProviderTagV2 `json:"tags,omitempty"`

	// GCP specific settings for the Data Centre. Cannot be provided with AWS or Azure settings.
	GcpSettings []ProviderGcpSettingsV2 `json:"gcpSettings,omitempty"`

	// Size of the nodes provisioned in the Data Centre. --AVAILABLE_NODE_SIZES_MARKER_V2_ZOOKEEPER--
	NodeSize string `json:"nodeSize"`

	//
	Nodes []NodeDetailsV2 `json:"nodes,omitempty"`

	// Enables Client ⇄ Node Encryption.
	ClientToServerEncryption bool `json:"clientToServerEncryption,omitempty"`

	CloudProvider CloudProviderEnumV2 `json:"cloudProvider"`

	// Azure specific settings for the Data Centre. Cannot be provided with AWS or GCP settings.
	AzureSettings []ProviderAzureSettingsV2 `json:"azureSettings,omitempty"`

	// A logical name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// ID of the Cluster Data Centre.
	Id string `json:"id,omitempty"`

	// Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.
	Region string `json:"region"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the \"Provider Account\" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`

	// Status of the Data Centre.
	Status string `json:"status,omitempty"`
}

// AssertApacheZookeeperDataCentreV2Required checks if the required fields are not zero-ed
func AssertApacheZookeeperDataCentreV2Required(obj ApacheZookeeperDataCentreV2) error {
	elements := map[string]interface{}{
		"numberOfNodes": obj.NumberOfNodes,
		"network":       obj.Network,
		"nodeSize":      obj.NodeSize,
		"cloudProvider": obj.CloudProvider,
		"name":          obj.Name,
		"region":        obj.Region,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
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

// AssertRecurseApacheZookeeperDataCentreV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ApacheZookeeperDataCentreV2 (e.g. [][]ApacheZookeeperDataCentreV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseApacheZookeeperDataCentreV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aApacheZookeeperDataCentreV2, ok := obj.(ApacheZookeeperDataCentreV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertApacheZookeeperDataCentreV2Required(aApacheZookeeperDataCentreV2)
	})
}
