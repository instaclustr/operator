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

// ClusterMaintenanceEventScheduleV2 - Information about a specific cluster maintenance event.
type ClusterMaintenanceEventScheduleV2 struct {

	// When this maintenance event is scheduled to start in UTC time
	ScheduledStartTime time.Time `json:"scheduledStartTime,omitempty"`

	// When this maintenance event is scheduled to end in UTC time
	ScheduledEndTime time.Time `json:"scheduledEndTime,omitempty"`

	// A text description of the activities being performed in this maintenance event
	Description string `json:"description,omitempty"`

	// ID of the cluster this maintenance event relates to
	ClusterId string `json:"clusterId,omitempty"`

	// Unique ID for this maintenance event
	Id string `json:"id,omitempty"`

	// The earliest limit for the scheduled start time in UTC time, i.e., scheduled start time will be after this value
	ScheduledStartTimeMin time.Time `json:"scheduledStartTimeMin,omitempty"`

	// If this event can still be rescheduled
	IsFinalized bool `json:"isFinalized,omitempty"`

	// The latest limit for the scheduled start time in UTC time, i.e., scheduled start time will be before this value
	ScheduledStartTimeMax time.Time `json:"scheduledStartTimeMax,omitempty"`
}

// AssertClusterMaintenanceEventScheduleV2Required checks if the required fields are not zero-ed
func AssertClusterMaintenanceEventScheduleV2Required(obj ClusterMaintenanceEventScheduleV2) error {
	return nil
}

// AssertClusterMaintenanceEventScheduleV2Constraints checks if the values respects the defined constraints
func AssertClusterMaintenanceEventScheduleV2Constraints(obj ClusterMaintenanceEventScheduleV2) error {
	return nil
}
