/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CommonClusterBackupEventsV2 struct {

	// List of data centres in the cluster
	ClusterDataCentres []CommonClusterBackupEventDataCentreV2 `json:"clusterDataCentres,omitempty"`

	// Name of the cluster
	Name string `json:"name,omitempty"`

	// ID of the Cluster
	Id string `json:"id,omitempty"`
}

// AssertCommonClusterBackupEventsV2Required checks if the required fields are not zero-ed
func AssertCommonClusterBackupEventsV2Required(obj CommonClusterBackupEventsV2) error {
	for _, el := range obj.ClusterDataCentres {
		if err := AssertCommonClusterBackupEventDataCentreV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertCommonClusterBackupEventsV2Constraints checks if the values respects the defined constraints
func AssertCommonClusterBackupEventsV2Constraints(obj CommonClusterBackupEventsV2) error {
	return nil
}
