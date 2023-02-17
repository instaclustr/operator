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
	ReplicationFactor int `json:"replicationFactor"`
}

// KafkaConnectSpec defines the desired state of KafkaConnect
type KafkaConnectSpec struct {
	Cluster       `json:",inline"`
	DataCentres   []*KafkaConnectDataCentre `json:"dataCentres"`
	TargetCluster []*TargetCluster          `json:"targetCluster"`

	// CustomConnectors defines the location for custom connector storage and access info.
	CustomConnectors []*CustomConnectors `json:"customConnectors,omitempty"`
}

// KafkaConnectStatus defines the observed state of KafkaConnect
type KafkaConnectStatus struct {
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
	ReplicationFactor int
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

		if ((int(newDC.NodesNumber)*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
		}
	}

	return nil
}

func (kdc *KafkaConnectDataCentre) newImmutableFields() *immutableKafkaConnectDCFields {
	return &immutableKafkaConnectDCFields{
		immutableDC: immutableDC{
			Name:                kdc.Name,
			Region:              kdc.Region,
			CloudProvider:       kdc.CloudProvider,
			ProviderAccountName: kdc.ProviderAccountName,
			Network:             kdc.Network,
		},
		ReplicationFactor: kdc.ReplicationFactor,
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

func (k *KafkaConnect) FromInst(iKCData []byte) (*KafkaConnect, error) {
	iKC := models.KafkaConnectCluster{}
	err := json.Unmarshal(iKCData, &iKC)
	if err != nil {
		return nil, err
	}

	return &KafkaConnect{
		TypeMeta:   k.TypeMeta,
		ObjectMeta: k.ObjectMeta,
		Spec:       k.Spec.FromInstAPI(iKC),
		Status:     k.Status.FromInstAPI(iKC),
	}, nil
}

func (ks *KafkaConnectSpec) FromInstAPI(iKC models.KafkaConnectCluster) KafkaConnectSpec {
	return KafkaConnectSpec{
		Cluster: Cluster{
			Name:                  iKC.Name,
			Version:               iKC.KafkaConnectVersion,
			PrivateNetworkCluster: iKC.PrivateNetworkCluster,
			SLATier:               iKC.SLATier,
			TwoFactorDelete:       ks.Cluster.TwoFactorDeleteFromInstAPI(iKC.TwoFactorDelete),
		},
		DataCentres:      ks.DCsFromInstAPI(iKC.DataCentres),
		TargetCluster:    ks.TargetClustersFromInstAPI(iKC.TargetCluster),
		CustomConnectors: ks.CustomConnectorsFromInstAPI(iKC.CustomConnectors),
	}
}

func (ks *KafkaConnectStatus) FromInstAPI(iKC models.KafkaConnectCluster) KafkaConnectStatus {
	return KafkaConnectStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iKC.ID,
			State:                         iKC.Status,
			DataCentres:                   ks.DCsFromInstAPI(iKC.DataCentres),
			CurrentClusterOperationStatus: iKC.CurrentClusterOperationStatus,
			MaintenanceEvents:             ks.MaintenanceEvents,
		},
	}
}

func (ks *KafkaConnectSpec) DCsFromInstAPI(iDCs []*models.KafkaConnectDataCentre) (dcs []*KafkaConnectDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &KafkaConnectDataCentre{
			DataCentre:        ks.Cluster.DCFromInstAPI(iDC.DataCentre),
			ReplicationFactor: iDC.ReplicationFactor,
		})
	}
	return
}

func (ks *KafkaConnectSpec) TargetClustersFromInstAPI(iClusters []*models.TargetCluster) (clusters []*TargetCluster) {
	for _, iCluster := range iClusters {
		clusters = append(clusters, &TargetCluster{
			ExternalCluster: ks.ExternalClustersFromInstAPI(iCluster.ExternalCluster),
			ManagedCluster:  ks.ManagedClustersFromInstAPI(iCluster.ManagedCluster),
		})
	}
	return
}

func (ks *KafkaConnectSpec) ExternalClustersFromInstAPI(iClusters []*models.ExternalCluster) (clusters []*ExternalCluster) {
	for _, iCluster := range iClusters {
		clusters = append(clusters, &ExternalCluster{
			SecurityProtocol:      iCluster.SecurityProtocol,
			SSLTruststorePassword: iCluster.SSLTruststorePassword,
			BootstrapServers:      iCluster.BootstrapServers,
			SASLJAASConfig:        iCluster.SASLJAASConfig,
			SASLMechanism:         iCluster.SASLMechanism,
			SSLProtocol:           iCluster.SSLProtocol,
			SSLEnabledProtocols:   iCluster.SSLEnabledProtocols,
			Truststore:            iCluster.Truststore,
		})
	}
	return
}

func (ks *KafkaConnectSpec) ManagedClustersFromInstAPI(iClusters []*models.ManagedCluster) (clusters []*ManagedCluster) {
	for _, iCluster := range iClusters {
		clusters = append(clusters, &ManagedCluster{
			TargetKafkaClusterID: iCluster.TargetKafkaClusterID,
			KafkaConnectVPCType:  iCluster.KafkaConnectVPCType,
		})
	}
	return
}

func (ks *KafkaConnectSpec) CustomConnectorsFromInstAPI(iConns []*models.CustomConnectors) (conns []*CustomConnectors) {
	for _, iConn := range iConns {
		conns = append(conns, &CustomConnectors{
			AzureConnectorSettings: ks.AzureConnectorSettingsFromInstAPI(iConn.AzureConnectorSettings),
			AWSConnectorSettings:   ks.AWSConnectorSettingsFromInstAPI(iConn.AWSConnectorSettings),
			GCPConnectorSettings:   ks.GCPConnectorSettingsFromInstAPI(iConn.GCPConnectorSettings),
		})
	}
	return
}

func (ks *KafkaConnectSpec) AzureConnectorSettingsFromInstAPI(iSettings []*models.AzureConnectorSettings) (settings []*AzureConnectorSettings) {
	for _, iCluster := range iSettings {
		settings = append(settings, &AzureConnectorSettings{
			StorageContainerName: iCluster.StorageContainerName,
			StorageAccountName:   iCluster.StorageAccountName,
			StorageAccountKey:    iCluster.StorageAccountKey,
		})
	}
	return
}

func (ks *KafkaConnectSpec) AWSConnectorSettingsFromInstAPI(iSettings []*models.AWSConnectorSettings) (settings []*AWSConnectorSettings) {
	for _, iCluster := range iSettings {
		settings = append(settings, &AWSConnectorSettings{
			S3RoleArn:    iCluster.S3RoleArn,
			SecretKey:    iCluster.SecretKey,
			AccessKey:    iCluster.AccessKey,
			S3BucketName: iCluster.S3BucketName,
		})
	}
	return
}

func (ks *KafkaConnectSpec) GCPConnectorSettingsFromInstAPI(iSettings []*models.GCPConnectorSettings) (settings []*GCPConnectorSettings) {
	for _, iCluster := range iSettings {
		settings = append(settings, &GCPConnectorSettings{
			PrivateKey:        iCluster.PrivateKeyID,
			ClientID:          iCluster.ClientID,
			ClientEmail:       iCluster.ClientEmail,
			ProjectID:         iCluster.ProjectID,
			StorageBucketName: iCluster.StorageBucketName,
			PrivateKeyID:      iCluster.PrivateKey,
		})
	}
	return
}

func (ks *KafkaConnectStatus) DCsFromInstAPI(iDCs []*models.KafkaConnectDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, ks.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func (ks *KafkaConnectSpec) IsEqual(kc KafkaConnectSpec) bool {
	return ks.Cluster.IsEqual(kc.Cluster) &&
		ks.AreDataCentresEqual(kc.DataCentres) &&
		ks.AreTargetClustersEqual(kc.TargetCluster) &&
		ks.AreCustomConnectorsEqual(kc.CustomConnectors)
}

func (ks *KafkaConnectSpec) AreDataCentresEqual(dcs []*KafkaConnectDataCentre) bool {
	if len(ks.DataCentres) != len(dcs) {
		return false
	}

	for i, iDC := range dcs {
		dataCentre := ks.DataCentres[i]
		if !dataCentre.IsEqual(iDC.DataCentre) ||
			iDC.ReplicationFactor != dataCentre.ReplicationFactor {
			return false
		}
	}

	return true
}

func (ks *KafkaConnectSpec) AreTargetClustersEqual(tClusters []*TargetCluster) bool {
	if len(ks.TargetCluster) != len(tClusters) {
		return false
	}

	for i, tCluster := range tClusters {
		cluster := ks.TargetCluster[i]
		if !cluster.AreExternalClustersEqual(tCluster.ExternalCluster) ||
			!cluster.AreManagedClustersEqual(tCluster.ManagedCluster) {
			return false
		}
	}

	return true
}

func (tc *TargetCluster) AreExternalClustersEqual(eClusters []*ExternalCluster) bool {
	if len(tc.ExternalCluster) != len(eClusters) {
		return false
	}

	for i, eCluster := range eClusters {
		cluster := tc.ExternalCluster[i]
		if *eCluster != *cluster {
			return false
		}
	}

	return true
}

func (tc *TargetCluster) AreManagedClustersEqual(mClusters []*ManagedCluster) bool {
	if len(tc.ManagedCluster) != len(mClusters) {
		return false
	}

	for i, mCluster := range mClusters {
		cluster := tc.ManagedCluster[i]
		if *mCluster != *cluster {
			return false
		}
	}

	return true
}

func (ks *KafkaConnectSpec) AreCustomConnectorsEqual(cConns []*CustomConnectors) bool {
	if len(ks.CustomConnectors) != len(cConns) {
		return false
	}

	for i, cConn := range cConns {
		conn := ks.CustomConnectors[i]
		if !conn.AreAzureConnectorSettingsEqual(cConn.AzureConnectorSettings) ||
			!conn.AreAWSConnectorSettingsEqual(cConn.AWSConnectorSettings) ||
			!conn.AreGCPConnectorSettingsEqual(cConn.GCPConnectorSettings) {
			return false
		}
	}

	return true
}

func (cc *CustomConnectors) AreAzureConnectorSettingsEqual(aSettings []*AzureConnectorSettings) bool {
	if len(cc.AzureConnectorSettings) != len(aSettings) {
		return false
	}

	for i, aSetting := range aSettings {
		settings := cc.AzureConnectorSettings[i]
		if *aSetting != *settings {
			return false
		}
	}

	return true
}

func (cc *CustomConnectors) AreAWSConnectorSettingsEqual(aSettings []*AWSConnectorSettings) bool {
	if len(cc.AWSConnectorSettings) != len(aSettings) {
		return false
	}

	for i, aSetting := range aSettings {
		settings := cc.AWSConnectorSettings[i]
		if *aSetting != *settings {
			return false
		}
	}

	return true
}

func (cc *CustomConnectors) AreGCPConnectorSettingsEqual(gSettings []*GCPConnectorSettings) bool {
	if len(cc.GCPConnectorSettings) != len(gSettings) {
		return false
	}

	for i, gSetting := range gSettings {
		settings := cc.GCPConnectorSettings[i]
		if *gSetting != *settings {
			return false
		}
	}

	return true
}

func (ks *KafkaConnectSpec) NewDCsUpdate() models.KafkaConnectAPIUpdate {
	return models.KafkaConnectAPIUpdate{
		DataCentres: ks.DCsToInstAPI(),
	}
}

func (ks *KafkaConnectSpec) ToInstAPI() models.KafkaConnectCluster {
	return models.KafkaConnectCluster{
		Name:                  ks.Name,
		KafkaConnectVersion:   ks.Version,
		PrivateNetworkCluster: ks.PrivateNetworkCluster,
		SLATier:               ks.SLATier,
		TwoFactorDelete:       ks.TwoFactorDeletesToInstAPI(),
		CustomConnectors:      ks.CustomConnectorsToInstAPI(),
		TargetCluster:         ks.TargetClustersToInstAPI(),
		DataCentres:           ks.DCsToInstAPI(),
	}
}

func (ks *KafkaConnectSpec) DCsToInstAPI() (iDCs []*models.KafkaConnectDataCentre) {
	for _, dc := range ks.DataCentres {
		iDCs = append(iDCs, dc.ToInstAPI())
	}
	return
}

func (kdc *KafkaConnectDataCentre) ToInstAPI() *models.KafkaConnectDataCentre {
	return &models.KafkaConnectDataCentre{
		DataCentre:        kdc.DataCentre.ToInstAPI(),
		ReplicationFactor: kdc.ReplicationFactor,
	}
}

func (ks *KafkaConnectSpec) CustomConnectorsToInstAPI() (iConns []*models.CustomConnectors) {
	for _, conn := range ks.CustomConnectors {
		iConns = append(iConns, conn.ToInstAPI())
	}
	return
}

func (cc *CustomConnectors) ToInstAPI() *models.CustomConnectors {
	return &models.CustomConnectors{
		AzureConnectorSettings: cc.AzureConnectorSettingsToInstAPI(),
		AWSConnectorSettings:   cc.AWSConnectorSettingsToInstAPI(),
		GCPConnectorSettings:   cc.GCPConnectorSettingsToInstAPI(),
	}
}

func (cc *CustomConnectors) AzureConnectorSettingsToInstAPI() (iSettings []*models.AzureConnectorSettings) {
	for _, settings := range cc.AzureConnectorSettings {
		iSettings = append(iSettings, &models.AzureConnectorSettings{
			StorageContainerName: settings.StorageContainerName,
			StorageAccountName:   settings.StorageAccountName,
			StorageAccountKey:    settings.StorageAccountKey,
		})
	}
	return
}

func (cc *CustomConnectors) AWSConnectorSettingsToInstAPI() (iSettings []*models.AWSConnectorSettings) {
	for _, settings := range cc.AWSConnectorSettings {
		iSettings = append(iSettings, &models.AWSConnectorSettings{
			S3RoleArn:    settings.S3RoleArn,
			SecretKey:    settings.SecretKey,
			AccessKey:    settings.AccessKey,
			S3BucketName: settings.S3BucketName,
		})
	}
	return
}

func (cc *CustomConnectors) GCPConnectorSettingsToInstAPI() (iSettings []*models.GCPConnectorSettings) {
	for _, settings := range cc.GCPConnectorSettings {
		iSettings = append(iSettings, &models.GCPConnectorSettings{
			PrivateKey:        settings.PrivateKey,
			ClientID:          settings.ClientID,
			ClientEmail:       settings.ClientEmail,
			ProjectID:         settings.ProjectID,
			StorageBucketName: settings.StorageBucketName,
			PrivateKeyID:      settings.PrivateKeyID,
		})
	}
	return
}

func (ks *KafkaConnectSpec) TargetClustersToInstAPI() (iClusters []*models.TargetCluster) {
	for _, cluster := range ks.TargetCluster {
		iClusters = append(iClusters, cluster.ToInstAPI())
	}
	return
}

func (tc *TargetCluster) ToInstAPI() *models.TargetCluster {
	return &models.TargetCluster{
		ExternalCluster: tc.ExternalClustersToInstAPI(),
		ManagedCluster:  tc.ManagedClustersToInstAPI(),
	}
}

func (tc *TargetCluster) ExternalClustersToInstAPI() (iClusters []*models.ExternalCluster) {
	for _, cluster := range tc.ExternalCluster {
		iClusters = append(iClusters, &models.ExternalCluster{
			SecurityProtocol:      cluster.SecurityProtocol,
			SSLTruststorePassword: cluster.SSLTruststorePassword,
			BootstrapServers:      cluster.BootstrapServers,
			SASLJAASConfig:        cluster.SASLJAASConfig,
			SASLMechanism:         cluster.SASLMechanism,
			SSLProtocol:           cluster.SSLProtocol,
			SSLEnabledProtocols:   cluster.SSLEnabledProtocols,
			Truststore:            cluster.Truststore,
		})
	}
	return
}

func (tc *TargetCluster) ManagedClustersToInstAPI() (iClusters []*models.ManagedCluster) {
	for _, cluster := range tc.ManagedCluster {
		iClusters = append(iClusters, &models.ManagedCluster{
			TargetKafkaClusterID: cluster.TargetKafkaClusterID,
			KafkaConnectVPCType:  cluster.KafkaConnectVPCType,
		})
	}
	return
}
