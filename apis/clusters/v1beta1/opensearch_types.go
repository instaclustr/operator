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
	"encoding/json"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

// +kubebuilder:object:generate:=false
type OpenSearchNodeTypes interface {
	OpenSearchDataNodes | OpenSearchDashboards | ClusterManagerNodes
}

// OpenSearchSpec defines the desired state of OpenSearch
type OpenSearchSpec struct {
	RestoreFrom              *OpenSearchRestoreFrom `json:"restoreFrom,omitempty"`
	Cluster                  `json:",inline"`
	DataCentres              []*OpenSearchDataCentre `json:"dataCentres,omitempty"`
	DataNodes                []*OpenSearchDataNodes  `json:"dataNodes,omitempty"`
	ICUPlugin                bool                    `json:"icuPlugin,omitempty"`
	AsynchronousSearchPlugin bool                    `json:"asynchronousSearchPlugin,omitempty"`
	KNNPlugin                bool                    `json:"knnPlugin,omitempty"`
	Dashboards               []*OpenSearchDashboards `json:"dashboards,omitempty"`
	ReportingPlugin          bool                    `json:"reportingPlugin,omitempty"`
	SQLPlugin                bool                    `json:"sqlPlugin,omitempty"`
	NotificationsPlugin      bool                    `json:"notificationsPlugin,omitempty"`
	AnomalyDetectionPlugin   bool                    `json:"anomalyDetectionPlugin,omitempty"`
	LoadBalancer             bool                    `json:"loadBalancer,omitempty"`
	ClusterManagerNodes      []*ClusterManagerNodes  `json:"clusterManagerNodes"`
	IndexManagementPlugin    bool                    `json:"indexManagementPlugin,omitempty"`
	AlertingPlugin           bool                    `json:"alertingPlugin,omitempty"`
	BundledUseOnly           bool                    `json:"bundleUseOnly,omitempty"`
	UserRefs                 []*UserReference        `json:"userRefs,omitempty"`
}

type OpenSearchDataCentre struct {
	DataCentre  `json:",inline"`
	PrivateLink bool `json:"privateLink,omitempty"`

	// ReplicationFactor is a number of racks to use when allocating data nodes.
	ReplicationFactor int `json:"replicationFactor"`
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

func (oss *OpenSearchSpec) ToInstAPI() *models.OpenSearchCluster {
	return &models.OpenSearchCluster{
		DataNodes:                oss.dataNodesToInstAPI(),
		PCIComplianceMode:        oss.PCICompliance,
		ICUPlugin:                oss.ICUPlugin,
		OpenSearchVersion:        oss.Version,
		AsynchronousSearchPlugin: oss.AsynchronousSearchPlugin,
		TwoFactorDelete:          oss.TwoFactorDeletesToInstAPI(),
		KNNPlugin:                oss.KNNPlugin,
		OpenSearchDashboards:     oss.dashboardsToInstAPI(),
		ReportingPlugin:          oss.ReportingPlugin,
		SQLPlugin:                oss.SQLPlugin,
		NotificationsPlugin:      oss.NotificationsPlugin,
		DataCentres:              oss.dcsToInstAPI(),
		AnomalyDetectionPlugin:   oss.AnomalyDetectionPlugin,
		LoadBalancer:             oss.LoadBalancer,
		PrivateNetworkCluster:    oss.PrivateNetworkCluster,
		Name:                     oss.Name,
		BundledUseOnly:           oss.BundledUseOnly,
		ClusterManagerNodes:      oss.clusterManagerNodesToInstAPI(),
		IndexManagementPlugin:    oss.IndexManagementPlugin,
		SLATier:                  oss.SLATier,
		AlertingPlugin:           oss.AlertingPlugin,
	}
}

func (oss *OpenSearchSpec) dcsToInstAPI() (iDCs []*models.OpenSearchDataCentre) {
	for _, dc := range oss.DataCentres {
		iDCs = append(iDCs, &models.OpenSearchDataCentre{
			DataCentre:    dc.DataCentre.ToInstAPI(),
			PrivateLink:   dc.PrivateLink,
			NumberOfRacks: dc.ReplicationFactor,
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

func (oss *OpenSearch) FromInstAPI(iData []byte) (*OpenSearch, error) {
	iOpenSearch := &models.OpenSearchCluster{}
	err := json.Unmarshal(iData, iOpenSearch)
	if err != nil {
		return nil, err
	}

	return &OpenSearch{
		TypeMeta:   oss.TypeMeta,
		ObjectMeta: oss.ObjectMeta,
		Spec:       oss.Spec.FromInstAPI(iOpenSearch),
		Status:     oss.Status.FromInstAPI(iOpenSearch),
	}, nil
}

func (oss *OpenSearchSpec) FromInstAPI(iOpenSearch *models.OpenSearchCluster) OpenSearchSpec {
	return OpenSearchSpec{
		Cluster: Cluster{
			Name:                  iOpenSearch.Name,
			Version:               iOpenSearch.OpenSearchVersion,
			PCICompliance:         iOpenSearch.PCIComplianceMode,
			PrivateNetworkCluster: iOpenSearch.PrivateNetworkCluster,
			SLATier:               iOpenSearch.SLATier,
			TwoFactorDelete:       oss.Cluster.TwoFactorDeleteFromInstAPI(iOpenSearch.TwoFactorDelete),
		},
		DataCentres:              oss.DCsFromInstAPI(iOpenSearch.DataCentres),
		DataNodes:                oss.DataNodesFromInstAPI(iOpenSearch.DataNodes),
		ICUPlugin:                oss.ICUPlugin,
		AsynchronousSearchPlugin: oss.AsynchronousSearchPlugin,
		KNNPlugin:                oss.KNNPlugin,
		Dashboards:               oss.DashboardsFromInstAPI(iOpenSearch.OpenSearchDashboards),
		ReportingPlugin:          oss.ReportingPlugin,
		SQLPlugin:                oss.SQLPlugin,
		NotificationsPlugin:      oss.NotificationsPlugin,
		AnomalyDetectionPlugin:   oss.AnomalyDetectionPlugin,
		LoadBalancer:             oss.LoadBalancer,
		ClusterManagerNodes:      oss.ClusterManagerNodesFromInstAPI(iOpenSearch.ClusterManagerNodes),
		IndexManagementPlugin:    oss.IndexManagementPlugin,
		AlertingPlugin:           oss.AlertingPlugin,
		BundledUseOnly:           oss.BundledUseOnly,
	}
}

func (oss *OpenSearchSpec) DCsFromInstAPI(iDCs []*models.OpenSearchDataCentre) (dcs []*OpenSearchDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &OpenSearchDataCentre{
			DataCentre:        oss.Cluster.DCFromInstAPI(iDC.DataCentre),
			PrivateLink:       iDC.PrivateLink,
			ReplicationFactor: iDC.NumberOfRacks,
		})
	}

	return
}

func (oss *OpenSearchSpec) DataNodesFromInstAPI(iDataNodes []*models.OpenSearchDataNodes) (dataNodes []*OpenSearchDataNodes) {
	for _, iNode := range iDataNodes {
		dataNodes = append(dataNodes, &OpenSearchDataNodes{
			NodeSize:    iNode.NodeSize,
			NodesNumber: iNode.NodeCount,
		})
	}
	return
}

func (oss *OpenSearchSpec) DashboardsFromInstAPI(iDashboards []*models.OpenSearchDashboards) (dashboards []*OpenSearchDashboards) {
	for _, iDashboard := range iDashboards {
		dashboards = append(dashboards, &OpenSearchDashboards{
			NodeSize:     iDashboard.NodeSize,
			OIDCProvider: iDashboard.OIDCProvider,
			Version:      iDashboard.Version,
		})
	}
	return
}

func (oss *OpenSearchSpec) ClusterManagerNodesFromInstAPI(iManagerNodes []*models.ClusterManagerNodes) (managerNodes []*ClusterManagerNodes) {
	for _, iNode := range iManagerNodes {
		managerNodes = append(managerNodes, &ClusterManagerNodes{
			NodeSize:         iNode.NodeSize,
			DedicatedManager: iNode.DedicatedManager,
		})
	}
	return
}

func (oss *OpenSearchStatus) FromInstAPI(iOpenSearch *models.OpenSearchCluster) OpenSearchStatus {
	return OpenSearchStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iOpenSearch.ID,
			State:                         iOpenSearch.Status,
			DataCentres:                   oss.DCsFromInstAPI(iOpenSearch.DataCentres),
			CurrentClusterOperationStatus: iOpenSearch.CurrentClusterOperationStatus,
			MaintenanceEvents:             oss.MaintenanceEvents,
		},
	}
}

func (oss *OpenSearchStatus) DCsFromInstAPI(iDCs []*models.OpenSearchDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, oss.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func (a *OpenSearchSpec) IsEqual(b OpenSearchSpec) bool {
	return a.Cluster.IsEqual(b.Cluster) &&
		a.IsTwoFactorDeleteEqual(b.TwoFactorDelete) &&
		areOpenSearchSettingsEqual[OpenSearchDataNodes](a.DataNodes, b.DataNodes) &&
		a.ICUPlugin == b.ICUPlugin &&
		a.AsynchronousSearchPlugin == b.AsynchronousSearchPlugin &&
		a.KNNPlugin == b.KNNPlugin &&
		areOpenSearchSettingsEqual[OpenSearchDashboards](a.Dashboards, b.Dashboards) &&
		a.ReportingPlugin == b.ReportingPlugin &&
		a.SQLPlugin == b.SQLPlugin &&
		a.NotificationsPlugin == b.NotificationsPlugin &&
		a.AnomalyDetectionPlugin == b.AnomalyDetectionPlugin &&
		a.LoadBalancer == b.LoadBalancer &&
		areOpenSearchSettingsEqual[ClusterManagerNodes](a.ClusterManagerNodes, b.ClusterManagerNodes) &&
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
		if !a[i].DataCentre.IsEqual(b[i].DataCentre) ||
			a[i].PrivateLink != b[i].PrivateLink ||
			a[i].ReplicationFactor != b[i].ReplicationFactor {
			return false
		}
	}

	return true
}

func areOpenSearchSettingsEqual[T OpenSearchNodeTypes](a, b []*T) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if *a[i] != *b[i] {
			return false
		}
	}

	return true
}

func (oss *OpenSearchSpec) ToInstAPIUpdate() models.OpenSearchInstAPIUpdateRequest {
	return models.OpenSearchInstAPIUpdateRequest{
		DataNodes:            oss.dataNodesToInstAPI(),
		OpenSearchDashboards: oss.dashboardsToInstAPI(),
		ClusterManagerNodes:  oss.clusterManagerNodesToInstAPI(),
	}
}

type OpenSearchRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterId"`

	// The display name of the restored cluster.
	ClusterNameOverride string `json:"clusterNameOverride,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCInfos []*OpenSearchRestoreCDCInfo `json:"cdcInfos,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// Only data for the specified indices will be restored, for the point in time.
	IndexNames string `json:"indexNames,omitempty"`

	// The cluster network for this cluster to be restored to.
	ClusterNetwork string `json:"clusterNetwork,omitempty"`
}

type OpenSearchRestoreCDCInfo struct {
	CDCID            string `json:"cdcId,omitempty"`
	RestoreToSameVPC bool   `json:"restoreToSameVpc,omitempty"`
	CustomVPCID      string `json:"customVpcId,omitempty"`
	CustomVPCNetwork string `json:"customVpcNetwork,omitempty"`
}

// OpenSearchStatus defines the observed state of OpenSearch
type OpenSearchStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
	return client.ObjectKeyFromObject(os).String() + "/" + jobName
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
			ClusterID:   os.Status.ID,
			ClusterKind: models.OsClusterKind,
		},
	}
}

func (oss *OpenSearchSpec) HasRestore() bool {
	if oss.RestoreFrom != nil && oss.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func init() {
	SchemeBuilder.Register(&OpenSearch{}, &OpenSearchList{})
}
