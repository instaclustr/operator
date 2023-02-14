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
	"fmt"
	"strconv"

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
	RacksNumber                  int32  `json:"racksNumber"`
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
	PrivateLink           *PrivateLink            `json:"privateLink,omitempty"`
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

func (oss *OpenSearchSpec) IsSpecEqual(instSpec *models.ClusterV1) bool {
	if len(oss.DataCentres) == 0 ||
		len(instSpec.DataCentres) != len(oss.DataCentres) ||
		oss.Version != instSpec.BundleVersion ||
		oss.SLATier != instSpec.SLATier ||
		oss.Name != instSpec.ClusterName {
		return false
	}

	for _, instDataCentre := range instSpec.DataCentres {
		for _, dataCentre := range oss.DataCentres {
			if dataCentre.Name == instDataCentre.CDCName {
				if !dataCentre.AreOptionsEqual(*instSpec.BundleOptions) ||
					!dataCentre.IsDataCentreEqual(instDataCentre, oss.PrivateLink) ||
					!dataCentre.IsClusterProviderEqual(instSpec.ClusterProvider) ||
					oss.PrivateNetworkCluster != instDataCentre.PrivateIPOnly {
					return false
				}

				break
			}
		}
	}

	return true
}

func (odc *OpenSearchDataCentre) AreOptionsEqual(instOptions models.BundleOptions) bool {
	if odc.ReportsPlugin != instOptions.ReportsPlugin ||
		odc.KNNPlugin != instOptions.KNNPlugin ||
		odc.NotificationsPlugin != instOptions.NotificationsPlugin ||
		odc.ICUPlugin != instOptions.ICUPlugin ||
		odc.AlertingPlugin != instOptions.AlertingPlugin ||
		odc.IndexManagementPlugin != instOptions.IndexManagementPlugin ||
		(instOptions.DedicatedMasterNodes && odc.NodeSize != instOptions.DataNodeSize) ||
		(!instOptions.DedicatedMasterNodes && odc.NodeSize != instOptions.MasterNodeSize) ||
		(odc.MasterNodeSize != "" && odc.MasterNodeSize != instOptions.MasterNodeSize) ||
		odc.OpenSearchDashboardsNodeSize != instOptions.OpenSearchDashboardsNodeSize {
		return false
	}

	return true
}

func (odc *OpenSearchDataCentre) IsDataCentreEqual(instDC *models.DataCentreV1, privateLink *PrivateLink) bool {
	if odc.Name != instDC.CDCName ||
		odc.Region != instDC.Name ||
		odc.CloudProvider != instDC.Provider ||
		odc.Network != instDC.CDCNetwork {
		return false
	}

	if instDC.PrivateLink != nil {
		for _, instLink := range instDC.PrivateLink.IAMPrincipalARNs {
			foundLink := false
			for _, k8sLink := range privateLink.IAMPrincipalARNs {
				if instLink == k8sLink {
					foundLink = true
					break
				}
			}

			if !foundLink {
				return false
			}
		}
	}

	return true
}

func (odc *OpenSearchDataCentre) IsClusterProviderEqual(instClusterProvider []*models.ClusterProviderV1) bool {
	for _, instProvider := range instClusterProvider {
		if odc.CloudProvider != instProvider.Name ||
			odc.ProviderAccountName != instProvider.AccountName {
			return false
		}

		for _, providerSettings := range odc.CloudProviderSettings {
			if instProvider.DiskEncryptionKey != providerSettings.DiskEncryptionKey ||
				instProvider.ResourceGroup != providerSettings.ResourceGroup ||
				instProvider.CustomVirtualNetworkID != providerSettings.CustomVirtualNetworkID {
				return false
			}

			for key, value := range instProvider.Tags {
				if odc.Tags[key] != value {
					return false
				}
			}
		}
	}
	return true
}

func (oss *OpenSearchSpec) SetSpecFromInst(instSpec *models.ClusterV1) bool {
	oss.Name = instSpec.ClusterName
	oss.Version = instSpec.BundleVersion
	oss.SLATier = instSpec.SLATier
	oss.PCICompliance = instSpec.PCICompliance != models.Disabled

	k8sDataCentres := []*OpenSearchDataCentre{}
	for _, instDC := range instSpec.DataCentres {
		oss.PrivateNetworkCluster = instDC.PrivateIPOnly
		if instDC.PrivateLink != nil {
			oss.PrivateLink = &PrivateLink{IAMPrincipalARNs: instDC.PrivateLink.IAMPrincipalARNs}
		}

		dcToAppend := &OpenSearchDataCentre{
			DataCentre: DataCentre{
				Name:          instDC.CDCName,
				Region:        instDC.Name,
				CloudProvider: instDC.Provider,
				Network:       instDC.CDCNetwork,
			},
			DedicatedMasterNodes:         instSpec.BundleOptions.DedicatedMasterNodes,
			OpenSearchDashboardsNodeSize: instSpec.BundleOptions.OpenSearchDashboardsNodeSize,
			IndexManagementPlugin:        instSpec.BundleOptions.IndexManagementPlugin,
			AlertingPlugin:               instSpec.BundleOptions.AlertingPlugin,
			ICUPlugin:                    instSpec.BundleOptions.ICUPlugin,
			KNNPlugin:                    instSpec.BundleOptions.KNNPlugin,
			NotificationsPlugin:          instSpec.BundleOptions.NotificationsPlugin,
			ReportsPlugin:                instSpec.BundleOptions.ReportsPlugin,
		}

		if instSpec.BundleOptions.DedicatedMasterNodes {
			dcToAppend.MasterNodeSize = instSpec.BundleOptions.MasterNodeSize
			dcToAppend.NodeSize = instSpec.BundleOptions.DataNodeSize
		} else {
			dcToAppend.NodeSize = instSpec.BundleOptions.MasterNodeSize
		}

		dcToAppend.SetCloudProviderSettingsAPIv1(instSpec.ClusterProvider)

		k8sDataCentres = append(k8sDataCentres, dcToAppend)
	}
	oss.DataCentres = k8sDataCentres

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

func (odc *OpenSearchDataCentre) newBundleInstAPIv1(version string) []*models.OpenSearchBundleV1 {
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

func (oss *OpenSearchSpec) ToInstAPIv1Creation() *models.OpenSearchClusterCreationV1 {
	if len(oss.DataCentres) == 0 {
		return nil
	}

	rackAlloc := &models.RackAllocationV1{
		NodesPerRack:  oss.DataCentres[0].NodesNumber,
		NumberOfRacks: oss.DataCentres[0].RacksNumber,
	}

	return &models.OpenSearchClusterCreationV1{
		ClusterV1: models.ClusterV1{
			ClusterName:           oss.Name,
			NodeSize:              oss.DataCentres[0].NodeSize,
			PrivateNetworkCluster: oss.PrivateNetworkCluster,
			SLATier:               oss.SLATier,
			DataCentre:            oss.DataCentres[0].Region,
			DataCentreCustomName:  oss.DataCentres[0].Name,
			ClusterNetwork:        oss.DataCentres[0].Network,
			RackAllocation:        rackAlloc,
		},
		Bundles:         oss.DataCentres[0].newBundleInstAPIv1(oss.Version),
		TwoFactorDelete: oss.TwoFactorDeleteToInstAPIv1(),
		Provider:        oss.DataCentres[0].providerToInstAPIv1(),
	}
}

func (odc *OpenSearchDataCentre) providerToInstAPIv1() *models.ClusterProviderV1 {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(odc.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = odc.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = odc.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = odc.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProviderV1{
		Name:                   odc.CloudProvider,
		AccountName:            odc.ProviderAccountName,
		Tags:                   odc.Tags,
		CustomVirtualNetworkID: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}

func (oss *OpenSearchStatus) IsEqual(instStatus *models.ClusterV1) bool {
	if oss.Status != instStatus.ClusterStatus ||
		!oss.AreDCsEqual(instStatus.DataCentres) {
		return false
	}

	return true
}

func (oss *OpenSearchStatus) AreDCsEqual(instDCs []*models.DataCentreV1) bool {
	if len(oss.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range oss.DataCentres {
			if instDC.ID == k8sDC.ID {
				if instDC.CDCStatus != k8sDC.Status ||
					!k8sDC.AreNodesEqualAPIv1(instDC.Nodes) {
					return false
				}

				break
			}
		}
	}

	return true
}

func init() {
	SchemeBuilder.Register(&OpenSearch{}, &OpenSearchList{})
}
