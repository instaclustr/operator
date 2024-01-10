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
	RestoreFrom         *CassandraRestoreFrom `json:"restoreFrom,omitempty"`
	OnPremisesSpec      *OnPremisesSpec       `json:"onPremisesSpec,omitempty"`
	Cluster             `json:",inline"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres,omitempty"`
	LuceneEnabled       bool                   `json:"luceneEnabled,omitempty"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth,omitempty"`
	BundledUseOnly      bool                   `json:"bundledUseOnly,omitempty"`
	UserRefs            References             `json:"userRefs,omitempty"`
	//+kubebuilder:validate:MaxItems:=1
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	ClusterStatus  `json:",inline"`
	AvailableUsers References `json:"availableUsers,omitempty"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ContinuousBackup               bool `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool `json:"clientToClusterEncryption"`
	ReplicationFactor              int  `json:"replicationFactor"`

	// Adds the specified version of Debezium Connector Cassandra to the Cassandra cluster
	// +kubebuilder:validation:MaxItems=1
	Debezium []DebeziumCassandraSpec `json:"debezium,omitempty"`
}

type DebeziumCassandraSpec struct {
	// KafkaVPCType with only VPC_PEERED supported
	KafkaVPCType      string                              `json:"kafkaVpcType"`
	KafkaTopicPrefix  string                              `json:"kafkaTopicPrefix"`
	KafkaDataCentreID string                              `json:"kafkaCdcId,omitempty"`
	ClusterRef        *clusterresourcesv1beta1.ClusterRef `json:"clusterRef,omitempty"`
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

func (c *Cassandra) FromInstAPI(iData []byte) (*Cassandra, error) {
	iCass := &models.CassandraCluster{}
	err := json.Unmarshal(iData, iCass)
	if err != nil {
		return nil, err
	}

	return &Cassandra{
		TypeMeta:   c.TypeMeta,
		ObjectMeta: c.ObjectMeta,
		Spec:       c.Spec.FromInstAPI(iCass),
		Status:     c.Status.FromInstAPI(iCass),
	}, nil
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
		ResizeSettings: resizeSettingsToInstAPI(cs.ResizeSettings),
	}
}

func (cs *CassandraSpec) FromInstAPI(iCass *models.CassandraCluster) CassandraSpec {
	return CassandraSpec{
		Cluster: Cluster{
			Name:                  iCass.Name,
			Version:               iCass.CassandraVersion,
			PCICompliance:         iCass.PCIComplianceMode,
			PrivateNetworkCluster: iCass.PrivateNetworkCluster,
			SLATier:               iCass.SLATier,
			TwoFactorDelete:       cs.Cluster.TwoFactorDeleteFromInstAPI(iCass.TwoFactorDelete),
			Description:           iCass.Description,
		},
		DataCentres:         cs.DCsFromInstAPI(iCass.DataCentres),
		LuceneEnabled:       iCass.LuceneEnabled,
		PasswordAndUserAuth: iCass.PasswordAndUserAuth,
		BundledUseOnly:      iCass.BundledUseOnly,
		ResizeSettings:      resizeSettingsFromInstAPI(iCass.ResizeSettings),
	}
}

func (cs *CassandraSpec) DebeziumFromInstAPI(iDebeziums []*models.Debezium) (dcs []DebeziumCassandraSpec) {
	var debeziums []DebeziumCassandraSpec
	for _, iDebezium := range iDebeziums {
		debeziums = append(debeziums, DebeziumCassandraSpec{
			KafkaVPCType:      iDebezium.KafkaVPCType,
			KafkaTopicPrefix:  iDebezium.KafkaTopicPrefix,
			KafkaDataCentreID: iDebezium.KafkaDataCentreID,
			Version:           iDebezium.Version,
		})
	}
	return debeziums
}

func (cs *CassandraSpec) DCsFromInstAPI(iDCs []*models.CassandraDataCentre) (dcs []*CassandraDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &CassandraDataCentre{
			DataCentre:                     cs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ContinuousBackup:               iDC.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: iDC.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      iDC.ClientToClusterEncryption,
			ReplicationFactor:              iDC.ReplicationFactor,
			Debezium:                       cs.DebeziumFromInstAPI(iDC.Debezium),
		})
	}
	return
}

func (cs *CassandraSpec) DCsToInstAPI() (iDCs []*models.CassandraDataCentre) {
	for _, dc := range cs.DataCentres {
		iDCs = append(iDCs, dc.ToInstAPI())
	}
	return
}

func (cs *CassandraSpec) ToInstAPI() *models.CassandraCluster {
	return &models.CassandraCluster{
		Name:                  cs.Name,
		CassandraVersion:      cs.Version,
		LuceneEnabled:         cs.LuceneEnabled,
		PasswordAndUserAuth:   cs.PasswordAndUserAuth,
		DataCentres:           cs.DCsToInstAPI(),
		SLATier:               cs.SLATier,
		PrivateNetworkCluster: cs.PrivateNetworkCluster,
		PCIComplianceMode:     cs.PCICompliance,
		TwoFactorDelete:       cs.TwoFactorDeletesToInstAPI(),
		BundledUseOnly:        cs.BundledUseOnly,
		Description:           cs.Description,
		ResizeSettings:        resizeSettingsToInstAPI(cs.ResizeSettings),
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

func (cs *CassandraSpec) IsEqual(spec CassandraSpec) bool {
	return cs.Cluster.IsEqual(spec.Cluster) &&
		cs.AreDCsEqual(spec.DataCentres) &&
		cs.LuceneEnabled == spec.LuceneEnabled &&
		cs.PasswordAndUserAuth == spec.PasswordAndUserAuth &&
		cs.BundledUseOnly == spec.BundledUseOnly
}

func (cs *CassandraSpec) AreDCsEqual(dcs []*CassandraDataCentre) bool {
	if len(cs.DataCentres) != len(dcs) {
		return false
	}

	for i, iDC := range dcs {
		dataCentre := cs.DataCentres[i]

		if iDC.Name != dataCentre.Name {
			continue
		}

		if !dataCentre.IsEqual(iDC.DataCentre) ||
			iDC.ClientToClusterEncryption != dataCentre.ClientToClusterEncryption ||
			iDC.PrivateIPBroadcastForDiscovery != dataCentre.PrivateIPBroadcastForDiscovery ||
			iDC.ContinuousBackup != dataCentre.ContinuousBackup ||
			iDC.ReplicationFactor != dataCentre.ReplicationFactor ||
			!dataCentre.DebeziumEquals(iDC) {
			return false
		}
	}

	return true
}

func (cs *CassandraStatus) FromInstAPI(iCass *models.CassandraCluster) CassandraStatus {
	return CassandraStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iCass.ID,
			State:                         iCass.Status,
			DataCentres:                   cs.DCsFromInstAPI(iCass.DataCentres),
			CurrentClusterOperationStatus: iCass.CurrentClusterOperationStatus,
			MaintenanceEvents:             cs.MaintenanceEvents,
		},
	}
}

func (cs *CassandraStatus) DCsFromInstAPI(iDCs []*models.CassandraDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, cs.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func (cdc *CassandraDataCentre) ToInstAPI() *models.CassandraDataCentre {
	return &models.CassandraDataCentre{
		DataCentre:                     cdc.DataCentre.ToInstAPI(),
		ClientToClusterEncryption:      cdc.ClientToClusterEncryption,
		ContinuousBackup:               cdc.ContinuousBackup,
		PrivateIPBroadcastForDiscovery: cdc.PrivateIPBroadcastForDiscovery,
		ReplicationFactor:              cdc.ReplicationFactor,
		Debezium:                       cdc.DebeziumToInstAPI(),
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
	if !c.Spec.PrivateNetworkCluster {
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
