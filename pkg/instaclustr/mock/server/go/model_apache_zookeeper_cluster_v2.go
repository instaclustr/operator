/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ApacheZookeeperClusterV2 - Definition of a managed Apache Zookeeper cluster that can be provisioned in Instaclustr.
type ApacheZookeeperClusterV2 struct {

	// List of data centre settings.
	DataCentres []ApacheZookeeperDataCentreV2 `json:"dataCentres"`

	CurrentClusterOperationStatus CurrentClusterOperationStatusV2 `json:"currentClusterOperationStatus,omitempty"`

	// Version of Apache Zookeeper to run on the cluster. Available versions: <ul> <li>`3.7.1`</li> </ul>
	ZookeeperVersion string `json:"zookeeperVersion"`

	// Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).
	PrivateNetworkCluster bool `json:"privateNetworkCluster"`

	// Name of the cluster.
	Name string `json:"name"`

	// A description of the cluster
	Description string `json:"description,omitempty"`

	TwoFactorDelete []TwoFactorDeleteSettingsV2 `json:"twoFactorDelete,omitempty"`

	// ID of the cluster.
	Id string `json:"id,omitempty"`

	SlaTier SlaTierV2 `json:"slaTier"`

	// Status of the cluster.
	Status string `json:"status,omitempty"`
}

// AssertApacheZookeeperClusterV2Required checks if the required fields are not zero-ed
func AssertApacheZookeeperClusterV2Required(obj ApacheZookeeperClusterV2) error {
	elements := map[string]interface{}{
		"dataCentres":           obj.DataCentres,
		"zookeeperVersion":      obj.ZookeeperVersion,
		"privateNetworkCluster": obj.PrivateNetworkCluster,
		"name":                  obj.Name,
		"slaTier":               obj.SlaTier,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.DataCentres {
		if err := AssertApacheZookeeperDataCentreV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.TwoFactorDelete {
		if err := AssertTwoFactorDeleteSettingsV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertApacheZookeeperClusterV2Constraints checks if the values respects the defined constraints
func AssertApacheZookeeperClusterV2Constraints(obj ApacheZookeeperClusterV2) error {
	return nil
}
