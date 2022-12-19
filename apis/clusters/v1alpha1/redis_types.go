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
	MasterNodes  int32 `json:"masterNodes"`
	ReplicaNodes int32 `json:"replicaNodes"`
	PasswordAuth bool  `json:"passwordAuth,omitempty"`
}

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	Cluster `json:",inline"`

	// Enables client to node encryption
	ClientEncryption      bool               `json:"clientEncryption"`
	PasswordAndUserAuth   bool               `json:"passwordAndUserAuth"`
	DataCentres           []*RedisDataCentre `json:"dataCentres"`
	ConcurrentResizes     int                `json:"concurrentResizes"`
	NotifySupportContacts bool               `json:"notifySupportContacts"`
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
			MasterNodes:  int(redisDC.MasterNodes),
			ReplicaNodes: int(redisDC.ReplicaNodes),
		}

		redisDC.SetCloudProviderSettings(&instDC.DataCentre)

		instTags := []*modelsv2.Tag{}
		for key, value := range redisDC.Tags {
			instTags = append(instTags, &modelsv2.Tag{
				Key:   key,
				Value: value,
			})
		}
		instDC.Tags = instTags

		instDCs = append(instDCs, instDC)
	}

	instSpec := &models.RedisCluster{
		ClientToNodeEncryption: rs.ClientEncryption,
		RedisVersion:           rs.Version,
		PCIComplianceMode:      rs.PCICompliance,
		PrivateNetworkCluster:  rs.PrivateNetworkCluster,
		PasswordAndUserAuth:    rs.PasswordAndUserAuth,
		Name:                   rs.Name,
		SLATier:                rs.SLATier,
		DataCentres:            instDCs,
	}

	return instSpec
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
