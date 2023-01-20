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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

type PgDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL options
	ClientEncryption           bool                          `json:"clientEncryption"`
	PostgresqlNodeCount        int32                         `json:"postgresqlNodeCount"`
	InterDataCentreReplication []*InterDataCentreReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*IntraDataCentreReplication `json:"intraDataCentreReplication"`
	// PGBouncer options
	PGBouncerVersion string `json:"pgBouncerVersion,omitempty"`
	PoolMode         string `json:"poolMode,omitempty"`
}

type InterDataCentreReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type IntraDataCentreReplication struct {
	ReplicationMode string `json:"replicationMode"`
}

type PgRestoreFrom struct {
	ClusterID           string              `json:"clusterId"`
	ClusterNameOverride string              `json:"clusterNameOverride,omitempty"`
	CDCInfos            []*PgRestoreCDCInfo `json:"cdcInfos,omitempty"`
	PointInTime         int64               `json:"pointInTime,omitempty"`
}

type PgRestoreCDCInfo struct {
	CDCID            string `json:"cdcId,omitempty"`
	RestoreToSameVPC bool   `json:"restoreToSameVpc,omitempty"`
	CustomVPCID      string `json:"customVpcId,omitempty"`
	CustomVPCNetwork string `json:"customVpcNetwork,omitempty"`
}

// PgSpec defines the desired state of PostgreSQL
type PgSpec struct {
	PgRestoreFrom         *PgRestoreFrom `json:"pgRestoreFrom,omitempty"`
	Cluster               `json:",inline"`
	DataCentres           []*PgDataCentre   `json:"dataCentres,omitempty"`
	ClusterConfigurations map[string]string `json:"clusterConfigurations,omitempty"`
	Description           string            `json:"description,omitempty"`
	SynchronousModeStrict bool              `json:"synchronousModeStrict,omitempty"`
}

// PgStatus defines the observed state of PostgreSQL
type PgStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgSpec   `json:"spec,omitempty"`
	Status PgStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}

func (pg *PostgreSQL) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(pg).String() + "/" + jobName
}

func (pg *PostgreSQL) NewPatch() client.Patch {
	old := pg.DeepCopy()
	return client.MergeFrom(old)
}

func (pg *PostgreSQL) NewBackupSpec(startTimestamp int) *clusterresourcesv1alpha1.ClusterBackup {
	return &clusterresourcesv1alpha1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1alpha1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.PgBackupPrefix + pg.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   pg.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: pg.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
			ClusterID:   pg.Status.ID,
			ClusterKind: models.PgClusterKind,
		},
	}
}

func (pgs *PgSpec) HasRestore() bool {
	if pgs.PgRestoreFrom != nil && pgs.PgRestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (pgs *PgSpec) ToInstAPI() *models.PGCluster {
	instTFD := pgs.TwoFactorDeletesToInstAPI()

	instDCs := pgs.DataCentresToInstAPI()
	return &models.PGCluster{
		Name:                  pgs.Name,
		PostgreSQLVersion:     pgs.Version,
		DataCentres:           instDCs,
		SynchronousModeStrict: pgs.SynchronousModeStrict,
		PrivateNetworkCluster: pgs.PrivateNetworkCluster,
		SLATier:               pgs.SLATier,
		TwoFactorDelete:       instTFD,
	}
}

func (pdc *PgDataCentre) ToInstAPI() *models.PGDataCentre {
	instDC := &models.PGDataCentre{
		DataCentre: modelsv2.DataCentre{
			Name:                pdc.Name,
			Network:             pdc.Network,
			NodeSize:            pdc.NodeSize,
			NumberOfNodes:       pdc.PostgresqlNodeCount,
			CloudProvider:       pdc.CloudProvider,
			Region:              pdc.Region,
			ProviderAccountName: pdc.ProviderAccountName,
		},
		PGBouncer:                 []*models.PGBouncer{},
		ClientToClusterEncryption: pdc.ClientEncryption,
	}

	pdc.CloudProviderSettingsToInstAPI(&instDC.DataCentre)

	pdc.TagsToInstAPI(&instDC.DataCentre)

	pdc.InterDCReplicationToInstAPI(instDC)

	pdc.IntraDCReplicationToInstAPI(instDC)

	if pdc.PGBouncerVersion != "" {
		pdc.PGBouncerToInstAPI(instDC)
	}

	return instDC
}

func (pdc *PgDataCentre) InterDCReplicationToInstAPI(instDC *models.PGDataCentre) {
	var instReplication []*models.PGInterDCReplication
	for _, interDCReplication := range pdc.InterDataCentreReplication {
		instReplication = append(instReplication, &models.PGInterDCReplication{
			IsPrimaryDataCentre: interDCReplication.IsPrimaryDataCentre,
		})
	}

	instDC.InterDataCentreReplication = instReplication
}

func (pdc *PgDataCentre) IntraDCReplicationToInstAPI(instDC *models.PGDataCentre) {
	var instReplication []*models.PGIntraDCReplication
	for _, interDCReplication := range pdc.IntraDataCentreReplication {
		instReplication = append(instReplication, &models.PGIntraDCReplication{
			ReplicationMode: interDCReplication.ReplicationMode,
		})
	}

	instDC.IntraDataCentreReplication = instReplication
}

func (pdc *PgDataCentre) PGBouncerToInstAPI(instDC *models.PGDataCentre) {
	instDC.PGBouncer = []*models.PGBouncer{
		{
			PGBouncerVersion: pdc.PGBouncerVersion,
			PoolMode:         pdc.PoolMode,
		},
	}
}

func (pgs *PgSpec) DataCentresToInstAPI() []*models.PGDataCentre {
	var instDCs []*models.PGDataCentre
	for _, k8sDC := range pgs.DataCentres {
		instDCs = append(instDCs, k8sDC.ToInstAPI())
	}
	return instDCs
}

func (pgs *PgSpec) SetSpecFromInstAPI(instCluster *models.PGStatus) {
	pgs.Name = instCluster.Name
	pgs.Version = instCluster.PostgreSQLVersion
	pgs.PCICompliance = instCluster.PCIComplianceMode
	pgs.PrivateNetworkCluster = instCluster.PrivateNetworkCluster
	pgs.SLATier = instCluster.SLATier
	pgs.SetTwoFactorDeletesFromInstAPI(instCluster.TwoFactorDelete)
	pgs.SynchronousModeStrict = instCluster.SynchronousModeStrict
	pgs.SetDCsFromInstAPI(instCluster.DataCentres)
}

func (pgs *PgSpec) SetDCsFromInstAPI(instDCs []*models.PGDataCentre) {
	var k8sDCs []*PgDataCentre
	for _, instDC := range instDCs {
		dc := &PgDataCentre{}
		dc.SetDCFromInstAPI(instDC)

		k8sDCs = append(k8sDCs, dc)
	}

	pgs.DataCentres = k8sDCs
}

func (pdc *PgDataCentre) SetDCFromInstAPI(instDC *models.PGDataCentre) {
	pdc.Name = instDC.Name
	pdc.Region = instDC.Region
	pdc.CloudProvider = instDC.CloudProvider
	pdc.ProviderAccountName = instDC.ProviderAccountName
	pdc.Network = instDC.Network
	pdc.NodeSize = instDC.NodeSize
	pdc.SetTagsFromInstAPI(instDC.Tags)
	pdc.SetCloudProviderSettingsFromInstAPI(&instDC.DataCentre)
	pdc.ClientEncryption = instDC.ClientToClusterEncryption
	pdc.PostgresqlNodeCount = instDC.NumberOfNodes
	pdc.SetInterDCReplicationsFromInstAPI(instDC.InterDataCentreReplication)
	pdc.SetIntraDCReplicationsFromInstAPI(instDC.IntraDataCentreReplication)
	pdc.SetPGBouncerFromInstAPI(instDC.PGBouncer)
}

func (pdc *PgDataCentre) SetInterDCReplicationsFromInstAPI(instIRs []*models.PGInterDCReplication) {
	k8sIRs := []*InterDataCentreReplication{}
	for _, instIR := range instIRs {
		k8sIR := &InterDataCentreReplication{}
		k8sIR.SetInterDCReplicationFromInstAPI(instIR)
		k8sIRs = append(k8sIRs, k8sIR)
	}
	pdc.InterDataCentreReplication = k8sIRs
}

func (ir *InterDataCentreReplication) SetInterDCReplicationFromInstAPI(instIR *models.PGInterDCReplication) {
	ir.IsPrimaryDataCentre = instIR.IsPrimaryDataCentre
}

func (pdc *PgDataCentre) SetIntraDCReplicationsFromInstAPI(instIRs []*models.PGIntraDCReplication) {
	k8sIRs := []*IntraDataCentreReplication{}
	for _, instIR := range instIRs {
		k8sIR := &IntraDataCentreReplication{}
		k8sIR.SetIntraDCReplicationFromInstAPI(instIR)
		k8sIRs = append(k8sIRs, k8sIR)
	}
	pdc.IntraDataCentreReplication = k8sIRs
}

func (ir *IntraDataCentreReplication) SetIntraDCReplicationFromInstAPI(instIR *models.PGIntraDCReplication) {
	ir.ReplicationMode = instIR.ReplicationMode
}

func (pdc *PgDataCentre) SetPGBouncerFromInstAPI(instPGBs []*models.PGBouncer) {
	for _, instPGB := range instPGBs {
		pdc.PGBouncerVersion = instPGB.PGBouncerVersion
		pdc.PoolMode = instPGB.PoolMode
	}
}

func (ps *PgStatus) SetStatusFromInstAPI(instCluster *models.PGStatus) {
	ps.ID = instCluster.ID
	ps.Status = instCluster.Status
	ps.CurrentClusterOperationStatus = instCluster.CurrentClusterOperationStatus
	ps.TwoFactorDeleteEnabled = len(instCluster.TwoFactorDelete) > 0
	ps.SetDCsFromInstAPI(instCluster.DataCentres)
}

func (ps *PgStatus) SetDCsFromInstAPI(instDCs []*models.PGDataCentre) {
	k8sDCs := []*DataCentreStatus{}
	for _, instDC := range instDCs {
		k8sDC := &DataCentreStatus{
			ID:         instDC.ID,
			Status:     instDC.Status,
			NodeNumber: instDC.NumberOfNodes,
		}

		k8sDC.SetNodesStatusFromInstAPI(instDC.Nodes)
		k8sDCs = append(k8sDCs, k8sDC)
	}
	ps.DataCentres = k8sDCs
}

func (ps *PgStatus) AreStatusesEqual(instStatus *models.PGStatus) bool {
	if ps.ID != instStatus.ID ||
		ps.Status != instStatus.Status ||
		ps.CurrentClusterOperationStatus != instStatus.CurrentClusterOperationStatus ||
		!ps.AreDCsEqual(instStatus.DataCentres) {
		return false
	}

	return true
}

func (ps *PgStatus) AreDCsEqual(instDCs []*models.PGDataCentre) bool {
	if len(ps.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range ps.DataCentres {
			if instDC.ID == k8sDC.ID {
				if instDC.Status != k8sDC.Status ||
					instDC.NumberOfNodes != k8sDC.NodeNumber ||
					!k8sDC.AreNodesEqual(instDC.Nodes) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (pgs *PgSpec) AreSpecsEqual(instCluster *models.PGStatus) bool {
	if pgs.Name != instCluster.Name ||
		pgs.Version != instCluster.PostgreSQLVersion ||
		pgs.PCICompliance != instCluster.PCIComplianceMode ||
		pgs.PrivateNetworkCluster != instCluster.PrivateNetworkCluster ||
		pgs.SLATier != instCluster.SLATier ||
		!pgs.IsTwoFactorDeleteEqual(instCluster.TwoFactorDelete) ||
		pgs.SynchronousModeStrict != instCluster.SynchronousModeStrict ||
		!pgs.AreDCsEqual(instCluster.DataCentres) {
		return false
	}

	return true
}

func (pgs *PgSpec) AreDCsEqual(instDCs []*models.PGDataCentre) bool {
	if len(pgs.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range pgs.DataCentres {
			if instDC.Name == k8sDC.Name {
				if k8sDC.PostgresqlNodeCount != instDC.NumberOfNodes ||
					k8sDC.Region != instDC.Region ||
					k8sDC.CloudProvider != instDC.CloudProvider ||
					k8sDC.ProviderAccountName != instDC.ProviderAccountName ||
					!k8sDC.AreCloudProviderSettingsEqual(instDC.AWSSettings, instDC.GCPSettings, instDC.AzureSettings) ||
					k8sDC.Network != instDC.Network ||
					k8sDC.NodeSize != instDC.NodeSize ||
					!k8sDC.AreTagsEqual(instDC.Tags) ||
					k8sDC.ClientEncryption != instDC.ClientToClusterEncryption ||
					!k8sDC.AreInterDCReplicationsEqual(instDC.InterDataCentreReplication) ||
					!k8sDC.AreIntraDCReplicationsEqual(instDC.IntraDataCentreReplication) ||
					!k8sDC.ArePGBouncersEqual(instDC.PGBouncer) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (pdc *PgDataCentre) AreInterDCReplicationsEqual(instIRs []*models.PGInterDCReplication) bool {
	if len(pdc.InterDataCentreReplication) != len(instIRs) {
		return false
	}

	for i, instIR := range instIRs {
		if instIR.IsPrimaryDataCentre != pdc.InterDataCentreReplication[i].IsPrimaryDataCentre {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) AreIntraDCReplicationsEqual(instIRs []*models.PGIntraDCReplication) bool {
	if len(pdc.IntraDataCentreReplication) != len(instIRs) {
		return false
	}

	for i, instIR := range instIRs {
		if instIR.ReplicationMode != pdc.IntraDataCentreReplication[i].ReplicationMode {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) ArePGBouncersEqual(instPGBs []*models.PGBouncer) bool {
	if pdc.PoolMode != "" &&
		pdc.PGBouncerVersion != "" &&
		len(instPGBs) == 0 {
		return false
	}

	for _, instPGB := range instPGBs {
		if instPGB.PoolMode != pdc.PoolMode ||
			instPGB.PGBouncerVersion != pdc.PGBouncerVersion {
			return false
		}
	}

	return true
}

// SetDefaultValues should be implemented using validation webhook
// TODO https://github.com/instaclustr/operator/issues/219
func (pgs *PgSpec) SetDefaultValues() {
	for _, dataCentre := range pgs.DataCentres {
		if dataCentre.ProviderAccountName == "" {
			dataCentre.ProviderAccountName = models.DefaultAccountName
		}

		if len(dataCentre.CloudProviderSettings) == 0 {
			dataCentre.CloudProviderSettings = append(dataCentre.CloudProviderSettings, &CloudProviderSettings{
				DiskEncryptionKey:      "",
				ResourceGroup:          "",
				CustomVirtualNetworkID: "",
			})
		}
	}
}
