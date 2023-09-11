/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type MonitoringResultMixedSchema struct {

	// Public IP of the node
	PublicIp string `json:"publicIp,omitempty"`

	// Private IP of the node
	PrivateIp string `json:"privateIp,omitempty"`

	Rack RackPrimaryKey `json:"rack,omitempty"`
}

// AssertMonitoringResultMixedSchemaRequired checks if the required fields are not zero-ed
func AssertMonitoringResultMixedSchemaRequired(obj MonitoringResultMixedSchema) error {
	if err := AssertRackPrimaryKeyRequired(obj.Rack); err != nil {
		return err
	}
	return nil
}

// AssertMonitoringResultMixedSchemaConstraints checks if the values respects the defined constraints
func AssertMonitoringResultMixedSchemaConstraints(obj MonitoringResultMixedSchema) error {
	return nil
}
