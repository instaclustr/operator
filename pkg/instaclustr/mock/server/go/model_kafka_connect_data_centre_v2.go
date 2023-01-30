/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaConnectDataCentreV2 -
type KafkaConnectDataCentreV2 struct {

	// Number of racks to use when allocating nodes.
	ReplicationFactor int32 `json:"replicationFactor"`

	// AWS specific settings for the Data Centre. Cannot be provided with GCP or Azure settings.
	AwsSettings []ProviderAwsSettingsV2 `json:"awsSettings,omitempty"`

	// Total number of nodes in the Data Centre. Must be a multiple of `replicationFactor`.
	NumberOfNodes int32 `json:"numberOfNodes"`

	// The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between `/12` and `/22` and must be part of a private address space.
	Network string `json:"network"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which  allow you to identify, categorize and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.
	Tags []ProviderTagV2 `json:"tags,omitempty"`

	// GCP specific settings for the Data Centre. Cannot be provided with AWS or Azure settings.
	GcpSettings []ProviderGcpSettingsV2 `json:"gcpSettings,omitempty"`

	// Size of the nodes provisioned in the Data Centre. Available node sizes: <details> <summary>*Google Cloud Platform* [__GCP__]</summary> <br> <details> <summary>*Central US (Iowa)* [__us-central1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Eastern Asia-Pacific (Taiwan)* [__asia-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Eastern South America (Brazil)* [__southamerica-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Eastern US (North Virginia)* [__us-east4__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Eastern US (South Carolina)* [__us-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Northeastern Asia-pacific (Japan)* [__asia-northeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Northeastern North America (Canada)* [__northamerica-northeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Northern Europe (Finland)* [__europe-north1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Southeastern Asia (Singapore)* [__asia-southeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Southeastern Australia (Sydney)* [__australia-southeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Southern Asia (India)* [__asia-south1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Belgium)* [__europe-west1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (England)* [__europe-west2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Germany)* [__europe-west3__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Netherlands)* [__europe-west4__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Zurich)* [__europe-west6__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western US (California)* [__us-west2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <details> <summary>*Western US (Oregon)* [__us-west1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-standard-16-40</td> </tr> <tr> <td>n1-standard-2-10</td> </tr> <tr> <td>n1-standard-4-10</td> </tr> <tr> <td>n1-standard-8-20</td> </tr> </table> <br> </details> <br> </details> <details> <summary>*Amazon Web Services* [__AWS_VPC__]</summary> <br> <details> <summary>*Africa (Cape Town)* [__AF_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>c5.2xlarge-20-gp2</td> </tr> <tr> <td>c5.4xlarge-40-gp2</td> </tr> <tr> <td>c5.xlarge-10-gp2</td> </tr> <tr> <td>r5.2xlarge-20-gp2</td> </tr> <tr> <td>r5.4xlarge-40-gp2</td> </tr> <tr> <td>r5.xlarge-10-gp2</td> </tr> <tr> <td>t3.medium-10-gp2</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Hong Kong)* [__AP_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Mumbai)* [__AP_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Seoul)* [__AP_NORTHEAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Singapore)* [__AP_SOUTHEAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Sydney)* [__AP_SOUTHEAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Tokyo)* [__AP_NORTHEAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Canada (Central)* [__CA_CENTRAL_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU Central (Frankfurt)* [__EU_CENTRAL_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU North (Stockholm)* [__EU_NORTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU South (Milan)* [__EU_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (Ireland)* [__EU_WEST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (London)* [__EU_WEST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (Paris)* [__EU_WEST_3__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Middle East (Bahrain)* [__ME_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>c5.2xlarge-20-gp2</td> </tr> <tr> <td>c5.4xlarge-40-gp2</td> </tr> <tr> <td>c5.xlarge-10-gp2</td> </tr> <tr> <td>r5.2xlarge-20-gp2</td> </tr> <tr> <td>r5.4xlarge-40-gp2</td> </tr> <tr> <td>r5.xlarge-10-gp2</td> </tr> <tr> <td>t3.medium-10-gp2</td> </tr> </table> <br> </details> <details> <summary>*South America (São Paulo)* [__SA_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US East (Northern Virginia)* [__US_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US East (Ohio)* [__US_EAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US West (Northern California)* [__US_WEST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US West (Oregon)* [__US_WEST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>KCN-DEV-t4g.medium-30</td> </tr> <tr> <td>KCN-PRD-c6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-c6g.large-80</td> </tr> <tr> <td>KCN-PRD-c6g.xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.2xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.4xlarge-80</td> </tr> <tr> <td>KCN-PRD-r6g.large-80</td> </tr> <tr> <td>KCN-PRD-r6g.xlarge-80</td> </tr> <tr> <td>c5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>c5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>c5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>r5.2xlarge-20-gp2 (deprecated)</td> </tr> <tr> <td>r5.4xlarge-40-gp2 (deprecated)</td> </tr> <tr> <td>r5.xlarge-10-gp2 (deprecated)</td> </tr> <tr> <td>t3.medium-10-gp2 (deprecated)</td> </tr> </table> <br> </details> <br> </details> <details> <summary>*Microsoft Azure* [__AZURE_AZ__]</summary> <br> <details> <summary>*Australia East (NSW)* [__AUSTRALIA_EAST__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*Canada Central (Toronto)* [__CANADA_CENTRAL__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*Central US (Iowa)* [__CENTRAL_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*East US (Virginia)* [__EAST_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*East US 2 (Virginia)* [__EAST_US_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> </table> <br> </details> <details> <summary>*North Europe (Ireland)* [__NORTH_EUROPE__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*South Central US (Texas)* [__SOUTH_CENTRAL_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*Southeast Asia (Singapore)* [__SOUTHEAST_ASIA__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*West Europe (Netherlands)* [__WEST_EUROPE__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D16s_v3-40</td> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> <tr> <td>Standard_D8s_v3-20</td> </tr> </table> <br> </details> <details> <summary>*West US 2 (Washington)* [__WEST_US_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_D2s_v3-10</td> </tr> <tr> <td>Standard_D4s_v3-10</td> </tr> </table> <br> </details> <br> </details>
	NodeSize string `json:"nodeSize"`

	//
	Nodes []NodeDetailsV2 `json:"nodes,omitempty"`

	CloudProvider CloudProviderEnumV2 `json:"cloudProvider"`

	// Azure specific settings for the Data Centre. Cannot be provided with AWS or GCP settings.
	AzureSettings []ProviderAzureSettingsV2 `json:"azureSettings,omitempty"`

	// A logical name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// ID of the Cluster Data Centre.
	Id string `json:"id,omitempty"`

	// Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.
	Region string `json:"region"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the \"Provider Account\" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`

	// Status of the Data Centre.
	Status string `json:"status,omitempty"`
}

// AssertKafkaConnectDataCentreV2Required checks if the required fields are not zero-ed
func AssertKafkaConnectDataCentreV2Required(obj KafkaConnectDataCentreV2) error {
	elements := map[string]interface{}{
		"replicationFactor": obj.ReplicationFactor,
		"numberOfNodes":     obj.NumberOfNodes,
		"network":           obj.Network,
		"nodeSize":          obj.NodeSize,
		"cloudProvider":     obj.CloudProvider,
		"name":              obj.Name,
		"region":            obj.Region,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.AwsSettings {
		if err := AssertProviderAwsSettingsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.Tags {
		if err := AssertProviderTagV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.GcpSettings {
		if err := AssertProviderGcpSettingsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.Nodes {
		if err := AssertNodeDetailsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.AzureSettings {
		if err := AssertProviderAzureSettingsV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseKafkaConnectDataCentreV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaConnectDataCentreV2 (e.g. [][]KafkaConnectDataCentreV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaConnectDataCentreV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaConnectDataCentreV2, ok := obj.(KafkaConnectDataCentreV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaConnectDataCentreV2Required(aKafkaConnectDataCentreV2)
	})
}
