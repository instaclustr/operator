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
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type CadenceDataCentre struct {
	DataCentre       `json:",inline"`
	ClientEncryption bool           `json:"clientEncryption,omitempty"`
	PrivateLink      []*PrivateLink `json:"privateLink,omitempty"`
}

type BundledCassandraSpec struct {
	NodeSize                       string `json:"nodeSize,omitempty"`
	NodesNumber                    int    `json:"nodesNumber,omitempty"`
	ReplicationFactor              int    `json:"replicationFactor,omitempty"`
	Network                        string `json:"network,omitempty"`
	PrivateIPBroadcastForDiscovery bool   `json:"privateIPBroadcastForDiscovery,omitempty"`
	PasswordAndUserAuth            bool   `json:"passwordAndUserAuth,omitempty"`
}

type BundledKafkaSpec struct {
	NodeSize          string `json:"nodeSize,omitempty"`
	NodesNumber       int    `json:"nodesNumber,omitempty"`
	Network           string `json:"network"`
	ReplicationFactor int    `json:"replicationFactor,omitempty"`
	PartitionsNumber  int    `json:"partitionsNumber,omitempty"`
}

type BundledOpenSearchSpec struct {
	NodeSize                  string `json:"nodeSize,omitempty"`
	ReplicationFactor         int    `json:"replicationFactor,omitempty"`
	NodesPerReplicationFactor int    `json:"nodesPerReplicationFactor,omitempty"`
	Network                   string `json:"network,omitempty"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	Cluster              `json:",inline"`
	DataCentres          []*CadenceDataCentre    `json:"dataCentres,omitempty"`
	Description          string                  `json:"description,omitempty"`
	UseCadenceWebAuth    bool                    `json:"useCadenceWebAuth"`
	AWSArchival          []*AWSArchival          `json:"awsArchival,omitempty"`
	StandardProvisioning []*StandardProvisioning `json:"standardProvisioning,omitempty"`
	SharedProvisioning   []*SharedProvisioning   `json:"sharedProvisioning,omitempty"`
	PackagedProvisioning []*PackagedProvisioning `json:"packagedProvisioning,omitempty"`
}

type AWSArchival struct {
	ArchivalS3URI            string `json:"archivalS3Uri"`
	ArchivalS3Region         string `json:"archivalS3Region"`
	AccessKeySecretNamespace string `json:"awsAccessKeySecretNamespace,omitempty"`
	AccessKeySecretName      string `json:"awsAccessKeySecretName,omitempty"`
}

type PackagedProvisioning struct {
	UseAdvancedVisibility bool                  `json:"useAdvancedVisibility"`
	BundledKafkaSpec      BundledKafkaSpec      `json:"bundledKafkaSpec,omitempty"`
	BundledOpenSearchSpec BundledOpenSearchSpec `json:"bundledOpenSearchSpec,omitempty"`
	BundledCassandraSpec  BundledCassandraSpec  `json:"bundledCassandraSpec,omitempty"`
}

type SharedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`
}

type StandardProvisioning struct {
	AdvancedVisibility []*AdvancedVisibility `json:"advancedVisibility,omitempty"`
	TargetCassandra    TargetCassandra       `json:"targetCassandra"`
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

type AdvancedVisibility struct {
	TargetKafka      TargetKafka      `json:"targetKafka"`
	TargetOpenSearch TargetOpenSearch `json:"targetOpenSearch"`
}

// CadenceStatus defines the observed state of Cadence
type CadenceStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

func (c *Cadence) NewClusterMetadataPatch() (client.Patch, error) {
	patchRequest := []*PatchRequest{}

	annotationsPayload, err := json.Marshal(c.Annotations)
	if err != nil {
		return nil, err
	}

	annotationsPatch := &PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(c.Finalizers)
	if err != nil {
		return nil, err
	}

	finzlizersPatch := &PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finzlizersPatch)

	patchPayload, err := json.Marshal(patchRequest)
	if err != nil {
		return nil, err
	}

	patch := client.RawPatch(types.JSONPatchType, patchPayload)

	return patch, nil
}

func (c *Cadence) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cadence) NewPatch() client.Patch {
	old := c.DeepCopy()
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
		SharedProvisioning:    sharedProvisioning,
		StandardProvisioning:  standardProvisioning,
	}, nil
}

func (cs *CadenceSpec) StandardProvisioningToInstAPI() []*models.CadenceStandardProvisioning {
	cadenceStandardProvisioning := []*models.CadenceStandardProvisioning{}

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
	sharedProvisioning := []*models.CadenceSharedProvisioning{}

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
	return
}

func (cs *CadenceSpec) DCsFromInstAPI(iDCs []*models.CadenceDataCentre) (dcs []*CadenceDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &CadenceDataCentre{
			DataCentre:       cs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ClientEncryption: iDC.ClientToClusterEncryption,
		})
	}
	return
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
	}
}

func (cs *CadenceStatus) DCsFromInstAPI(iDCs []*models.CadenceDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, cs.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}

func (cs *CadenceSpec) NewDCsUpdate() models.CadenceClusterAPIUpdate {
	return models.CadenceClusterAPIUpdate{
		DataCentres: cs.DCsToInstAPI(),
	}
}
