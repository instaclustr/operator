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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	models2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

type SchemaRegistry struct {
	Version string `json:"version"`
}

type RestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	UseLocalSchemaRegistry               bool   `json:"useLocalSchemaRegistry,omitempty"`
	SchemaRegistryServerURL              string `json:"schemaRegistryServerUrl,omitempty"`
	SchemaRegistryUsername               string `json:"schemaRegistryUsername,omitempty"`
	SchemaRegistryPassword               string `json:"schemaRegistryPassword,omitempty"`
	Version                              string `json:"version"`
}

type DedicatedZookeeper struct {
	// Size of the nodes provisioned as dedicated Zookeeper nodes.
	NodeSize string `json:"nodeSize"`

	// Number of dedicated Zookeeper node count, it must be 3 or 5.
	NodesNumber int32 `json:"nodesNumber"`
}

type KarapaceRestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	Version                              string `json:"version"`
}

type KarapaceSchemaRegistry struct {
	Version string `json:"version"`
}

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Cluster        `json:",inline"`
	SchemaRegistry []*SchemaRegistry `json:"schemaRegistry,omitempty"`

	// ReplicationFactorNumber to use for new topic.
	// Also represents the number of racks to use when allocating nodes.
	ReplicationFactorNumber int32 `json:"replicationFactorNumber"`

	// PartitionsNumber number of partitions to use when created new topics.
	PartitionsNumber          int32              `json:"partitionsNumber"`
	RestProxy                 []*RestProxy       `json:"restProxy,omitempty"`
	AllowDeleteTopics         bool               `json:"allowDeleteTopics"`
	AutoCreateTopics          bool               `json:"autoCreateTopics"`
	ClientToClusterEncryption bool               `json:"clientToClusterEncryption"`
	DataCentres               []*KafkaDataCentre `json:"dataCentres"`

	// Provision additional dedicated nodes for Apache Zookeeper to run on.
	// Zookeeper nodes will be co-located with Kafka if this is not provided
	DedicatedZookeeper                []*DedicatedZookeeper     `json:"dedicatedZookeeper,omitempty"`
	ClientBrokerAuthWithMTLS          bool                      `json:"clientBrokerAuthWithMtls,omitempty"`
	ClientAuthBrokerWithoutEncryption bool                      `json:"clientAuthBrokerWithoutEncryption,omitempty"`
	ClientAuthBrokerWithEncryption    bool                      `json:"clientAuthBrokerWithEncryption,omitempty"`
	KarapaceRestProxy                 []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
	BundledUseOnly                    bool                      `json:"bundledUseOnly,omitempty"`
}

type KafkaDataCentre struct {
	DataCentre  `json:",inline"`
	PrivateLink []*KafkaPrivateLink `json:"privateLink,omitempty"`
}

type KafkaPrivateLink struct {
	AdvertisedHostname string `json:"advertisedHostname"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kafka is the Schema for the kafkas API
type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSpec   `json:"spec,omitempty"`
	Status KafkaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaList contains a list of Kafka
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func (k *Kafka) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *Kafka) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (k *KafkaSpec) ToInstAPI() *models.KafkaInstAPICreateRequest {
	return &models.KafkaInstAPICreateRequest{
		SchemaRegistry:                    k.schemaRegistryToInstAPI(),
		RestProxy:                         k.restProxyToInstAPI(),
		PCIComplianceMode:                 k.PCICompliance,
		DefaultReplicationFactor:          k.ReplicationFactorNumber,
		DefaultNumberOfPartitions:         k.PartitionsNumber,
		TwoFactorDelete:                   k.TwoFactorDeletesToInstAPI(),
		AllowDeleteTopics:                 k.AllowDeleteTopics,
		AutoCreateTopics:                  k.AutoCreateTopics,
		ClientToClusterEncryption:         k.ClientToClusterEncryption,
		DedicatedZookeeper:                k.dedicatedZookeeperToInstAPI(),
		PrivateNetworkCluster:             k.PrivateNetworkCluster,
		Name:                              k.Name,
		SLATier:                           k.SLATier,
		KafkaVersion:                      k.Version,
		KafkaDataCentre:                   k.dcToInstAPI(),
		ClientBrokerAuthWithMTLS:          k.ClientBrokerAuthWithMTLS,
		ClientAuthBrokerWithoutEncryption: k.ClientAuthBrokerWithoutEncryption,
		ClientAuthBrokerWithEncryption:    k.ClientAuthBrokerWithEncryption,
		BundledUseOnly:                    k.BundledUseOnly,
		KarapaceRestProxy:                 k.karapaceRestProxyToInstAPI(),
		KarapaceSchemaRegistry:            k.karapaceSchemaRegistryToInstAPI(),
	}
}

func (k *KafkaSpec) schemaRegistryToInstAPI() []*models.SchemaRegistry {
	var instaSchemas []*models.SchemaRegistry
	for _, schema := range k.SchemaRegistry {
		instaSchemas = append(instaSchemas, &models.SchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func (k *KafkaSpec) karapaceSchemaRegistryToInstAPI() []*models.KarapaceSchemaRegistry {
	var instaSchemas []*models.KarapaceSchemaRegistry
	for _, schema := range k.KarapaceSchemaRegistry {
		instaSchemas = append(instaSchemas, &models.KarapaceSchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func (k *KafkaSpec) restProxyToInstAPI() []*models.RestProxy {
	var instaRestProxies []*models.RestProxy
	for _, proxy := range k.RestProxy {
		instaRestProxies = append(instaRestProxies, &models.RestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			UseLocalSchemaRegistry:               proxy.UseLocalSchemaRegistry,
			SchemaRegistryServerURL:              proxy.SchemaRegistryServerURL,
			SchemaRegistryUsername:               proxy.SchemaRegistryUsername,
			SchemaRegistryPassword:               proxy.SchemaRegistryPassword,
			Version:                              proxy.Version,
		})
	}

	return instaRestProxies
}

func (k *KafkaSpec) karapaceRestProxyToInstAPI() []*models.KarapaceRestProxy {
	var instaRestProxies []*models.KarapaceRestProxy
	for _, proxy := range k.KarapaceRestProxy {
		instaRestProxies = append(instaRestProxies, &models.KarapaceRestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			Version:                              proxy.Version,
		})
	}

	return instaRestProxies
}

func (k *KafkaSpec) dedicatedZookeeperToInstAPI() []*models.DedicatedZookeeper {
	var instaZookeepers []*models.DedicatedZookeeper
	for _, zookeeper := range k.DedicatedZookeeper {
		instaZookeepers = append(instaZookeepers, &models.DedicatedZookeeper{
			ZookeeperNodeSize:  zookeeper.NodeSize,
			ZookeeperNodeCount: zookeeper.NodesNumber,
		})
	}

	return instaZookeepers
}

func (k *KafkaDataCentre) privateLinkToInstAPI() []*models.KafkaPrivateLink {
	var instaPrivateLink []*models.KafkaPrivateLink
	for _, link := range k.PrivateLink {
		instaPrivateLink = append(instaPrivateLink, &models.KafkaPrivateLink{
			AdvertisedHostname: link.AdvertisedHostname,
		})
	}

	return instaPrivateLink
}

func (k *KafkaSpec) dcToInstAPI() []*models.KafkaDataCentre {
	var instaDCs []*models.KafkaDataCentre
	for _, crdDC := range k.DataCentres {
		dc := &models2.DataCentre{
			Name:                crdDC.Name,
			Network:             crdDC.Network,
			NodeSize:            crdDC.NodeSize,
			NumberOfNodes:       crdDC.NodesNumber,
			CloudProvider:       crdDC.CloudProvider,
			Region:              crdDC.Region,
			ProviderAccountName: crdDC.ProviderAccountName,
		}
		crdDC.CloudProviderSettingsToInstAPI(dc)
		crdDC.TagsToInstAPI(dc)

		kafkaDC := &models.KafkaDataCentre{
			DataCentre:  *dc,
			PrivateLink: crdDC.privateLinkToInstAPI(),
		}

		instaDCs = append(instaDCs, kafkaDC)
	}

	return instaDCs
}

func (k *KafkaSpec) ToInstAPIUpdate() *models.KafkaInstAPIUpdateRequest {
	newKafka := &models.KafkaInstAPIUpdateRequest{}
	newKafka.DataCentre = k.dcToInstAPI()
	newKafka.DedicatedZookeeper = k.dedicatedZookeeperToInstAPIUpdate()

	return newKafka
}

func (k *KafkaSpec) dedicatedZookeeperToInstAPIUpdate() []*models.DedicatedZookeeperUpdate {
	var instaZookeepers []*models.DedicatedZookeeperUpdate
	for _, zookeeper := range k.DedicatedZookeeper {
		instaZookeepers = append(instaZookeepers, &models.DedicatedZookeeperUpdate{
			ZookeeperNodeSize: zookeeper.NodeSize,
		})
	}

	return instaZookeepers
}

func (k *KafkaStatus) IsEqual(instStatus *models.KafkaCluster) bool {
	if k.Status != instStatus.Status ||
		k.CurrentClusterOperationStatus != instStatus.CurrentClusterOperationStatus ||
		!k.AreDCsEqual(instStatus.DataCentres) {
		return false
	}

	return true

}

func (k *KafkaStatus) AreDCsEqual(instDCs []*models.KafkaDataCentre) bool {
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

func (k *KafkaStatus) SetFromInst(instStatus *models.KafkaCluster) {
	k.Status = instStatus.Status
	k.CurrentClusterOperationStatus = instStatus.CurrentClusterOperationStatus
	k.SetDCsFromInst(instStatus.DataCentres)
}

func (k *KafkaStatus) SetDCsFromInst(instDCs []*models.KafkaDataCentre) {
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

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
