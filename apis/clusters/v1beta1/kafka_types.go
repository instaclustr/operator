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

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
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
	GenericClusterSpec `json:",inline"`

	// ReplicationFactor to use for new topic.
	// Also represents the number of racks to use when allocating nodes.
	ReplicationFactor int `json:"replicationFactor"`

	// PartitionsNumber number of partitions to use when created new topics.
	PartitionsNumber int `json:"partitionsNumber"`

	AllowDeleteTopics         bool       `json:"allowDeleteTopics"`
	AutoCreateTopics          bool       `json:"autoCreateTopics"`
	ClientToClusterEncryption bool       `json:"clientToClusterEncryption"`
	ClientBrokerAuthWithMTLS  bool       `json:"clientBrokerAuthWithMtls,omitempty"`
	BundledUseOnly            bool       `json:"bundledUseOnly,omitempty"`
	PCICompliance             bool       `json:"pciCompliance,omitempty"`
	UserRefs                  References `json:"userRefs,omitempty" dcomparisonSkip:"true"`

	// Provision additional dedicated nodes for Apache Zookeeper to run on.
	// Zookeeper nodes will be co-located with Kafka if this is not provided
	DedicatedZookeeper []*DedicatedZookeeper `json:"dedicatedZookeeper,omitempty"`
	//+kubebuilder:validation:MinItems:=1
	//+kubebuilder:validation:MaxItems:=1
	DataCentres            []*KafkaDataCentre        `json:"dataCentres"`
	SchemaRegistry         []*SchemaRegistry         `json:"schemaRegistry,omitempty"`
	RestProxy              []*RestProxy              `json:"restProxy,omitempty"`
	KarapaceRestProxy      []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
	Kraft                  []*Kraft                  `json:"kraft,omitempty"`
	ResizeSettings         GenericResizeSettings     `json:"resizeSettings,omitempty" dcomparisonSkip:"true"`
}

type Kraft struct {
	ControllerNodeCount int `json:"controllerNodeCount"`
}

type KafkaDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	NodesNumber int    `json:"nodesNumber"`
	NodeSize    string `json:"nodeSize"`

	PrivateLink []*PrivateLink `json:"privateLink,omitempty"`
}

func (dc *KafkaDataCentre) Equals(o *KafkaDataCentre) bool {
	return dc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		slices.EqualsPtr(dc.PrivateLink, o.PrivateLink) &&
		dc.NodeSize == o.NodeSize &&
		dc.NodesNumber == o.NodesNumber
}

func (ksdc *KafkaDataCentre) FromInstAPI(instaModel *models.KafkaDataCentre) {
	ksdc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)
	ksdc.PrivateLinkFromInstAPI(instaModel.PrivateLink)

	ksdc.NodeSize = instaModel.NodeSize
	ksdc.NodesNumber = instaModel.NumberOfNodes
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	GenericStatus `json:",inline"`

	AvailableUsers References               `json:"availableUsers,omitempty"`
	DataCentres    []*KafkaDataCentreStatus `json:"dataCentres,omitempty"`
}

type KafkaDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	Nodes       []*Node             `json:"nodes,omitempty"`
	PrivateLink PrivateLinkStatuses `json:"privateLink,omitempty"`
}

func (s *KafkaDataCentreStatus) Equals(o *KafkaDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		nodesEqual(s.Nodes, o.Nodes) &&
		slices.EqualsPtr(s.PrivateLink, o.PrivateLink)
}

func (s *KafkaStatus) Equals(o *KafkaStatus) bool {
	return s.GenericStatus.Equals(&o.GenericStatus) &&
		s.DCsEqual(o)
}

func (s *KafkaStatus) DCsEqual(o *KafkaStatus) bool {
	if len(s.DataCentres) != len(o.DataCentres) {
		return false
	}

	sMap := map[string]*KafkaDataCentreStatus{}
	for _, dc := range s.DataCentres {
		sMap[dc.Name] = dc
	}

	for _, dc := range o.DataCentres {
		sDC, ok := sMap[dc.Name]
		if !ok {
			return false
		}

		if !sDC.Equals(dc) {
			return false
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
//+kubebuilder:printcolumn:name="Node count",type="string",JSONPath=".status.nodeCount"

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
	return k.Kind + "/" + client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *Kafka) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (k *KafkaSpec) ToInstAPI() *models.KafkaCluster {
	return &models.KafkaCluster{
		GenericClusterFields:      k.GenericClusterSpec.ToInstAPI(),
		SchemaRegistry:            k.schemaRegistryToInstAPI(),
		RestProxy:                 k.restProxyToInstAPI(),
		DefaultReplicationFactor:  k.ReplicationFactor,
		DefaultNumberOfPartitions: k.PartitionsNumber,
		AllowDeleteTopics:         k.AllowDeleteTopics,
		AutoCreateTopics:          k.AutoCreateTopics,
		ClientToClusterEncryption: k.ClientToClusterEncryption,
		DedicatedZookeeper:        k.dedicatedZookeeperToInstAPI(),
		KafkaVersion:              k.Version,
		DataCentres:               k.dcToInstAPI(),
		ClientBrokerAuthWithMtls:  k.ClientBrokerAuthWithMTLS,
		BundledUseOnly:            k.BundledUseOnly,
		Kraft:                     k.kraftToInstAPI(),
		KarapaceRestProxy:         k.karapaceRestProxyToInstAPI(),
		KarapaceSchemaRegistry:    k.karapaceSchemaRegistryToInstAPI(),
		ResizeSettings:            k.ResizeSettings.ToInstAPI(),
		PCIComplianceMode:         k.PCICompliance,
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
			GenericDataCentreFields: crdDC.GenericDataCentreSpec.ToInstAPI(),
			PrivateLink:             crdDC.privateLinkToInstAPI(),
			NumberOfNodes:           crdDC.NodesNumber,
			NodeSize:                crdDC.NodeSize,
		})
	}
	return
}

func (k *KafkaSpec) ToInstAPIUpdate() *models.KafkaInstAPIUpdateRequest {
	return &models.KafkaInstAPIUpdateRequest{
		DataCentre:         k.dcToInstAPI(),
		DedicatedZookeeper: k.dedicatedZookeeperToInstAPIUpdate(),
		ResizeSettings:     k.ResizeSettings.ToInstAPI(),
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

func (k *Kafka) FromInstAPI(instaModel *models.KafkaCluster) {
	k.Spec.FromInstAPI(instaModel)
	k.Status.FromInstAPI(instaModel)
}

func (ks *KafkaSpec) FromInstAPI(instaModel *models.KafkaCluster) {
	ks.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields, instaModel.KafkaVersion)
	ks.ResizeSettings.FromInstAPI(instaModel.ResizeSettings)

	ks.ReplicationFactor = instaModel.DefaultReplicationFactor
	ks.PartitionsNumber = instaModel.DefaultNumberOfPartitions
	ks.AllowDeleteTopics = instaModel.AllowDeleteTopics
	ks.AutoCreateTopics = instaModel.AutoCreateTopics
	ks.ClientToClusterEncryption = instaModel.ClientToClusterEncryption
	ks.ClientBrokerAuthWithMTLS = instaModel.ClientBrokerAuthWithMtls
	ks.BundledUseOnly = instaModel.BundledUseOnly
	ks.PCICompliance = instaModel.PCIComplianceMode

	ks.DCsFromInstAPI(instaModel.DataCentres)
	ks.kraftFromInstAPI(instaModel.Kraft)
	ks.restProxyFromInstAPI(instaModel.RestProxy)
	ks.schemaRegistryFromInstAPI(instaModel.SchemaRegistry)
	ks.karapaceRestProxyFromInstAPI(instaModel.KarapaceRestProxy)
	ks.dedicatedZookeeperFromInstAPI(instaModel.DedicatedZookeeper)
	ks.karapaceSchemaRegistryFromInstAPI(instaModel.KarapaceSchemaRegistry)
}

func (ks *KafkaStatus) FromInstAPI(instaModel *models.KafkaCluster) {
	ks.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	ks.DCsFromInstAPI(instaModel.DataCentres)
	ks.GetNodeCount()
}

func (ks *KafkaStatus) GetNodeCount() {
	var total, running int
	for _, dc := range ks.DataCentres {
		for _, node := range dc.Nodes {
			total++
			if node.Status == models.RunningStatus {
				running++
			}
		}
	}
	ks.NodeCount = fmt.Sprintf("%v/%v", running, total)
}

func (ks *KafkaSpec) DCsFromInstAPI(instaModels []*models.KafkaDataCentre) {
	ks.DataCentres = make([]*KafkaDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := KafkaDataCentre{}
		dc.FromInstAPI(instaModel)
		ks.DataCentres[i] = &dc
	}
}

func (dc *KafkaDataCentre) PrivateLinkFromInstAPI(instaModels []*models.PrivateLink) {
	dc.PrivateLink = make([]*PrivateLink, len(instaModels))
	for i, instaModel := range instaModels {
		dc.PrivateLink[i] = &PrivateLink{
			AdvertisedHostname: instaModel.AdvertisedHostname,
		}
	}
}

func (ks *KafkaSpec) schemaRegistryFromInstAPI(instaModels []*models.SchemaRegistry) {
	ks.SchemaRegistry = make([]*SchemaRegistry, len(instaModels))
	for i, instaModel := range instaModels {
		ks.SchemaRegistry[i] = &SchemaRegistry{
			Version: instaModel.Version,
		}
	}
}

func (ks *KafkaSpec) restProxyFromInstAPI(instaModels []*models.RestProxy) {
	ks.RestProxy = make([]*RestProxy, len(instaModels))
	for i, instaModel := range instaModels {
		ks.RestProxy[i] = &RestProxy{
			IntegrateRestProxyWithSchemaRegistry: instaModel.IntegrateRestProxyWithSchemaRegistry,
			UseLocalSchemaRegistry:               instaModel.UseLocalSchemaRegistry,
			SchemaRegistryServerURL:              instaModel.SchemaRegistryServerURL,
			SchemaRegistryUsername:               instaModel.SchemaRegistryUsername,
			SchemaRegistryPassword:               instaModel.SchemaRegistryPassword,
			Version:                              instaModel.Version,
		}
	}
}

func (ks *KafkaSpec) dedicatedZookeeperFromInstAPI(instaModels []*models.DedicatedZookeeper) {
	ks.DedicatedZookeeper = make([]*DedicatedZookeeper, len(instaModels))
	for i, instaModel := range instaModels {
		ks.DedicatedZookeeper[i] = &DedicatedZookeeper{
			NodeSize:    instaModel.ZookeeperNodeSize,
			NodesNumber: instaModel.ZookeeperNodeCount,
		}
	}
}

func (ks *KafkaSpec) karapaceRestProxyFromInstAPI(instaModels []*models.KarapaceRestProxy) {
	ks.KarapaceRestProxy = make([]*KarapaceRestProxy, len(instaModels))
	for i, instaModel := range instaModels {
		ks.KarapaceRestProxy[i] = &KarapaceRestProxy{
			IntegrateRestProxyWithSchemaRegistry: instaModel.IntegrateRestProxyWithSchemaRegistry,
			Version:                              instaModel.Version,
		}
	}
}

func (ks *KafkaSpec) kraftFromInstAPI(instaModels []*models.Kraft) {
	ks.Kraft = make([]*Kraft, len(instaModels))
	for i, instaModel := range instaModels {
		ks.Kraft[i] = &Kraft{
			ControllerNodeCount: instaModel.ControllerNodeCount,
		}
	}
}

func (ks *KafkaSpec) karapaceSchemaRegistryFromInstAPI(instaModels []*models.KarapaceSchemaRegistry) {
	ks.KarapaceSchemaRegistry = make([]*KarapaceSchemaRegistry, len(instaModels))
	for i, instaModel := range instaModels {
		ks.KarapaceSchemaRegistry[i] = &KarapaceSchemaRegistry{
			Version: instaModel.Version,
		}
	}
}

func (ks *KafkaStatus) DCsFromInstAPI(instaModels []*models.KafkaDataCentre) {
	ks.DataCentres = make([]*KafkaDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := KafkaDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		ks.DataCentres[i] = &dc
	}
}

func (s *KafkaDataCentreStatus) FromInstAPI(instaModel *models.KafkaDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	s.PrivateLink.FromInstAPI(instaModel.PrivateLink)
	s.Nodes = nodesFromInstAPI(instaModel.Nodes)
}

func (a *KafkaSpec) IsEqual(b KafkaSpec) bool {
	return a.GenericClusterSpec.Equals(&b.GenericClusterSpec) &&
		a.ReplicationFactor == b.ReplicationFactor &&
		a.PartitionsNumber == b.PartitionsNumber &&
		a.AllowDeleteTopics == b.AllowDeleteTopics &&
		a.AutoCreateTopics == b.AutoCreateTopics &&
		a.ClientToClusterEncryption == b.ClientToClusterEncryption &&
		a.ClientBrokerAuthWithMTLS == b.ClientBrokerAuthWithMTLS &&
		a.BundledUseOnly == b.BundledUseOnly &&
		slices.EqualsPtr(a.SchemaRegistry, b.SchemaRegistry) &&
		slices.EqualsPtr(a.RestProxy, b.RestProxy) &&
		slices.EqualsPtr(a.KarapaceRestProxy, b.KarapaceRestProxy) &&
		slices.EqualsPtr(a.Kraft, b.Kraft) &&
		slices.EqualsPtr(a.KarapaceSchemaRegistry, b.KarapaceSchemaRegistry) &&
		slices.EqualsPtr(a.DedicatedZookeeper, b.DedicatedZookeeper) &&
		a.areDCsEqual(b.DataCentres)
}

func (c *Kafka) GetSpec() KafkaSpec { return c.Spec }

func (c *Kafka) IsSpecEqual(spec KafkaSpec) bool {
	return c.Spec.IsEqual(spec)
}

func (rs *KafkaSpec) areDCsEqual(b []*KafkaDataCentre) bool {
	if len(rs.DataCentres) != len(b) {
		return false
	}

	sMap := map[string]*KafkaDataCentre{}
	for _, dc := range rs.DataCentres {
		sMap[dc.Name] = dc
	}

	for _, dc := range b {
		sDC, ok := sMap[dc.Name]
		if !ok {
			return false
		}

		if !sDC.Equals(dc) {
			return false
		}
	}

	return true
}

func (k *Kafka) GetUserRefs() References {
	return k.Spec.UserRefs
}

func (k *Kafka) SetUserRefs(refs References) {
	k.Spec.UserRefs = refs
}

func (k *Kafka) GetAvailableUsers() References {
	return k.Status.AvailableUsers
}

func (k *Kafka) SetAvailableUsers(users References) {
	k.Status.AvailableUsers = users
}

func (k *Kafka) GetClusterID() string {
	return k.Status.ID
}

func (k *Kafka) GetDataCentreID(cdcName string) string {
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

func (k *Kafka) SetClusterID(id string) {
	k.Status.ID = id
}

func (k *Kafka) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !k.Spec.PrivateNetwork {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.KafkaClient,
				Port: models.Port9092,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port9092,
				},
			},
			{
				Name: models.KafkaControlPlane,
				Port: models.Port9093,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port9093,
				},
			},
		}
		if k.Spec.ClientToClusterEncryption {
			sslPort := k8scorev1.ServicePort{
				Name: models.KafkaBroker,
				Port: models.Port9094,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port9094,
				},
			}
			exposePorts = append(exposePorts, sslPort)
		}
	}
	return exposePorts
}

func (k *Kafka) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.KafkaClient,
			Port: models.Port9092,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port9092,
			},
		},
		{
			Name: models.KafkaControlPlane,
			Port: models.Port9093,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port9093,
			},
		},
	}
	if k.Spec.ClientToClusterEncryption {
		kafkaBrokerPort := k8scorev1.ServicePort{
			Name: models.KafkaBroker,
			Port: models.Port9094,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port9094,
			},
		}
		headlessPorts = append(headlessPorts, kafkaBrokerPort)
	}
	return headlessPorts
}
