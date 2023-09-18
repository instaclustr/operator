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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

// +kubebuilder:object:generate:=false
type KafkaAddons interface {
	SchemaRegistry | RestProxy | KarapaceSchemaRegistry | KarapaceRestProxy | DedicatedZookeeper | PrivateLink | Kraft
}

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
	Cluster        `json:",inline"`
	SchemaRegistry []*SchemaRegistry `json:"schemaRegistry,omitempty"`

	// ReplicationFactor to use for new topic.
	// Also represents the number of racks to use when allocating nodes.
	ReplicationFactor int `json:"replicationFactor"`

	// PartitionsNumber number of partitions to use when created new topics.
	PartitionsNumber          int          `json:"partitionsNumber"`
	RestProxy                 []*RestProxy `json:"restProxy,omitempty"`
	AllowDeleteTopics         bool         `json:"allowDeleteTopics"`
	AutoCreateTopics          bool         `json:"autoCreateTopics"`
	ClientToClusterEncryption bool         `json:"clientToClusterEncryption"`
	//+kubebuilder:validation:MinItems:=1
	//+kubebuilder:validation:MaxItems:=1
	DataCentres []*KafkaDataCentre `json:"dataCentres"`

	// Provision additional dedicated nodes for Apache Zookeeper to run on.
	// Zookeeper nodes will be co-located with Kafka if this is not provided
	DedicatedZookeeper       []*DedicatedZookeeper     `json:"dedicatedZookeeper,omitempty"`
	ClientBrokerAuthWithMTLS bool                      `json:"clientBrokerAuthWithMtls,omitempty"`
	KarapaceRestProxy        []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry   []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
	BundledUseOnly           bool                      `json:"bundledUseOnly,omitempty"`
	UserRefs                 []*UserReference          `json:"userRefs,omitempty"`
	Kraft                    []*Kraft                  `json:"kraft,omitempty"`
}

type Kraft struct {
	ControllerNodeCount int `json:"controllerNodeCount"`
}

type KafkaDataCentre struct {
	DataCentre  `json:",inline"`
	PrivateLink []*PrivateLink `json:"privateLink,omitempty"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.status.id`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`

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

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}

func (k *Kafka) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *Kafka) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (k *KafkaSpec) ToInstAPI() *models.KafkaCluster {
	return &models.KafkaCluster{
		SchemaRegistry:            k.schemaRegistryToInstAPI(),
		RestProxy:                 k.restProxyToInstAPI(),
		PCIComplianceMode:         k.PCICompliance,
		DefaultReplicationFactor:  k.ReplicationFactor,
		DefaultNumberOfPartitions: k.PartitionsNumber,
		TwoFactorDelete:           k.TwoFactorDeletesToInstAPI(),
		AllowDeleteTopics:         k.AllowDeleteTopics,
		AutoCreateTopics:          k.AutoCreateTopics,
		ClientToClusterEncryption: k.ClientToClusterEncryption,
		DedicatedZookeeper:        k.dedicatedZookeeperToInstAPI(),
		PrivateNetworkCluster:     k.PrivateNetworkCluster,
		Name:                      k.Name,
		SLATier:                   k.SLATier,
		KafkaVersion:              k.Version,
		DataCentres:               k.dcToInstAPI(),
		ClientBrokerAuthWithMtls:  k.ClientBrokerAuthWithMTLS,
		BundledUseOnly:            k.BundledUseOnly,
		Kraft:                     k.kraftToInstAPI(),
		KarapaceRestProxy:         k.karapaceRestProxyToInstAPI(),
		KarapaceSchemaRegistry:    k.karapaceSchemaRegistryToInstAPI(),
	}
}

func (k *KafkaSpec) schemaRegistryToInstAPI() (iSchemas []*models.SchemaRegistry) {
	for _, schema := range k.SchemaRegistry {
		iSchemas = append(iSchemas, &models.SchemaRegistry{
			Version: schema.Version,
		})
	}

	return
}

func (k *KafkaSpec) karapaceSchemaRegistryToInstAPI() (iSchemas []*models.KarapaceSchemaRegistry) {
	for _, schema := range k.KarapaceSchemaRegistry {
		iSchemas = append(iSchemas, &models.KarapaceSchemaRegistry{
			Version: schema.Version,
		})
	}

	return
}

func (k *KafkaSpec) restProxyToInstAPI() (iRestProxies []*models.RestProxy) {
	for _, proxy := range k.RestProxy {
		iRestProxies = append(iRestProxies, &models.RestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			UseLocalSchemaRegistry:               proxy.UseLocalSchemaRegistry,
			SchemaRegistryServerURL:              proxy.SchemaRegistryServerURL,
			SchemaRegistryUsername:               proxy.SchemaRegistryUsername,
			SchemaRegistryPassword:               proxy.SchemaRegistryPassword,
			Version:                              proxy.Version,
		})
	}

	return
}

func (k *KafkaSpec) karapaceRestProxyToInstAPI() (iRestProxies []*models.KarapaceRestProxy) {
	for _, proxy := range k.KarapaceRestProxy {
		iRestProxies = append(iRestProxies, &models.KarapaceRestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			Version:                              proxy.Version,
		})
	}

	return
}

func (k *KafkaSpec) kraftToInstAPI() (iKraft []*models.Kraft) {
	for _, kraft := range k.Kraft {
		iKraft = append(iKraft, &models.Kraft{
			ControllerNodeCount: kraft.ControllerNodeCount,
		})
	}
	return
}

func (k *KafkaSpec) dedicatedZookeeperToInstAPI() (iZookeepers []*models.DedicatedZookeeper) {
	for _, zookeeper := range k.DedicatedZookeeper {
		iZookeepers = append(iZookeepers, &models.DedicatedZookeeper{
			ZookeeperNodeSize:  zookeeper.NodeSize,
			ZookeeperNodeCount: zookeeper.NodesNumber,
		})
	}

	return
}

func (k *KafkaDataCentre) privateLinkToInstAPI() (iPrivateLink []*models.PrivateLink) {
	for _, link := range k.PrivateLink {
		iPrivateLink = append(iPrivateLink, &models.PrivateLink{
			AdvertisedHostname: link.AdvertisedHostname,
		})
	}

	return
}

func (k *KafkaSpec) dcToInstAPI() (iDCs []*models.KafkaDataCentre) {
	for _, crdDC := range k.DataCentres {
		iDCs = append(iDCs, &models.KafkaDataCentre{
			DataCentre:  crdDC.DataCentre.ToInstAPI(),
			PrivateLink: crdDC.privateLinkToInstAPI(),
		})
	}
	return
}

func (k *KafkaSpec) ToInstAPIUpdate() *models.KafkaInstAPIUpdateRequest {
	return &models.KafkaInstAPIUpdateRequest{
		DataCentre:         k.dcToInstAPI(),
		DedicatedZookeeper: k.dedicatedZookeeperToInstAPIUpdate(),
	}
}

func (k *KafkaSpec) dedicatedZookeeperToInstAPIUpdate() (iZookeepers []*models.DedicatedZookeeperUpdate) {
	for _, zookeeper := range k.DedicatedZookeeper {
		iZookeepers = append(iZookeepers, &models.DedicatedZookeeperUpdate{
			ZookeeperNodeSize: zookeeper.NodeSize,
		})
	}

	return
}

func (k *Kafka) FromInstAPI(iData []byte) (*Kafka, error) {
	iKafka := &models.KafkaCluster{}
	err := json.Unmarshal(iData, iKafka)
	if err != nil {
		return nil, err
	}

	return &Kafka{
		TypeMeta:   k.TypeMeta,
		ObjectMeta: k.ObjectMeta,
		Spec:       k.Spec.FromInstAPI(iKafka),
		Status:     k.Status.FromInstAPI(iKafka),
	}, nil
}

func (ks *KafkaSpec) FromInstAPI(iKafka *models.KafkaCluster) KafkaSpec {
	return KafkaSpec{
		Cluster: Cluster{
			Name:                  iKafka.Name,
			Version:               iKafka.KafkaVersion,
			PCICompliance:         iKafka.PCIComplianceMode,
			PrivateNetworkCluster: iKafka.PrivateNetworkCluster,
			SLATier:               iKafka.SLATier,
			TwoFactorDelete:       ks.Cluster.TwoFactorDeleteFromInstAPI(iKafka.TwoFactorDelete),
		},
		SchemaRegistry:            ks.SchemaRegistryFromInstAPI(iKafka.SchemaRegistry),
		ReplicationFactor:         iKafka.DefaultReplicationFactor,
		PartitionsNumber:          iKafka.DefaultNumberOfPartitions,
		RestProxy:                 ks.RestProxyFromInstAPI(iKafka.RestProxy),
		AllowDeleteTopics:         iKafka.AllowDeleteTopics,
		AutoCreateTopics:          iKafka.AutoCreateTopics,
		ClientToClusterEncryption: iKafka.ClientToClusterEncryption,
		DataCentres:               ks.DCsFromInstAPI(iKafka.DataCentres),
		DedicatedZookeeper:        ks.DedicatedZookeeperFromInstAPI(iKafka.DedicatedZookeeper),
		ClientBrokerAuthWithMTLS:  iKafka.ClientBrokerAuthWithMtls,
		KarapaceRestProxy:         ks.KarapaceRestProxyFromInstAPI(iKafka.KarapaceRestProxy),
		Kraft:                     ks.kraftFromInstAPI(iKafka.Kraft),
		KarapaceSchemaRegistry:    ks.KarapaceSchemaRegistryFromInstAPI(iKafka.KarapaceSchemaRegistry),
		BundledUseOnly:            iKafka.BundledUseOnly,
	}
}

func (ks *KafkaStatus) FromInstAPI(iKafka *models.KafkaCluster) KafkaStatus {
	return KafkaStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iKafka.ID,
			State:                         iKafka.Status,
			DataCentres:                   ks.DCsFromInstAPI(iKafka.DataCentres),
			CurrentClusterOperationStatus: iKafka.CurrentClusterOperationStatus,
			MaintenanceEvents:             ks.MaintenanceEvents,
		},
	}
}

func (ks *KafkaSpec) DCsFromInstAPI(iDCs []*models.KafkaDataCentre) (dcs []*KafkaDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &KafkaDataCentre{
			DataCentre:  ks.Cluster.DCFromInstAPI(iDC.DataCentre),
			PrivateLink: ks.PrivateLinkFromInstAPI(iDC.PrivateLink),
		})
	}
	return
}

func (ks *KafkaSpec) PrivateLinkFromInstAPI(iPLs []*models.PrivateLink) (pls []*PrivateLink) {
	for _, iPL := range iPLs {
		pls = append(pls, &PrivateLink{
			AdvertisedHostname: iPL.AdvertisedHostname,
		})
	}
	return
}

func (ks *KafkaSpec) SchemaRegistryFromInstAPI(iSRs []*models.SchemaRegistry) (srs []*SchemaRegistry) {
	for _, iSR := range iSRs {
		srs = append(srs, &SchemaRegistry{
			Version: iSR.Version,
		})
	}
	return
}

func (ks *KafkaSpec) RestProxyFromInstAPI(iRPs []*models.RestProxy) (rps []*RestProxy) {
	for _, iRP := range iRPs {
		rps = append(rps, &RestProxy{
			IntegrateRestProxyWithSchemaRegistry: iRP.IntegrateRestProxyWithSchemaRegistry,
			UseLocalSchemaRegistry:               iRP.UseLocalSchemaRegistry,
			SchemaRegistryServerURL:              iRP.SchemaRegistryServerURL,
			SchemaRegistryUsername:               iRP.SchemaRegistryUsername,
			SchemaRegistryPassword:               iRP.SchemaRegistryPassword,
			Version:                              iRP.Version,
		})
	}
	return
}

func (ks *KafkaSpec) DedicatedZookeeperFromInstAPI(iDZs []*models.DedicatedZookeeper) (dzs []*DedicatedZookeeper) {
	for _, iDZ := range iDZs {
		dzs = append(dzs, &DedicatedZookeeper{
			NodeSize:    iDZ.ZookeeperNodeSize,
			NodesNumber: iDZ.ZookeeperNodeCount,
		})
	}
	return
}

func (ks *KafkaSpec) KarapaceRestProxyFromInstAPI(iKRPs []*models.KarapaceRestProxy) (krps []*KarapaceRestProxy) {
	for _, iKRP := range iKRPs {
		krps = append(krps, &KarapaceRestProxy{
			IntegrateRestProxyWithSchemaRegistry: iKRP.IntegrateRestProxyWithSchemaRegistry,
			Version:                              iKRP.Version,
		})
	}
	return
}

func (ks *KafkaSpec) kraftFromInstAPI(iKraft []*models.Kraft) (kraft []*Kraft) {
	for _, ikraft := range iKraft {
		kraft = append(kraft, &Kraft{
			ControllerNodeCount: ikraft.ControllerNodeCount,
		})
	}
	return
}

func (ks *KafkaSpec) KarapaceSchemaRegistryFromInstAPI(iKSRs []*models.KarapaceSchemaRegistry) (ksrs []*KarapaceSchemaRegistry) {
	for _, iKSR := range iKSRs {
		ksrs = append(ksrs, &KarapaceSchemaRegistry{
			Version: iKSR.Version,
		})
	}
	return
}

func (ks *KafkaStatus) DCsFromInstAPI(iDCs []*models.KafkaDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dc := ks.DCFromInstAPI(iDC.DataCentre)
		dc.PrivateLink = privateLinkStatusesFromInstAPI(iDC.PrivateLink)
		dcs = append(dcs, dc)
	}
	return dcs
}

func (a *KafkaSpec) IsEqual(b KafkaSpec) bool {
	return a.Cluster.IsEqual(b.Cluster) &&
		a.ReplicationFactor == b.ReplicationFactor &&
		a.PartitionsNumber == b.PartitionsNumber &&
		a.AllowDeleteTopics == b.AllowDeleteTopics &&
		a.AutoCreateTopics == b.AutoCreateTopics &&
		a.ClientToClusterEncryption == b.ClientToClusterEncryption &&
		a.ClientBrokerAuthWithMTLS == b.ClientBrokerAuthWithMTLS &&
		a.BundledUseOnly == b.BundledUseOnly &&
		isKafkaAddonsEqual[SchemaRegistry](a.SchemaRegistry, b.SchemaRegistry) &&
		isKafkaAddonsEqual[RestProxy](a.RestProxy, b.RestProxy) &&
		isKafkaAddonsEqual[KarapaceRestProxy](a.KarapaceRestProxy, b.KarapaceRestProxy) &&
		isKafkaAddonsEqual[Kraft](a.Kraft, b.Kraft) &&
		isKafkaAddonsEqual[KarapaceSchemaRegistry](a.KarapaceSchemaRegistry, b.KarapaceSchemaRegistry) &&
		isKafkaAddonsEqual[DedicatedZookeeper](a.DedicatedZookeeper, b.DedicatedZookeeper) &&
		a.areDCsEqual(b.DataCentres) &&
		a.IsTwoFactorDeleteEqual(b.TwoFactorDelete)
}

func (rs *KafkaSpec) areDCsEqual(b []*KafkaDataCentre) bool {
	a := rs.DataCentres
	if len(a) != len(b) {
		return false
	}

	for i := range b {
		if a[i].Name != b[i].Name {
			continue
		}

		if !a[i].DataCentre.IsEqual(b[i].DataCentre) ||
			!isKafkaAddonsEqual[PrivateLink](a[i].PrivateLink, b[i].PrivateLink) {
			return false
		}
	}

	return true
}
