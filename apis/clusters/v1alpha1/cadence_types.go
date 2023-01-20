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

	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

type CadenceDataCentre struct {
	DataCentre       `json:",inline"`
	CadenceNodeCount int  `json:"cadenceNodeCount"`
	ClientEncryption bool `json:"clientEncryption,omitempty"`
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
	Cluster                     `json:",inline"`
	DataCentres                 []*CadenceDataCentre  `json:"dataCentres,omitempty"`
	ConcurrentResizes           int                   `json:"concurrentResizes,omitempty"`
	NotifySupportContacts       bool                  `json:"notifySupportContacts,omitempty"`
	Description                 string                `json:"description,omitempty"`
	ProvisioningType            string                `json:"provisioningType"`
	BundledCassandraSpec        BundledCassandraSpec  `json:"bundledCassandraSpec,omitempty"`
	UseAdvancedVisibility       bool                  `json:"useAdvancedVisibility,omitempty"`
	BundledKafkaSpec            BundledKafkaSpec      `json:"bundledKafkaSpec,omitempty"`
	BundledOpenSearchSpec       BundledOpenSearchSpec `json:"bundledOpenSearchSpec,omitempty"`
	UseCadenceWebAuth           bool                  `json:"useCadenceWebAuth,omitempty"`
	ArchivalS3URI               string                `json:"archivalS3Uri,omitempty"`
	ArchivalS3Region            string                `json:"archivalS3Region,omitempty"`
	AWSAccessKeySecretNamespace string                `json:"awsAccessKeySecretNamespace,omitempty"`
	AWSAccessKeySecretName      string                `json:"awsAccessKeySecretName,omitempty"`
	EnableArchival              bool                  `json:"enableArchival,omitempty"`
	TargetCassandraCDCID        string                `json:"targetCassandraCdcId,omitempty"`
	TargetCassandraVPCType      string                `json:"targetCassandraVpcType,omitempty"`
	TargetKafkaCDCID            string                `json:"targetKafkaCdcId,omitempty"`
	TargetKafkaVPCType          string                `json:"targetKafkaVpcType,omitempty"`
	TargetOpenSearchCDCID       string                `json:"targetOpenSearchCdcId,omitempty"`
	TargetOpenSearchVPCType     string                `json:"targetOpenSearchVpcType,omitempty"`
}

// CadenceSpec defines the observed state of Cadence
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

func (cs *CadenceSpec) ToInstAPIv1(ctx context.Context, k8sClient client.Client) (*models.CadenceClusterAPIv1, error) {
	cadenceBundles, err := cs.bundlesToInstAPIv1(ctx, k8sClient)
	if err != nil {
		return nil, err
	}

	var twoFactorDelete *models.TwoFactorDelete
	if len(cs.TwoFactorDelete) > 0 {
		twoFactorDelete = cs.TwoFactorDelete[0].ToInstAPIv1()
	}

	if len(cs.DataCentres) < 1 {
		return nil, models.ZeroDataCentres
	}

	return &models.CadenceClusterAPIv1{
		Cluster: models.Cluster{
			ClusterName:           cs.Name,
			NodeSize:              cs.DataCentres[0].NodeSize,
			PrivateNetworkCluster: cs.PrivateNetworkCluster,
			SLATier:               cs.SLATier,
			Provider:              cs.DataCentres[0].providerToInstAPIv1(),
			TwoFactorDelete:       twoFactorDelete,
			RackAllocation:        cs.DataCentres[0].rackAllocationToInstAPIv1(),
			DataCentre:            cs.DataCentres[0].Region,
			DataCentreCustomName:  cs.DataCentres[0].Name,
			ClusterNetwork:        cs.DataCentres[0].Network,
		},
		Bundles: cadenceBundles,
	}, nil
}

func (cs *CadenceSpec) bundlesToInstAPIv1(ctx context.Context, k8sClient client.Client) ([]*models.CadenceBundleAPIv1, error) {
	if len(cs.DataCentres) < 1 {
		return nil, models.ZeroDataCentres
	}
	dataCentre := cs.DataCentres[0]

	var AWSAccessKeyID string
	var AWSSecretAccessKey string

	if cs.EnableArchival {
		awsCredsSecret := &v1.Secret{}
		awsSecretNamespacedName := types.NamespacedName{Name: cs.AWSAccessKeySecretName, Namespace: cs.AWSAccessKeySecretNamespace}
		err := k8sClient.Get(ctx, awsSecretNamespacedName, awsCredsSecret)
		if err != nil {
			return nil, err
		}

		keyID := awsCredsSecret.Data[models.AWSAccessKeyID]
		secretAccessKey := awsCredsSecret.Data[models.AWSSecretAccessKey]
		AWSAccessKeyID = string(keyID[:len(keyID)-1])
		AWSSecretAccessKey = string(secretAccessKey[:len(secretAccessKey)-1])
	}

	cadenceBundle := &models.CadenceBundleAPIv1{
		Bundle: models.Bundle{
			Bundle:  models.Cadence,
			Version: cs.Version,
		},
		Options: &models.CadenceBundleOptionsAPIv1{
			UseAdvancedVisibility:   cs.UseAdvancedVisibility,
			UseCadenceWebAuth:       cs.UseCadenceWebAuth,
			ClientEncryption:        dataCentre.ClientEncryption,
			TargetCassandraCDCID:    cs.TargetCassandraCDCID,
			TargetCassandraVPCType:  cs.TargetCassandraVPCType,
			TargetKafkaCDCID:        cs.TargetKafkaCDCID,
			TargetKafkaVPCType:      cs.TargetKafkaVPCType,
			TargetOpenSearchCDCID:   cs.TargetOpenSearchCDCID,
			TargetOpenSearchVPCType: cs.TargetOpenSearchVPCType,
			EnableArchival:          cs.EnableArchival,
			ArchivalS3URI:           cs.ArchivalS3URI,
			ArchivalS3Region:        cs.ArchivalS3Region,
			AWSAccessKeyID:          AWSAccessKeyID,
			AWSSecretAccessKey:      AWSSecretAccessKey,
			CadenceNodeCount:        dataCentre.CadenceNodeCount,
			ProvisioningType:        cs.ProvisioningType,
		},
	}
	cadenceBundles := []*models.CadenceBundleAPIv1{cadenceBundle}

	return cadenceBundles, nil
}

func (dataCentre *CadenceDataCentre) rackAllocationToInstAPIv1() *models.RackAllocation {
	return &models.RackAllocation{
		NodesPerRack:  dataCentre.NodesNumber,
		NumberOfRacks: dataCentre.RacksNumber,
	}
}

func (twoFactorDelete *TwoFactorDelete) ToInstAPIv1() *models.TwoFactorDelete {
	return &models.TwoFactorDelete{
		DeleteVerifyEmail: twoFactorDelete.Email,
		DeleteVerifyPhone: twoFactorDelete.Phone,
	}
}

func (cadence *Cadence) NewClusterMetadataPatch() (client.Patch, error) {
	patchRequest := []*PatchRequest{}

	annotationsPayload, err := json.Marshal(cadence.Annotations)
	if err != nil {
		return nil, err
	}

	annotationsPatch := &PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.AnnotationsPath,
		Value:     json.RawMessage(annotationsPayload),
	}
	patchRequest = append(patchRequest, annotationsPatch)

	finalizersPayload, err := json.Marshal(cadence.Finalizers)
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

func (cs *CadenceSpec) GetUpdatedFields(oldSpec *CadenceSpec) models.CadenceUpdatedFields {
	updatedFields := models.CadenceUpdatedFields{}

	if cs.DataCentres[0].NodeSize != oldSpec.DataCentres[0].NodeSize {
		updatedFields.NodeSizeUpdated = true
	}

	if cs.Description != oldSpec.Description {
		updatedFields.DescriptionUpdated = true
	}

	if len(cs.TwoFactorDelete) != 0 {
		if cs.TwoFactorDelete[0] != oldSpec.TwoFactorDelete[0] {
			updatedFields.TwoFactorDeleteUpdated = true
		}
	}

	return updatedFields
}

func (c *Cadence) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(c).String() + "/" + jobName
}

func (c *Cadence) NewPatch() client.Patch {
	old := c.DeepCopy()
	return client.MergeFrom(old)
}

func (cs *CadenceSpec) ToInstAPIv2(ctx context.Context, k8sClient client.Client) (*models.CadenceClusterAPIv2, error) {
	var AWSArchival *models.AWSArchival
	var err error
	if cs.EnableArchival {
		AWSArchival, err = cs.ArchivalToAPIv2(ctx, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("cannot convert archival data to APIv2 format: %v", err)
		}
	}

	instDataCentres := []*models.CadenceDataCentre{}
	for _, dataCentre := range cs.DataCentres {
		instDataCentres = append(instDataCentres, dataCentre.ToAPIv2())
	}

	var sharedProvisioning []*models.CadenceSharedProvisioning
	var standardProvisioning []*models.CadenceStandardProvisioning
	switch cs.ProvisioningType {
	case models.SharedProvisioningType:
		sharedProvisioning = cs.SharedProvisioningToAPIv2()
	case models.StandardProvisioningType, models.PackagedProvisioningType:
		standardProvisioning, err = cs.StandardProvisioningToAPIv2()
		if err != nil {
			return nil, err
		}
	}

	return &models.CadenceClusterAPIv2{
		CadenceVersion:        cs.Version,
		CadenceDataCentres:    instDataCentres,
		Name:                  cs.Name,
		PCIComplianceMode:     cs.PCICompliance,
		TwoFactorDelete:       cs.TwoFactorDeleteToAPIv2(),
		UseCadenceWebAuth:     cs.UseCadenceWebAuth,
		PrivateNetworkCluster: cs.PrivateNetworkCluster,
		SLATier:               cs.SLATier,
		AWSArchival:           AWSArchival,
		SharedProvisioning:    sharedProvisioning,
		StandardProvisioning:  standardProvisioning,
	}, nil
}

func (cs *CadenceSpec) StandardProvisioningToAPIv2() ([]*models.CadenceStandardProvisioning, error) {
	stdProvisioning := &models.CadenceStandardProvisioning{
		TargetCassandra: &models.TargetCassandra{
			DependencyCDCID:   cs.TargetCassandraCDCID,
			DependencyVPCType: cs.TargetCassandraVPCType,
		},
	}

	if cs.UseAdvancedVisibility {
		if cs.TargetKafkaVPCType == "" ||
			cs.TargetKafkaCDCID == "" ||
			cs.TargetOpenSearchVPCType == "" ||
			cs.TargetOpenSearchCDCID == "" {
			return nil, models.ErrEmptyAdvancedVisibility
		}

		stdProvisioning.AdvancedVisibility = []*models.AdvancedVisibility{
			{
				TargetKafka: &models.TargetKafka{
					DependencyCDCID:   cs.TargetKafkaCDCID,
					DependencyVPCType: cs.TargetKafkaVPCType,
				},
				TargetOpenSearch: &models.TargetOpenSearch{
					DependencyCDCID:   cs.TargetOpenSearchCDCID,
					DependencyVPCType: cs.TargetOpenSearchVPCType,
				},
			},
		}
	}

	return []*models.CadenceStandardProvisioning{stdProvisioning}, nil
}

func (cs *CadenceSpec) SharedProvisioningToAPIv2() []*models.CadenceSharedProvisioning {
	return []*models.CadenceSharedProvisioning{
		{
			UseAdvancedVisibility: cs.UseAdvancedVisibility,
		},
	}
}

func (cs *CadenceSpec) TwoFactorDeleteToAPIv2() []*modelsv2.TwoFactorDelete {
	var twoFactorDeleteAPIv2 []*modelsv2.TwoFactorDelete

	for _, twoFactorDelete := range cs.TwoFactorDelete {
		twoFactorDeleteAPIv2 = append(twoFactorDeleteAPIv2, twoFactorDelete.ToInstAPI())
	}

	return twoFactorDeleteAPIv2
}

func (cs *CadenceSpec) ArchivalToAPIv2(ctx context.Context, k8sClient client.Client) (*models.AWSArchival, error) {
	AWSAccessKeyID, AWSSecretAccessKey, err := cs.GetSecret(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get aws creds secret: %v", err)
	}

	return &models.AWSArchival{
		ArchivalS3Region:   cs.ArchivalS3Region,
		ArchivalS3URI:      cs.ArchivalS3URI,
		AWSAccessKeyID:     AWSAccessKeyID,
		AWSSecretAccessKey: AWSSecretAccessKey,
	}, nil
}

func (cs *CadenceSpec) GetSecret(ctx context.Context, k8sClient client.Client) (string, string, error) {
	var err error
	awsCredsSecret := &v1.Secret{}
	awsSecretNamespacedName := types.NamespacedName{Name: cs.AWSAccessKeySecretName, Namespace: cs.AWSAccessKeySecretNamespace}
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

func (cdc *CadenceDataCentre) ToAPIv2() *models.CadenceDataCentre {
	cadenceDC := &models.CadenceDataCentre{
		ClientToClusterEncryption: cdc.ClientEncryption,
		DataCentre: modelsv2.DataCentre{
			CloudProvider:       cdc.CloudProvider,
			Name:                cdc.Name,
			Network:             cdc.Network,
			NodeSize:            cdc.NodeSize,
			NumberOfNodes:       int32(cdc.CadenceNodeCount),
			Region:              cdc.Region,
			ProviderAccountName: cdc.ProviderAccountName,
		},
	}

	cdc.TagsToInstAPI(&cadenceDC.DataCentre)

	cdc.CloudProviderSettingsToInstAPI(&cadenceDC.DataCentre)

	return cadenceDC
}
