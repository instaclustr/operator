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
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

type PgDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL options
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer options
	PoolMode string `json:"poolMode,omitempty"`
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
	PGBouncerVersion      string            `json:"pgBouncerVersion,omitempty"`
	DataCentres           []*PgDataCentre   `json:"dataCentres,omitempty"`
	ConcurrentResizes     int               `json:"concurrentResizes,omitempty"`
	NotifySupportContacts bool              `json:"notifySupportContacts,omitempty"`
	ClusterConfigurations map[string]string `json:"clusterConfigurations,omitempty"`
	Description           string            `json:"description,omitempty"`
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
			Annotations: map[string]string{models.StartAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: pg.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
			ClusterID:   pg.Status.ID,
			ClusterKind: models.PgClusterKind,
		},
	}
}

func (pgs *PgSpec) IsSpecEqual(instSpec *models.ClusterSpec, instClusterConfig map[string]string) bool {
	if len(pgs.DataCentres) == 0 ||
		len(instSpec.DataCentres) != len(pgs.DataCentres) ||
		pgs.Version != instSpec.BundleVersion ||
		pgs.SLATier != instSpec.SLATier ||
		!pgs.AreAddonBundlesEqual(instSpec.AddonBundles) ||
		!pgs.IsClusterConfigEqual(instClusterConfig) ||
		pgs.PrivateNetworkCluster != instSpec.DataCentres[0].PrivateIPOnly ||
		pgs.Name != instSpec.ClusterName {
		return false
	}

	for _, instDataCentre := range instSpec.DataCentres {
		for _, dataCentre := range pgs.DataCentres {
			if dataCentre.Name == instDataCentre.CDCName {
				if !dataCentre.AreOptionsEqual(instSpec.BundleOptions) ||
					!dataCentre.IsDataCentreEqual(instDataCentre) {
					return false
				}

				if !dataCentre.IsClusterProviderEqual(instSpec.ClusterProvider) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (pdc *PgDataCentre) AreOptionsEqual(instBundleOptions models.BundleOptions) bool {
	if pdc.ClientEncryption != instBundleOptions.ClientEncryption ||
		pdc.ReplicationMode != instBundleOptions.ReplicationMode ||
		pdc.SynchronousModeStrict != instBundleOptions.SynchronousModeStrict ||
		pdc.PostgresqlNodeCount != instBundleOptions.PostgresqlNodeCount {
		return false
	}

	return true
}

func (pdc *PgDataCentre) IsDataCentreEqual(instDC *models.DataCentreSpec) bool {
	if int(pdc.NodesNumber) != len(instDC.Nodes) ||
		pdc.Name != instDC.CDCName ||
		pdc.Network != instDC.CDCNetwork ||
		pdc.Region != instDC.Name ||
		pdc.CloudProvider != instDC.Provider ||
		pdc.ClientEncryption != instDC.ClientEncryption {
		return false
	}

	for _, instNode := range instDC.Nodes {
		if pdc.NodeSize != instNode.Size {
			return false
		}
		break
	}

	return true
}

func (pgs *PgSpec) AreAddonBundlesEqual(instAddons []*models.AddonBundle) bool {
	for _, instAddon := range instAddons {
		if instAddon.Bundle == modelsv1.PgBouncer {
			if instAddon.Version != pgs.PGBouncerVersion {
				return false
			}
			for _, dataCentre := range pgs.DataCentres {
				if instAddon.Options.PoolMode != dataCentre.PoolMode {
					return false
				}
			}
		}
	}

	return true
}

func (pgs *PgSpec) IsClusterConfigEqual(instClusterConfig map[string]string) bool {
	for key, value := range instClusterConfig {
		if pgs.ClusterConfigurations[key] != value {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) IsClusterProviderEqual(instProviders []*models.ClusterProvider) bool {
	for i, instProvider := range instProviders {
		if pdc.CloudProvider != instProvider.Name ||
			pdc.ProviderAccountName != instProvider.AccountName {
			return false
		}

		for key, value := range instProvider.Tags {
			if pdc.Tags[key] != value {
				return false
			}
		}

		if len(pdc.CloudProviderSettings) != len(instProviders) ||
			pdc.CloudProviderSettings[i].CustomVirtualNetworkID != instProvider.CustomVirtualNetworkId ||
			pdc.CloudProviderSettings[i].DiskEncryptionKey != instProvider.DiskEncryptionKey ||
			pdc.CloudProviderSettings[i].ResourceGroup != instProvider.ResourceGroup {
		}
	}

	return true
}

func (pgs *PgSpec) SetSpecFromInst(instSpec *models.ClusterSpec, clusterConfig map[string]string) {
	pgs.Name = instSpec.ClusterName
	pgs.Version = instSpec.BundleVersion
	pgs.SLATier = instSpec.SLATier
	pgs.PCICompliance = instSpec.PCICompliance != models.Disabled
	pgs.ClusterConfigurations = clusterConfig

	for _, instDC := range instSpec.DataCentres {
		pgs.PrivateNetworkCluster = instDC.PrivateIPOnly

		dataCentre := &PgDataCentre{
			ClientEncryption: instDC.ClientEncryption,
			DataCentre: DataCentre{
				Name:    instDC.CDCName,
				Network: instDC.CDCNetwork,
				Region:  instDC.Name,
			},
			PostgresqlNodeCount:   instSpec.BundleOptions.PostgresqlNodeCount,
			ReplicationMode:       instSpec.BundleOptions.ReplicationMode,
			SynchronousModeStrict: instSpec.BundleOptions.SynchronousModeStrict,
		}

		if len(instDC.Nodes) != 0 {
			dataCentre.NodeSize = instDC.Nodes[0].Size
		}

		for _, instProvider := range instSpec.ClusterProvider {
			dataCentre.CloudProvider = instProvider.Name
			dataCentre.ProviderAccountName = instProvider.AccountName
			dataCentre.Tags = instProvider.Tags
			dataCentre.CloudProviderSettings = append(dataCentre.CloudProviderSettings, &CloudProviderSettings{
				ResourceGroup:          instProvider.ResourceGroup,
				CustomVirtualNetworkID: instProvider.CustomVirtualNetworkId,
				DiskEncryptionKey:      instProvider.DiskEncryptionKey,
			})
		}

		pgs.DataCentres = append(pgs.DataCentres, dataCentre)
	}

	for _, addonBundle := range instSpec.AddonBundles {
		if addonBundle.Bundle == modelsv1.PgBouncer {
			pgs.PGBouncerVersion = addonBundle.Version
			for _, dataCentre := range pgs.DataCentres {
				dataCentre.PoolMode = addonBundle.Options.PoolMode
			}
		}
	}
}

func (pgs *PgSpec) Update(instSpec *models.ClusterSpec, clusterConfigurations map[string]string) {
	pgs.ClusterConfigurations = clusterConfigurations
	for _, instDC := range instSpec.DataCentres {
		if len(instDC.Nodes) == 0 {
			continue
		}

		for _, dataCentre := range pgs.DataCentres {
			if instDC.CDCName == dataCentre.Name {
				dataCentre.NodeSize = instDC.Nodes[0].Size

				providerSettings := []*CloudProviderSettings{}
				for _, instProvider := range instSpec.ClusterProvider {
					providerSettings = append(providerSettings, &CloudProviderSettings{
						CustomVirtualNetworkID: instProvider.CustomVirtualNetworkId,
						ResourceGroup:          instProvider.ResourceGroup,
						DiskEncryptionKey:      instProvider.DiskEncryptionKey,
					})
				}

				break
			}
		}
	}
}

func (pgs *PgSpec) HasRestoreFilled() bool {
	if pgs.PgRestoreFrom != nil && pgs.PgRestoreFrom.ClusterID != "" {
		return true
	}

	return false
}
