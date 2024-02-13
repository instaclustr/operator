/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

type OpenSearchCluster struct {
	GenericClusterFields `json:",inline"`

	OpenSearchVersion        string                   `json:"opensearchVersion"`
	ICUPlugin                bool                     `json:"icuPlugin"`
	AsynchronousSearchPlugin bool                     `json:"asynchronousSearchPlugin"`
	KNNPlugin                bool                     `json:"knnPlugin"`
	ReportingPlugin          bool                     `json:"reportingPlugin"`
	SQLPlugin                bool                     `json:"sqlPlugin"`
	NotificationsPlugin      bool                     `json:"notificationsPlugin"`
	AnomalyDetectionPlugin   bool                     `json:"anomalyDetectionPlugin"`
	LoadBalancer             bool                     `json:"loadBalancer"`
	BundledUseOnly           bool                     `json:"bundledUseOnly"`
	IndexManagementPlugin    bool                     `json:"indexManagementPlugin"`
	AlertingPlugin           bool                     `json:"alertingPlugin"`
	PCIComplianceMode        bool                     `json:"pciComplianceMode"`
	DataCentres              []*OpenSearchDataCentre  `json:"dataCentres"`
	DataNodes                []*OpenSearchDataNodes   `json:"dataNodes,omitempty"`
	OpenSearchDashboards     []*OpenSearchDashboards  `json:"opensearchDashboards,omitempty"`
	ClusterManagerNodes      []*ClusterManagerNodes   `json:"clusterManagerNodes"`
	IngestNodes              []*OpenSearchIngestNodes `json:"ingestNodes,omitempty"`
	ResizeSettings           []*ResizeSettings        `json:"resizeSettings,omitempty"`

	LoadBalancerConnectionURL            string `json:"loadBalancerConnectionURL,omitempty"`
	PrivateEndpoint                      string `json:"privateEndpoint,omitempty"`
	PublicEndpoint                       string `json:"publicEndpoint,omitempty"`
	IngestNodesLoadBalancerConnectionURL string `json:"ingestNodesLoadBalancerConnectionURL,omitempty"`
}

type OpenSearchIngestNodes struct {
	NodeSize  string `json:"nodeSize"`
	NodeCount int    `json:"nodeCount"`
}

type OpenSearchDataNodes struct {
	NodeSize  string `json:"nodeSize"`
	NodeCount int    `json:"nodeCount"`
}

type OpenSearchDashboards struct {
	NodeSize     string `json:"nodeSize"`
	OIDCProvider string `json:"oidcProvider,omitempty"`
	Version      string `json:"version"`
}

type OpenSearchDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	PrivateLink   bool `json:"privateLink"`
	NumberOfRacks int  `json:"numberOfRacks"`

	PrivateLinkEndpointServiceName string           `json:"privateLinkEndpointServiceName,omitempty"`
	PrivateLinkRestAPIURL          string           `json:"privateLinkRestApiUrl,omitempty"`
	Nodes                          []OpenSearchNode `json:"nodes,omitempty"`
}

type OpenSearchNode struct {
	Node `json:",inline"`

	PublicEndpoint           string `json:"publicEndpoint,omitempty"`
	PrivateEndpoint          string `json:"privateEndpoint,omitempty"`
	PrivateLinkConnectionURL string `json:"privateLinkConnectionUrl,omitempty"`
}

type ClusterManagerNodes struct {
	NodeSize         string `json:"nodeSize"`
	DedicatedManager bool   `json:"dedicatedManager"`
}

type OpenSearchInstAPIUpdateRequest struct {
	DataNodes             []*OpenSearchDataNodes   `json:"dataNodes,omitempty"`
	OpenSearchDashboards  []*OpenSearchDashboards  `json:"opensearchDashboards,omitempty"`
	ClusterManagerNodes   []*ClusterManagerNodes   `json:"clusterManagerNodes"`
	ResizeSettings        []*ResizeSettings        `json:"resizeSettings,omitempty"`
	OpenSearchIngestNodes []*OpenSearchIngestNodes `json:"ingestNodes,omitempty"`
}
