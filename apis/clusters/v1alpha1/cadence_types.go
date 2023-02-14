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

	finalizersPatch := &PatchRequest{
		Operation: models.ReplaceOperation,
		Path:      models.FinalizersPath,
		Value:     json.RawMessage(finalizersPayload),
	}
	patchRequest = append(patchRequest, finalizersPatch)

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

func (cs *CadenceSpec) ToInstAPI(ctx context.Context, k8sClient client.Client) (*models.CadenceAPIv2, error) {
	var AWSArchival []*models.AWSArchival
	var err error
	if cs.EnableArchival {
		AWSArchival, err = cs.ArchivalToInstAPI(ctx, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("cannot convert archival data to APIv2 format: %v", err)
		}
	}

	var sharedProvisioning []*models.CadenceSharedProvisioning
	var standardProvisioning []*models.CadenceStandardProvisioning
	switch cs.ProvisioningType {
	case models.SharedProvisioningType:
		sharedProvisioning = cs.SharedProvisioningToInstAPI()
	case models.StandardProvisioningType, models.PackagedProvisioningType:
		standardProvisioning, err = cs.StandardProvisioningToInstAPI()
		if err != nil {
			return nil, err
		}
	}

	return &models.CadenceAPIv2{
		CadenceVersion:        cs.Version,
		DataCentres:           cs.DataCentresToInstAPI(),
		Name:                  cs.Name,
		PCIComplianceMode:     cs.PCICompliance,
		TwoFactorDelete:       cs.TwoFactorDeletesToInstAPI(),
		UseCadenceWebAuth:     cs.UseCadenceWebAuth,
		PrivateNetworkCluster: cs.PrivateNetworkCluster,
		SLATier:               cs.SLATier,
		AWSArchival:           AWSArchival,
		SharedProvisioning:    sharedProvisioning,
		StandardProvisioning:  standardProvisioning,
	}, nil
}

func (cs *CadenceSpec) StandardProvisioningToInstAPI() ([]*models.CadenceStandardProvisioning, error) {
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

func (cs *CadenceSpec) SharedProvisioningToInstAPI() []*models.CadenceSharedProvisioning {
	return []*models.CadenceSharedProvisioning{
		{
			UseAdvancedVisibility: cs.UseAdvancedVisibility,
		},
	}
}

func (cs *CadenceSpec) ArchivalToInstAPI(ctx context.Context, k8sClient client.Client) ([]*models.AWSArchival, error) {
	AWSAccessKeyID, AWSSecretAccessKey, err := cs.GetSecret(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("cannot get aws creds secret: %v", err)
	}

	return []*models.AWSArchival{
		{
			ArchivalS3Region:   cs.ArchivalS3Region,
			ArchivalS3URI:      cs.ArchivalS3URI,
			AWSAccessKeyID:     AWSAccessKeyID,
			AWSSecretAccessKey: AWSSecretAccessKey,
		},
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

	var awsAccessKeyID string
	var awsSecretAccessKey string
	keyID := awsCredsSecret.Data[models.AWSAccessKeyID]
	secretAccessKey := awsCredsSecret.Data[models.AWSSecretAccessKey]
	awsAccessKeyID = string(keyID[:len(keyID)-1])
	awsSecretAccessKey = string(secretAccessKey[:len(secretAccessKey)-1])

	return awsAccessKeyID, awsSecretAccessKey, err
}

func (cdc *CadenceDataCentre) ToInstAPI() *models.CadenceDataCentre {
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

func (cs *CadenceSpec) DataCentresToInstAPI() []*models.CadenceDataCentre {
	var instDCs []*models.CadenceDataCentre
	for _, k8sDC := range cs.DataCentres {
		instDCs = append(instDCs, k8sDC.ToInstAPI())
	}
	return instDCs
}

func (cst *CadenceStatus) SetStatusFromInst(instStatus *models.CadenceAPIv2) {
	cst.ID = instStatus.ID
	cst.Status = instStatus.Status
	cst.CurrentClusterOperationStatus = instStatus.CurrentClusterOperationStatus
	cst.SetDCsStatusFromInst(instStatus.DataCentres)
}

func (cst *CadenceStatus) SetDCsStatusFromInst(instDCs []*models.CadenceDataCentre) {
	dcs := []*DataCentreStatus{}
	for _, instDC := range instDCs {
		dcToAppend := &DataCentreStatus{
			ID:         instDC.ID,
			Status:     instDC.Status,
			NodeNumber: instDC.NumberOfNodes,
		}
		dcToAppend.SetNodesStatusFromInstAPI(instDC.Nodes)
		dcs = append(dcs, dcToAppend)
	}
	cst.DataCentres = dcs
}

func (cst *CadenceStatus) AreStatusesEqual(instCadence *models.CadenceAPIv2) bool {
	if cst.Status != instCadence.Status ||
		cst.CurrentClusterOperationStatus != instCadence.CurrentClusterOperationStatus ||
		!cst.AreDCsEqual(instCadence.DataCentres) {
		return false
	}

	return true
}

func (cst *CadenceStatus) AreDCsEqual(instDCs []*models.CadenceDataCentre) bool {
	if len(cst.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range cst.DataCentres {
			if instDC.ID == k8sDC.ID {
				if instDC.Status != k8sDC.Status ||
					instDC.NumberOfNodes != k8sDC.NodeNumber ||
					!k8sDC.AreNodesEqual(instDC.Nodes) {
					return false
				}

				break
			}
		}
	}

	return true
}
