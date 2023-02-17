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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

type RedisDataCentre struct {
	DataCentre  `json:",inline"`
	MasterNodes int `json:"masterNodes"`
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
	ClientEncryption    bool               `json:"clientEncryption,omitempty"`
	PasswordAndUserAuth bool               `json:"passwordAndUserAuth,omitempty"`
	DataCentres         []*RedisDataCentre `json:"dataCentres,omitempty"`
	Description         string             `json:"description,omitempty"`
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

func (rs *RedisSpec) ToInstAPI() *models.RedisCluster {
	return &models.RedisCluster{
		Name:                   rs.Name,
		RedisVersion:           rs.Version,
		ClientToNodeEncryption: rs.ClientEncryption,
		PCIComplianceMode:      rs.PCICompliance,
		PrivateNetworkCluster:  rs.PrivateNetworkCluster,
		PasswordAndUserAuth:    rs.PasswordAndUserAuth,
		SLATier:                rs.SLATier,
		DataCentres:            rs.DCsToInstAPI(),
		TwoFactorDelete:        rs.TwoFactorDeletesToInstAPI(),
	}
}

func (rs *RedisSpec) DCsToInstAPI() (iDCs []*models.RedisDataCentre) {
	for _, redisDC := range rs.DataCentres {
		iSettings := redisDC.CloudProviderSettingsToInstAPI()
		iDC := &models.RedisDataCentre{
			DataCentre: models.DataCentre{
				Name:                redisDC.Name,
				Network:             redisDC.Network,
				NodeSize:            redisDC.NodeSize,
				AWSSettings:         iSettings.AWSSettings,
				GCPSettings:         iSettings.GCPSettings,
				AzureSettings:       iSettings.AzureSettings,
				Tags:                redisDC.TagsToInstAPI(),
				CloudProvider:       redisDC.CloudProvider,
				Region:              redisDC.Region,
				ProviderAccountName: redisDC.ProviderAccountName,
			},
			MasterNodes:  redisDC.MasterNodes,
			ReplicaNodes: redisDC.NodesNumber,
		}
		iDCs = append(iDCs, iDC)
	}
	return
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

func (rs *RedisSpec) IsEqual(iRedis RedisSpec) bool {
	return rs.Cluster.IsEqual(iRedis.Cluster) &&
		iRedis.ClientEncryption == rs.ClientEncryption &&
		iRedis.PasswordAndUserAuth == rs.PasswordAndUserAuth &&
		rs.AreDCsEqual(iRedis.DataCentres) &&
		rs.IsTwoFactorDeleteEqual(iRedis.TwoFactorDelete)
}

func (rs *RedisSpec) AreDCsEqual(iDCs []*RedisDataCentre) bool {
	if len(iDCs) != len(rs.DataCentres) {
		return false
	}

	for i, iDC := range iDCs {
		dataCentre := rs.DataCentres[i]
		if !dataCentre.IsEqual(iDC.DataCentre) ||
			iDC.MasterNodes != dataCentre.MasterNodes {
			return false
		}
	}

	return true
}

func (r *Redis) FromInstAPI(iData []byte) (*Redis, error) {
	iRedis := models.RedisCluster{}
	err := json.Unmarshal(iData, &iRedis)
	if err != nil {
		return nil, err
	}

	return &Redis{
		TypeMeta:   r.TypeMeta,
		ObjectMeta: r.ObjectMeta,
		Spec:       r.Spec.FromInstAPI(iRedis),
		Status:     r.Status.FromInstAPI(iRedis),
	}, nil
}

func (rs *RedisSpec) FromInstAPI(iRedis models.RedisCluster) RedisSpec {
	return RedisSpec{
		Cluster: Cluster{
			Name:                  iRedis.Name,
			Version:               iRedis.RedisVersion,
			PCICompliance:         iRedis.PCIComplianceMode,
			PrivateNetworkCluster: iRedis.PrivateNetworkCluster,
			SLATier:               iRedis.SLATier,
			TwoFactorDelete:       rs.Cluster.TwoFactorDeleteFromInstAPI(iRedis.TwoFactorDelete),
		},
		ClientEncryption:    iRedis.ClientToNodeEncryption,
		PasswordAndUserAuth: iRedis.PasswordAndUserAuth,
		DataCentres:         rs.DCsFromInstAPI(iRedis.DataCentres),
		Description:         rs.Description,
	}
}

func (rs *RedisSpec) DCsFromInstAPI(iDCs []*models.RedisDataCentre) (dcs []*RedisDataCentre) {
	for _, iDC := range iDCs {
		iDC.NumberOfNodes = iDC.ReplicaNodes
		dcs = append(dcs, &RedisDataCentre{
			DataCentre:  rs.Cluster.DCFromInstAPI(iDC.DataCentre),
			MasterNodes: iDC.MasterNodes,
		})
	}
	return
}

func (rs *RedisStatus) FromInstAPI(iRedis models.RedisCluster) RedisStatus {
	return RedisStatus{
		ClusterStatus{
			ID:                            iRedis.ID,
			State:                         iRedis.Status,
			DataCentres:                   rs.DCsFromInstAPI(iRedis.DataCentres),
			CurrentClusterOperationStatus: iRedis.CurrentClusterOperationStatus,
			MaintenanceEvents:             rs.MaintenanceEvents,
		},
	}
}

func (rs *RedisStatus) DCsFromInstAPI(iDCs []*models.RedisDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, rs.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
