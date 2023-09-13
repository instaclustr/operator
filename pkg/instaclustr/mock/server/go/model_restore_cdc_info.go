/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type RestoreCdcInfo struct {
	CdcId string `json:"cdcId,omitempty"`

	RestoreToSameVpc bool `json:"restoreToSameVpc,omitempty"`

	CustomVpcId string `json:"customVpcId,omitempty"`

	CustomVpcNetwork string `json:"customVpcNetwork,omitempty"`
}

// AssertRestoreCdcInfoRequired checks if the required fields are not zero-ed
func AssertRestoreCdcInfoRequired(obj RestoreCdcInfo) error {
	return nil
}

// AssertRestoreCdcInfoConstraints checks if the values respects the defined constraints
func AssertRestoreCdcInfoConstraints(obj RestoreCdcInfo) error {
	return nil
}