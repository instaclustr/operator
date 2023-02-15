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

type RedisDataCentre struct {
	DataCentre   `json:",inline"`
	MasterNodes  int  `json:"masterNodes"`
	PasswordAuth bool `json:"passwordAuth,omitempty"`
}

type RedisRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterId"`

	// The display name of the restored cluster.
	ClusterNameOverride string `json:"clusterNameOverride,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCInfos []*RedisRestoreCDCInfo `json:"cdcInfos,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// Only data for the specified indices will be restored, for the point in time.
	IndexNames string `json:"indexNames,omitempty"`

	// The cluster network for this cluster to be restored to.
	ClusterNetwork string `json:"clusterNetwork,omitempty"`
}

type RedisRestoreCDCInfo struct {
	CDCID            string `json:"cdcId,omitempty"`
	RestoreToSameVPC bool   `json:"restoreToSameVpc,omitempty"`
	CustomVPCID      string `json:"customVpcId,omitempty"`
	CustomVPCNetwork string `json:"customVpcNetwork,omitempty"`
}

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	RestoreFrom *RedisRestoreFrom `json:"restoreFrom,omitempty"`
	Cluster     `json:",inline"`

	// Enables client to node encryption
	ClientEncryption      bool               `json:"clientEncryption,omitempty"`
	PasswordAndUserAuth   bool               `json:"passwordAndUserAuth,omitempty"`
	DataCentres           []*RedisDataCentre `json:"dataCentres,omitempty"`
	ConcurrentResizes     int                `json:"concurrentResizes,omitempty"`
	NotifySupportContacts bool               `json:"notifySupportContacts,omitempty"`
	Description           string             `json:"description,omitempty"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Redis is the Schema for the redis API
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func (r *Redis) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(r).String() + "/" + jobName
}

func (r *Redis) NewPatch() client.Patch {
	old := r.DeepCopy()
	return client.MergeFrom(old)
}

func (r *Redis) NewBackupSpec(startTimestamp int) *clusterresourcesv1alpha1.ClusterBackup {
	return &clusterresourcesv1alpha1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1alpha1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + r.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   r.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: r.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
			ClusterID:   r.Status.ID,
			ClusterKind: models.RedisClusterKind,
		},
	}
}

func (rs *RedisSpec) ToInstAPIv2() *models.RedisCluster {
	instDCs := rs.DCsToInstAPI()
	instTwoFactorDelete := rs.TwoFactorDeletesToInstAPI()

	instSpec := &models.RedisCluster{
		Name:                   rs.Name,
		RedisVersion:           rs.Version,
		ClientToNodeEncryption: rs.ClientEncryption,
		PCIComplianceMode:      rs.PCICompliance,
		PrivateNetworkCluster:  rs.PrivateNetworkCluster,
		PasswordAndUserAuth:    rs.PasswordAndUserAuth,
		SLATier:                rs.SLATier,
		DataCentres:            instDCs,
		TwoFactorDelete:        instTwoFactorDelete,
	}

	return instSpec
}

func (rs *RedisSpec) DCsToInstAPI() []*models.RedisDataCentre {
	instDCs := []*models.RedisDataCentre{}
	for _, redisDC := range rs.DataCentres {
		instDC := &models.RedisDataCentre{
			DataCentre: modelsv2.DataCentre{
				Name:                redisDC.Name,
				Network:             redisDC.Network,
				NodeSize:            redisDC.NodeSize,
				CloudProvider:       redisDC.CloudProvider,
				Region:              redisDC.Region,
				ProviderAccountName: redisDC.ProviderAccountName,
			},
			MasterNodes:  redisDC.MasterNodes,
			ReplicaNodes: int(redisDC.NodesNumber),
		}

		redisDC.CloudProviderSettingsToInstAPI(&instDC.DataCentre)

		redisDC.TagsToInstAPI(&instDC.DataCentre)

		instDCs = append(instDCs, instDC)
	}
	return instDCs
}

func (rs *RedisSpec) DCsToInstAPIUpdate() *models.RedisDataCentreUpdate {
	return &models.RedisDataCentreUpdate{
		DataCentres: rs.DCsToInstAPI(),
	}
}

func (rs *RedisSpec) HasRestore() bool {
	if rs.RestoreFrom != nil && rs.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (rs *RedisSpec) IsEqual(instSpec *models.RedisCluster) bool {
	if instSpec.ClientToNodeEncryption != rs.ClientEncryption ||
		instSpec.RedisVersion != rs.Version ||
		instSpec.PCIComplianceMode != rs.PCICompliance ||
		instSpec.PrivateNetworkCluster != rs.PrivateNetworkCluster ||
		instSpec.PasswordAndUserAuth != rs.PasswordAndUserAuth ||
		instSpec.Name != rs.Name ||
		instSpec.SLATier != rs.SLATier ||
		!rs.AreDCsEqual(instSpec.DataCentres) ||
		!rs.IsTwoFactorDeleteEqual(instSpec.TwoFactorDelete) {
		return false
	}

	return true
}

func (rs *RedisSpec) AreDCsEqual(instDCs []*models.RedisDataCentre) bool {
	if len(instDCs) != len(rs.DataCentres) {
		return false
	}

	for _, instDC := range instDCs {
		for _, dataCentre := range rs.DataCentres {
			if dataCentre.Name == instDC.Name {
				if instDC.Network != dataCentre.Network ||
					instDC.NodeSize != dataCentre.NodeSize ||
					instDC.MasterNodes != dataCentre.MasterNodes ||
					instDC.ReplicaNodes != int(dataCentre.NodesNumber) ||
					instDC.Region != dataCentre.Region ||
					instDC.ProviderAccountName != dataCentre.ProviderAccountName ||
					!dataCentre.AreCloudProviderSettingsEqual(instDC.AWSSettings, instDC.GCPSettings, instDC.AzureSettings) ||
					!dataCentre.AreTagsEqual(instDC.Tags) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (rs *RedisSpec) SetFromInstAPI(instSpec *models.RedisCluster) {
	rs.ClientEncryption = instSpec.ClientToNodeEncryption
	rs.Version = instSpec.RedisVersion
	rs.PCICompliance = instSpec.PCIComplianceMode
	rs.PrivateNetworkCluster = instSpec.PrivateNetworkCluster
	rs.PasswordAndUserAuth = instSpec.PasswordAndUserAuth
	rs.Name = instSpec.Name
	rs.SLATier = instSpec.SLATier

	rs.SetTwoFactorDeletesFromInstAPI(instSpec.TwoFactorDelete)

	rs.SetDCsFromInstAPI(instSpec.DataCentres)
}

func (rs *RedisSpec) SetDCsFromInstAPI(instDCs []*models.RedisDataCentre) {
	dataCentres := []*RedisDataCentre{}
	for _, instDC := range instDCs {
		redisDC := &RedisDataCentre{
			DataCentre: DataCentre{
				Name:                instDC.Name,
				Region:              instDC.Region,
				CloudProvider:       instDC.CloudProvider,
				ProviderAccountName: instDC.ProviderAccountName,
				Network:             instDC.Network,
				NodeSize:            instDC.NodeSize,
				NodesNumber:         int32(instDC.ReplicaNodes),
			},
			MasterNodes: instDC.MasterNodes,
		}

		redisDC.SetCloudProviderSettingsFromInstAPI(&instDC.DataCentre)

		redisDC.SetTagsFromInstAPI(instDC.Tags)
		dataCentres = append(dataCentres, redisDC)
	}
	rs.DataCentres = dataCentres
}

func (rs *RedisStatus) AreEqual(instRedis *models.RedisCluster) bool {
	if rs.Status != instRedis.Status ||
		rs.CurrentClusterOperationStatus != instRedis.CurrentClusterOperationStatus ||
		!rs.AreDCsEqual(instRedis.DataCentres) {
		return false
	}

	return true
}

func (rs *RedisStatus) AreDCsEqual(instDCs []*models.RedisDataCentre) bool {
	if len(instDCs) != len(rs.DataCentres) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range rs.DataCentres {
			if instDC.ID == k8sDC.ID {
				if instDC.Status != k8sDC.Status ||
					k8sDC.AreNodesEqual(instDC.Nodes) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (rs *RedisStatus) SetFromInstAPI(instRedis *models.RedisCluster) {
	rs.Status = instRedis.Status
	rs.CurrentClusterOperationStatus = instRedis.CurrentClusterOperationStatus
	rs.SetDCsFromInstAPI(instRedis.DataCentres)
}

func (rs *RedisStatus) SetDCsFromInstAPI(instDCs []*models.RedisDataCentre) {
	newDCs := []*DataCentreStatus{}
	for _, instDC := range instDCs {
		dc := &DataCentreStatus{
			ID:     instDC.ID,
			Status: instDC.Status,
		}
		dc.SetNodesFromInstAPI(instDC.Nodes)
		newDCs = append(newDCs, dc)
	}
	rs.DataCentres = newDCs
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
