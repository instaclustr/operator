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
	"strconv"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type RedisDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	NodeSize     string `json:"nodeSize"`
	MasterNodes  int    `json:"masterNodes"`
	ReplicaNodes int    `json:"replicaNodes,omitempty"`

	//+kubebuilder:validation:Minimum:=0
	//+kubebuilder:validation:Maximum:=5
	// ReplicationFactor defines how many replica nodes should be created for each master node
	// (e.a. if there are 3 masterNodes and replicationFactor 1 then it creates 1 replicaNode for each accordingly).
	ReplicationFactor int `json:"replicationFactor,omitempty"`

	PrivateLink PrivateLinkSpec `json:"privateLink,omitempty"`
}

type RedisRestoreFrom struct {
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

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	GenericClusterSpec `json:",inline"`

	RestoreFrom *RedisRestoreFrom `json:"restoreFrom,omitempty"`

	// Enables client to node encryption
	ClientEncryption bool `json:"clientEncryption"`
	// Enables Password Authentication and User Authorization
	PasswordAndUserAuth bool `json:"passwordAndUserAuth"`
	PCICompliance       bool `json:"pciCompliance,omitempty"`

	//+kubebuilder:validation:MaxItems:=2
	DataCentres []*RedisDataCentre `json:"dataCentres"`

	ResizeSettings GenericResizeSettings `json:"resizeSettings,omitempty"`
	UserRefs       References            `json:"userRefs,omitempty"`
}

type RedisDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	Nodes       []*Node             `json:"nodes"`
	PrivateLink PrivateLinkStatuses `json:"privateLink"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	GenericStatus `json:",inline"`

	DataCentres    []*RedisDataCentreStatus `json:"dataCentres,omitempty"`
	AvailableUsers References               `json:"availableUsers,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

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
	return r.Kind + "/" + client.ObjectKeyFromObject(r).String() + "/" + jobName
}

func (r *Redis) NewPatch() client.Patch {
	old := r.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (r *Redis) NewBackupSpec(startTimestamp int) *clusterresourcesv1beta1.ClusterBackup {
	return &clusterresourcesv1beta1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1beta1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + r.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   r.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: r.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1beta1.ClusterBackupSpec{
			ClusterRef: &clusterresourcesv1beta1.ClusterRef{
				Name:        r.Name,
				Namespace:   r.Namespace,
				ClusterKind: models.RedisClusterKind,
			},
		},
	}
}

func (c *Redis) GetSpec() RedisSpec { return c.Spec }

func (c *Redis) IsSpecEqual(spec RedisSpec) bool {
	return c.Spec.IsEqual(&spec)
}

func (rs *RedisSpec) ToInstAPI() *models.RedisCluster {
	return &models.RedisCluster{
		GenericClusterFields:   rs.GenericClusterSpec.ToInstAPI(),
		RedisVersion:           rs.Version,
		ClientToNodeEncryption: rs.ClientEncryption,
		PasswordAndUserAuth:    rs.PasswordAndUserAuth,
		PCIComplianceMode:      rs.PCICompliance,
		DataCentres:            rs.DCsToInstAPI(),
	}
}

func (r *Redis) RestoreInfoToInstAPI(restoreData *RedisRestoreFrom) any {
	iRestore := struct {
		CDCConfigs          []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`
		PointInTime         int64               `json:"pointInTime,omitempty"`
		ClusterID           string              `json:"clusterId,omitempty"`
		RestoredClusterName string              `json:"restoredClusterName,omitempty"`
	}{

		CDCConfigs:          restoreData.CDCConfigs,
		PointInTime:         restoreData.PointInTime,
		ClusterID:           restoreData.ClusterID,
		RestoredClusterName: restoreData.RestoredClusterName,
	}

	return iRestore
}

func (rs *RedisSpec) DCsToInstAPI() (iDCs []*models.RedisDataCentre) {
	for _, redisDC := range rs.DataCentres {
		iDC := &models.RedisDataCentre{
			GenericDataCentreFields: redisDC.GenericDataCentreSpec.ToInstAPI(),
			NodeSize:                redisDC.NodeSize,
			MasterNodes:             redisDC.MasterNodes,
			ReplicaNodes:            redisDC.ReplicaNodes,
			ReplicationFactor:       redisDC.ReplicationFactor,
			PrivateLink:             privateLinksToInstAPI(redisDC.PrivateLink),
		}
		iDCs = append(iDCs, iDC)
	}
	return
}

func (rs *RedisSpec) DCsUpdateToInstAPI() *models.RedisDataCentreUpdate {
	return &models.RedisDataCentreUpdate{
		DataCentres:    rs.DCsToInstAPI(),
		ResizeSettings: rs.ResizeSettings.ToInstAPI(),
	}
}

func (rs *RedisSpec) HasRestore() bool {
	if rs.RestoreFrom != nil && rs.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (rs *RedisSpec) IsEqual(o *RedisSpec) bool {
	return rs.GenericClusterSpec.Equals(&o.GenericClusterSpec) &&
		o.ClientEncryption == rs.ClientEncryption &&
		o.PasswordAndUserAuth == rs.PasswordAndUserAuth &&
		rs.DCsEqual(o.DataCentres)
}

func (rdc *RedisDataCentre) Equals(o *RedisDataCentre) bool {
	return rdc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		rdc.NodeSize == o.NodeSize &&
		rdc.ReplicaNodes == o.ReplicaNodes &&
		rdc.MasterNodes == o.MasterNodes &&
		slices.EqualsPtr(rdc.PrivateLink, o.PrivateLink)
}

func (rs *RedisSpec) DCsEqual(o []*RedisDataCentre) bool {
	if len(o) != len(rs.DataCentres) {
		return false
	}

	m := map[string]*RedisDataCentre{}
	for _, dc := range rs.DataCentres {
		m[dc.Name] = dc
	}

	for _, dc := range o {
		mDC, ok := m[dc.Name]
		if !ok {
			return false
		}

		if !mDC.Equals(dc) {
			return false
		}
	}

	return true
}

func (r *Redis) FromInstAPI(instaModel *models.RedisCluster) {
	r.Spec.FromInstAPI(instaModel)
	r.Status.FromInstAPI(instaModel)
}

func (rs *RedisSpec) FromInstAPI(instaModel *models.RedisCluster) {
	rs.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields, instaModel.RedisVersion)

	rs.ClientEncryption = instaModel.ClientToNodeEncryption
	rs.PasswordAndUserAuth = instaModel.PasswordAndUserAuth
	rs.PCICompliance = instaModel.PCIComplianceMode

	rs.DCsFromInstAPI(instaModel.DataCentres)
}

func (rs *RedisSpec) DCsFromInstAPI(instaModels []*models.RedisDataCentre) {
	rs.DataCentres = make([]*RedisDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := RedisDataCentre{}
		dc.FromInstAPI(instaModel)
		rs.DataCentres[i] = &dc
	}
}

func (rdc *RedisDataCentre) FromInstAPI(instaModel *models.RedisDataCentre) {
	rdc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)

	rdc.NodeSize = instaModel.NodeSize
	rdc.MasterNodes = instaModel.MasterNodes
	rdc.ReplicaNodes = instaModel.ReplicaNodes
	rdc.ReplicationFactor = instaModel.ReplicationFactor

	rdc.PrivateLink.FromInstAPI(instaModel.PrivateLink)
}

func (rs *RedisStatus) FromInstAPI(instaModel *models.RedisCluster) {
	rs.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	rs.DCsFromInstAPI(instaModel.DataCentres)
}

func (rs *RedisStatus) DCsFromInstAPI(instaModels []*models.RedisDataCentre) {
	rs.DataCentres = make([]*RedisDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := RedisDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		rs.DataCentres[i] = &dc
	}
}

func (s *RedisDataCentreStatus) FromInstAPI(instaModel *models.RedisDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	s.PrivateLink.FromInstAPI(instaModel.PrivateLink)
	s.Nodes = nodesFromInstAPI(instaModel.Nodes)
}

func (s *RedisDataCentreStatus) Equals(o *RedisDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		nodesEqual(s.Nodes, o.Nodes) &&
		slices.EqualsPtr(s.PrivateLink, o.PrivateLink)
}

func (rs *RedisStatus) Equals(o *RedisStatus) bool {
	return rs.GenericStatus.Equals(&o.GenericStatus) &&
		rs.DCsEqual(o.DataCentres)
}

func (rs *RedisStatus) DCsEqual(o []*RedisDataCentreStatus) bool {
	if len(rs.DataCentres) != len(o) {
		return false
	}

	m := map[string]*RedisDataCentreStatus{}
	for _, dc := range rs.DataCentres {
		m[dc.Name] = dc
	}

	for _, dc := range o {
		mDC, ok := m[dc.Name]
		if !ok {
			return false
		}

		if !mDC.Equals(dc) {
			return false
		}
	}

	return true
}

func (s *RedisStatus) ToOnPremises() ClusterStatus {
	dc := &DataCentreStatus{
		ID:    s.DataCentres[0].ID,
		Nodes: s.DataCentres[0].Nodes,
	}

	return ClusterStatus{
		ID:          s.ID,
		DataCentres: []*DataCentreStatus{dc},
	}
}

func (r *Redis) GetUserRefs() References {
	return r.Spec.UserRefs
}

func (r *Redis) SetUserRefs(refs References) {
	r.Spec.UserRefs = refs
}

func (r *Redis) GetAvailableUsers() References {
	return r.Status.AvailableUsers
}

func (r *Redis) SetAvailableUsers(users References) {
	r.Status.AvailableUsers = users
}

func (r *Redis) GetClusterID() string {
	return r.Status.ID
}

func (r *Redis) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return r.Status.DataCentres[0].ID
	}
	for _, cdc := range r.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}

func (r *Redis) SetClusterID(id string) {
	r.Status.ID = id
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}

func (r *Redis) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !r.Spec.PrivateNetwork {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.RedisDB,
				Port: models.Port6379,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port6379,
				},
			},
			{
				Name: models.RedisBus,
				Port: models.Port16379,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port16379,
				},
			},
		}
	}
	return exposePorts
}

func (r *Redis) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.RedisDB,
			Port: models.Port6379,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port6379,
			},
		},
		{
			Name: models.RedisBus,
			Port: models.Port16379,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port16379,
			},
		},
	}
	return headlessPorts
}
