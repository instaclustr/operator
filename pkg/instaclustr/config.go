package instaclustr

import "time"

const (
	DefaultTimeout  = time.Second * 60
	OperatorVersion = "k8s v0.0.1"
)

// constants for API v2
const (
	CassandraEndpoint                    = "/cluster-management/v2/resources/applications/cassandra/clusters/v2/"
	KafkaEndpoint                        = "/cluster-management/v2/resources/applications/kafka/clusters/v2/"
	KafkaConnectEndpoint                 = "/cluster-management/v2/resources/applications/kafka-connect/clusters/v2/"
	KafkaTopicEndpoint                   = "/cluster-management/v2/resources/applications/kafka/topics/v2/"
	KafkaTopicConfigsUpdateEndpoint      = "/cluster-management/v2/operations/applications/kafka/topics/v2/%s/modify-configs/v2"
	AWSPeeringEndpoint                   = "/cluster-management/v2/resources/providers/aws/vpc-peers/v2/"
	AzurePeeringEndpoint                 = "/cluster-management/v2/data-sources/providers/azure/vnet-peers/v2/"
	ClusterNetworkFirewallRuleEndpoint   = "/cluster-management/v2/resources/network-firewall-rules/v2/"
	AWSSecurityGroupFirewallRuleEndpoint = "/cluster-management/v2/resources/providers/aws/security-group-firewall-rules/v2/"
	GCPPeeringEndpoint                   = "/cluster-management/v2/resources/providers/gcp/vpc-peers/v2/"
)

// constants for API v1
const (
	AddDataCentresEndpoint                 = "/cluster-data-centres"
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
