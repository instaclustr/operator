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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/utils/strings/slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

type OpenSearchDataCentre struct {
	DataCentre                   `json:",inline"`
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
	AlertingPlugin               bool   `json:"alertingPlugin,omitempty"`
	ICUPlugin                    bool   `json:"icuPlugin,omitempty"`
	KNNPlugin                    bool   `json:"knnPlugin,omitempty"`
	NotificationsPlugin          bool   `json:"notificationsPlugin,omitempty"`
	ReportsPlugin                bool   `json:"reportsPlugin,omitempty"`
	RacksNumber                  int    `json:"racksNumber"`
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

// OpenSearchSpec defines the desired state of OpenSearch
type OpenSearchSpec struct {
	RestoreFrom           *OpenSearchRestoreFrom `json:"restoreFrom,omitempty"`
	Cluster               `json:",inline"`
	DataCentres           []*OpenSearchDataCentre `json:"dataCentres,omitempty"`
	ConcurrentResizes     int                     `json:"concurrentResizes,omitempty"`
	NotifySupportContacts bool                    `json:"notifySupportContacts,omitempty"`
	Description           string                  `json:"description,omitempty"`
	PrivateLink           *PrivateLinkV1          `json:"privateLink,omitempty"`
	BundledUseOnly        bool                    `json:"bundleUseOnly,omitempty"`
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

type immutableOpenSearchFields struct {
	//TODO Add version validation when APIv2 lifecycle is implemented
	Name                  string
	PCICompliance         bool
	PrivateNetworkCluster bool
	SLATier               string
}

type immutableOpenSearchDCFields struct {
	immutableDC
	specificOpenSearchDCFields
}

type specificOpenSearchDCFields struct {
	DedicatedMasterNodes  bool
	IndexManagementPlugin bool
	AlertingPlugin        bool
	ICUPlugin             bool
	KNNPlugin             bool
	NotificationsPlugin   bool
	ReportsPlugin         bool
}

func (os *OpenSearch) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(os).String() + "/" + jobName
}

func (os *OpenSearch) NewPatch() client.Patch {
	old := os.DeepCopy()
	return client.MergeFrom(old)
}

func (os *OpenSearch) NewBackupSpec(startTimestamp int) *clusterresourcesv1alpha1.ClusterBackup {
	return &clusterresourcesv1alpha1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1alpha1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + os.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   os.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: os.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
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

func (os *OpenSearch) FromInstAPIv1(iData []byte) (*OpenSearch, error) {
	iOs := models.OpenSearchClusterV1{}
	err := json.Unmarshal(iData, &iOs)
	if err != nil {
		return nil, err
	}

	return &OpenSearch{
		TypeMeta:   os.TypeMeta,
		ObjectMeta: os.ObjectMeta,
		Spec:       os.Spec.FromInstAPIv1(iOs),
		Status:     os.Status.FromInstAPIv1(iOs),
	}, nil
}

func (oss *OpenSearchSpec) FromInstAPIv1(iOs models.OpenSearchClusterV1) OpenSearchSpec {
	os := OpenSearchSpec{
		Cluster: Cluster{
			Name:                  iOs.ClusterName,
			Version:               iOs.BundleVersion,
			PCICompliance:         iOs.PCICompliance != models.Disabled,
			PrivateNetworkCluster: iOs.PrivateNetworkCluster,
			SLATier:               iOs.SLATier,
		},
		DataCentres:           oss.DCsFromInstAPIv1(iOs),
		ConcurrentResizes:     oss.ConcurrentResizes,
		NotifySupportContacts: oss.NotifySupportContacts,
		Description:           oss.Description,
		BundledUseOnly:        iOs.BundledUseOnlyCluster,
	}
	if len(iOs.DataCentre) != 0 &&
		iOs.DataCentres[0].PrivateLink != nil {
		os.PrivateLink = &PrivateLinkV1{
			IAMPrincipalARNs: iOs.DataCentres[0].PrivateLink.IAMPrincipalARNs,
		}
	}
	return os
}

func (oss *OpenSearchSpec) DCsFromInstAPIv1(iOs models.OpenSearchClusterV1) (dcs []*OpenSearchDataCentre) {
	for _, iDC := range iOs.DataCentres {
		var provider *models.ClusterProviderV1
		for _, iProvider := range iOs.ClusterProvider {
			if iProvider.Name == iDC.Provider {
				provider = iProvider
				break
			}
		}
		dataCentre := DataCentre{
			Name:                iDC.CDCName,
			Region:              iDC.Name,
			CloudProvider:       iDC.Provider,
			ProviderAccountName: provider.AccountName,
			CloudProviderSettings: []*CloudProviderSettings{
				{
					CustomVirtualNetworkID: provider.CustomVirtualNetworkID,
					ResourceGroup:          provider.ResourceGroup,
					DiskEncryptionKey:      provider.DiskEncryptionKey,
				},
			},
			NodeSize:    iOs.BundleOptions.DataNodeSize,
			Network:     iDC.CDCNetwork,
			NodesNumber: iDC.NodeCount,
			Tags:        provider.Tags,
		}
		if iOs.BundleOptions.DataNodeSize == "" {
			dataCentre.NodeSize = iOs.BundleOptions.MasterNodeSize
			iOs.BundleOptions.MasterNodeSize = ""
		}

		dcs = append(dcs, &OpenSearchDataCentre{
			DataCentre:                   dataCentre,
			DedicatedMasterNodes:         iOs.BundleOptions.DedicatedMasterNodes,
			MasterNodeSize:               iOs.BundleOptions.MasterNodeSize,
			OpenSearchDashboardsNodeSize: iOs.BundleOptions.OpenSearchDashboardsNodeSize,
			IndexManagementPlugin:        iOs.BundleOptions.IndexManagementPlugin,
			AlertingPlugin:               iOs.BundleOptions.AlertingPlugin,
			ICUPlugin:                    iOs.BundleOptions.ICUPlugin,
			KNNPlugin:                    iOs.BundleOptions.KNNPlugin,
			NotificationsPlugin:          iOs.BundleOptions.NotificationsPlugin,
			ReportsPlugin:                iOs.BundleOptions.ReportsPlugin,
		})
	}
	return
}

func (oss *OpenSearchStatus) FromInstAPIv1(iOs models.OpenSearchClusterV1) OpenSearchStatus {
	return OpenSearchStatus{
		ClusterStatus: ClusterStatus{
			ID:                iOs.ID,
			State:             iOs.ClusterStatus,
			DataCentres:       oss.DCsFromInstAPIv1(iOs.DataCentres),
			MaintenanceEvents: oss.MaintenanceEvents,
			CDCID:             iOs.CDCID,
		},
	}
}

func (oss *OpenSearchStatus) DCsFromInstAPIv1(iDCs []*models.OpenSearchDataCentreV1) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &DataCentreStatus{
			ID:         iDC.ID,
			Status:     iDC.CDCStatus,
			Nodes:      oss.NodesFromInstAPIv1(iDC.Nodes),
			NodeNumber: iDC.NodeCount,
		})
	}
	return
}

func (oss *OpenSearchSpec) IsEqual(iSpec OpenSearchSpec) bool {
	return oss.Cluster.IsEqual(iSpec.Cluster) &&
		oss.AreDCsEqual(iSpec.DataCentres) &&
		oss.IsPrivateLinkEqual(iSpec.PrivateLink)
}

func (oss *OpenSearchSpec) AreDCsEqual(dcs []*OpenSearchDataCentre) bool {
	if len(oss.DataCentres) != len(dcs) {
		return false
	}

	for i, bDC := range dcs {
		aDC := oss.DataCentres[i]
		if !aDC.IsEqual(bDC.DataCentre) ||
			aDC.DedicatedMasterNodes != bDC.DedicatedMasterNodes ||
			aDC.MasterNodeSize != bDC.MasterNodeSize ||
			aDC.OpenSearchDashboardsNodeSize != bDC.OpenSearchDashboardsNodeSize ||
			aDC.IndexManagementPlugin != bDC.IndexManagementPlugin ||
			aDC.AlertingPlugin != bDC.AlertingPlugin ||
			aDC.ICUPlugin != bDC.ICUPlugin ||
			aDC.KNNPlugin != bDC.KNNPlugin ||
			aDC.NotificationsPlugin != bDC.NotificationsPlugin ||
			aDC.ReportsPlugin != bDC.ReportsPlugin ||
			aDC.RacksNumber != bDC.RacksNumber {
			return false
		}
	}

	return true
}

func (oss *OpenSearchSpec) IsPrivateLinkEqual(iLink *PrivateLinkV1) bool {
	if oss.PrivateLink == iLink {
		return true
	}

	if oss.PrivateLink != nil &&
		len(oss.PrivateLink.IAMPrincipalARNs) != len(iLink.IAMPrincipalARNs) {
		return false
	}

	for _, arn := range oss.PrivateLink.IAMPrincipalARNs {
		if !slices.Contains(iLink.IAMPrincipalARNs, arn) {
			return false
		}
	}

	return true
}

func (oss *OpenSearchSpec) ValidateImmutableFieldsUpdate(oldSpec OpenSearchSpec) error {
	newImmutableFields := oss.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable spec fields: old spec: %+v: new spec: %+v", oldImmutableFields, newImmutableFields)
	}

	err := validateTwoFactorDelete(oss.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = oss.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	return nil
}

func (oss *OpenSearchSpec) newImmutableFields() *immutableOpenSearchFields {
	return &immutableOpenSearchFields{
		Name:                  oss.Name,
		PCICompliance:         oss.PCICompliance,
		PrivateNetworkCluster: oss.PrivateNetworkCluster,
		SLATier:               oss.SLATier,
	}
}

func (odc *OpenSearchDataCentre) newImmutableFields() *immutableOpenSearchDCFields {
	return &immutableOpenSearchDCFields{
		immutableDC: immutableDC{
			Name:                odc.Name,
			Region:              odc.Region,
			CloudProvider:       odc.CloudProvider,
			ProviderAccountName: odc.ProviderAccountName,
			Network:             odc.Network,
		},
		specificOpenSearchDCFields: specificOpenSearchDCFields{
			DedicatedMasterNodes:  odc.DedicatedMasterNodes,
			IndexManagementPlugin: odc.IndexManagementPlugin,
			AlertingPlugin:        odc.AlertingPlugin,
			ICUPlugin:             odc.ICUPlugin,
			KNNPlugin:             odc.KNNPlugin,
			NotificationsPlugin:   odc.NotificationsPlugin,
			ReportsPlugin:         odc.ReportsPlugin,
		},
	}
}

func (oss *OpenSearchSpec) validateImmutableDataCentresFieldsUpdate(oldSpec OpenSearchSpec) error {
	if len(oss.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range oss.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		// TODO add CloudProviderSettings immutable fields validation when APIv2 is implemented
	}

	return nil
}

func (odc *OpenSearchDataCentre) bundleToInstAPIv1(version string) []*models.OpenSearchBundleV1 {
	var newBundles []*models.OpenSearchBundleV1
	bundle := &models.OpenSearchBundleV1{
		Bundle: models.Bundle{
			Bundle:  models.OpenSearchV1,
			Version: version,
		},
		Options: &models.OpenSearchBundleOptionsV1{
			DedicatedMasterNodes:         odc.DedicatedMasterNodes,
			MasterNodeSize:               odc.MasterNodeSize,
			OpenSearchDashboardsNodeSize: odc.OpenSearchDashboardsNodeSize,
			IndexManagementPlugin:        odc.IndexManagementPlugin,
			ICUPlugin:                    odc.ICUPlugin,
			KNNPlugin:                    odc.KNNPlugin,
			NotificationsPlugin:          odc.NotificationsPlugin,
			ReportsPlugin:                odc.ReportsPlugin,
		},
	}
	newBundles = append(newBundles, bundle)

	return newBundles
}

func (odc *OpenSearchDataCentre) providerToInstAPIv1() *models.ClusterProviderV1 {
	var iCustomVirtualNetworkID string
	var iResourceGroup string
	var iDiskEncryptionKey string
	if len(odc.CloudProviderSettings) > 0 {
		iCustomVirtualNetworkID = odc.CloudProviderSettings[0].CustomVirtualNetworkID
		iResourceGroup = odc.CloudProviderSettings[0].ResourceGroup
		iDiskEncryptionKey = odc.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProviderV1{
		Name:                   odc.CloudProvider,
		AccountName:            odc.ProviderAccountName,
		Tags:                   odc.Tags,
		CustomVirtualNetworkID: iCustomVirtualNetworkID,
		ResourceGroup:          iResourceGroup,
		DiskEncryptionKey:      iDiskEncryptionKey,
	}
}

func (oss *OpenSearchSpec) NewCreateRequestInstAPIv1() (iOs models.OpenSearchCreateAPIv1) {
	iOs.ClusterName = oss.Name
	iOs.PrivateNetworkCluster = oss.PrivateNetworkCluster
	iOs.SLATier = oss.SLATier
	iOs.BundledUseOnlyCluster = oss.BundledUseOnly
	iOs.PCICompliantCluster = oss.PCICompliance
	iOs.TwoFactorDelete = oss.Cluster.TwoFactorDeleteToInstAPIv1()
	if len(oss.DataCentres) != 0 {
		dc := oss.DataCentres[0]
		iOs.NodeSize = dc.NodeSize
		iOs.ClusterNetwork = dc.Network
		iOs.DataCentre = dc.Region
		iOs.DataCentreCustomName = dc.Name
		iOs.Bundles = dc.bundleToInstAPIv1(oss.Version)
		iOs.Provider = dc.providerToInstAPIv1()
		iOs.RackAllocation = &models.RackAllocationV1{
			NumberOfRacks: dc.RacksNumber,
			NodesPerRack:  dc.NodesNumber,
		}
	}
	return
}

func init() {
	SchemeBuilder.Register(&OpenSearch{}, &OpenSearchList{})
}
