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
	ClusterStatus            `json:",inline"`
	DataNodes                []*OpenSearchDataNodes  `json:"dataNodes,omitempty"`
	PCIComplianceMode        bool                    `json:"pciComplianceMode"`
	ICUPlugin                bool                    `json:"icuPlugin"`
	OpenSearchVersion        string                  `json:"opensearchVersion"`
	AsynchronousSearchPlugin bool                    `json:"asynchronousSearchPlugin"`
	TwoFactorDelete          []*TwoFactorDelete      `json:"twoFactorDelete,omitempty"`
	KNNPlugin                bool                    `json:"knnPlugin"`
	OpenSearchDashboards     []*OpenSearchDashboards `json:"opensearchDashboards,omitempty"`
	ReportingPlugin          bool                    `json:"reportingPlugin"`
	SQLPlugin                bool                    `json:"sqlPlugin"`
	NotificationsPlugin      bool                    `json:"notificationsPlugin"`
	DataCentres              []*OpenSearchDataCentre `json:"dataCentres"`
	AnomalyDetectionPlugin   bool                    `json:"anomalyDetectionPlugin"`
	LoadBalancer             bool                    `json:"loadBalancer"`
	PrivateNetworkCluster    bool                    `json:"privateNetworkCluster"`
	Name                     string                  `json:"name"`
	BundledUseOnly           bool                    `json:"bundledUseOnly"`
	ClusterManagerNodes      []*ClusterManagerNodes  `json:"clusterManagerNodes"`
	IndexManagementPlugin    bool                    `json:"indexManagementPlugin"`
	SLATier                  string                  `json:"slaTier,omitempty"`
	AlertingPlugin           bool                    `json:"alertingPlugin"`
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
	DataCentre    `json:",inline"`
	PrivateLink   bool `json:"privateLink"`
	NumberOfRacks int  `json:"numberOfRacks"`
}

type ClusterManagerNodes struct {
	NodeSize         string `json:"nodeSize"`
	DedicatedManager bool   `json:"dedicatedManager"`
}

type OpenSearchInstAPIUpdateRequest struct {
	DataNodes            []*OpenSearchDataNodes  `json:"dataNodes,omitempty"`
	OpenSearchDashboards []*OpenSearchDashboards `json:"opensearchDashboards,omitempty"`
	ClusterManagerNodes  []*ClusterManagerNodes  `json:"clusterManagerNodes"`
}
