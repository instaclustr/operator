/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ClusterEvents struct {

	// Cluster ID
	Id string `json:"id,omitempty"`

	// Cluster name
	Name string `json:"name,omitempty"`

	ClusterDataCentres []ClusterDataCentre `json:"clusterDataCentres,omitempty"`
}

// AssertClusterEventsRequired checks if the required fields are not zero-ed
func AssertClusterEventsRequired(obj ClusterEvents) error {
	for _, el := range obj.ClusterDataCentres {
		if err := AssertClusterDataCentreRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertClusterEventsConstraints checks if the values respects the defined constraints
func AssertClusterEventsConstraints(obj ClusterEvents) error {
	return nil
}