/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// GetAllMirrorsResponseItem - list of all active mirroring configurations
type GetAllMirrorsResponseItem struct {

	// Id of the mirroring config.
	Id string `json:"id,omitempty"`

	// The base name for the connectors of this mirror
	ConnectorName string `json:"connectorName,omitempty"`

	// The original regular expression used to define which topics to mirror
	TopicsRegex string `json:"topicsRegex,omitempty"`
}

// AssertGetAllMirrorsResponseItemRequired checks if the required fields are not zero-ed
func AssertGetAllMirrorsResponseItemRequired(obj GetAllMirrorsResponseItem) error {
	return nil
}

// AssertGetAllMirrorsResponseItemConstraints checks if the values respects the defined constraints
func AssertGetAllMirrorsResponseItemConstraints(obj GetAllMirrorsResponseItem) error {
	return nil
}
