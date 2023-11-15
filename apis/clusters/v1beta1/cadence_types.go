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
	"regexp"

	k8scorev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
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
//+kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.status.id`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`

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
	return client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cadence) NewPatch() client.Patch {
	old := c.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
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

type immutableCadenceFields struct {
	immutableCluster
	UseCadenceWebAuth bool
}

type immutableCadenceDCFields struct {
	immutableDC      immutableDC
	ClientEncryption bool
}

func (cs *CadenceSpec) newImmutableFields() *immutableCadenceFields {
	return &immutableCadenceFields{
		immutableCluster:  cs.Cluster.newImmutableFields(),
		UseCadenceWebAuth: cs.UseCadenceWebAuth,
	}
}

func (cs *CadenceSpec) validateUpdate(oldSpec CadenceSpec) error {
	newImmutableFields := cs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldSpec, cs)
	}

	err := cs.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	err = validateTwoFactorDelete(cs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = cs.validateAWSArchival(oldSpec.AWSArchival)
	if err != nil {
		return err
	}

	err = cs.validateStandardProvisioning(oldSpec.StandardProvisioning)
	if err != nil {
		return err
	}

	err = cs.validateSharedProvisioning(oldSpec.SharedProvisioning)
	if err != nil {
		return err
	}

	err = cs.validatePackagedProvisioning(oldSpec.PackagedProvisioning)
	if err != nil {
		return err
	}

	err = cs.validateTargetsPrimaryCadence(&oldSpec)
	if err != nil {
		return err
	}

	return nil
}

func (cs *CadenceSpec) validateAWSArchival(old []*AWSArchival) error {
	if len(cs.AWSArchival) != len(old) {
		return models.ErrImmutableAWSArchival
	}

	for i, arch := range cs.AWSArchival {
		if *arch != *old[i] {
			return models.ErrImmutableAWSArchival
		}
	}

	return nil
}

func (cs *CadenceSpec) validateStandardProvisioning(old []*StandardProvisioning) error {
	if len(cs.StandardProvisioning) != len(old) {
		return models.ErrImmutableStandardProvisioning
	}

	for i, sp := range cs.StandardProvisioning {
		if *sp.TargetCassandra != *old[i].TargetCassandra {
			return models.ErrImmutableStandardProvisioning
		}

		err := sp.validateAdvancedVisibility(sp.AdvancedVisibility, old[i].AdvancedVisibility)
		if err != nil {
			return err
		}

	}

	return nil
}

func (sp *StandardProvisioning) validateAdvancedVisibility(new, old []*AdvancedVisibility) error {
	if len(old) != len(new) {
		return models.ErrImmutableAdvancedVisibility
	}

	for i, av := range new {
		if *old[i].TargetKafka != *av.TargetKafka ||
			*old[i].TargetOpenSearch != *av.TargetOpenSearch {
			return models.ErrImmutableAdvancedVisibility
		}
	}

	return nil
}

func (cs *CadenceSpec) validateSharedProvisioning(old []*SharedProvisioning) error {
	if len(cs.SharedProvisioning) != len(old) {
		return models.ErrImmutableSharedProvisioning
	}

	for i, sp := range cs.SharedProvisioning {
		if *sp != *old[i] {
			return models.ErrImmutableSharedProvisioning
		}
	}

	return nil
}

func (cs *CadenceSpec) validatePackagedProvisioning(old []*PackagedProvisioning) error {
	if len(cs.PackagedProvisioning) != len(old) {
		return models.ErrImmutablePackagedProvisioning
	}

	for i, pp := range cs.PackagedProvisioning {
		if pp.UseAdvancedVisibility {
			if pp.BundledKafkaSpec == nil || pp.BundledOpenSearchSpec == nil {
				return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must not be empty because UseAdvancedVisibility is set to true")
			}
			if *pp.BundledKafkaSpec != *old[i].BundledKafkaSpec {
				return models.ErrImmutablePackagedProvisioning
			}
			if *pp.BundledOpenSearchSpec != *old[i].BundledOpenSearchSpec {
				return models.ErrImmutablePackagedProvisioning
			}
		} else {
			if pp.BundledKafkaSpec != nil || pp.BundledOpenSearchSpec != nil {
				return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must be empty because UseAdvancedVisibility is set to false")
			}
		}

		if *pp.BundledCassandraSpec != *old[i].BundledCassandraSpec ||
			pp.UseAdvancedVisibility != old[i].UseAdvancedVisibility {
			return models.ErrImmutablePackagedProvisioning
		}
	}

	return nil
}

func (cs *CadenceSpec) validateTargetsPrimaryCadence(old *CadenceSpec) error {
	if len(cs.TargetPrimaryCadence) != len(old.TargetPrimaryCadence) {
		return fmt.Errorf("targetPrimaryCadence is immutable")
	}

	for _, oldTarget := range old.TargetPrimaryCadence {
		for _, newTarget := range cs.TargetPrimaryCadence {
			if *oldTarget != *newTarget {
				return fmt.Errorf("targetPrimaryCadence is immutable")
			}
		}
	}

	return nil
}

func (cdc *CadenceDataCentre) newImmutableFields() *immutableCadenceDCFields {
	return &immutableCadenceDCFields{
		immutableDC: immutableDC{
			Name:                cdc.Name,
			Region:              cdc.Region,
			CloudProvider:       cdc.CloudProvider,
			ProviderAccountName: cdc.ProviderAccountName,
			Network:             cdc.Network,
		},
		ClientEncryption: cdc.ClientEncryption,
	}
}

func (cs *CadenceSpec) validateImmutableDataCentresFieldsUpdate(oldSpec CadenceSpec) error {
	if len(cs.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range cs.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		if newDC.NodesNumber < oldDC.NodesNumber {
			return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldDC.NodesNumber)
		}

		err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		err = newDC.validatePrivateLink(oldDC.PrivateLink)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cdc *CadenceDataCentre) validatePrivateLink(old []*PrivateLink) error {
	if len(cdc.PrivateLink) != len(old) {
		return models.ErrImmutablePrivateLink
	}

	for i, pl := range cdc.PrivateLink {
		if *pl != *old[i] {
			return models.ErrImmutablePrivateLink
		}
	}

	return nil
}

func (aws *AWSArchival) validate() error {
	if !validation.Contains(aws.ArchivalS3Region, models.AWSRegions) {
		return fmt.Errorf("AWS Region: %s is unavailable, available regions: %v",
			aws.ArchivalS3Region, models.AWSRegions)
	}

	s3URIMatched, err := regexp.Match(models.S3URIRegExp, []byte(aws.ArchivalS3URI))
	if !s3URIMatched || err != nil {
		return fmt.Errorf("the provided S3 URI: %s is not correct and must fit the pattern: %s. %v",
			aws.ArchivalS3URI, models.S3URIRegExp, err)
	}

	return nil
}

func (sp *StandardProvisioning) validate() error {
	for _, av := range sp.AdvancedVisibility {
		if !validation.Contains(av.TargetOpenSearch.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target OpenSearch Dependency VPC Type: %s is unavailable, available options: %v",
				av.TargetOpenSearch.DependencyVPCType, models.DependencyVPCs)
		}

		osDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(av.TargetOpenSearch.DependencyCDCID))
		if !osDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target OpenSearch Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				av.TargetOpenSearch.DependencyCDCID, models.UUIDStringRegExp, err)
		}

		if !validation.Contains(av.TargetKafka.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target Kafka Dependency VPC Type: %s is unavailable, available options: %v",
				av.TargetKafka.DependencyVPCType, models.DependencyVPCs)
		}

		kafkaDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(av.TargetKafka.DependencyCDCID))
		if !kafkaDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target Kafka Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				av.TargetKafka.DependencyCDCID, models.UUIDStringRegExp, err)
		}

		if !validation.Contains(sp.TargetCassandra.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target Kafka Dependency VPC Type: %s is unavailable, available options: %v",
				sp.TargetCassandra.DependencyVPCType, models.DependencyVPCs)
		}

		cassandraDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(sp.TargetCassandra.DependencyCDCID))
		if !cassandraDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target Kafka Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				sp.TargetCassandra.DependencyCDCID, models.UUIDStringRegExp, err)
		}
	}

	return nil
}

func (b *BundledKafkaSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(b.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", b.Network, err)
	}

	err = validateReplicationFactor(models.KafkaReplicationFactors, b.ReplicationFactor)
	if err != nil {
		return err
	}

	if ((b.NodesNumber*b.ReplicationFactor)/b.ReplicationFactor)%b.ReplicationFactor != 0 {
		return fmt.Errorf("kafka: number of nodes must be a multiple of replication factor: %v", b.ReplicationFactor)
	}

	return nil
}

func (c *BundledCassandraSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(c.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", c.Network, err)
	}

	err = validateReplicationFactor(models.CassandraReplicationFactors, c.ReplicationFactor)
	if err != nil {
		return err
	}

	if ((c.NodesNumber*c.ReplicationFactor)/c.ReplicationFactor)%c.ReplicationFactor != 0 {
		return fmt.Errorf("cassandra: number of nodes must be a multiple of replication factor: %v", c.ReplicationFactor)
	}

	return nil
}

func (o *BundledOpenSearchSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(o.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", o.Network, err)
	}

	err = validateReplicationFactor(models.OpenSearchReplicationFactors, o.ReplicationFactor)
	if err != nil {
		return err
	}

	return nil
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
				Name: models.CadenceWeb,
				Port: models.Port8088,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port8088,
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
