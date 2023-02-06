/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaConnectCustomConnectorsV2 -
type KafkaConnectCustomConnectorsV2 struct {

	// Defines the information to access custom connectors located in an azure storage container. Cannot be provided if custom connectors are stored in GCP or AWS.
	AzureConnectorSettings []KafkaConnectCustomConnectorsAzureV2 `json:"azureConnectorSettings,omitempty"`

	// Defines the information to access custom connectors located in a S3 bucket. Cannot be provided if custom connectors are stored in GCP or AZURE. Access could be provided via Access and Secret key pair or IAM Role ARN. If neither is provided, access policy is defaulted to be provided later.
	AwsConnectorSettings []KafkaConnectCustomConnectorsAwsV2 `json:"awsConnectorSettings,omitempty"`

	// Defines the information to access custom connectors located in a gcp storage container. Cannot be provided if custom connectors are stored in AWS or AZURE.
	GcpConnectorSettings []KafkaConnectCustomConnectorsGcpV2 `json:"gcpConnectorSettings,omitempty"`
}

// AssertKafkaConnectCustomConnectorsV2Required checks if the required fields are not zero-ed
func AssertKafkaConnectCustomConnectorsV2Required(obj KafkaConnectCustomConnectorsV2) error {
	for _, el := range obj.AzureConnectorSettings {
		if err := AssertKafkaConnectCustomConnectorsAzureV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.AwsConnectorSettings {
		if err := AssertKafkaConnectCustomConnectorsAwsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.GcpConnectorSettings {
		if err := AssertKafkaConnectCustomConnectorsGcpV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseKafkaConnectCustomConnectorsV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaConnectCustomConnectorsV2 (e.g. [][]KafkaConnectCustomConnectorsV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaConnectCustomConnectorsV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaConnectCustomConnectorsV2, ok := obj.(KafkaConnectCustomConnectorsV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaConnectCustomConnectorsV2Required(aKafkaConnectCustomConnectorsV2)
	})
}