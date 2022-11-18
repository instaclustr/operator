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

	"github.com/instaclustr/operator/pkg/models"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ZookeeperDataCentre struct {
	Name                     string                   `json:"name,omitempty"`
	Region                   string                   `json:"region"`
	CloudProvider            string                   `json:"cloudProvider"`
	ProviderAccountName      string                   `json:"providerAccountName,omitempty"`
	CloudProviderSettings    []*CloudProviderSettings `json:"cloudProviderSettings,omitempty"`
	Network                  string                   `json:"network"`
	NodeSize                 string                   `json:"nodeSize"`
	NodesNumber              int32                    `json:"nodesNumber"`
	Tags                     map[string]string        `json:"tags,omitempty"`
	ClientToServerEncryption bool                     `json:"clientToServerEncryption"`
}

// ZookeeperSpec defines the desired state of Zookeeper
type ZookeeperSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name                  string                 `json:"name"`
	Version               string                 `json:"version"`
	SLATier               string                 `json:"slaTier"`
	TwoFactorDelete       []*TwoFactorDelete     `json:"twoFactorDelete,omitempty"`
	PrivateNetworkCluster bool                   `json:"privateNetworkCluster,omitempty"`
	DataCentres           []*ZookeeperDataCentre `json:"dataCentres"`
}

// ZookeeperStatus defines the observed state of Zookeeper
type ZookeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

func (k *Zookeeper) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *Zookeeper) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}
