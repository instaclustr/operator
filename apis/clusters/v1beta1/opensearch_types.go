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

package v1beta1

import (
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

// OpenSearchSpec defines the desired state of OpenSearch
type OpenSearchSpec struct {
	GenericClusterSpec `json:",inline"`

	RestoreFrom              *OpenSearchRestoreFrom  `json:"restoreFrom,omitempty"`
	DataCentres              []*OpenSearchDataCentre `json:"dataCentres,omitempty"`
	DataNodes                []*OpenSearchDataNodes  `json:"dataNodes,omitempty"`
	Dashboards               []*OpenSearchDashboards `json:"opensearchDashboards,omitempty"`
	ClusterManagerNodes      []*ClusterManagerNodes  `json:"clusterManagerNodes"`
	ICUPlugin                bool                    `json:"icuPlugin,omitempty"`
	AsynchronousSearchPlugin bool                    `json:"asynchronousSearchPlugin,omitempty"`
	KNNPlugin                bool                    `json:"knnPlugin,omitempty"`
	ReportingPlugin          bool                    `json:"reportingPlugin,omitempty"`
	SQLPlugin                bool                    `json:"sqlPlugin,omitempty"`
	NotificationsPlugin      bool                    `json:"notificationsPlugin,omitempty"`
	AnomalyDetectionPlugin   bool                    `json:"anomalyDetectionPlugin,omitempty"`
	LoadBalancer             bool                    `json:"loadBalancer,omitempty"`
	IndexManagementPlugin    bool                    `json:"indexManagementPlugin,omitempty"`
	AlertingPlugin           bool                    `json:"alertingPlugin,omitempty"`
	BundledUseOnly           bool                    `json:"bundledUseOnly,omitempty"`
	UserRefs                 References              `json:"userRefs,omitempty"`
	//+kubuilder:validation:MaxItems:=1
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty"`
	//+kubuilder:validation:MaxItems:=1
	IngestNodes []*OpenSearchIngestNodes `json:"ingestNodes,omitempty"`
}

func (s *OpenSearchSpec) FromInstAPI(instaModel *models.OpenSearchCluster) {
	s.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields)

	s.Version = instaModel.OpenSearchVersion
	s.ICUPlugin = instaModel.ICUPlugin
	s.AsynchronousSearchPlugin = instaModel.AsynchronousSearchPlugin
	s.KNNPlugin = instaModel.KNNPlugin
	s.ReportingPlugin = instaModel.ReportingPlugin
	s.SQLPlugin = instaModel.SQLPlugin
	s.NotificationsPlugin = instaModel.NotificationsPlugin
	s.AnomalyDetectionPlugin = instaModel.AnomalyDetectionPlugin
	s.LoadBalancer = instaModel.LoadBalancer
	s.IndexManagementPlugin = instaModel.IndexManagementPlugin
	s.AlertingPlugin = instaModel.AlertingPlugin
	s.BundledUseOnly = instaModel.BundledUseOnly

	s.dcsFromInstAPI(instaModel.DataCentres)
	s.dataNodesFromInstAPI(instaModel.DataNodes)
	s.dashboardsFromInstAPI(instaModel.OpenSearchDashboards)
	s.ingestNodesFromInstAPI(instaModel.IngestNodes)
	s.clusterManagerNodesFromInstAPI(instaModel.ClusterManagerNodes)
}

func (s *OpenSearchSpec) dcsFromInstAPI(dcModels []*models.OpenSearchDataCentre) {
	s.DataCentres = make([]*OpenSearchDataCentre, 0, len(dcModels))
	for _, model := range dcModels {
		dc := &OpenSearchDataCentre{}
		dc.FromInstAPI(model)
		s.DataCentres = append(s.DataCentres, dc)
	}
}

func (dc *OpenSearchDataCentre) FromInstAPI(instaModel *models.OpenSearchDataCentre) {
	dc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)

	dc.PrivateLink = instaModel.PrivateLink
	dc.NumberOfRacks = instaModel.NumberOfRacks
}

func (s *OpenSearchSpec) dataNodesFromInstAPI(instaModels []*models.OpenSearchDataNodes) {
	s.DataNodes = make([]*OpenSearchDataNodes, 0, len(instaModels))
	for _, model := range instaModels {
		s.DataNodes = append(s.DataNodes, &OpenSearchDataNodes{
			NodeSize:    model.NodeSize,
			NodesNumber: model.NodeCount,
		})
	}
}

func (s *OpenSearchSpec) dashboardsFromInstAPI(instaModels []*models.OpenSearchDashboards) {
	s.Dashboards = make([]*OpenSearchDashboards, 0, len(instaModels))
	for _, model := range instaModels {
		s.Dashboards = append(s.Dashboards, &OpenSearchDashboards{
			NodeSize:     model.NodeSize,
			OIDCProvider: model.OIDCProvider,
			Version:      model.Version,
		})
	}
}

func (s *OpenSearchSpec) ingestNodesFromInstAPI(instaModels []*models.OpenSearchIngestNodes) {
	s.IngestNodes = make([]*OpenSearchIngestNodes, 0, len(instaModels))
	for _, model := range instaModels {
		s.IngestNodes = append(s.IngestNodes, &OpenSearchIngestNodes{
			NodeSize:  model.NodeSize,
			NodeCount: model.NodeCount,
		})
	}
}

func (s *OpenSearchSpec) clusterManagerNodesFromInstAPI(instaModels []*models.ClusterManagerNodes) {
	s.ClusterManagerNodes = make([]*ClusterManagerNodes, 0, len(instaModels))
	for _, model := range instaModels {
		s.ClusterManagerNodes = append(s.ClusterManagerNodes, &ClusterManagerNodes{
			NodeSize:         model.NodeSize,
			DedicatedManager: model.DedicatedManager,
		})
	}
}

type OpenSearchDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	PrivateLink   bool `json:"privateLink,omitempty"`
	NumberOfRacks int  `json:"numberOfRacks,omitempty"`
}

func (dc *OpenSearchDataCentre) Equals(o *OpenSearchDataCentre) bool {
	return dc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		dc.PrivateLink == o.PrivateLink &&
		dc.NumberOfRacks == o.NumberOfRacks
}

type OpenSearchDataNodes struct {
	NodeSize    string `json:"nodeSize"`
	NodesNumber int    `json:"nodesNumber"`
}

type OpenSearchDashboards struct {
	NodeSize     string `json:"nodeSize"`
	OIDCProvider string `json:"oidcProvider,omitempty"`
	Version      string `json:"version"`
}

type ClusterManagerNodes struct {
	NodeSize         string `json:"nodeSize"`
	DedicatedManager bool   `json:"dedicatedManager"`
}

type OpenSearchIngestNodes struct {
	NodeSize  string `json:"nodeSize"`
	NodeCount int    `json:"nodeCount"`
}

func (oss *OpenSearchSpec) ToInstAPI() *models.OpenSearchCluster {
	return &models.OpenSearchCluster{
		GenericClusterFields:     oss.GenericClusterSpec.ToInstAPI(),
		ResizeSettings:           resizeSettingsToInstAPI(oss.ResizeSettings),
		DataNodes:                oss.dataNodesToInstAPI(),
		ICUPlugin:                oss.ICUPlugin,
		OpenSearchVersion:        oss.Version,
		AsynchronousSearchPlugin: oss.AsynchronousSearchPlugin,
		KNNPlugin:                oss.KNNPlugin,
		OpenSearchDashboards:     oss.dashboardsToInstAPI(),
		ReportingPlugin:          oss.ReportingPlugin,
		SQLPlugin:                oss.SQLPlugin,
		NotificationsPlugin:      oss.NotificationsPlugin,
		DataCentres:              oss.dcsToInstAPI(),
		AnomalyDetectionPlugin:   oss.AnomalyDetectionPlugin,
		LoadBalancer:             oss.LoadBalancer,
		BundledUseOnly:           oss.BundledUseOnly,
		ClusterManagerNodes:      oss.clusterManagerNodesToInstAPI(),
		IndexManagementPlugin:    oss.IndexManagementPlugin,
		AlertingPlugin:           oss.AlertingPlugin,
		IngestNodes:              oss.ingestNodesToInstAPI(),
	}
}

func (oss *OpenSearchSpec) dcsToInstAPI() (iDCs []*models.OpenSearchDataCentre) {
	for _, dc := range oss.DataCentres {
		iDCs = append(iDCs, &models.OpenSearchDataCentre{
			GenericDataCentreFields: dc.GenericDataCentreSpec.ToInstAPI(),
			PrivateLink:             dc.PrivateLink,
			NumberOfRacks:           dc.NumberOfRacks,
		})
	}

	return
}

func tagsToInstAPI(k8sTags map[string]string) (iTags []*models.Tag) {
	for key, value := range k8sTags {
		iTags = append(iTags, &models.Tag{
			Key:   key,
			Value: value,
		})
	}

	return
}

func (oss *OpenSearchSpec) dashboardsToInstAPI() (iDashboards []*models.OpenSearchDashboards) {
	for _, dashboard := range oss.Dashboards {
		iDashboards = append(iDashboards, &models.OpenSearchDashboards{
			NodeSize:     dashboard.NodeSize,
			OIDCProvider: dashboard.OIDCProvider,
			Version:      dashboard.Version,
		})
	}

	return
}

func (oss *OpenSearchSpec) clusterManagerNodesToInstAPI() (iManagerNodes []*models.ClusterManagerNodes) {
	for _, managerNodes := range oss.ClusterManagerNodes {
		iManagerNodes = append(iManagerNodes, &models.ClusterManagerNodes{
			NodeSize:         managerNodes.NodeSize,
			DedicatedManager: managerNodes.DedicatedManager,
		})
	}

	return
}

func (oss *OpenSearchSpec) dataNodesToInstAPI() (iDataNodes []*models.OpenSearchDataNodes) {
	for _, dataNode := range oss.DataNodes {
		iDataNodes = append(iDataNodes, &models.OpenSearchDataNodes{
			NodeSize:  dataNode.NodeSize,
			NodeCount: dataNode.NodesNumber,
		})
	}

	return
}

func (oss *OpenSearchSpec) ingestNodesToInstAPI() (iIngestNodes []*models.OpenSearchIngestNodes) {
	for _, ingestNode := range oss.IngestNodes {
		iIngestNodes = append(iIngestNodes, &models.OpenSearchIngestNodes{
			NodeSize:  ingestNode.NodeSize,
			NodeCount: ingestNode.NodeCount,
		})
	}

	return
}

func (oss *OpenSearch) FromInstAPI(model *models.OpenSearchCluster) {
	oss.Spec.FromInstAPI(model)
	oss.Status.FromInstAPI(model)
}

func (oss *OpenSearch) RestoreInfoToInstAPI(restoreData *OpenSearchRestoreFrom) any {
	iRestore := struct {
		RestoredClusterName string              `json:"restoredClusterName,omitempty"`
		CDCConfigs          []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`
		PointInTime         int64               `json:"pointInTime,omitempty"`
		IndexNames          string              `json:"indexNames,omitempty"`
		ClusterID           string              `json:"clusterId,omitempty"`
	}{
		RestoredClusterName: restoreData.RestoredClusterName,
		CDCConfigs:          restoreData.CDCConfigs,
		PointInTime:         restoreData.PointInTime,
		IndexNames:          restoreData.IndexNames,
		ClusterID:           restoreData.ClusterID,
	}

	return iRestore
}

func tagsFromInstAPI(iTags []*models.Tag) map[string]string {
	newTags := map[string]string{}
	for _, iTag := range iTags {
		newTags[iTag.Key] = iTag.Value
	}
	return newTags
}

func cloudProviderSettingsFromInstAPI(iDC *models.GenericDataCentreFields) (settings []*CloudProviderSettings) {
	switch iDC.CloudProvider {
	case models.AWSVPC:
		for _, awsSetting := range iDC.AWSSettings {
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID: awsSetting.CustomVirtualNetworkID,
				DiskEncryptionKey:      awsSetting.EBSEncryptionKey,
				BackupBucket:           awsSetting.BackupBucket,
			})
		}
	case models.GCP:
		for _, gcpSetting := range iDC.GCPSettings {
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID:    gcpSetting.CustomVirtualNetworkID,
				DisableSnapshotAutoExpiry: gcpSetting.DisableSnapshotAutoExpiry,
			})
		}
	case models.AZUREAZ:
		for _, azureSetting := range iDC.AzureSettings {
			settings = append(settings, &CloudProviderSettings{
				ResourceGroup: azureSetting.ResourceGroup,
			})
		}
	}

	return settings
}

func (c *OpenSearch) GetSpec() OpenSearchSpec { return c.Spec }

func (c *OpenSearch) IsSpecEqual(spec OpenSearchSpec) bool {
	return c.Spec.IsEqual(spec)
}

func (a *OpenSearchSpec) IsEqual(b OpenSearchSpec) bool {
	return a.GenericClusterSpec.Equals(&b.GenericClusterSpec) &&
		slices.EqualsPtr(a.DataNodes, b.DataNodes) &&
		slices.EqualsPtr(a.Dashboards, b.Dashboards) &&
		slices.EqualsPtr(a.ClusterManagerNodes, b.ClusterManagerNodes) &&
		slices.EqualsPtr(a.IngestNodes, b.IngestNodes) &&
		a.ICUPlugin == b.ICUPlugin &&
		a.AsynchronousSearchPlugin == b.AsynchronousSearchPlugin &&
		a.KNNPlugin == b.KNNPlugin &&
		a.ReportingPlugin == b.ReportingPlugin &&
		a.SQLPlugin == b.SQLPlugin &&
		a.NotificationsPlugin == b.NotificationsPlugin &&
		a.AnomalyDetectionPlugin == b.AnomalyDetectionPlugin &&
		a.LoadBalancer == b.LoadBalancer &&
		a.IndexManagementPlugin == b.IndexManagementPlugin &&
		a.AlertingPlugin == b.AlertingPlugin &&
		a.BundledUseOnly == b.BundledUseOnly &&
		a.areDCsEqual(b.DataCentres)
}

func (oss *OpenSearchSpec) areDCsEqual(b []*OpenSearchDataCentre) bool {
	a := oss.DataCentres
	if len(a) != len(b) {
		return false
	}

	for i := range b {
		if a[i].Name != b[i].Name {
			continue
		}

		if !a[i].Equals(b[i]) {
			return false
		}
	}

	return true
}

func areTagsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for bKey, bValue := range b {
		if aValue, exists := a[bKey]; !exists || aValue != bValue {
			return false
		}
	}

	return true
}

func (oss *OpenSearchSpec) ToInstAPIUpdate() models.OpenSearchInstAPIUpdateRequest {
	return models.OpenSearchInstAPIUpdateRequest{
		DataNodes:             oss.dataNodesToInstAPI(),
		OpenSearchDashboards:  oss.dashboardsToInstAPI(),
		ClusterManagerNodes:   oss.clusterManagerNodesToInstAPI(),
		ResizeSettings:        resizeSettingsToInstAPI(oss.ResizeSettings),
		OpenSearchIngestNodes: oss.ingestNodesToInstAPI(),
	}
}

type OpenSearchRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterId"`

	// The display name of the restored cluster.
	RestoredClusterName string `json:"restoredClusterName,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCConfigs []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// Only data for the specified indices will be restored, for the point in time.
	IndexNames string `json:"indexNames,omitempty"`
}

// OpenSearchStatus defines the observed state of OpenSearch
type OpenSearchStatus struct {
	GenericStatus `json:",inline"`

	DataCentres []OpenSearchDataCentreStatus `json:"dataCentres,omitempty"`

	LoadBalancerConnectionURL            string `json:"loadBalancerConnectionURL,omitempty"`
	IngestNodesLoadBalancerConnectionURL string `json:"ingestNodesLoadBalancerConnectionURL,omitempty"`
	PrivateEndpoint                      string `json:"privateEndpoint,omitempty"`
	PublicEndpoint                       string `json:"publicEndpoint,omitempty"`

	AvailableUsers References `json:"availableUsers,omitempty"`
}

func (oss *OpenSearchStatus) FromInstAPI(instaModel *models.OpenSearchCluster) {
	oss.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)

	oss.PrivateEndpoint = instaModel.PrivateEndpoint
	oss.PublicEndpoint = instaModel.PublicEndpoint
	oss.IngestNodesLoadBalancerConnectionURL = instaModel.IngestNodesLoadBalancerConnectionURL
	oss.LoadBalancerConnectionURL = instaModel.LoadBalancerConnectionURL

	oss.DataCentres = make([]OpenSearchDataCentreStatus, 0, len(instaModel.DataCentres))
	for _, dc := range instaModel.DataCentres {
		d := OpenSearchDataCentreStatus{}
		d.FromInstAPI(dc)
		oss.DataCentres = append(oss.DataCentres, d)
	}
}

type OpenSearchDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	PrivateLinkEndpointServiceName string `json:"privateLinkEndpointServiceName,omitempty"`
	PrivateLinkRestAPIURL          string `json:"privateLinkRestApiUrl,omitempty"`

	Nodes []OpenSearchNodeStatus `json:"nodes,omitempty"`
}

func (s *OpenSearchDataCentreStatus) FromInstAPI(instaModel *models.OpenSearchDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)

	s.PrivateLinkRestAPIURL = instaModel.PrivateLinkRestAPIURL
	s.PrivateLinkEndpointServiceName = instaModel.PrivateLinkEndpointServiceName

	s.Nodes = make([]OpenSearchNodeStatus, 0, len(instaModel.Nodes))
	for _, node := range instaModel.Nodes {
		n := OpenSearchNodeStatus{}
		n.FromInstAPI(&node)
		s.Nodes = append(s.Nodes, n)
	}
}

type OpenSearchNodeStatus struct {
	Node `json:",inline"`

	PublicEndpoint           string `json:"publicEndpoint,omitempty"`
	PrivateEndpoint          string `json:"privateEndpoint,omitempty"`
	PrivateLinkConnectionURL string `json:"privateLinkConnectionUrl,omitempty"`
}

func (s *OpenSearchNodeStatus) FromInstAPI(instaModel *models.OpenSearchNode) {
	s.Node.FromInstAPI(&instaModel.Node)

	s.PrivateLinkConnectionURL = instaModel.PrivateLinkConnectionURL
	s.PrivateEndpoint = instaModel.PrivateEndpoint
	s.PublicEndpoint = instaModel.PublicEndpoint
}

func (s *OpenSearchNodeStatus) Equals(o *OpenSearchNodeStatus) bool {
	return s.Node.Equals(&o.Node) &&
		s.PublicEndpoint == o.PublicEndpoint &&
		s.PrivateEndpoint == o.PrivateEndpoint &&
		s.PrivateLinkConnectionURL == o.PrivateLinkConnectionURL
}

func (s *OpenSearchDataCentreStatus) Equals(o *OpenSearchDataCentreStatus) bool {
	if len(s.Nodes) != len(o.Nodes) {
		return false
	}

	compare := map[string]OpenSearchNodeStatus{}
	for _, k8sNode := range s.Nodes {
		compare[k8sNode.ID] = k8sNode
	}

	for _, iNode := range o.Nodes {
		k8sNode, ok := compare[iNode.ID]
		if !ok {
			return false
		}

		if !k8sNode.Equals(&iNode) {
			return false
		}
	}

	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		s.PrivateLinkRestAPIURL == o.PrivateLinkRestAPIURL &&
		s.PrivateLinkEndpointServiceName == o.PrivateLinkEndpointServiceName
}

func (oss *OpenSearchStatus) DataCentreEquals(s *OpenSearchStatus) bool {
	if len(oss.DataCentres) != len(s.DataCentres) {
		return false
	}

	compare := map[string]OpenSearchDataCentreStatus{}
	for _, k8sDC := range oss.DataCentres {
		compare[k8sDC.ID] = k8sDC
	}

	for _, iDC := range s.DataCentres {
		k8sDC, ok := compare[iDC.ID]
		if !ok {
			return false
		}

		if !k8sDC.Equals(&iDC) {
			return false
		}
	}

	return true
}

func (oss *OpenSearchStatus) Equals(o *OpenSearchStatus) bool {
	return oss.GenericStatus.Equals(&o.GenericStatus) &&
		oss.PublicEndpoint == o.PublicEndpoint &&
		oss.PrivateEndpoint == o.PrivateEndpoint &&
		oss.LoadBalancerConnectionURL == o.LoadBalancerConnectionURL &&
		oss.IngestNodesLoadBalancerConnectionURL == o.IngestNodesLoadBalancerConnectionURL &&
		oss.DataCentreEquals(o)
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// OpenSearch is the Schema for the opensearches API
type OpenSearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenSearchSpec   `json:"spec,omitempty"`
	Status OpenSearchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenSearchList contains a list of OpenSearch
type OpenSearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenSearch `json:"items"`
}

func (os *OpenSearch) GetJobID(jobName string) string {
	return os.Kind + "/" + client.ObjectKeyFromObject(os).String() + "/" + jobName
}

func (os *OpenSearch) NewPatch() client.Patch {
	old := os.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (os *OpenSearch) NewBackupSpec(startTimestamp int) *clusterresourcesv1beta1.ClusterBackup {
	return &clusterresourcesv1beta1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1beta1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + os.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   os.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: os.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1beta1.ClusterBackupSpec{
			ClusterRef: &clusterresourcesv1beta1.ClusterRef{
				Name:        os.Name,
				Namespace:   os.Namespace,
				ClusterKind: models.OsClusterKind,
			},
		},
	}
}

func (oss *OpenSearchSpec) HasRestore() bool {
	if oss.RestoreFrom != nil && oss.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (oss *OpenSearchSpec) validateResizeSettings(nodesNumber int) error {
	for _, rs := range oss.ResizeSettings {
		if rs.Concurrency > nodesNumber {
			return fmt.Errorf("resizeSettings.concurrency cannot be greater than number of nodes: %v", nodesNumber)
		}
	}

	return nil
}

func (oss *OpenSearch) GetUserRefs() References {
	return oss.Spec.UserRefs
}

func (oss *OpenSearch) SetUserRefs(refs References) {
	oss.Spec.UserRefs = refs
}

func (oss *OpenSearch) GetAvailableUsers() References {
	return oss.Status.AvailableUsers
}

func (oss *OpenSearch) SetAvailableUsers(users References) {
	oss.Status.AvailableUsers = users
}

func (oss *OpenSearch) GetClusterID() string {
	return oss.Status.ID
}

func (oss *OpenSearch) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return oss.Status.DataCentres[0].ID
	}
	for _, cdc := range oss.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}

func (oss *OpenSearch) SetClusterID(id string) {
	oss.Status.ID = id
}

func init() {
	SchemeBuilder.Register(&OpenSearch{}, &OpenSearchList{})
}
