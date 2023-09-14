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
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

type Spark struct {
	Version string `json:"version"`
}

type CassandraRestoreDC struct {
	CDCID            string `json:"cdcId,omitempty"`
	RestoreToSameVPC bool   `json:"restoreToSameVpc,omitempty"`
	CustomVPCID      string `json:"customVpcId,omitempty"`
	CustomVPCNetwork string `json:"customVpcNetwork,omitempty"`
}

type CassandraRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterID"`

	// The display name of the restored cluster.
	ClusterNameOverride string `json:"clusterNameOverride,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCInfos []CassandraRestoreDC `json:"cdcInfos,omitempty"`

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
	Cluster             `json:",inline"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres,omitempty"`
	LuceneEnabled       bool                   `json:"luceneEnabled,omitempty"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth,omitempty"`
	Spark               []*Spark               `json:"spark,omitempty"`
	BundledUseOnly      bool                   `json:"bundledUseOnly,omitempty"`
	UserRefs            []*UserReference       `json:"userRefs,omitempty"`
	//+kubebuilder:validate:MaxItems:=1
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	ClusterStatus `json:",inline"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ContinuousBackup               bool `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool `json:"clientToClusterEncryption"`
	ReplicationFactor              int  `json:"replicationFactor"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

type immutableCassandraFields struct {
	specificCassandra
	immutableCluster
}

type specificCassandra struct {
	LuceneEnabled       bool
	PasswordAndUserAuth bool
}

type immutableCassandraDCFields struct {
	immutableDC
	specificCassandraDC
}

type specificCassandraDC struct {
	replicationFactor              int
	continuousBackup               bool
	privateIpBroadcastForDiscovery bool
	clientToClusterEncryption      bool
}

func (c *Cassandra) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(c).String() + "/" + jobName
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
			ClusterID:   c.Status.ID,
			ClusterKind: models.CassandraClusterKind,
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

func (cs *CassandraSpec) newImmutableFields() *immutableCassandraFields {
	return &immutableCassandraFields{
		specificCassandra: specificCassandra{
			LuceneEnabled:       cs.LuceneEnabled,
			PasswordAndUserAuth: cs.PasswordAndUserAuth,
		},
		immutableCluster: cs.Cluster.newImmutableFields(),
	}
}

func (cs *CassandraSpec) validateUpdate(oldSpec CassandraSpec) error {
	newImmutableFields := cs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldImmutableFields, newImmutableFields)
	}

	err := cs.validateDataCentresUpdate(oldSpec)
	if err != nil {
		return err
	}
	err = validateTwoFactorDelete(cs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}
	err = validateSpark(cs.Spark, oldSpec.Spark)
	if err != nil {
		return err
	}

	for _, dc := range cs.DataCentres {
		err = cs.validateResizeSettings(dc.NodesNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *CassandraSpec) validateDataCentresUpdate(oldSpec CassandraSpec) error {
	if len(cs.DataCentres) < len(oldSpec.DataCentres) {
		return models.ErrDecreasedDataCentresNumber
	}

	for _, newDC := range cs.DataCentres {
		var exists bool
		for _, oldDC := range oldSpec.DataCentres {
			if oldDC.Name == newDC.Name {
				newDCImmutableFields := newDC.newImmutableFields()
				oldDCImmutableFields := oldDC.newImmutableFields()

				if *newDCImmutableFields != *oldDCImmutableFields {
					return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
				}

				if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
					return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
				}

				if newDC.NodesNumber < oldDC.NodesNumber {
					return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldDC.NodesNumber)
				}

				err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
				if err != nil {
					return err
				}

				err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
				if err != nil {
					return err
				}

				exists = true
				break
			}
		}

		if !exists {
			err := newDC.DataCentre.ValidateCreation()
			if err != nil {
				return err
			}

			if !cs.PrivateNetworkCluster && newDC.PrivateIPBroadcastForDiscovery {
				return fmt.Errorf("cannot use private ip broadcast for discovery on public network cluster")
			}

			err = validateReplicationFactor(models.CassandraReplicationFactors, newDC.ReplicationFactor)
			if err != nil {
				return err
			}

			if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
				return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
			}

			return nil

		}
	}

	return nil
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
		},
		DataCentres:         cs.DCsFromInstAPI(iCass.DataCentres),
		LuceneEnabled:       iCass.LuceneEnabled,
		PasswordAndUserAuth: iCass.PasswordAndUserAuth,
		Spark:               cs.SparkFromInstAPI(iCass.Spark),
		BundledUseOnly:      iCass.BundledUseOnly,
		ResizeSettings:      resizeSettingsFromInstAPI(iCass.ResizeSettings),
	}
}

func (cs *CassandraSpec) DCsFromInstAPI(iDCs []*models.CassandraDataCentre) (dcs []*CassandraDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &CassandraDataCentre{
			DataCentre:                     cs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ContinuousBackup:               iDC.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: iDC.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      iDC.ClientToClusterEncryption,
			ReplicationFactor:              iDC.ReplicationFactor,
		})
	}
	return
}

func (cs *CassandraSpec) SparkFromInstAPI(iSparks []*models.Spark) (sparks []*Spark) {
	for _, iSpark := range iSparks {
		sparks = append(sparks, &Spark{
			Version: iSpark.Version,
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
		Spark:                 cs.SparkToInstAPI(),
		DataCentres:           cs.DCsToInstAPI(),
		SLATier:               cs.SLATier,
		PrivateNetworkCluster: cs.PrivateNetworkCluster,
		PCIComplianceMode:     cs.PCICompliance,
		TwoFactorDelete:       cs.TwoFactorDeletesToInstAPI(),
		BundledUseOnly:        cs.BundledUseOnly,
		ResizeSettings:        resizeSettingsToInstAPI(cs.ResizeSettings),
	}
}

func (cs *CassandraSpec) SparkToInstAPI() (iSparks []*models.Spark) {
	for _, spark := range cs.Spark {
		iSparks = append(iSparks, &models.Spark{
			Version: spark.Version,
		})
	}
	return
}

func (cs *CassandraSpec) IsEqual(spec CassandraSpec) bool {
	return cs.Cluster.IsEqual(spec.Cluster) &&
		cs.AreDCsEqual(spec.DataCentres) &&
		cs.LuceneEnabled == spec.LuceneEnabled &&
		cs.PasswordAndUserAuth == spec.PasswordAndUserAuth &&
		cs.IsSparkEqual(spec.Spark)
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
			iDC.ReplicationFactor != dataCentre.ReplicationFactor {
			return false
		}
	}

	return true
}

func (cs *CassandraSpec) IsSparkEqual(sparks []*Spark) bool {
	if len(cs.Spark) != len(sparks) {
		return false
	}

	for i, spark := range sparks {
		if cs.Spark[i].Version != spark.Version {
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
	}
}

func (cdc *CassandraDataCentre) newImmutableFields() *immutableCassandraDCFields {
	return &immutableCassandraDCFields{
		immutableDC{
			Name:                cdc.Name,
			Region:              cdc.Region,
			CloudProvider:       cdc.CloudProvider,
			ProviderAccountName: cdc.ProviderAccountName,
			Network:             cdc.Network,
		},
		specificCassandraDC{
			replicationFactor:              cdc.ReplicationFactor,
			continuousBackup:               cdc.ContinuousBackup,
			privateIpBroadcastForDiscovery: cdc.PrivateIPBroadcastForDiscovery,
			clientToClusterEncryption:      cdc.ClientToClusterEncryption,
		},
	}
}

func (c *CassandraSpec) validateResizeSettings(nodeNumber int) error {
	for _, rs := range c.ResizeSettings {
		if rs.Concurrency > nodeNumber {
			return fmt.Errorf("resizeSettings.concurrency cannot be greater than number of nodes: %v", nodeNumber)
		}
	}

	return nil
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}
