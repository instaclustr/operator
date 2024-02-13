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
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresource "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type TargetCluster struct {
	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`

	// Details to connect to an Instaclustr managed cluster. Cannot be provided if targeting an external cluster.
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
	TargetKafkaClusterID string                      `json:"targetKafkaClusterId,omitempty"`
	ClusterRef           *clusterresource.ClusterRef `json:"clusterRef,omitempty"`

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
	SecretKey string `json:"secretKey,omitempty"`

	// AWS Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.
	AccessKey string `json:"accessKey,omitempty"`

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
	GenericDataCentreSpec `json:",inline"`

	NodeSize          string `json:"nodeSize"`
	NumberOfNodes     int    `json:"numberOfNodes"`
	ReplicationFactor int    `json:"replicationFactor"`
}

// KafkaConnectSpec defines the desired state of KafkaConnect
type KafkaConnectSpec struct {
	GenericClusterSpec `json:",inline"`

	DataCentres   []*KafkaConnectDataCentre `json:"dataCentres"`
	TargetCluster []*TargetCluster          `json:"targetCluster"`

	// CustomConnectors defines the location for custom connector storage and access info.
	CustomConnectors []*CustomConnectors `json:"customConnectors,omitempty"`
}

// KafkaConnectStatus defines the observed state of KafkaConnect
type KafkaConnectStatus struct {
	GenericStatus `json:",inline"`

	DataCentres []*KafkaConnectDataCentreStatus `json:"dataCentres,omitempty"`

	DefaultUserSecretRef *Reference `json:"defaultUserSecretRef,omitempty"`
}

type KafkaConnectDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	NumberOfNodes        int        `json:"numberOfNodes,omitempty"`
	Nodes                []*Node    `json:"nodes,omitempty"`
	DefaultUserSecretRef *Reference `json:"defaultUserSecretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

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
	return k.Kind + "/" + client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *KafkaConnect) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (k *KafkaConnect) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return k.Status.DataCentres[0].ID
	}
	for _, cdc := range k.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}
func (k *KafkaConnect) GetClusterID() string {
	return k.Status.ID
}

func init() {
	SchemeBuilder.Register(&KafkaConnect{}, &KafkaConnectList{})
}

func (k *KafkaConnect) FromInstAPI(instaModel *models.KafkaConnectCluster) {
	k.Spec.FromInstAPI(instaModel)
	k.Status.FromInstAPI(instaModel)
}

func (ks *KafkaConnectSpec) FromInstAPI(instaModel *models.KafkaConnectCluster) {
	ks.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields)

	ks.Version = instaModel.KafkaConnectVersion

	ks.DCsFromInstAPI(instaModel.DataCentres)
	ks.TargetClustersFromInstAPI(instaModel.TargetCluster)
	ks.CustomConnectorsFromInstAPI(instaModel.CustomConnectors)
}

func (ks *KafkaConnectStatus) FromInstAPI(instaModel *models.KafkaConnectCluster) {
	ks.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	ks.DCsFromInstAPI(instaModel.DataCentres)
}

func (ks *KafkaConnectSpec) DCsFromInstAPI(instaModels []*models.KafkaConnectDataCentre) {
	ks.DataCentres = make([]*KafkaConnectDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := KafkaConnectDataCentre{}
		dc.FromInstAPI(instaModel)
		ks.DataCentres[i] = &dc
	}
}

func (k *KafkaConnectDataCentre) FromInstAPI(instaModel *models.KafkaConnectDataCentre) {
	k.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)

	k.NodeSize = instaModel.NodeSize
	k.NumberOfNodes = instaModel.NumberOfNodes
	k.ReplicationFactor = instaModel.ReplicationFactor
}

func (k *KafkaConnectDataCentre) Equals(o *KafkaConnectDataCentre) bool {
	return k.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		k.NumberOfNodes == o.NumberOfNodes &&
		k.ReplicationFactor == o.ReplicationFactor &&
		k.NodeSize == o.NodeSize
}

func (ks *KafkaConnectSpec) TargetClustersFromInstAPI(instaModels []*models.TargetCluster) {
	targetCluster := make([]*TargetCluster, len(instaModels))
	for i, instaModel := range instaModels {
		targetCluster[i] = &TargetCluster{
			ExternalCluster: ks.ExternalClustersFromInstAPI(instaModel.ExternalCluster),
			ManagedCluster:  ks.ManagedClustersFromInstAPI(instaModel.ManagedCluster),
		}
	}
	ks.TargetCluster = targetCluster
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
	var clusterRef *clusterresource.ClusterRef
	if managedCluster := ks.GetManagedCluster(); managedCluster != nil {
		clusterRef = managedCluster.ClusterRef
	}

	for _, iCluster := range iClusters {
		clusters = append(clusters, &ManagedCluster{
			TargetKafkaClusterID: iCluster.TargetKafkaClusterID,
			KafkaConnectVPCType:  iCluster.KafkaConnectVPCType,
			ClusterRef:           clusterRef,
		})
	}
	return
}

func (ks *KafkaConnectSpec) CustomConnectorsFromInstAPI(instaModels []*models.CustomConnectors) {
	ks.CustomConnectors = make([]*CustomConnectors, len(instaModels))
	for i, iConn := range instaModels {
		ks.CustomConnectors[i] = &CustomConnectors{
			AzureConnectorSettings: ks.AzureConnectorSettingsFromInstAPI(iConn.AzureConnectorSettings),
			AWSConnectorSettings:   ks.AWSConnectorSettingsFromInstAPI(iConn.AWSConnectorSettings),
			GCPConnectorSettings:   ks.GCPConnectorSettingsFromInstAPI(iConn.GCPConnectorSettings),
		}
	}
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

func (ks *KafkaConnectStatus) DCsFromInstAPI(instaModels []*models.KafkaConnectDataCentre) {
	ks.DataCentres = make([]*KafkaConnectDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := KafkaConnectDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		ks.DataCentres[i] = &dc
	}
}

func (c *KafkaConnect) GetSpec() KafkaConnectSpec { return c.Spec }

func (c *KafkaConnect) IsSpecEqual(spec KafkaConnectSpec) bool {
	return c.Spec.Equals(&spec)
}

func (ks *KafkaConnectSpec) Equals(o *KafkaConnectSpec) bool {
	return ks.GenericClusterSpec.Equals(&o.GenericClusterSpec) &&
		ks.DCsEqual(o.DataCentres) &&
		ks.TargetClustersEqual(o.TargetCluster) &&
		ks.CustomConnectorsEqual(o.CustomConnectors)
}

func (ks *KafkaConnectSpec) DCsEqual(o []*KafkaConnectDataCentre) bool {
	if len(ks.DataCentres) != len(o) {
		return false
	}

	m := map[string]*KafkaConnectDataCentre{}
	for _, dc := range ks.DataCentres {
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

func (ks *KafkaConnectSpec) TargetClustersEqual(o []*TargetCluster) bool {
	if len(ks.TargetCluster) != len(o) {
		return false
	}

	for i := range o {

		if !slices.EqualsPtr(ks.TargetCluster[i].ExternalCluster, o[i].ExternalCluster) ||
			!ks.managedClustersEqual(ks.TargetCluster[i].ManagedCluster, o[i].ManagedCluster) {
			return false
		}
	}

	return true
}

func (ks *KafkaConnectSpec) managedClustersEqual(m1, m2 []*ManagedCluster) bool {
	if len(m1) != len(m2) {
		return false
	}

	for i := range m1 {
		if m1[i].TargetKafkaClusterID != m2[i].TargetKafkaClusterID ||
			m1[i].KafkaConnectVPCType != m2[i].KafkaConnectVPCType {
			return false
		}
	}

	return true
}

func (ks *KafkaConnectSpec) CustomConnectorsEqual(o []*CustomConnectors) bool {
	if len(ks.CustomConnectors) != len(o) {
		return false
	}

	for i := range ks.CustomConnectors {
		if !slices.EqualsPtr(ks.CustomConnectors[i].AWSConnectorSettings, o[i].AWSConnectorSettings) ||
			!slices.EqualsPtr(ks.CustomConnectors[i].GCPConnectorSettings, o[i].GCPConnectorSettings) ||
			!slices.EqualsPtr(ks.CustomConnectors[i].AzureConnectorSettings, o[i].AzureConnectorSettings) {
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

func (ks *KafkaConnectSpec) ToInstAPI() *models.KafkaConnectCluster {
	return &models.KafkaConnectCluster{
		GenericClusterFields: ks.GenericClusterSpec.ToInstAPI(),
		KafkaConnectVersion:  ks.Version,
		CustomConnectors:     ks.CustomConnectorsToInstAPI(),
		TargetCluster:        ks.TargetClustersToInstAPI(),
		DataCentres:          ks.DCsToInstAPI(),
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
		GenericDataCentreFields: kdc.GenericDataCentreSpec.ToInstAPI(),
		NodeSize:                kdc.NodeSize,
		NumberOfNodes:           kdc.NumberOfNodes,
		ReplicationFactor:       kdc.ReplicationFactor,
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

func (k *KafkaConnect) NewDefaultUserSecret(username, password string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.SecretKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, k.Name),
			Namespace: k.Namespace,
			Labels: map[string]string{
				models.ControlledByLabel:  k.Name,
				models.DefaultSecretLabel: "true",
			},
		},
		StringData: map[string]string{
			models.Username: username,
			models.Password: password,
		},
	}
}

func (k *KafkaConnect) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !k.Spec.PrivateNetwork {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.KafkaConnectAPI,
				Port: models.Port8083,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port8083,
				},
			},
		}
	}
	return exposePorts
}

func (k *KafkaConnect) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.KafkaConnectAPI,
			Port: models.Port8083,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port8083,
			},
		},
	}
	return headlessPorts
}

func (s *KafkaConnectDataCentreStatus) FromInstAPI(instaModel *models.KafkaConnectDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	s.NumberOfNodes = instaModel.NumberOfNodes
	s.Nodes = nodesFromInstAPI(instaModel.Nodes)
}

func (s *KafkaConnectDataCentreStatus) Equals(o *KafkaConnectDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		nodesEqual(s.Nodes, o.Nodes) &&
		s.NumberOfNodes == o.NumberOfNodes
}

func (ks *KafkaConnectSpec) GetManagedCluster() *ManagedCluster {
	if len(ks.TargetCluster) < 1 {
		return nil
	}

	if len(ks.TargetCluster[0].ManagedCluster) < 1 || ks.TargetCluster[0].ManagedCluster[0] == nil {
		return nil
	}

	return ks.TargetCluster[0].ManagedCluster[0]
}

func (ks *KafkaConnectStatus) Equals(o *KafkaConnectStatus) bool {
	return ks.GenericStatus.Equals(&o.GenericStatus) &&
		ks.DCsEqual(o.DataCentres)
}

func (ks *KafkaConnectStatus) DCsEqual(o []*KafkaConnectDataCentreStatus) bool {
	if len(ks.DataCentres) != len(o) {
		return false
	}

	m := map[string]*KafkaConnectDataCentreStatus{}
	for _, dc := range ks.DataCentres {
		m[dc.ID] = dc
	}

	for _, dc := range o {
		mDC, ok := m[dc.ID]
		if !ok {
			return false
		}

		if !mDC.Equals(dc) {
			return false
		}
	}

	return true
}
