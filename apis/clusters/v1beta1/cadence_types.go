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
	"context"
	"encoding/json"
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type CadenceDataCentre struct {
	DataCentre       `json:",inline"`
	ClientEncryption bool           `json:"clientEncryption,omitempty"`
	PrivateLink      []*PrivateLink `json:"privateLink,omitempty"`
}

type BundledCassandraSpec struct {
	NodeSize                       string `json:"nodeSize"`
	NodesNumber                    int    `json:"nodesNumber"`
	ReplicationFactor              int    `json:"replicationFactor"`
	Network                        string `json:"network"`
	PrivateIPBroadcastForDiscovery bool   `json:"privateIPBroadcastForDiscovery"`
	PasswordAndUserAuth            bool   `json:"passwordAndUserAuth"`
}

type BundledKafkaSpec struct {
	NodeSize          string `json:"nodeSize"`
	NodesNumber       int    `json:"nodesNumber"`
	Network           string `json:"network"`
	ReplicationFactor int    `json:"replicationFactor"`
	PartitionsNumber  int    `json:"partitionsNumber"`
}

type BundledOpenSearchSpec struct {
	NodeSize          string `json:"nodeSize"`
	ReplicationFactor int    `json:"replicationFactor"`
	Network           string `json:"network"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	Cluster        `json:",inline"`
	OnPremisesSpec *OnPremisesSpec `json:"onPremisesSpec,omitempty"`
	//+kubebuilder:validation:MinItems:=1
	//+kubebuilder:validation:MaxItems:=1
	DataCentres          []*CadenceDataCentre    `json:"dataCentres"`
	UseCadenceWebAuth    bool                    `json:"useCadenceWebAuth"`
	AWSArchival          []*AWSArchival          `json:"awsArchival,omitempty"`
	StandardProvisioning []*StandardProvisioning `json:"standardProvisioning,omitempty"`
	SharedProvisioning   []*SharedProvisioning   `json:"sharedProvisioning,omitempty"`
	PackagedProvisioning []*PackagedProvisioning `json:"packagedProvisioning,omitempty"`
	TargetPrimaryCadence []*TargetCadence        `json:"targetPrimaryCadence,omitempty"`
	ResizeSettings       []*ResizeSettings       `json:"resizeSettings,omitempty"`
	UseHTTPAPI           bool                    `json:"useHttpApi,omitempty"`
}

type AWSArchival struct {
	ArchivalS3URI            string `json:"archivalS3Uri"`
	ArchivalS3Region         string `json:"archivalS3Region"`
	AccessKeySecretNamespace string `json:"awsAccessKeySecretNamespace"`
	AccessKeySecretName      string `json:"awsAccessKeySecretName"`
}

type PackagedProvisioning struct {
	UseAdvancedVisibility bool                   `json:"useAdvancedVisibility"`
	BundledKafkaSpec      *BundledKafkaSpec      `json:"bundledKafkaSpec,omitempty"`
	BundledOpenSearchSpec *BundledOpenSearchSpec `json:"bundledOpenSearchSpec,omitempty"`
	BundledCassandraSpec  *BundledCassandraSpec  `json:"bundledCassandraSpec"`
}

type SharedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`
}

type StandardProvisioning struct {
	AdvancedVisibility []*AdvancedVisibility `json:"advancedVisibility,omitempty"`
	TargetCassandra    *TargetCassandra      `json:"targetCassandra"`
}

type TargetCassandra struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type TargetKafka struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type TargetOpenSearch struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type TargetCadence struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type AdvancedVisibility struct {
	TargetKafka      *TargetKafka      `json:"targetKafka"`
	TargetOpenSearch *TargetOpenSearch `json:"targetOpenSearch"`
}

// CadenceStatus defines the observed state of Cadence
type CadenceStatus struct {
	ClusterStatus          `json:",inline"`
	TargetSecondaryCadence []*TargetCadence `json:"targetSecondaryCadence,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// Cadence is the Schema for the cadences API
type Cadence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CadenceSpec   `json:"spec,omitempty"`
	Status CadenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CadenceList contains a list of Cadence
type CadenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cadence `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cadence{}, &CadenceList{})
}

func (c *Cadence) GetJobID(jobName string) string {
	return c.Kind + "/" + client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cadence) NewPatch() client.Patch {
	old := c.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (c *Cadence) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return c.Status.DataCentres[0].ID
	}
	for _, cdc := range c.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}
func (c *Cadence) GetClusterID() string {
	return c.Status.ID
}

func (cs *CadenceSpec) ToInstAPI(ctx context.Context, k8sClient client.Client) (*models.CadenceCluster, error) {
	awsArchival, err := cs.ArchivalToInstAPI(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("cannot convert archival data to APIv2 format: %v", err)
	}

	sharedProvisioning := cs.SharedProvisioningToInstAPI()

	var standardProvisioning []*models.CadenceStandardProvisioning
	if len(cs.StandardProvisioning) != 0 || len(cs.PackagedProvisioning) != 0 {
		standardProvisioning = cs.StandardProvisioningToInstAPI()
	}

	return &models.CadenceCluster{
		CadenceVersion:        cs.Version,
		DataCentres:           cs.DCsToInstAPI(),
		Name:                  cs.Name,
		PCIComplianceMode:     cs.PCICompliance,
		TwoFactorDelete:       cs.TwoFactorDeletesToInstAPI(),
		UseCadenceWebAuth:     cs.UseCadenceWebAuth,
		PrivateNetworkCluster: cs.PrivateNetworkCluster,
		SLATier:               cs.SLATier,
		AWSArchival:           awsArchival,
		Description:           cs.Description,
		SharedProvisioning:    sharedProvisioning,
		StandardProvisioning:  standardProvisioning,
		TargetPrimaryCadence:  cs.TargetCadenceToInstAPI(),
		ResizeSettings:        resizeSettingsToInstAPI(cs.ResizeSettings),
		UseHTTPAPI:            cs.UseHTTPAPI,
	}, nil
}

func (cs *CadenceSpec) StandardProvisioningToInstAPI() []*models.CadenceStandardProvisioning {
	var cadenceStandardProvisioning []*models.CadenceStandardProvisioning

	for _, standardProvisioning := range cs.StandardProvisioning {
		targetCassandra := standardProvisioning.TargetCassandra

		stdProvisioning := &models.CadenceStandardProvisioning{
			TargetCassandra: &models.TargetCassandra{
				DependencyCDCID:   targetCassandra.DependencyCDCID,
				DependencyVPCType: targetCassandra.DependencyVPCType,
			},
		}

		for _, advancedVisibility := range standardProvisioning.AdvancedVisibility {
			targetKafka := advancedVisibility.TargetKafka
			targetOpenSearch := advancedVisibility.TargetOpenSearch

			stdProvisioning.AdvancedVisibility = []*models.AdvancedVisibility{
				{
					TargetKafka: &models.TargetKafka{
						DependencyCDCID:   targetKafka.DependencyCDCID,
						DependencyVPCType: targetKafka.DependencyVPCType,
					},
					TargetOpenSearch: &models.TargetOpenSearch{
						DependencyCDCID:   targetOpenSearch.DependencyCDCID,
						DependencyVPCType: targetOpenSearch.DependencyVPCType,
					},
				},
			}
		}

		cadenceStandardProvisioning = append(cadenceStandardProvisioning, stdProvisioning)
	}

	return cadenceStandardProvisioning
}

func (cs *CadenceSpec) SharedProvisioningToInstAPI() []*models.CadenceSharedProvisioning {
	var sharedProvisioning []*models.CadenceSharedProvisioning

	for _, sp := range cs.SharedProvisioning {
		sharedProvisioning = append(sharedProvisioning, &models.CadenceSharedProvisioning{
			UseAdvancedVisibility: sp.UseAdvancedVisibility,
		})
	}

	return sharedProvisioning
}

func (cs *CadenceSpec) ArchivalToInstAPI(ctx context.Context, k8sClient client.Client) ([]*models.AWSArchival, error) {
	awsArchival := []*models.AWSArchival{}

	for _, aws := range cs.AWSArchival {
		AWSAccessKeyID, AWSSecretAccessKey, err := getSecret(ctx, k8sClient, aws)
		if err != nil {
			return nil, fmt.Errorf("cannot get aws creds secret: %v", err)
		}

		awsArchival = append(awsArchival, &models.AWSArchival{
			ArchivalS3Region:   aws.ArchivalS3Region,
			ArchivalS3URI:      aws.ArchivalS3URI,
			AWSAccessKeyID:     AWSAccessKeyID,
			AWSSecretAccessKey: AWSSecretAccessKey,
		})
	}

	return awsArchival, nil
}

func getSecret(ctx context.Context, k8sClient client.Client, aws *AWSArchival) (string, string, error) {
	var err error
	awsCredsSecret := &v1.Secret{}
	awsSecretNamespacedName := types.NamespacedName{
		Name:      aws.AccessKeySecretName,
		Namespace: aws.AccessKeySecretNamespace,
	}
	err = k8sClient.Get(ctx, awsSecretNamespacedName, awsCredsSecret)
	if err != nil {
		return "", "", err
	}

	var AWSAccessKeyID string
	var AWSSecretAccessKey string
	keyID := awsCredsSecret.Data[models.AWSAccessKeyID]
	secretAccessKey := awsCredsSecret.Data[models.AWSSecretAccessKey]
	AWSAccessKeyID = string(keyID[:len(keyID)-1])
	AWSSecretAccessKey = string(secretAccessKey[:len(secretAccessKey)-1])

	return AWSAccessKeyID, AWSSecretAccessKey, err
}

func (cdc *CadenceDataCentre) ToInstAPI() *models.CadenceDataCentre {
	cloudProviderSettings := cdc.CloudProviderSettingsToInstAPI()
	return &models.CadenceDataCentre{
		PrivateLink:               cdc.privateLinkToInstAPI(),
		ClientToClusterEncryption: cdc.ClientEncryption,
		DataCentre: models.DataCentre{
			CloudProvider:       cdc.CloudProvider,
			Name:                cdc.Name,
			Network:             cdc.Network,
			NodeSize:            cdc.NodeSize,
			NumberOfNodes:       cdc.NodesNumber,
			Region:              cdc.Region,
			ProviderAccountName: cdc.ProviderAccountName,
			AWSSettings:         cloudProviderSettings.AWSSettings,
			GCPSettings:         cloudProviderSettings.GCPSettings,
			AzureSettings:       cloudProviderSettings.AzureSettings,
			Tags:                cdc.TagsToInstAPI(),
		},
	}
}

func (cd *CadenceDataCentre) privateLinkToInstAPI() (iPrivateLink []*models.PrivateLink) {
	for _, link := range cd.PrivateLink {
		iPrivateLink = append(iPrivateLink, &models.PrivateLink{
			AdvertisedHostname: link.AdvertisedHostname,
		})
	}

	return
}

func (cs *CadenceSpec) TargetCadenceToInstAPI() []*models.TargetCadence {
	var targets []*models.TargetCadence
	for _, target := range cs.TargetPrimaryCadence {
		targets = append(targets, &models.TargetCadence{
			DependencyCDCID:   target.DependencyCDCID,
			DependencyVPCType: target.DependencyVPCType,
		})
	}

	return targets
}

func (cs *CadenceSpec) DCsToInstAPI() (iDCs []*models.CadenceDataCentre) {
	for _, dc := range cs.DataCentres {
		iDCs = append(iDCs, dc.ToInstAPI())
	}

	return
}

func (c *Cadence) FromInstAPI(iData []byte) (*Cadence, error) {
	iCad := &models.CadenceCluster{}
	err := json.Unmarshal(iData, iCad)
	if err != nil {
		return nil, err
	}

	return &Cadence{
		TypeMeta:   c.TypeMeta,
		ObjectMeta: c.ObjectMeta,
		Spec:       c.Spec.FromInstAPI(iCad),
		Status:     c.Status.FromInstAPI(iCad),
	}, nil
}

func (cs *CadenceSpec) FromInstAPI(iCad *models.CadenceCluster) (spec CadenceSpec) {
	spec.DataCentres = cs.DCsFromInstAPI(iCad.DataCentres)
	spec.ResizeSettings = resizeSettingsFromInstAPI(iCad.ResizeSettings)
	spec.TwoFactorDelete = cs.Cluster.TwoFactorDeleteFromInstAPI(iCad.TwoFactorDelete)
	spec.Description = iCad.Description

	return
}

func (cs *CadenceSpec) DCsFromInstAPI(iDCs []*models.CadenceDataCentre) (dcs []*CadenceDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &CadenceDataCentre{
			DataCentre:       cs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ClientEncryption: iDC.ClientToClusterEncryption,
			PrivateLink:      cs.PrivateLinkFromInstAPI(iDC.PrivateLink),
		})
	}

	return
}

func (cs *CadenceSpec) PrivateLinkFromInstAPI(iPLs []*models.PrivateLink) (pls []*PrivateLink) {
	for _, iPL := range iPLs {
		pls = append(pls, &PrivateLink{
			AdvertisedHostname: iPL.AdvertisedHostname,
		})
	}
	return
}

func (cs *CadenceSpec) IsEqual(spec CadenceSpec) bool {
	return cs.AreDCsEqual(spec.DataCentres)
}

func (cs *CadenceSpec) AreDCsEqual(dcs []*CadenceDataCentre) bool {
	if len(cs.DataCentres) != len(dcs) {
		return false
	}

	for i, iDC := range dcs {
		dataCentre := cs.DataCentres[i]

		if iDC.Name != dataCentre.Name {
			continue
		}

		if !dataCentre.IsEqual(iDC.DataCentre) ||
			iDC.ClientEncryption != dataCentre.ClientEncryption ||
			!arePrivateLinksEqual(dataCentre.PrivateLink, iDC.PrivateLink) {
			return false
		}
	}

	return true
}

func (cs *CadenceStatus) FromInstAPI(iCad *models.CadenceCluster) CadenceStatus {
	return CadenceStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iCad.ID,
			State:                         iCad.Status,
			DataCentres:                   cs.DCsFromInstAPI(iCad.DataCentres),
			CurrentClusterOperationStatus: iCad.CurrentClusterOperationStatus,
			MaintenanceEvents:             cs.MaintenanceEvents,
		},
		TargetSecondaryCadence: cs.SecondaryTargetsFromInstAPI(iCad),
	}
}

func (cs *CadenceStatus) SecondaryTargetsFromInstAPI(iCad *models.CadenceCluster) []*TargetCadence {
	var targets []*TargetCadence
	for _, target := range iCad.TargetSecondaryCadence {
		targets = append(targets, &TargetCadence{
			DependencyCDCID:   target.DependencyCDCID,
			DependencyVPCType: target.DependencyVPCType,
		})
	}

	return targets
}

func (cs *CadenceStatus) DCsFromInstAPI(iDCs []*models.CadenceDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dc := cs.ClusterStatus.DCFromInstAPI(iDC.DataCentre)
		dc.PrivateLink = privateLinkStatusesFromInstAPI(iDC.PrivateLink)
		dcs = append(dcs, dc)
	}

	return
}

func (cs *CadenceSpec) NewDCsUpdate() models.CadenceClusterAPIUpdate {
	return models.CadenceClusterAPIUpdate{
		DataCentres: cs.DCsToInstAPI(),
	}
}

func (c *Cadence) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !c.Spec.PrivateNetworkCluster {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.CadenceTChannel,
				Port: models.Port7933,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7933,
				},
			},
			{
				Name: models.CadenceHHTPAPI,
				Port: models.Port8088,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port8088,
				},
			},
			{
				Name: models.CadenceWeb,
				Port: models.Port443,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port443,
				},
			},
		}
		if c.Spec.DataCentres[0].ClientEncryption {
			sslPort := k8scorev1.ServicePort{
				Name: models.CadenceGRPC,
				Port: models.Port7833,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port7833,
				},
			}
			exposePorts = append(exposePorts, sslPort)
		}
	}
	return exposePorts
}

func (c *Cadence) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.CadenceTChannel,
			Port: models.Port7933,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port7933,
			},
		},
	}
	if c.Spec.DataCentres[0].ClientEncryption {
		sslPort := k8scorev1.ServicePort{
			Name: models.CadenceGRPC,
			Port: models.Port7833,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port7833,
			},
		}
		headlessPorts = append(headlessPorts, sslPort)
	}
	return headlessPorts
}
