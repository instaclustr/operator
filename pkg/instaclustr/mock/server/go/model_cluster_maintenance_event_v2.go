/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"time"
)

// ClusterMaintenanceEventV2 - Information about a specific cluster maintenance event.
type ClusterMaintenanceEventV2 struct {

	// A text description of the activities being performed in this maintenance event
	Description string `json:"description,omitempty"`

	// The time work on this maintenance event started in UTC time
	StartTime time.Time `json:"startTime,omitempty"`

	// ID of the cluster this maintenance event relates to
	ClusterId string `json:"clusterId,omitempty"`

	// The time work on this maintenance event ended in UTC time
	EndTime time.Time `json:"endTime,omitempty"`

	// Unique ID for this maintenance event
	Id string `json:"id,omitempty"`

	// Maintenance outcome
	Outcome string `json:"outcome,omitempty"`
}

// AssertClusterMaintenanceEventV2Required checks if the required fields are not zero-ed
func AssertClusterMaintenanceEventV2Required(obj ClusterMaintenanceEventV2) error {
	return nil
}

// AssertClusterMaintenanceEventV2Constraints checks if the values respects the defined constraints
func AssertClusterMaintenanceEventV2Constraints(obj ClusterMaintenanceEventV2) error {
	return nil
}
