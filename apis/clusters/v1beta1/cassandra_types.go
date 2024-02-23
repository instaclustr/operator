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
)

type CassandraRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterID"`

	// The display name of the restored cluster.
	RestoredClusterName string `json:"restoredClusterName,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCConfigs []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// Only data for the specified tables will be restored, for the point in time.
	KeyspaceTables string `json:"keyspaceTables,omitempty"`

	// The cluster network for this cluster to be restored to.
	ClusterNetwork string `json:"clusterNetwork,omitempty"`
}

// CassandraSpec defines the desired state of Cassandra
type CassandraSpec struct {
	GenericClusterSpec `json:",inline"`

	RestoreFrom         *CassandraRestoreFrom  `json:"restoreFrom,omitempty"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres,omitempty"`
	LuceneEnabled       bool                   `json:"luceneEnabled,omitempty"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth,omitempty"`
	BundledUseOnly      bool                   `json:"bundledUseOnly,omitempty"`
	PCICompliance       bool                   `json:"pciCompliance,omitempty"`
	UserRefs            References             `json:"userRefs,omitempty" dcomparisonSkip:"true"`
	ResizeSettings      GenericResizeSettings  `json:"resizeSettings,omitempty" dcomparisonSkip:"true"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	GenericStatus `json:",inline"`
	DataCentres   []*CassandraDataCentreStatus `json:"dataCentres,omitempty"`

	AvailableUsers References `json:"availableUsers,omitempty"`
}

func (s *CassandraStatus) ToOnPremises() ClusterStatus {
	dc := &DataCentreStatus{
		ID:    s.DataCentres[0].ID,
		Nodes: s.DataCentres[0].Nodes,
	}

	return ClusterStatus{
		ID:          s.ID,
		DataCentres: []*DataCentreStatus{dc},
	}
}

func (s *CassandraStatus) Equals(o *CassandraStatus) bool {
	return s.GenericStatus.Equals(&o.GenericStatus) &&
		s.DataCentresEqual(o)
}

func (s *CassandraStatus) DataCentresEqual(o *CassandraStatus) bool {
	if len(s.DataCentres) != len(o.DataCentres) {
		return false
	}

	sMap := map[string]*CassandraDataCentreStatus{}
	for _, dc := range s.DataCentres {
		sMap[dc.Name] = dc
	}

	for _, oDC := range o.DataCentres {
		sDC, ok := sMap[oDC.Name]
		if !ok {
			return false
		}

		if !sDC.Equals(oDC) {
			return false
		}
	}

	return true
}

type CassandraDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	ContinuousBackup               bool   `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool   `json:"privateIpBroadcastForDiscovery"`
	PrivateLink                    bool   `json:"privateLink,omitempty"`
	ClientToClusterEncryption      bool   `json:"clientToClusterEncryption"`
	ReplicationFactor              int    `json:"replicationFactor"`
	NodesNumber                    int    `json:"nodesNumber"`
	NodeSize                       string `json:"nodeSize"`
	// Adds the specified version of Debezium Connector Cassandra to the Cassandra cluster
	// +kubebuilder:validation:MaxItems=1
	Debezium      []*DebeziumCassandraSpec `json:"debezium,omitempty"`
	ShotoverProxy []*ShotoverProxySpec     `json:"shotoverProxy,omitempty"`
}

func (dc *CassandraDataCentre) Equals(o *CassandraDataCentre) bool {
	return dc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		dc.ContinuousBackup == o.ContinuousBackup &&
		dc.PrivateIPBroadcastForDiscovery == o.PrivateIPBroadcastForDiscovery &&
		dc.PrivateLink == o.PrivateLink &&
		dc.ClientToClusterEncryption == o.ClientToClusterEncryption &&
		dc.ReplicationFactor == o.ReplicationFactor &&
		dc.NodesNumber == o.NodesNumber &&
		dc.NodeSize == o.NodeSize &&
		dc.DebeziumEquals(o) &&
		dc.ShotoverProxyEquals(o)
}

type CassandraDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`
	Nodes                   []*Node `json:"nodes"`
}

func (s *CassandraDataCentreStatus) Equals(o *CassandraDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		nodesEqual(s.Nodes, o.Nodes)
}

func (s *CassandraDataCentreStatus) FromInstAPI(instModel *models.CassandraDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instModel.GenericDataCentreFields)

	s.Nodes = make([]*Node, len(instModel.Nodes))
	for i, instNode := range instModel.Nodes {
		node := &Node{}
		node.FromInstAPI(instNode)
		s.Nodes[i] = node
	}
}

type ShotoverProxySpec struct {
	NodeSize string `json:"nodeSize"`
}

type DebeziumCassandraSpec struct {
	// KafkaVPCType with only VPC_PEERED supported
	KafkaVPCType      string                              `json:"kafkaVpcType"`
	KafkaTopicPrefix  string                              `json:"kafkaTopicPrefix"`
	KafkaDataCentreID string                              `json:"kafkaCdcId,omitempty"`
	ClusterRef        *clusterresourcesv1beta1.ClusterRef `json:"clusterRef,omitempty" dcomparisonSkip:"true"`
	Version           string                              `json:"version"`
}

func (d *CassandraDataCentre) DebeziumToInstAPI() []*models.Debezium {
	var instDebezium []*models.Debezium
	for _, k8sDebezium := range d.Debezium {
		instDebezium = append(instDebezium, &models.Debezium{
			KafkaVPCType:      k8sDebezium.KafkaVPCType,
			KafkaTopicPrefix:  k8sDebezium.KafkaTopicPrefix,
			KafkaDataCentreID: k8sDebezium.KafkaDataCentreID,
			Version:           k8sDebezium.Version,
		})
	}
	return instDebezium
}

func (d *CassandraDataCentre) ShotoverProxyToInstaAPI() []*models.ShotoverProxy {
	var instaShotoverProxy []*models.ShotoverProxy
	for _, k8sShotoverProxy := range d.ShotoverProxy {
		instaShotoverProxy = append(instaShotoverProxy, &models.ShotoverProxy{
			NodeSize: k8sShotoverProxy.NodeSize,
		})
	}
	return instaShotoverProxy
}

func (d *CassandraDataCentre) DebeziumEquals(new *CassandraDataCentre) bool {
	if len(d.Debezium) != len(new.Debezium) {
		return false
	}

	for _, oldDbz := range d.Debezium {
		for _, newDbz := range new.Debezium {
			if newDbz.Version != oldDbz.Version ||
				newDbz.KafkaTopicPrefix != oldDbz.KafkaTopicPrefix ||
				newDbz.KafkaVPCType != oldDbz.KafkaVPCType {
				return false
			}
		}
	}

	return true
}

func (d *CassandraDataCentre) ShotoverProxyEquals(new *CassandraDataCentre) bool {
	if len(d.ShotoverProxy) != len(new.ShotoverProxy) {
		return false
	}

	for _, oldSP := range d.ShotoverProxy {
		for _, newSP := range new.ShotoverProxy {
			if newSP.NodeSize != oldSP.NodeSize {
				return false
			}
		}
	}
	return true
}

func (c *CassandraSpec) IsEqual(o *CassandraSpec) bool {
	return c.GenericClusterSpec.Equals(&o.GenericClusterSpec) &&
		c.AreDCsEqual(o.DataCentres) &&
		c.LuceneEnabled == o.LuceneEnabled &&
		c.PasswordAndUserAuth == o.PasswordAndUserAuth &&
		c.BundledUseOnly == o.BundledUseOnly
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// Cassandra is the Schema for the cassandras API
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CassandraList contains a list of Cassandra
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

func (c *Cassandra) GetJobID(jobName string) string {
	return c.Kind + "/" + client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cassandra) NewPatch() client.Patch {
	old := c.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (c *Cassandra) NewBackupSpec(startTimestamp int) *clusterresourcesv1beta1.ClusterBackup {
	return &clusterresourcesv1beta1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1beta1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + c.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   c.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: c.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1beta1.ClusterBackupSpec{
			ClusterRef: &clusterresourcesv1beta1.ClusterRef{
				Name:        c.Name,
				Namespace:   c.Namespace,
				ClusterKind: models.CassandraClusterKind,
			},
		},
	}
}

func (c *Cassandra) FromInstAPI(instModel *models.CassandraCluster) {
	c.Spec.FromInstAPI(instModel)
	c.Status.FromInstAPI(instModel)
}

func (cs *CassandraSpec) HasRestore() bool {
	if cs.RestoreFrom != nil && cs.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (cs *CassandraSpec) DCsUpdateToInstAPI() models.CassandraClusterAPIUpdate {
	return models.CassandraClusterAPIUpdate{
		DataCentres:    cs.DCsToInstAPI(),
		ResizeSettings: cs.ResizeSettings.ToInstAPI(),
	}
}

func (cs *CassandraSpec) FromInstAPI(instModel *models.CassandraCluster) {
	cs.GenericClusterSpec.FromInstAPI(&instModel.GenericClusterFields, instModel.CassandraVersion)

	cs.LuceneEnabled = instModel.LuceneEnabled
	cs.PasswordAndUserAuth = instModel.PasswordAndUserAuth
	cs.BundledUseOnly = instModel.BundledUseOnly
	cs.PCICompliance = instModel.PCIComplianceMode
	cs.ResizeSettings.FromInstAPI(instModel.ResizeSettings)

	cs.dcsFromInstAPI(instModel.DataCentres)
}

func (cs *CassandraSpec) dcsFromInstAPI(instModels []*models.CassandraDataCentre) {
	dcs := make([]*CassandraDataCentre, len(instModels))
	for i, instModel := range instModels {
		dc := &CassandraDataCentre{}
		dcs[i] = dc

		if index := cs.getDCIndexByName(instModel.Name); index > -1 {
			dc.Debezium = cs.DataCentres[index].Debezium
		}

		dc.FromInstAPI(instModel)
	}

	cs.DataCentres = dcs
}

func (c *CassandraSpec) getDCIndexByName(name string) int {
	for i, dc := range c.DataCentres {
		if dc.Name == name {
			return i
		}
	}

	return -1
}

func (d *CassandraDataCentre) FromInstAPI(instModel *models.CassandraDataCentre) {
	d.GenericDataCentreSpec.FromInstAPI(&instModel.GenericDataCentreFields)

	d.ContinuousBackup = instModel.ContinuousBackup
	d.PrivateIPBroadcastForDiscovery = instModel.PrivateIPBroadcastForDiscovery
	d.PrivateLink = instModel.PrivateLink
	d.ClientToClusterEncryption = instModel.ClientToClusterEncryption
	d.ReplicationFactor = instModel.ReplicationFactor
	d.NodesNumber = instModel.NumberOfNodes
	d.NodeSize = instModel.NodeSize

	d.debeziumFromInstAPI(instModel.Debezium)
	d.shotoverProxyFromInstAPI(instModel.ShotoverProxy)
}

func (cs *CassandraDataCentre) debeziumFromInstAPI(instModels []*models.Debezium) {
	debezium := make([]*DebeziumCassandraSpec, len(instModels))
	for i, instModel := range instModels {
		debezium[i] = &DebeziumCassandraSpec{
			KafkaVPCType:      instModel.KafkaVPCType,
			KafkaTopicPrefix:  instModel.KafkaTopicPrefix,
			KafkaDataCentreID: instModel.KafkaDataCentreID,
			Version:           instModel.Version,
		}

		if len(cs.Debezium)-1 >= i {
			debezium[i].ClusterRef = cs.Debezium[i].ClusterRef
		}
	}
	cs.Debezium = debezium
}

func (cs *CassandraDataCentre) shotoverProxyFromInstAPI(instModels []*models.ShotoverProxy) {
	cs.ShotoverProxy = make([]*ShotoverProxySpec, len(instModels))
	for i, instModel := range instModels {
		cs.ShotoverProxy[i] = &ShotoverProxySpec{
			NodeSize: instModel.NodeSize,
		}
	}
}

func (cs *CassandraSpec) DCsToInstAPI() (instaModels []*models.CassandraDataCentre) {
	for _, dc := range cs.DataCentres {
		instaModels = append(instaModels, dc.ToInstAPI())
	}
	return
}

func (cs *CassandraSpec) ToInstAPI() *models.CassandraCluster {
	return &models.CassandraCluster{
		GenericClusterFields: cs.GenericClusterSpec.ToInstAPI(),
		CassandraVersion:     cs.Version,
		LuceneEnabled:        cs.LuceneEnabled,
		PasswordAndUserAuth:  cs.PasswordAndUserAuth,
		BundledUseOnly:       cs.BundledUseOnly,
		PCIComplianceMode:    cs.PCICompliance,
		DataCentres:          cs.DCsToInstAPI(),
		ResizeSettings:       cs.ResizeSettings.ToInstAPI(),
	}
}

func (c *Cassandra) RestoreInfoToInstAPI(restoreData *CassandraRestoreFrom) any {
	iRestore := struct {
		RestoredClusterName string              `json:"restoredClusterName,omitempty"`
		CDCConfigs          []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`
		PointInTime         int64               `json:"pointInTime,omitempty"`
		KeyspaceTables      string              `json:"keyspaceTables,omitempty"`
		ClusterNetwork      string              `json:"clusterNetwork,omitempty"`
		ClusterID           string              `json:"clusterId,omitempty"`
	}{
		CDCConfigs:          restoreData.CDCConfigs,
		RestoredClusterName: restoreData.RestoredClusterName,
		PointInTime:         restoreData.PointInTime,
		KeyspaceTables:      restoreData.KeyspaceTables,
		ClusterNetwork:      restoreData.ClusterNetwork,
		ClusterID:           restoreData.ClusterID,
	}

	return iRestore
}

func (c *Cassandra) GetSpec() CassandraSpec { return c.Spec }

func (c *Cassandra) IsSpecEqual(spec CassandraSpec) bool {
	return c.Spec.IsEqual(&spec)
}

func (cs *CassandraSpec) AreDCsEqual(dcs []*CassandraDataCentre) bool {
	if len(cs.DataCentres) != len(dcs) {
		return false
	}

	k8sDCs := map[string]*CassandraDataCentre{}
	for _, dc := range cs.DataCentres {
		k8sDCs[dc.Name] = dc
	}

	for _, instDC := range dcs {
		k8sDC, ok := k8sDCs[instDC.Name]
		if !ok {
			return false
		}

		if !k8sDC.Equals(instDC) {
			return false
		}
	}

	return true
}

func (cs *CassandraStatus) FromInstAPI(instModel *models.CassandraCluster) {
	cs.GenericStatus.FromInstAPI(&instModel.GenericClusterFields)
	cs.dcsFromInstAPI(instModel.DataCentres)
}

func (cs *CassandraStatus) dcsFromInstAPI(instModels []*models.CassandraDataCentre) {
	cs.DataCentres = make([]*CassandraDataCentreStatus, len(instModels))
	for i, instModel := range instModels {
		dc := &CassandraDataCentreStatus{}
		dc.FromInstAPI(instModel)
		cs.DataCentres[i] = dc
	}
}

func (cdc *CassandraDataCentre) ToInstAPI() *models.CassandraDataCentre {
	return &models.CassandraDataCentre{
		GenericDataCentreFields:        cdc.GenericDataCentreSpec.ToInstAPI(),
		ClientToClusterEncryption:      cdc.ClientToClusterEncryption,
		PrivateLink:                    cdc.PrivateLink,
		ContinuousBackup:               cdc.ContinuousBackup,
		PrivateIPBroadcastForDiscovery: cdc.PrivateIPBroadcastForDiscovery,
		ReplicationFactor:              cdc.ReplicationFactor,
		NodeSize:                       cdc.NodeSize,
		NumberOfNodes:                  cdc.NodesNumber,
		Debezium:                       cdc.DebeziumToInstAPI(),
		ShotoverProxy:                  cdc.ShotoverProxyToInstaAPI(),
	}
}

func (c *Cassandra) GetAvailableUsers() References {
	return c.Status.AvailableUsers
}

func (c *Cassandra) SetAvailableUsers(users References) {
	c.Status.AvailableUsers = users
}

func (c *Cassandra) GetUserRefs() References {
	return c.Spec.UserRefs
}

func (c *Cassandra) SetUserRefs(refs References) {
	c.Spec.UserRefs = refs
}

func (c *Cassandra) GetClusterID() string {
	return c.Status.ID
}

func (c *Cassandra) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return c.Status.DataCentres[0].ID
	}
	for _, cdc := range c.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}

func (c *Cassandra) SetClusterID(id string) {
	c.Status.ID = id
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}

func (c *Cassandra) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !c.Spec.PrivateNetwork {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.CassandraInterNode,
				Port: models.Port7000,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7000,
				},
			},
			{
				Name: models.CassandraCQL,
				Port: models.Port9042,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port9042,
				},
			},
			{
				Name: models.CassandraJMX,
				Port: models.Port7199,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7199,
				},
			},
		}
		if c.Spec.DataCentres[0].ClientToClusterEncryption {
			sslPort := k8scorev1.ServicePort{
				Name: models.CassandraSSL,
				Port: models.Port7001,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7001,
				},
			}
			exposePorts = append(exposePorts, sslPort)
		}
	}
	return exposePorts
}

func (c *Cassandra) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.CassandraInterNode,
			Port: models.Port7000,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port7000,
			},
		},
		{
			Name: models.CassandraCQL,
			Port: models.Port9042,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port9042,
			},
		},
	}
	if c.Spec.DataCentres[0].ClientToClusterEncryption {
		sslPort := k8scorev1.ServicePort{
			Name: models.CassandraSSL,
			Port: models.Port7001,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port7001,
			},
		}
		headlessPorts = append(headlessPorts, sslPort)
	}
	return headlessPorts
}
