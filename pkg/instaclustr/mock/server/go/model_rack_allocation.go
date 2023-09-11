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
	"errors"
)

type RackAllocation struct {

	// Number of racks to use when allocating nodes
	NumberOfRacks int32 `json:"numberOfRacks,omitempty"`

	// Number of nodes per rack
	NodesPerRack int32 `json:"nodesPerRack,omitempty"`
}

// AssertRackAllocationRequired checks if the required fields are not zero-ed
func AssertRackAllocationRequired(obj RackAllocation) error {
	return nil
}

// AssertRackAllocationConstraints checks if the values respects the defined constraints
func AssertRackAllocationConstraints(obj RackAllocation) error {
	if obj.NumberOfRacks < 2 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.NumberOfRacks > 5 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	if obj.NodesPerRack < 1 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.NodesPerRack > 100 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	return nil
}
