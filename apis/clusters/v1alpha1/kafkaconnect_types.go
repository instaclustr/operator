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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type TargetCluster struct {

	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`

	// Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.
	ManagedCluster []*ManagedCluster `json:"managedCluster,omitempty"`
}

type ExternalCluster struct {
	SecurityProtocol      string `json:"securityProtocol,omitempty"`
	SSLTruststorePassword string `json:"sslTruststorePassword,omitempty"`
	BootstrapServers      string `json:"bootstrapServers,omitempty"`
	SASLJAASConfig        string `json:"saslJaasConfig,omitempty"`
	SASLMechanism         string `json:"saslMechanism,omitempty"`
	SSLProtocol           string `json:"sslProtocol,omitempty"`
	SSLEnabledProtocols   string `json:"sslEnabledProtocols,omitempty"`
	Truststore            string `json:"truststore,omitempty"`
}

type ManagedCluster struct {
	TargetKafkaClusterID string `json:"targetKafkaClusterId"`

	// 	Available options are KAFKA_VPC, VPC_PEERED, SEPARATE_VPC
	KafkaConnectVPCType string `json:"kafkaConnectVpcType"`
}

type CustomConnectors struct {
	// Settings to access custom connectors located in an azure storage container.
	AzureConnectorSettings []*AzureConnectorSettings `json:"azureConnectorSettings,omitempty"`

	// Settings to access custom connectors located in a S3 bucket.
	AWSConnectorSettings []*AWSConnectorSettings `json:"awsConnectorSettings,omitempty"`

	// Settings to access custom connectors located in a gcp storage container.
	GCPConnectorSettings []*GCPConnectorSettings `json:"gcpConnectorSettings,omitempty"`
}

type AzureConnectorSettings struct {
	// Azure storage container name for Kafka Connect custom connector.
	StorageContainerName string `json:"storageContainerName"`

	// Azure storage account name to access your Azure bucket for Kafka Connect custom connector.
	StorageAccountName string `json:"storageAccountName"`

	// Azure storage account key to access your Azure bucket for Kafka Connect custom connector.
	StorageAccountKey string `json:"storageAccountKey"`
}

type AWSConnectorSettings struct {
	// AWS Identity Access Management role that is used for accessing your specified S3 bucket for Kafka Connect custom connector.
	S3RoleArn string `json:"s3RoleArn"`

	// AWS Secret Key associated with the Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.
	SecretKey string `json:"secretKey"`

	// AWS Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.
	AccessKey string `json:"accessKey"`

	// S3 bucket name for Kafka Connect custom connector.
	S3BucketName string `json:"s3BucketName"`
}

type GCPConnectorSettings struct {
	PrivateKey        string `json:"privateKey"`
	ClientID          string `json:"clientId"`
	ClientEmail       string `json:"clientEmail"`
	ProjectID         string `json:"projectId"`
	StorageBucketName string `json:"storageBucketName"`
	PrivateKeyID      string `json:"privateKeyId"`
}

type KafkaConnectDataCentre struct {
	DataCentre        `json:",inline"`
	ReplicationFactor int32 `json:"replicationFactor"`
}

// KafkaConnectSpec defines the desired state of KafkaConnect
type KafkaConnectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Cluster       `json:",inline"`
	DataCentres   []*KafkaConnectDataCentre `json:"dataCentres"`
	TargetCluster []*TargetCluster          `json:"targetCluster"`

	// CustomConnectors defines the location for custom connector storage and access info.
	CustomConnectors []*CustomConnectors `json:"customConnectors,omitempty"`
}

// KafkaConnectStatus defines the observed state of KafkaConnect
type KafkaConnectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaConnect is the Schema for the kafkaconnects API
type KafkaConnect struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaConnectSpec   `json:"spec,omitempty"`
	Status KafkaConnectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaConnectList contains a list of KafkaConnect
type KafkaConnectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaConnect `json:"items"`
}

func (k *KafkaConnect) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *KafkaConnect) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&KafkaConnect{}, &KafkaConnectList{})
}

type immutableKafkaConnectFields struct {
	immutableCluster
}

type immutableKafkaConnectDCFields struct {
	immutableDC       immutableDC
	ReplicationFactor int32
}

func (kc *KafkaConnectSpec) newImmutableFields() *immutableKafkaConnectFields {
	return &immutableKafkaConnectFields{
		immutableCluster: kc.Cluster.newImmutableFields(),
	}
}

func (kc *KafkaConnectSpec) validateUpdate(oldSpec KafkaConnectSpec) error {
	newImmutableFields := kc.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldSpec, kc)
	}

	err := kc.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	err = kc.validateImmutableTargetClusterFieldsUpdate(kc.TargetCluster, oldSpec.TargetCluster)
	if err != nil {
		return err
	}

	err = validateTwoFactorDelete(kc.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}

func (kc *KafkaConnectSpec) validateImmutableDataCentresFieldsUpdate(oldSpec KafkaConnectSpec) error {
	if len(kc.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range kc.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
		}
	}

	return nil
}

func (kcdc *KafkaConnectDataCentre) newImmutableFields() *immutableKafkaConnectDCFields {
	return &immutableKafkaConnectDCFields{
		immutableDC: immutableDC{
			Name:                kcdc.Name,
			Region:              kcdc.Region,
			CloudProvider:       kcdc.CloudProvider,
			ProviderAccountName: kcdc.ProviderAccountName,
			Network:             kcdc.Network,
		},
		ReplicationFactor: kcdc.ReplicationFactor,
	}
}

func (kc *KafkaConnectSpec) validateImmutableTargetClusterFieldsUpdate(new, old []*TargetCluster) error {
	if len(new) == 0 && len(old) == 0 {
		return models.ErrImmutableTargetCluster
	}

	if len(old) != len(new) {
		return models.ErrImmutableTargetCluster
	}

	for _, index := range new {
		for _, elem := range old {
			err := validateImmutableExternalClusterFields(index, elem)
			if err != nil {
				return err
			}

			err = validateImmutableManagedClusterFields(index, elem)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateImmutableExternalClusterFields(new, old *TargetCluster) error {
	for _, index := range new.ExternalCluster {
		for _, elem := range old.ExternalCluster {
			if *index != *elem {
				return models.ErrImmutableExternalCluster
			}
		}
	}
	return nil
}

func validateImmutableManagedClusterFields(new, old *TargetCluster) error {
	for _, index := range new.ManagedCluster {
		for _, elem := range old.ManagedCluster {
			if *index != *elem {
				return models.ErrImmutableManagedCluster
			}
		}
	}
	return nil
}

func (k *KafkaConnectStatus) IsEqual(instStatus *models.KafkaConnectCluster) bool {
	if k.Status != instStatus.Status ||
		k.CurrentClusterOperationStatus != instStatus.CurrentClusterOperationStatus ||
		!k.AreDCsEqual(instStatus.DataCentres) {
		return false
	}

	return true

}

func (k *KafkaConnectStatus) AreDCsEqual(instDCs []*models.KafkaConnectDataCentre) bool {
	if len(k.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range k.DataCentres {
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

func (k *KafkaConnectStatus) SetFromInst(instStatus *models.KafkaConnectCluster) {
	k.Status = instStatus.Status
	k.CurrentClusterOperationStatus = instStatus.CurrentClusterOperationStatus
	k.SetDCsFromInst(instStatus.DataCentres)
}

func (k *KafkaConnectStatus) SetDCsFromInst(instDCs []*models.KafkaConnectDataCentre) {
	var dcs []*DataCentreStatus
	for _, instDC := range instDCs {
		dc := &DataCentreStatus{
			ID:         instDC.ID,
			Status:     instDC.Status,
			NodeNumber: instDC.NumberOfNodes,
		}
		dc.SetNodesFromInstAPI(instDC.Nodes)
		dcs = append(dcs, dc)
	}
	k.DataCentres = dcs
}
