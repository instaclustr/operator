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
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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
	ClusterID           string               `json:"clusterID"`
	ClusterNameOverride string               `json:"clusterNameOverride,omitempty"`
	CDCInfos            []CassandraRestoreDC `json:"cdcInfos,omitempty"`
	PointInTime         int64                `json:"pointInTime,omitempty"`
	KeyspaceTables      string               `json:"keyspaceTables,omitempty"`
}

// CassandraSpec defines the desired state of Cassandra
type CassandraSpec struct {
	RestoreFrom         *CassandraRestoreFrom `json:"restoreFrom,omitempty"`
	Cluster             `json:",inline"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres,omitempty"`
	LuceneEnabled       bool                   `json:"luceneEnabled,omitempty"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth,omitempty"`
	Spark               []*Spark               `json:"spark,omitempty"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClusterStatus `json:",inline"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ContinuousBackup               bool `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool `json:"clientToClusterEncryption"`
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
	SpecificFields specificCassandraFields
	Cluster        immutableClusterFields
}

type specificCassandraFields struct {
	LuceneEnabled       bool
	PasswordAndUserAuth bool
}

func (c *Cassandra) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cassandra) NewPatch() client.Patch {
	old := c.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (c *Cassandra) NewBackupSpec(startTimestamp int) *clusterresourcesv1alpha1.ClusterBackup {
	return &clusterresourcesv1alpha1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1alpha1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.SnapshotUploadPrefix + c.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   c.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: c.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
			ClusterID:   c.Status.ID,
			ClusterKind: models.CassandraClusterKind,
		},
	}
}

func (cs *CassandraSpec) IsSpecEqual(instSpec *modelsv2.CassandraCluster) bool {
	if cs.Name != instSpec.Name ||
		cs.Version != instSpec.CassandraVersion ||
		cs.PCICompliance != instSpec.PCIComplianceMode ||
		cs.PrivateNetworkCluster != instSpec.PrivateNetworkCluster ||
		cs.SLATier != instSpec.SLATier ||
		!cs.IsTwoFactorDeleteEqual(instSpec.TwoFactorDeletes) ||
		!cs.AreDataCentresEqual(instSpec.DataCentres) ||
		cs.LuceneEnabled != instSpec.LuceneEnabled ||
		cs.PasswordAndUserAuth != instSpec.PasswordAndUserAuth ||
		!cs.IsSparkEqual(instSpec.Spark) {
		return false
	}

	return true
}

func (cs *CassandraSpec) AreDataCentresEqual(instDataCentres []*modelsv2.CassandraDataCentre) bool {
	if len(cs.DataCentres) != len(instDataCentres) {
		return false
	}

	for _, instDC := range instDataCentres {
		for _, dataCentre := range cs.DataCentres {
			if instDC.Name == dataCentre.Name {
				if instDC.ClientToClusterEncryption != dataCentre.ClientToClusterEncryption ||
					instDC.PrivateIPBroadcastForDiscovery != dataCentre.PrivateIPBroadcastForDiscovery ||
					instDC.ContinuousBackup != dataCentre.ContinuousBackup ||
					instDC.CloudProvider != dataCentre.CloudProvider ||
					instDC.NodeSize != dataCentre.NodeSize ||
					instDC.ProviderAccountName != dataCentre.ProviderAccountName ||
					instDC.Region != dataCentre.Region ||
					instDC.Network != dataCentre.Network ||
					instDC.NumberOfNodes != dataCentre.NodesNumber ||
					instDC.ReplicationFactor != dataCentre.RacksNumber ||
					!dataCentre.AreTagsEqual(instDC.Tags) ||
					!dataCentre.AreCloudProviderSettingsEqual(instDC.AWSSettings, instDC.GCPSettings, instDC.AzureSettings) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (cs *CassandraSpec) IsSparkEqual(instSparks []*modelsv2.Spark) bool {
	if len(cs.Spark) != len(instSparks) {
		return false
	}

	for i, instSpark := range instSparks {
		if cs.Spark[i].Version != instSpark.Version {
			return false
		}
	}

	return true
}

func (cs *CassandraSpec) SetSpecFromInst(instSpec *modelsv2.CassandraCluster) {
	cs.Name = instSpec.Name
	cs.Version = instSpec.CassandraVersion
	cs.PCICompliance = instSpec.PCIComplianceMode
	cs.PrivateNetworkCluster = instSpec.PrivateNetworkCluster
	cs.SLATier = instSpec.SLATier
	cs.LuceneEnabled = instSpec.LuceneEnabled
	cs.PasswordAndUserAuth = instSpec.PasswordAndUserAuth

	twoFactorDeletes := []*TwoFactorDelete{}
	for _, instTFD := range instSpec.TwoFactorDeletes {
		twoFactorDeletes = append(twoFactorDeletes, &TwoFactorDelete{
			Phone: instTFD.ConfirmationPhoneNumber,
			Email: instTFD.ConfirmationEmail,
		})
	}
	cs.TwoFactorDelete = twoFactorDeletes

	dataCentres := []*CassandraDataCentre{}
	for _, instDC := range instSpec.DataCentres {
		cassDC := &CassandraDataCentre{
			DataCentre: DataCentre{
				Name:                instDC.Name,
				Region:              instDC.Region,
				CloudProvider:       instDC.CloudProvider,
				ProviderAccountName: instDC.ProviderAccountName,
				Network:             instDC.Network,
				NodeSize:            instDC.NodeSize,
				RacksNumber:         instDC.ReplicationFactor,
				NodesNumber:         instDC.NumberOfNodes,
			},
			ContinuousBackup:               instDC.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: instDC.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      instDC.ClientToClusterEncryption,
		}

		cassDC.SetCloudProviderSettingsFromInstAPI(&instDC.DataCentre)

		tags := map[string]string{}
		for _, tag := range instDC.Tags {
			tags[tag.Key] = tag.Value
		}
		cassDC.Tags = tags
		dataCentres = append(dataCentres, cassDC)
	}
	cs.DataCentres = dataCentres

	sparks := []*Spark{}
	for _, instSpark := range instSpec.Spark {
		sparks = append(sparks, &Spark{
			Version: instSpark.Version,
		})
	}
	cs.Spark = sparks
}

func (cs *CassandraSpec) HasRestore() bool {
	if cs.RestoreFrom != nil && cs.RestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (cs *CassandraSpec) NewImmutableFieldsToValidate() *immutableCassandraFields {
	return &immutableCassandraFields{
		specificCassandraFields{
			LuceneEnabled:       cs.LuceneEnabled,
			PasswordAndUserAuth: cs.PasswordAndUserAuth,
		},
		immutableClusterFields{
			Name:                  cs.Name,
			Version:               cs.Version,
			PCICompliance:         cs.PCICompliance,
			PrivateNetworkCluster: cs.PrivateNetworkCluster,
			SLATier:               cs.SLATier,
		},
	}
}

func (cs *CassandraSpec) ValidateUpdate(oldCassandraSpec CassandraSpec) error {
	newImmutableFields := cs.NewImmutableFieldsToValidate()
	oldImmutableFields := oldCassandraSpec.NewImmutableFieldsToValidate()

	if newImmutableFields.Cluster != oldImmutableFields.Cluster ||
		newImmutableFields.SpecificFields != oldImmutableFields.SpecificFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldCassandraSpec, cs)
	}
	err := validateTwoFactorDelete(cs.TwoFactorDelete, oldCassandraSpec.TwoFactorDelete)
	if err != nil {
		return err
	}
	err = validateSpark(cs.Spark, oldCassandraSpec.Spark)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}
