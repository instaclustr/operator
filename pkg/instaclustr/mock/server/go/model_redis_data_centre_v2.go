/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RedisDataCentreV2 -
type RedisDataCentreV2 struct {

	// AWS specific settings for the Data Centre. Cannot be provided with GCP or Azure settings.
	AwsSettings []ProviderAwsSettingsV2 `json:"awsSettings,omitempty"`

	// The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between `/12` and `/22` and must be part of a private address space.
	Network string `json:"network"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which  allow you to identify, categorize and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.
	Tags []ProviderTagV2 `json:"tags,omitempty"`

	// GCP specific settings for the Data Centre. Cannot be provided with AWS or Azure settings.
	GcpSettings []ProviderGcpSettingsV2 `json:"gcpSettings,omitempty"`

	// Size of the nodes provisioned in the Data Centre. Available node sizes: <details> <summary>*Amazon Web Services* [__AWS_VPC__]</summary> <br> <details> <summary>*Africa (Cape Town)* [__AF_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>t3.medium-80-r</td> </tr> <tr> <td>t3.small-20-r</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Hong Kong)* [__AP_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Mumbai)* [__AP_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Seoul)* [__AP_NORTHEAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Singapore)* [__AP_SOUTHEAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Sydney)* [__AP_SOUTHEAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Asia Pacific (Tokyo)* [__AP_NORTHEAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Canada (Central)* [__CA_CENTRAL_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU Central (Frankfurt)* [__EU_CENTRAL_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU North (Stockholm)* [__EU_NORTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU South (Milan)* [__EU_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (Ireland)* [__EU_WEST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (London)* [__EU_WEST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*EU West (Paris)* [__EU_WEST_3__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*Middle East (Bahrain)* [__ME_SOUTH_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>t3.medium-80-r</td> </tr> <tr> <td>t3.small-20-r</td> </tr> </table> <br> </details> <details> <summary>*South America (São Paulo)* [__SA_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US East (Northern Virginia)* [__US_EAST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US East (Ohio)* [__US_EAST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US West (Northern California)* [__US_WEST_1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <details> <summary>*US West (Oregon)* [__US_WEST_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>r5.2xlarge-400-r</td> </tr> <tr> <td>r5.4xlarge-600-r</td> </tr> <tr> <td>r5.8xlarge-1000-r</td> </tr> <tr> <td>r5.large-100-r</td> </tr> <tr> <td>r5.xlarge-200-r</td> </tr> <tr> <td>r6g.2xlarge-400-r</td> </tr> <tr> <td>r6g.4xlarge-600-r</td> </tr> <tr> <td>r6g.large-100-r</td> </tr> <tr> <td>r6g.xlarge-200-r</td> </tr> <tr> <td>RDS-DEV-t4g.medium-80</td> </tr> <tr> <td>RDS-DEV-t4g.small-20</td> </tr> <tr> <td>t3.medium-80-r (deprecated)</td> </tr> <tr> <td>t3.small-20-r (deprecated)</td> </tr> </table> <br> </details> <br> </details> <details> <summary>*Microsoft Azure* [__AZURE_AZ__]</summary> <br> <details> <summary>*Australia East (NSW)* [__AUSTRALIA_EAST__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*Canada Central (Toronto)* [__CANADA_CENTRAL__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*Central US (Iowa)* [__CENTRAL_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*East US (Virginia)* [__EAST_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*East US 2 (Virginia)* [__EAST_US_2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*North Europe (Ireland)* [__NORTH_EUROPE__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*South Central US (Texas)* [__SOUTH_CENTRAL_US__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*Southeast Asia (Singapore)* [__SOUTHEAST_ASIA__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <details> <summary>*West Europe (Netherlands)* [__WEST_EUROPE__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>Standard_DS2_v2-64-r</td> </tr> <tr> <td>Standard_E16s_v3-800-r</td> </tr> <tr> <td>Standard_E2s_v3-100-r</td> </tr> <tr> <td>Standard_E4s_v3-200-r</td> </tr> <tr> <td>Standard_E8s_v3-400-r</td> </tr> </table> <br> </details> <br> </details> <details> <summary>*Google Cloud Platform* [__GCP__]</summary> <br> <details> <summary>*Central US (Iowa)* [__us-central1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Eastern Asia-Pacific (Taiwan)* [__asia-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Eastern South America (Brazil)* [__southamerica-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Eastern US (North Virginia)* [__us-east4__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Eastern US (South Carolina)* [__us-east1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Northeastern Asia-pacific (Japan)* [__asia-northeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Northeastern North America (Canada)* [__northamerica-northeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Northern Europe (Finland)* [__europe-north1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Southeastern Asia (Singapore)* [__asia-southeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Southeastern Australia (Sydney)* [__australia-southeast1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Southern Asia (India)* [__asia-south1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Belgium)* [__europe-west1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (England)* [__europe-west2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Germany)* [__europe-west3__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Netherlands)* [__europe-west4__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western Europe (Zurich)* [__europe-west6__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western US (California)* [__us-west2__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <details> <summary>*Western US (Oregon)* [__us-west1__]</summary> <br> <table> <tr> <th>Node Size</th> </tr> <tr> <td>n1-highmem-16-600-r</td> </tr> <tr> <td>n1-highmem-2-100-r</td> </tr> <tr> <td>n1-highmem-4-200-r</td> </tr> <tr> <td>n1-highmem-8-400-r</td> </tr> <tr> <td>n1-standard-1-30-r</td> </tr> </table> <br> </details> <br> </details>
	NodeSize string `json:"nodeSize"`

	//
	Nodes []NodeDetailsV2 `json:"nodes,omitempty"`

	CloudProvider CloudProviderEnumV2 `json:"cloudProvider"`

	// Azure specific settings for the Data Centre. Cannot be provided with AWS or GCP settings.
	AzureSettings []ProviderAzureSettingsV2 `json:"azureSettings,omitempty"`

	// Total number of master nodes in the Data Centre.
	MasterNodes int32 `json:"masterNodes"`

	// A logical name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// Total number of replica nodes in the Data Centre.
	ReplicaNodes int32 `json:"replicaNodes"`

	// ID of the Cluster Data Centre.
	Id string `json:"id,omitempty"`

	// Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.
	Region string `json:"region"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the \"Provider Account\" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`

	//
	PrivateLink []*PrivateLinkSettingsV2 `json:"privateLink,omitempty"`

	// Status of the Data Centre.
	Status string `json:"status,omitempty"`
}

// AssertRedisDataCentreV2Required checks if the required fields are not zero-ed
func AssertRedisDataCentreV2Required(obj RedisDataCentreV2) error {
	elements := map[string]interface{}{
		"network":       obj.Network,
		"nodeSize":      obj.NodeSize,
		"cloudProvider": obj.CloudProvider,
		"masterNodes":   obj.MasterNodes,
		"name":          obj.Name,
		"replicaNodes":  obj.ReplicaNodes,
		"region":        obj.Region,
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

// AssertRecurseRedisDataCentreV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of RedisDataCentreV2 (e.g. [][]RedisDataCentreV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseRedisDataCentreV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aRedisDataCentreV2, ok := obj.(RedisDataCentreV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertRedisDataCentreV2Required(aRedisDataCentreV2)
	})
}
