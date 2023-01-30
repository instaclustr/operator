/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// MongodbClusterV2 - Definition of a managed Kafka cluster that can be provisioned in Instaclustr.
type MongodbClusterV2 struct {

	// Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/).
	PciComplianceMode bool `json:"pciComplianceMode"`

	// List of data centre settings.
	DataCentres []MongodbDataCentreV2 `json:"dataCentres"`

	CurrentClusterOperationStatus CurrentClusterOperationStatusV2 `json:"currentClusterOperationStatus,omitempty"`

	// Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).
	PrivateNetworkCluster bool `json:"privateNetworkCluster"`

	// Name of the cluster.
	Name string `json:"name"`

	//
	TwoFactorDelete []TwoFactorDeleteSettingsV2 `json:"twoFactorDelete,omitempty"`

	// ID of the cluster.
	Id string `json:"id,omitempty"`

	SlaTier SlaTierV2 `json:"slaTier"`

	// Version of MongoDB to run on the cluster. Available versions: <ul> </ul>
	MongodbVersion string `json:"mongodbVersion"`

	// Status of the cluster.
	Status string `json:"status,omitempty"`
}

// AssertMongodbClusterV2Required checks if the required fields are not zero-ed
func AssertMongodbClusterV2Required(obj MongodbClusterV2) error {
	elements := map[string]interface{}{
		"pciComplianceMode":     obj.PciComplianceMode,
		"dataCentres":           obj.DataCentres,
		"privateNetworkCluster": obj.PrivateNetworkCluster,
		"name":                  obj.Name,
		"slaTier":               obj.SlaTier,
		"mongodbVersion":        obj.MongodbVersion,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.DataCentres {
		if err := AssertMongodbDataCentreV2Required(el); err != nil {
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

// AssertRecurseMongodbClusterV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of MongodbClusterV2 (e.g. [][]MongodbClusterV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseMongodbClusterV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aMongodbClusterV2, ok := obj.(MongodbClusterV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertMongodbClusterV2Required(aMongodbClusterV2)
	})
}
