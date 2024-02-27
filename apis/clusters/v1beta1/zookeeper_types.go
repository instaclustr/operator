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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type ZookeeperDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	NodesNumber              int      `json:"nodesNumber"`
	NodeSize                 string   `json:"nodeSize"`
	ClientToServerEncryption bool     `json:"clientToServerEncryption"`
	EnforceAuthEnabled       bool     `json:"enforceAuthEnabled,omitempty"`
	EnforceAuthSchemes       []string `json:"enforceAuthSchemes,omitempty"`
}

// ZookeeperSpec defines the desired state of Zookeeper
type ZookeeperSpec struct {
	GenericClusterSpec `json:",inline"`
	DataCentres        []*ZookeeperDataCentre `json:"dataCentres"`
}

// ZookeeperStatus defines the observed state of Zookeeper
type ZookeeperStatus struct {
	GenericStatus        `json:",inline"`
	DataCentres          []*ZookeeperDataCentreStatus `json:"dataCentres,omitempty"`
	DefaultUserSecretRef *Reference                   `json:"defaultUserSecretRef,omitempty"`
}

type ZookeeperDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`
	Nodes                   []*Node `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Node count",type="string",JSONPath=".status.nodeCount"

// Zookeeper is the Schema for the zookeepers API
type Zookeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperSpec   `json:"spec,omitempty"`
	Status ZookeeperStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ZookeeperList contains a list of Zookeeper
type ZookeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zookeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zookeeper{}, &ZookeeperList{})
}

func (z *Zookeeper) GetJobID(jobName string) string {
	return z.Kind + "/" + client.ObjectKeyFromObject(z).String() + "/" + jobName
}

func (z *Zookeeper) NewPatch() client.Patch {
	old := z.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (z *Zookeeper) FromInstAPI(instaModel *models.ZookeeperCluster) {
	z.Spec.FromInstAPI(instaModel)
	z.Status.FromInstAPI(instaModel)
}

func (z *Zookeeper) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return z.Status.DataCentres[0].ID
	}
	for _, cdc := range z.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}

func (z *Zookeeper) GetClusterID() string {
	return z.Status.ID
}

func (zs *ZookeeperSpec) FromInstAPI(instaModel *models.ZookeeperCluster) {
	zs.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields, instaModel.ZookeeperVersion)
	zs.DCsFromInstAPI(instaModel.DataCentres)
}

func (zs *ZookeeperStatus) FromInstAPI(instaModel *models.ZookeeperCluster) {
	zs.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	zs.DCsFromInstAPI(instaModel.DataCentres)
	zs.NodeCount = zs.GetNodeCount(instaModel.DataCentres)
}

func (zs *ZookeeperStatus) GetNodeCount(dcs []*models.ZookeeperDataCentre) string {
	var total, running int
	for _, dc := range dcs {
		for _, node := range dc.Nodes {
			total++
			if node.Status == models.RunningStatus {
				running++
			}
		}
	}
	return fmt.Sprintf("%v/%v", running, total)
}

func (zs *ZookeeperSpec) DCsFromInstAPI(instaModels []*models.ZookeeperDataCentre) {
	dcs := make([]*ZookeeperDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &ZookeeperDataCentre{}
		dc.FromInstAPI(instaModel)
		dcs[i] = dc
	}
	zs.DataCentres = dcs
}

func (zs *ZookeeperStatus) DCsFromInstAPI(instaModels []*models.ZookeeperDataCentre) {
	dcs := make([]*ZookeeperDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &ZookeeperDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		dcs[i] = dc
	}
	zs.DataCentres = dcs
}

func (zs *ZookeeperSpec) ToInstAPI() *models.ZookeeperCluster {
	return &models.ZookeeperCluster{
		GenericClusterFields: zs.GenericClusterSpec.ToInstAPI(),
		ZookeeperVersion:     zs.Version,
		DataCentres:          zs.DCsToInstAPI(),
	}
}

func (zs *ZookeeperSpec) DCsToInstAPI() (dcs []*models.ZookeeperDataCentre) {
	for _, k8sDC := range zs.DataCentres {
		dcs = append(dcs, k8sDC.ToInstAPI())
	}
	return dcs
}

func (zdc *ZookeeperDataCentre) ToInstAPI() *models.ZookeeperDataCentre {
	return &models.ZookeeperDataCentre{
		GenericDataCentreFields:  zdc.GenericDataCentreSpec.ToInstAPI(),
		ClientToServerEncryption: zdc.ClientToServerEncryption,
		EnforceAuthSchemes:       zdc.EnforceAuthSchemes,
		EnforceAuthEnabled:       zdc.EnforceAuthEnabled,
		NumberOfNodes:            zdc.NodesNumber,
		NodeSize:                 zdc.NodeSize,
	}
}

func (z *Zookeeper) GetSpec() ZookeeperSpec { return z.Spec }

func (z *Zookeeper) IsSpecEqual(spec ZookeeperSpec) bool { return z.Spec.IsEqual(spec) }

func (a *ZookeeperSpec) IsEqual(b ZookeeperSpec) bool {
	return a.GenericClusterSpec.Equals(&b.GenericClusterSpec) &&
		a.DCsEqual(b.DataCentres)
}

func (rs *ZookeeperSpec) DCsEqual(o []*ZookeeperDataCentre) bool {
	if len(rs.DataCentres) != len(o) {
		return false
	}

	m := map[string]*ZookeeperDataCentre{}
	for _, dc := range rs.DataCentres {
		m[dc.Name] = dc
	}

	for _, iDC := range o {
		dc, ok := m[iDC.Name]
		if !ok || !dc.Equals(iDC) {
			return false
		}
	}

	return true
}

func (zdc *ZookeeperDataCentre) FromInstAPI(instaModel *models.ZookeeperDataCentre) {
	zdc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)
	zdc.NodeSize = instaModel.NodeSize
	zdc.NodesNumber = instaModel.NumberOfNodes
	zdc.ClientToServerEncryption = instaModel.ClientToServerEncryption
	zdc.EnforceAuthEnabled = instaModel.EnforceAuthEnabled
	zdc.EnforceAuthSchemes = instaModel.EnforceAuthSchemes
}

func (s *ZookeeperDataCentreStatus) FromInstAPI(instaModel *models.ZookeeperDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	s.Nodes = nodesFromInstAPI(instaModel.Nodes)
}

func (zdc *ZookeeperDataCentre) Equals(o *ZookeeperDataCentre) bool {
	return zdc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		zdc.NodesNumber == o.NodesNumber &&
		zdc.NodeSize == o.NodeSize &&
		zdc.ClientToServerEncryption == o.ClientToServerEncryption &&
		zdc.EnforceAuthEnabled == o.EnforceAuthEnabled &&
		slices.Equals(zdc.EnforceAuthSchemes, o.EnforceAuthSchemes)
}

func (zs *ZookeeperStatus) Equals(o *ZookeeperStatus) bool {
	return zs.GenericStatus.Equals(&o.GenericStatus) &&
		zs.DCsEquals(o.DataCentres)
}

func (zs *ZookeeperStatus) DCsEquals(o []*ZookeeperDataCentreStatus) bool {
	if len(zs.DataCentres) != len(o) {
		return false
	}

	m := map[string]*ZookeeperDataCentreStatus{}
	for _, dc := range zs.DataCentres {
		m[dc.Name] = dc
	}

	for _, iDC := range o {
		dc, ok := m[iDC.Name]
		if !ok || !dc.Equals(iDC) {
			return false
		}
	}

	return true
}

func (s *ZookeeperDataCentreStatus) Equals(o *ZookeeperDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		nodesEqual(s.Nodes, o.Nodes)
}
