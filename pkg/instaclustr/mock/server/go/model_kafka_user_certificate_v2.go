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

// KafkaUserCertificateV2 - Certificate signing request.
type KafkaUserCertificateV2 struct {

	// Date certificate expires.
	ExpiryDate string `json:"expiryDate,omitempty"`

	// Certificate signing request.
	Csr string `json:"csr"`

	// The Kafka username
	KafkaUsername string `json:"kafkaUsername"`

	// ID of the kafka cluster
	ClusterId string `json:"clusterId"`

	// ID of the certificate.
	Id string `json:"id,omitempty"`

	// Generated client signed certificate.
	SignedCertificate string `json:"signedCertificate,omitempty"`

	// Number of months for which the certificate will be valid.
	ValidPeriod int32 `json:"validPeriod"`
}

// AssertKafkaUserCertificateV2Required checks if the required fields are not zero-ed
func AssertKafkaUserCertificateV2Required(obj KafkaUserCertificateV2) error {
	elements := map[string]interface{}{
		"csr":           obj.Csr,
		"kafkaUsername": obj.KafkaUsername,
		"clusterId":     obj.ClusterId,
		"validPeriod":   obj.ValidPeriod,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKafkaUserCertificateV2Constraints checks if the values respects the defined constraints
func AssertKafkaUserCertificateV2Constraints(obj KafkaUserCertificateV2) error {
	if obj.ValidPeriod < 3 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.ValidPeriod > 120 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	return nil
}
