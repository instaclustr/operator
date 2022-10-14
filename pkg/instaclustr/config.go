package instaclustr

import "time"

const (
	DefaultTimeout  = time.Second * 60
	OperatorVersion = "k8s v0.0.1"
)

// constants for API v2
const (
	CassandraEndpoint = "/cluster-management/v2/resources/applications/cassandra/clusters/v2/"
	KafkaEndpoint     = "/cluster-management/v2/resources/applications/kafka/clusters/v2/"
)

// constants for API v1
const (
	ClusterConfigurationsEndpoint          = "/configurations"
	ClusterConfigurationsParameterEndpoint = "/parameter/"
	ActiveOnly                             = "activeOnly=true"
	ClusterEvents                          = "/events"
	TerraformDescription                   = "/terraform-description"
	ClustersCreationEndpoint               = "/provisioning/v1/extended/"

	// ClustersEndpoint is used for GET, DELETE and UPDATE clusters
	ClustersEndpointV1 = "/provisioning/v1/"

	// ClustersResizeEndpoint is used for nodes resizing in a cluster data centre.
	// Example: fmt.Sprintf("%s/provisioning/v1/%s/%s/resize", serverHostname, clusterID, dataCentreID)
	ClustersResizeEndpoint = "%s/provisioning/v1/%s/%s/resize"
)
