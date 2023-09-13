/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CreateInstaclustrMirrorRequestSchemaOnlyClass struct {

	// Indicates that your source cluster is an instaclustr managed cluster
	SourceType string `json:"sourceType,omitempty"`

	// Whether to rename topics as they are mirrored, by prefixing the source alias to the topic name
	UseRenaming bool `json:"useRenaming,omitempty"`

	// Id of your source instaclustr Kafka cluster
	KafkaSource string `json:"kafkaSource,omitempty"`

	// Whether or not to connect to  your source cluster's private IP addresses
	UsePrivateIps bool `json:"usePrivateIps,omitempty"`

	// Alias to use for your source Kafka cluster. This will be used to rename topics if that option is turned on
	SourceAlias string `json:"sourceAlias,omitempty"`

	// Regex to select which topics to mirror
	TopicsRegex string `json:"topicsRegex,omitempty"`

	// Maximum number of tasks for Kafka Connect to use
	MaxTasks int32 `json:"maxTasks,omitempty"`
}

// AssertCreateInstaclustrMirrorRequestSchemaOnlyClassRequired checks if the required fields are not zero-ed
func AssertCreateInstaclustrMirrorRequestSchemaOnlyClassRequired(obj CreateInstaclustrMirrorRequestSchemaOnlyClass) error {
	return nil
}

// AssertCreateInstaclustrMirrorRequestSchemaOnlyClassConstraints checks if the values respects the defined constraints
func AssertCreateInstaclustrMirrorRequestSchemaOnlyClassConstraints(obj CreateInstaclustrMirrorRequestSchemaOnlyClass) error {
	return nil
}