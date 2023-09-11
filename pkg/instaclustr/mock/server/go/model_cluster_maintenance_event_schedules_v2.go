/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ClusterMaintenanceEventSchedulesV2 struct {
	MaintenanceEvents []ClusterMaintenanceEventScheduleV2 `json:"maintenanceEvents,omitempty"`
}

// AssertClusterMaintenanceEventSchedulesV2Required checks if the required fields are not zero-ed
func AssertClusterMaintenanceEventSchedulesV2Required(obj ClusterMaintenanceEventSchedulesV2) error {
	for _, el := range obj.MaintenanceEvents {
		if err := AssertClusterMaintenanceEventScheduleV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertClusterMaintenanceEventSchedulesV2Constraints checks if the values respects the defined constraints
func AssertClusterMaintenanceEventSchedulesV2Constraints(obj ClusterMaintenanceEventSchedulesV2) error {
	return nil
}
