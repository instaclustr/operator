/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CreateMirrorRequest struct {
	UseRenaming bool `json:"useRenaming,omitempty"`

	SourceType string `json:"sourceType,omitempty"`

	KafkaSource string `json:"kafkaSource,omitempty"`

	UsePrivateIps bool `json:"usePrivateIps,omitempty"`

	SourceAlias string `json:"sourceAlias,omitempty"`

	SourceConnectionProperties string `json:"sourceConnectionProperties,omitempty"`

	TopicsRegex string `json:"topicsRegex,omitempty"`

	MaxTasks int32 `json:"maxTasks,omitempty"`
}

// AssertCreateMirrorRequestRequired checks if the required fields are not zero-ed
func AssertCreateMirrorRequestRequired(obj CreateMirrorRequest) error {
	return nil
}

// AssertCreateMirrorRequestConstraints checks if the values respects the defined constraints
func AssertCreateMirrorRequestConstraints(obj CreateMirrorRequest) error {
	return nil
}
