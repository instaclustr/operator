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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type ZookeeperDataCentre struct {
	DataCentre               `json:",inline"`
	ClientToServerEncryption bool     `json:"clientToServerEncryption"`
	EnforceAuthSchemes       []string `json:"enforceAuthSchemes,omitempty"`
	EnforceAuthEnabled       bool     `json:"enforceAuthEnabled,omitempty"`
}

// ZookeeperSpec defines the desired state of Zookeeper
type ZookeeperSpec struct {
	Cluster     `json:",inline"`
	DataCentres []*ZookeeperDataCentre `json:"dataCentres"`
}

// ZookeeperStatus defines the observed state of Zookeeper
type ZookeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

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

func (z *Zookeeper) FromInstAPI(iData []byte) (*Zookeeper, error) {
	iZook := &models.ZookeeperCluster{}
	err := json.Unmarshal(iData, iZook)
	if err != nil {
		return nil, err
	}

	return &Zookeeper{
		TypeMeta:   z.TypeMeta,
		ObjectMeta: z.ObjectMeta,
		Spec:       z.Spec.FromInstAPI(iZook),
		Status:     z.Status.FromInstAPI(iZook),
	}, nil
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

func (zs *ZookeeperSpec) FromInstAPI(iZook *models.ZookeeperCluster) ZookeeperSpec {
	return ZookeeperSpec{
		Cluster: Cluster{
			Name:                  iZook.Name,
			Version:               iZook.ZookeeperVersion,
			Description:           iZook.Description,
			PrivateNetworkCluster: iZook.PrivateNetworkCluster,
			SLATier:               iZook.SLATier,
			TwoFactorDelete:       zs.Cluster.TwoFactorDeleteFromInstAPI(iZook.TwoFactorDelete),
		},
		DataCentres: zs.DCsFromInstAPI(iZook.DataCentres),
	}
}

func (zs *ZookeeperStatus) FromInstAPI(iZook *models.ZookeeperCluster) ZookeeperStatus {
	return ZookeeperStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iZook.ID,
			State:                         iZook.Status,
			DataCentres:                   zs.DCsFromInstAPI(iZook.DataCentres),
			CurrentClusterOperationStatus: iZook.CurrentClusterOperationStatus,
			MaintenanceEvents:             zs.MaintenanceEvents,
		},
	}
}

func (zs *ZookeeperSpec) DCsFromInstAPI(iDCs []*models.ZookeeperDataCentre) (dcs []*ZookeeperDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &ZookeeperDataCentre{
			DataCentre:               zs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ClientToServerEncryption: iDC.ClientToServerEncryption,
			EnforceAuthSchemes:       iDC.EnforceAuthSchemes,
			EnforceAuthEnabled:       iDC.EnforceAuthEnabled,
		})
	}
	return
}

func (zs *ZookeeperStatus) DCsFromInstAPI(iDCs []*models.ZookeeperDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, zs.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func (zs *ZookeeperSpec) ToInstAPI() *models.ZookeeperCluster {
	return &models.ZookeeperCluster{
		Name:                  zs.Name,
		ZookeeperVersion:      zs.Version,
		PrivateNetworkCluster: zs.PrivateNetworkCluster,
		SLATier:               zs.SLATier,
		TwoFactorDelete:       zs.Cluster.TwoFactorDeletesToInstAPI(),
		DataCentres:           zs.DCsToInstAPI(),
		Description:           zs.Description,
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
		DataCentre:               zdc.DataCentre.ToInstAPI(),
		ClientToServerEncryption: zdc.ClientToServerEncryption,
		EnforceAuthSchemes:       zdc.EnforceAuthSchemes,
		EnforceAuthEnabled:       zdc.EnforceAuthEnabled,
	}
}

func (z *Zookeeper) GetSpec() ZookeeperSpec { return z.Spec }

func (z *Zookeeper) IsSpecEqual(spec ZookeeperSpec) bool { return z.Spec.IsEqual(spec) }

func (a *ZookeeperSpec) IsEqual(b ZookeeperSpec) bool {
	return a.Cluster.IsEqual(b.Cluster) &&
		a.areDCsEqual(b.DataCentres)
}

func (rs *ZookeeperSpec) areDCsEqual(b []*ZookeeperDataCentre) bool {
	a := rs.DataCentres
	if len(a) != len(b) {
		return false
	}

	for i := range b {
		if a[i].Name != b[i].Name {
			continue
		}

		if !a[i].DataCentre.IsEqual(b[i].DataCentre) ||
			a[i].ClientToServerEncryption != b[i].ClientToServerEncryption {
			return false
		}
	}

	return true
}

func (a *Zookeeper) NewDefaultUserSecret(username, password string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.SecretKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, a.Name),
			Namespace: a.Namespace,
			Labels: map[string]string{
				models.ControlledByLabel:  a.Name,
				models.DefaultSecretLabel: "true",
			},
		},
		StringData: map[string]string{
			models.Username: username,
			models.Password: password,
		},
	}
}
