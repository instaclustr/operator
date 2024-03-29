/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RestoreCdcInfoSchema - An optional list of Cluster Data Centres for which custom VPC settings will be used.<br><br>This property must be populated for all Cluster Data Centres or none at all.
type RestoreCdcInfoSchema struct {

	// Cluster Data Centre ID
	CdcId string `json:"cdcId,omitempty"`

	// Determines if the restored cluster will be allocated to the existing VPC.<br><br>Either restoreToSameVpc or customVpcId must be provided.
	RestoreToSameVpc bool `json:"restoreToSameVpc,omitempty"`

	// Custom VPC ID to which the restored cluster will be allocated.<br><br>Either restoreToSameVpc or customVpcId must be provided.
	CustomVpcId string `json:"customVpcId,omitempty"`

	// CIDR block in which the cluster will be allocated for a custom VPC.
	CustomVpcNetwork string `json:"customVpcNetwork,omitempty"`
}

// AssertRestoreCdcInfoSchemaRequired checks if the required fields are not zero-ed
func AssertRestoreCdcInfoSchemaRequired(obj RestoreCdcInfoSchema) error {
	return nil
}

// AssertRestoreCdcInfoSchemaConstraints checks if the values respects the defined constraints
func AssertRestoreCdcInfoSchemaConstraints(obj RestoreCdcInfoSchema) error {
	return nil
}
