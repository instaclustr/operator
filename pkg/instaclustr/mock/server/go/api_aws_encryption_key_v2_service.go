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
	"context"
	"errors"
	"net/http"
)

// AWSEncryptionKeyV2APIService is a service that implements the logic for the AWSEncryptionKeyV2APIServicer
// This service should implement the business logic for every endpoint for the AWSEncryptionKeyV2API API.
// Include any external packages or services that will be required by this service.
type AWSEncryptionKeyV2APIService struct {
}

// NewAWSEncryptionKeyV2APIService creates a default api service
func NewAWSEncryptionKeyV2APIService() AWSEncryptionKeyV2APIServicer {
	return &AWSEncryptionKeyV2APIService{}
}

// ClusterManagementV2DataSourcesProvidersAwsEncryptionKeysV2Get - List all AWS encryption keys
func (s *AWSEncryptionKeyV2APIService) ClusterManagementV2DataSourcesProvidersAwsEncryptionKeysV2Get(ctx context.Context) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesProvidersAwsEncryptionKeysV2Get with the required logic for this service method.
	// Add api_aws_encryption_key_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []AccountAwsEncryptionKeysV2{}) or use other options such as http.Ok ...
	// return Response(200, []AccountAwsEncryptionKeysV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesProvidersAwsEncryptionKeysV2Get method not implemented")
}

// ClusterManagementV2OperationsProvidersAwsEncryptionKeysV2EncryptionKeyIdValidateV2Get - Validate encryption Key
func (s *AWSEncryptionKeyV2APIService) ClusterManagementV2OperationsProvidersAwsEncryptionKeysV2EncryptionKeyIdValidateV2Get(ctx context.Context, encryptionKeyId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsProvidersAwsEncryptionKeysV2EncryptionKeyIdValidateV2Get with the required logic for this service method.
	// Add api_aws_encryption_key_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, EncryptionKeyValidationResponseV2{}) or use other options such as http.Ok ...
	// return Response(200, EncryptionKeyValidationResponseV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsProvidersAwsEncryptionKeysV2EncryptionKeyIdValidateV2Get method not implemented")
}

// ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdDelete - Delete encryption key
func (s *AWSEncryptionKeyV2APIService) ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdDelete(ctx context.Context, encryptionKeyId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdDelete with the required logic for this service method.
	// Add api_aws_encryption_key_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	// return Response(204, nil),nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdDelete method not implemented")
}

// ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdGet - Get encryption key details
func (s *AWSEncryptionKeyV2APIService) ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdGet(ctx context.Context, encryptionKeyId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdGet with the required logic for this service method.
	// Add api_aws_encryption_key_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, AwsEncryptionKeyV2{}) or use other options such as http.Ok ...
	// return Response(200, AwsEncryptionKeyV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2EncryptionKeyIdGet method not implemented")
}

// ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2Post - Add an AWS KMS encryption key
func (s *AWSEncryptionKeyV2APIService) ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2Post(ctx context.Context, awsEncryptionKeyV2 AwsEncryptionKeyV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2Post with the required logic for this service method.
	// Add api_aws_encryption_key_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, AwsEncryptionKeyV2{}) or use other options such as http.Ok ...
	// return Response(202, AwsEncryptionKeyV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAwsEncryptionKeysV2Post method not implemented")
}
