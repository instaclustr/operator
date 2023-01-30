/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaAclListV2 - List of access control lists for a Kafka cluster.
type KafkaAclListV2 struct {

	// List of ACLs for the given principal.
	Acls []KafkaAclV2 `json:"acls"`

	// This is the principal without the `User:` prefix.
	UserQuery string `json:"userQuery"`

	// UUID of the Kafka cluster.
	ClusterId string `json:"clusterId"`

	// Instaclustr identifier for the ACL list for a principal. The value of this property has the form: [clusterId]_[principalUserQuery] The user query is the principal value without the leading `User:`.
	Id string `json:"id,omitempty"`
}

// AssertKafkaAclListV2Required checks if the required fields are not zero-ed
func AssertKafkaAclListV2Required(obj KafkaAclListV2) error {
	elements := map[string]interface{}{
		"acls":      obj.Acls,
		"userQuery": obj.UserQuery,
		"clusterId": obj.ClusterId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Acls {
		if err := AssertKafkaAclV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseKafkaAclListV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaAclListV2 (e.g. [][]KafkaAclListV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaAclListV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaAclListV2, ok := obj.(KafkaAclListV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaAclListV2Required(aKafkaAclListV2)
	})
}
