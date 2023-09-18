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
)

// ClusterMaintenanceEventsV2APIService is a service that implements the logic for the ClusterMaintenanceEventsV2APIServicer
// This service should implement the business logic for every endpoint for the ClusterMaintenanceEventsV2API API.
// Include any external packages or services that will be required by this service.
type ClusterMaintenanceEventsV2APIService struct {
}

// NewClusterMaintenanceEventsV2APIService creates a default api service
func NewClusterMaintenanceEventsV2APIService() ClusterMaintenanceEventsV2APIServicer {
	return &ClusterMaintenanceEventsV2APIService{}
}

// ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2InProgressV2Get -
func (s *ClusterMaintenanceEventsV2APIService) ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2InProgressV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2InProgressV2Get with the required logic for this service method.
	// Add api_cluster_maintenance_events_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []ClusterMaintenanceEventsV2{}) or use other options such as http.Ok ...
	return Response(200, []ClusterMaintenanceEventsV2{{
		MaintenanceEvents: []ClusterMaintenanceEventV2{
			{
				Id: CreatedID,
			},
		},
	},
	}), nil
}

// ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2PastV2Get -
func (s *ClusterMaintenanceEventsV2APIService) ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2PastV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2PastV2Get with the required logic for this service method.
	// Add api_cluster_maintenance_events_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []ClusterMaintenanceEventsV2{}) or use other options such as http.Ok ...
	return Response(200, []ClusterMaintenanceEventsV2{{
		MaintenanceEvents: []ClusterMaintenanceEventV2{
			{
				Id: CreatedID,
			},
		},
	},
	}), nil
}

// ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2UpcomingV2Get -
func (s *ClusterMaintenanceEventsV2APIService) ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2UpcomingV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesClusterClusterIdMaintenanceEventsV2UpcomingV2Get with the required logic for this service method.
	// Add api_cluster_maintenance_events_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []ClusterMaintenanceEventSchedulesV2{}) or use other options such as http.Ok ...
	return Response(200, []ClusterMaintenanceEventsV2{{
		MaintenanceEvents: []ClusterMaintenanceEventV2{
			{
				Id: CreatedID,
			},
		},
	},
	}), nil
}

// ClusterManagementV2OperationsMaintenanceEventsMaintenanceEventIdV2RescheduleMaintenanceEventV2Put - Reschedule a maintenance event
func (s *ClusterMaintenanceEventsV2APIService) ClusterManagementV2OperationsMaintenanceEventsMaintenanceEventIdV2RescheduleMaintenanceEventV2Put(ctx context.Context, maintenanceEventId string, clusterMaintenanceEventScheduleUpdateV2 ClusterMaintenanceEventScheduleUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsMaintenanceEventsMaintenanceEventIdV2RescheduleMaintenanceEventV2Put with the required logic for this service method.
	// Add api_cluster_maintenance_events_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, ClusterMaintenanceEventScheduleV2{}) or use other options such as http.Ok ...
	return Response(200, ClusterMaintenanceEventScheduleV2{}), nil
}
