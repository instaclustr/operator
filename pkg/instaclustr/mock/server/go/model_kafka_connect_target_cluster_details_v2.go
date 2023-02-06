/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaConnectTargetClusterDetailsV2 -
type KafkaConnectTargetClusterDetailsV2 struct {

	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.
	ExternalCluster []KafkaConnectExternalTargetClusterDetailsV2 `json:"externalCluster,omitempty"`

	// Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.
	ManagedCluster []KafkaConnectManagedTargetClusterDetailsV2 `json:"managedCluster,omitempty"`
}

// AssertKafkaConnectTargetClusterDetailsV2Required checks if the required fields are not zero-ed
func AssertKafkaConnectTargetClusterDetailsV2Required(obj KafkaConnectTargetClusterDetailsV2) error {
	for _, el := range obj.ExternalCluster {
		if err := AssertKafkaConnectExternalTargetClusterDetailsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.ManagedCluster {
		if err := AssertKafkaConnectManagedTargetClusterDetailsV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseKafkaConnectTargetClusterDetailsV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaConnectTargetClusterDetailsV2 (e.g. [][]KafkaConnectTargetClusterDetailsV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaConnectTargetClusterDetailsV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaConnectTargetClusterDetailsV2, ok := obj.(KafkaConnectTargetClusterDetailsV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaConnectTargetClusterDetailsV2Required(aKafkaConnectTargetClusterDetailsV2)
	})
}