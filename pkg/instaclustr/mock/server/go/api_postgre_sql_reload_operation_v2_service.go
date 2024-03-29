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

// PostgreSQLReloadOperationV2APIService is a service that implements the logic for the PostgreSQLReloadOperationV2APIServicer
// This service should implement the business logic for every endpoint for the PostgreSQLReloadOperationV2API API.
// Include any external packages or services that will be required by this service.
type PostgreSQLReloadOperationV2APIService struct {
}

// NewPostgreSQLReloadOperationV2APIService creates a default api service
func NewPostgreSQLReloadOperationV2APIService() PostgreSQLReloadOperationV2APIServicer {
	return &PostgreSQLReloadOperationV2APIService{}
}

// ClusterManagementV2OperationsApplicationsPostgresqlClustersV2ClusterIdReloadGet - Get cluster reload operations
func (s *PostgreSQLReloadOperationV2APIService) ClusterManagementV2OperationsApplicationsPostgresqlClustersV2ClusterIdReloadGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsPostgresqlClustersV2ClusterIdReloadGet with the required logic for this service method.
	// Add api_postgre_sql_reload_operation_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, PostgresqlReloadOperationsV2{}) or use other options such as http.Ok ...
	// return Response(202, PostgresqlReloadOperationsV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsPostgresqlClustersV2ClusterIdReloadGet method not implemented")
}

// ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadGet - Get node reload operation
func (s *PostgreSQLReloadOperationV2APIService) ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadGet(ctx context.Context, nodeId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadGet with the required logic for this service method.
	// Add api_postgre_sql_reload_operation_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, PostgresqlReloadOperationV2{}) or use other options such as http.Ok ...
	// return Response(202, PostgresqlReloadOperationV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadGet method not implemented")
}

// ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadPost - Trigger a node reload operation
func (s *PostgreSQLReloadOperationV2APIService) ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadPost(ctx context.Context, nodeId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadPost with the required logic for this service method.
	// Add api_postgre_sql_reload_operation_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, PostgresqlReloadOperationV2{}) or use other options such as http.Ok ...
	// return Response(202, PostgresqlReloadOperationV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsPostgresqlNodesV2NodeIdReloadPost method not implemented")
}
