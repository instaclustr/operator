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
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type CadenceDataCentre struct {
	GenericDataCentreSpec `json:",inline"`

	NodesNumber      int    `json:"nodesNumber"`
	NodeSize         string `json:"nodeSize"`
	ClientEncryption bool   `json:"clientEncryption,omitempty"`

	PrivateLink PrivateLinkSpec `json:"privateLink,omitempty"`
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
	NodeSize      string `json:"nodeSize"`
	NumberOfRacks int    `json:"numberOfRacks"`
	Network       string `json:"network"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	GenericClusterSpec `json:",inline"`

	//+kubebuilder:validation:MinItems:=1
	//+kubebuilder:validation:MaxItems:=1
	DataCentres []*CadenceDataCentre `json:"dataCentres"`

	//+kubebuilder:validation:MaxItems:=1
	AWSArchival []*AWSArchival `json:"awsArchival,omitempty"`

	//+kubebuilder:validation:MaxItems:=1
	StandardProvisioning []*StandardProvisioning `json:"standardProvisioning,omitempty"`

	//+kubebuilder:validation:MaxItems:=1
	SharedProvisioning []*SharedProvisioning `json:"sharedProvisioning,omitempty"`

	//+kubebuilder:validation:MaxItems:=1
	PackagedProvisioning []*PackagedProvisioning `json:"packagedProvisioning,omitempty"`

	//+kubebuilder:validation:MaxItems:=1
	TargetPrimaryCadence []*CadenceDependencyTarget `json:"targetPrimaryCadence,omitempty"`

	ResizeSettings GenericResizeSettings `json:"resizeSettings,omitempty" dcomparisonSkip:"true"`

	UseCadenceWebAuth bool `json:"useCadenceWebAuth"`
	UseHTTPAPI        bool `json:"useHttpApi,omitempty"`
	PCICompliance     bool `json:"pciCompliance,omitempty"`
}

type AWSArchival struct {
	ArchivalS3URI            string `json:"archivalS3Uri"`
	ArchivalS3Region         string `json:"archivalS3Region"`
	AccessKeySecretNamespace string `json:"awsAccessKeySecretNamespace"`
	AccessKeySecretName      string `json:"awsAccessKeySecretName"`
}

type PackagedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`

	// +kubebuilder:validation:Enum=Developer;Production-Starter;Production-Small
	SolutionSize string `json:"solutionSize"`
}

type SharedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`
}

type StandardProvisioning struct {
	//+kubebuilder:validation:MaxItems:=1
	AdvancedVisibility []*AdvancedVisibility    `json:"advancedVisibility,omitempty"`
	TargetCassandra    *CadenceDependencyTarget `json:"targetCassandra"`
}

type CadenceDependencyTarget struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type AdvancedVisibility struct {
	TargetKafka      *CadenceDependencyTarget `json:"targetKafka"`
	TargetOpenSearch *CadenceDependencyTarget `json:"targetOpenSearch"`
}

// CadenceStatus defines the observed state of Cadence
type CadenceStatus struct {
	GenericStatus `json:",inline"`

	PackagedProvisioningClusterRefs []*Reference               `json:"packagedProvisioningClusterRefs,omitempty"`
	DataCentres                     []*CadenceDataCentreStatus `json:"dataCentres,omitempty"`
	TargetSecondaryCadence          []*CadenceDependencyTarget `json:"targetSecondaryCadence,omitempty"`
}

type CadenceDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	Nodes       []*Node             `json:"nodes,omitempty"`
	PrivateLink PrivateLinkStatuses `json:"privateLink,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Node count",type="string",JSONPath=".status.nodeCount"

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
		GenericClusterFields: cs.GenericClusterSpec.ToInstAPI(),
		CadenceVersion:       cs.Version,
		DataCentres:          cs.DCsToInstAPI(),
		PCIComplianceMode:    cs.PCICompliance,
		TwoFactorDelete:      cs.TwoFactorDeleteToInstAPI(),
		UseCadenceWebAuth:    cs.UseCadenceWebAuth,
		AWSArchival:          awsArchival,
		SharedProvisioning:   sharedProvisioning,
		StandardProvisioning: standardProvisioning,
		TargetPrimaryCadence: cs.TargetCadenceToInstAPI(),
		ResizeSettings:       resizeSettingsToInstAPI(cs.ResizeSettings),
		UseHTTPAPI:           cs.UseHTTPAPI,
	}, nil
}

func (cs *CadenceSpec) StandardProvisioningToInstAPI() []*models.CadenceStandardProvisioning {
	var cadenceStandardProvisioning []*models.CadenceStandardProvisioning

	for _, standardProvisioning := range cs.StandardProvisioning {
		targetCassandra := standardProvisioning.TargetCassandra

		stdProvisioning := &models.CadenceStandardProvisioning{
			TargetCassandra: &models.CadenceDependencyTarget{
				DependencyCDCID:   targetCassandra.DependencyCDCID,
				DependencyVPCType: targetCassandra.DependencyVPCType,
			},
		}

		for _, advancedVisibility := range standardProvisioning.AdvancedVisibility {
			targetKafka := advancedVisibility.TargetKafka
			targetOpenSearch := advancedVisibility.TargetOpenSearch

			stdProvisioning.AdvancedVisibility = []*models.AdvancedVisibility{
				{
					TargetKafka: &models.CadenceDependencyTarget{
						DependencyCDCID:   targetKafka.DependencyCDCID,
						DependencyVPCType: targetKafka.DependencyVPCType,
					},
					TargetOpenSearch: &models.CadenceDependencyTarget{
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
	awsCredsSecret := &k8scorev1.Secret{}
	awsSecretNamespacedName := types.NamespacedName{
		Name:      aws.AccessKeySecretName,
		Namespace: aws.AccessKeySecretNamespace,
	}

	err := k8sClient.Get(ctx, awsSecretNamespacedName, awsCredsSecret)
	if err != nil {
		return "", "", err
	}

	keyID := awsCredsSecret.Data[models.AWSAccessKeyID]
	secretAccessKey := awsCredsSecret.Data[models.AWSSecretAccessKey]
	AWSAccessKeyID := string(keyID[:len(keyID)-1])
	AWSSecretAccessKey := string(secretAccessKey[:len(secretAccessKey)-1])

	return AWSAccessKeyID, AWSSecretAccessKey, nil
}

func (cdc *CadenceDataCentre) ToInstAPI() *models.CadenceDataCentre {
	return &models.CadenceDataCentre{
		GenericDataCentreFields:   cdc.GenericDataCentreSpec.ToInstAPI(),
		PrivateLink:               cdc.PrivateLink.ToInstAPI(),
		ClientToClusterEncryption: cdc.ClientEncryption,
		NumberOfNodes:             cdc.NodesNumber,
		NodeSize:                  cdc.NodeSize,
	}
}

func (cs *CadenceSpec) TargetCadenceToInstAPI() []*models.CadenceDependencyTarget {
	var targets []*models.CadenceDependencyTarget
	for _, target := range cs.TargetPrimaryCadence {
		targets = append(targets, &models.CadenceDependencyTarget{
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

func (c *Cadence) FromInstAPI(instaModel *models.CadenceCluster) {
	c.Spec.FromInstAPI(instaModel)
	c.Status.FromInstAPI(instaModel)
}

func (cs *CadenceSpec) FromInstAPI(instaModel *models.CadenceCluster) {
	cs.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields, instaModel.CadenceVersion)
	cs.ResizeSettings.FromInstAPI(instaModel.ResizeSettings)

	cs.DCsFromInstAPI(instaModel.DataCentres)
	cs.StandardProvisioningFromInstAPI(instaModel.StandardProvisioning)
	cs.SharedProvisioningFromInstAPI(instaModel.SharedProvisioning)
	cs.TargetPrimaryCadenceFromInstAPI(instaModel.TargetPrimaryCadence)

	// TODO AWS ARCHIVAL

	cs.UseHTTPAPI = instaModel.UseHTTPAPI
	cs.UseCadenceWebAuth = instaModel.UseCadenceWebAuth
	cs.PCICompliance = instaModel.PCIComplianceMode
}

func (cs *CadenceSpec) DCsFromInstAPI(instaModels []*models.CadenceDataCentre) {
	dcs := make([]*CadenceDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &CadenceDataCentre{}
		dc.FromInstAPI(instaModel)
		dcs[i] = dc
	}
	cs.DataCentres = dcs
}

func (cs *CadenceSpec) PrivateLinkFromInstAPI(iPLs []*models.PrivateLink) (pls []*PrivateLink) {
	for _, iPL := range iPLs {
		pls = append(pls, &PrivateLink{
			AdvertisedHostname: iPL.AdvertisedHostname,
		})
	}
	return
}

func (c *Cadence) GetSpec() CadenceSpec { return c.Spec }

func (c *Cadence) IsSpecEqual(spec CadenceSpec) bool {
	return c.Spec.Equals(&spec.GenericClusterSpec) &&
		c.Spec.DCsEqual(spec.DataCentres)
}

func (cs *CadenceSpec) DCsEqual(dcs []*CadenceDataCentre) bool {
	if len(cs.DataCentres) != len(dcs) {
		return false
	}

	m := map[string]*CadenceDataCentre{}
	for _, dc := range cs.DataCentres {
		m[dc.Name] = dc
	}

	for _, dc := range dcs {
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

func (cs *CadenceStatus) FromInstAPI(instaModel *models.CadenceCluster) {
	cs.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	cs.targetSecondaryFromInstAPI(instaModel.TargetSecondaryCadence)
	cs.DCsFromInstAPI(instaModel.DataCentres)
	cs.GetNodeCount()
}

func (cs *CadenceStatus) GetNodeCount() {
	var total, running int
	for _, dc := range cs.DataCentres {
		for _, node := range dc.Nodes {
			total++
			if node.Status == models.RunningStatus {
				running++
			}
		}
	}
	cs.NodeCount = fmt.Sprintf("%v/%v", running, total)
}

func (cs *CadenceStatus) targetSecondaryFromInstAPI(instaModels []*models.CadenceDependencyTarget) {
	cs.TargetSecondaryCadence = make([]*CadenceDependencyTarget, len(instaModels))
	for i, target := range instaModels {
		cs.TargetSecondaryCadence[i] = &CadenceDependencyTarget{
			DependencyCDCID:   target.DependencyCDCID,
			DependencyVPCType: target.DependencyVPCType,
		}
	}
}

func (cs *CadenceStatus) DCsFromInstAPI(instaModels []*models.CadenceDataCentre) {
	cs.DataCentres = make([]*CadenceDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &CadenceDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		cs.DataCentres[i] = dc
	}
}

func (cs *CadenceSpec) NewDCsUpdate() models.CadenceClusterAPIUpdate {
	return models.CadenceClusterAPIUpdate{
		DataCentres:    cs.DCsToInstAPI(),
		ResizeSettings: cs.ResizeSettings.ToInstAPI(),
	}
}

func (c *Cadence) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !c.Spec.PrivateNetwork {
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

func (cdc *CadenceDataCentreStatus) FromInstAPI(instaModel *models.CadenceDataCentre) {
	cdc.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	cdc.Nodes = nodesFromInstAPI(instaModel.Nodes)
	cdc.PrivateLink.FromInstAPI(instaModel.PrivateLink)
}

func (cdc *CadenceDataCentre) Equals(o *CadenceDataCentre) bool {
	return cdc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		cdc.NodesNumber == o.NodesNumber &&
		cdc.NodeSize == o.NodeSize &&
		cdc.ClientEncryption == o.ClientEncryption &&
		slices.EqualsPtr(cdc.PrivateLink, o.PrivateLink)
}

func (cdc *CadenceDataCentre) FromInstAPI(instaModel *models.CadenceDataCentre) {
	cdc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)
	cdc.PrivateLink.FromInstAPI(instaModel.PrivateLink)

	cdc.NodesNumber = instaModel.NumberOfNodes
	cdc.NodeSize = instaModel.NodeSize
	cdc.ClientEncryption = instaModel.ClientToClusterEncryption
}

func (cs *CadenceStatus) Equals(o *CadenceStatus) bool {
	return cs.GenericStatus.Equals(&o.GenericStatus) &&
		cs.DCsEqual(o.DataCentres) &&
		slices.EqualsPtr(cs.TargetSecondaryCadence, o.TargetSecondaryCadence)
}

func (cs *CadenceStatus) DCsEqual(o []*CadenceDataCentreStatus) bool {
	if len(cs.DataCentres) != len(o) {
		return false
	}

	m := map[string]*CadenceDataCentreStatus{}
	for _, dc := range cs.DataCentres {
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

func (cdc *CadenceDataCentreStatus) Equals(o *CadenceDataCentreStatus) bool {
	return cdc.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		cdc.PrivateLink.Equal(o.PrivateLink) &&
		nodesEqual(cdc.Nodes, o.Nodes)
}

func (c *CadenceSpec) CalculateNodeSize(cloudProvider, solution, app string) string {
	if appSizes, ok := models.SolutionSizesMap[cloudProvider]; ok {
		if solutionMap, ok := appSizes[app]; ok {
			return solutionMap[solution]
		}
	}
	return ""
}

func (cs *CadenceSpec) StandardProvisioningFromInstAPI(instaModels []*models.CadenceStandardProvisioning) {
	provisioning := make([]*StandardProvisioning, len(instaModels))
	for i, instaModel := range instaModels {
		p := &StandardProvisioning{}
		p.FromInstAPI(instaModel)
		provisioning[i] = p
	}
	cs.StandardProvisioning = provisioning
}

func (s *StandardProvisioning) FromInstAPI(instaModel *models.CadenceStandardProvisioning) {
	s.AdvancedVisibility = make([]*AdvancedVisibility, len(instaModel.AdvancedVisibility))
	for i, visibility := range instaModel.AdvancedVisibility {
		vis := &AdvancedVisibility{
			TargetKafka: &CadenceDependencyTarget{
				DependencyCDCID:   visibility.TargetKafka.DependencyCDCID,
				DependencyVPCType: visibility.TargetKafka.DependencyVPCType,
			},
			TargetOpenSearch: &CadenceDependencyTarget{
				DependencyCDCID:   visibility.TargetOpenSearch.DependencyCDCID,
				DependencyVPCType: visibility.TargetOpenSearch.DependencyVPCType,
			},
		}
		s.AdvancedVisibility[i] = vis
	}

	if instaModel.TargetCassandra != nil {
		s.TargetCassandra = &CadenceDependencyTarget{
			DependencyCDCID:   instaModel.TargetCassandra.DependencyCDCID,
			DependencyVPCType: instaModel.TargetCassandra.DependencyVPCType,
		}
	}
}

func (cs *CadenceSpec) SharedProvisioningFromInstAPI(instaModels []*models.CadenceSharedProvisioning) {
	cs.SharedProvisioning = make([]*SharedProvisioning, len(instaModels))
	for i, instaModel := range instaModels {
		cs.SharedProvisioning[i] = &SharedProvisioning{
			UseAdvancedVisibility: instaModel.UseAdvancedVisibility,
		}
	}
}

func (cs *CadenceSpec) TargetPrimaryCadenceFromInstAPI(instaModels []*models.CadenceDependencyTarget) {
	cs.TargetPrimaryCadence = make([]*CadenceDependencyTarget, len(instaModels))
	for i, instaModel := range instaModels {
		cs.TargetPrimaryCadence[i] = &CadenceDependencyTarget{
			DependencyCDCID:   instaModel.DependencyCDCID,
			DependencyVPCType: instaModel.DependencyVPCType,
		}
	}
}
