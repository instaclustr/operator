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
	"sync"
)

// ApacheCassandraProvisioningV2APIService is a service that implements the logic for the ApacheCassandraProvisioningV2APIServicer
// This service should implement the business logic for every endpoint for the ApacheCassandraProvisioningV2API API.
// Include any external packages or services that will be required by this service.
type ApacheCassandraProvisioningV2APIService struct {
	mu       sync.RWMutex
	clusters map[string]*CassandraClusterV2
}

// NewApacheCassandraProvisioningV2APIService creates a default api service
func NewApacheCassandraProvisioningV2APIService() ApacheCassandraProvisioningV2APIServicer {
	return &ApacheCassandraProvisioningV2APIService{clusters: map[string]*CassandraClusterV2{}}
}

// ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get - List recent cluster backup events.
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, CommonClusterBackupEventsV2{}) or use other options such as http.Ok ...
	// return Response(200, CommonClusterBackupEventsV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get method not implemented")
}

// ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post - Manually trigger cluster backup.
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, {}) or use other options such as http.Ok ...
	// return Response(202, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post method not implemented")
}

// ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post - Trigger a Cassandra Cluster Restore
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post(ctx context.Context, cassandraClusterRestoreV2 CassandraClusterRestoreV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, CassandraClusterRestoreV2{}) or use other options such as http.Ok ...
	// return Response(202, CassandraClusterRestoreV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post method not implemented")
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete - Delete cluster
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.clusters[clusterId]
	if !exists {
		return Response(http.StatusNotFound, nil), nil
	}

	delete(s.clusters, clusterId)

	return Response(http.StatusNoContent, nil), nil
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet - Get Cassandra cluster details.
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.clusters[clusterId]
	if !exists {
		return Response(http.StatusNotFound, nil), nil
	}

	c.Status = RUNNING

	return Response(http.StatusOK, c), nil
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut - Update Cassandra Cluster Details
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut(ctx context.Context, clusterId string, cassandraClusterUpdateV2 CassandraClusterUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.clusters[clusterId]
	if !exists {
		return Response(http.StatusNotFound, nil), nil
	}

	newNode := []NodeDetailsV2{{
		Rack:          "us-east-1a",
		NodeSize:      cassandraClusterUpdateV2.DataCentres[0].NodeSize,
		PublicAddress: "54.146.160.89",
	}}

	c.DataCentres[0].Nodes = newNode

	return Response(http.StatusAccepted, nil), nil
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post - Create a Cassandra cluster.
func (s *ApacheCassandraProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post(ctx context.Context, cassandraClusterV2 CassandraClusterV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post with the required logic for this service method.
	// Add api_apache_cassandra_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	s.mu.Lock()
	defer s.mu.Unlock()

	newCassandra := &CassandraClusterV2{}

	newCassandra = &cassandraClusterV2
	newCassandra.Id = cassandraClusterV2.Name + CreatedID
	newCassandra.DataCentres[0].Nodes = []NodeDetailsV2{{NodeSize: cassandraClusterV2.DataCentres[0].NodeSize}}

	for i := range newCassandra.DataCentres {
		newCassandra.DataCentres[i].Id = newCassandra.DataCentres[i].Name + "-" + CreatedID
	}

	s.clusters[newCassandra.Id] = newCassandra

	return Response(http.StatusAccepted, newCassandra), nil
}
