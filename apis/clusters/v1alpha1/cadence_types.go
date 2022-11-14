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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type CadenceBundleOptions struct {
	EnableArchival          bool   `json:"enableArchival,omitempty"`
	TargetCassandraCDCID    string `json:"targetCassandraCdcId,omitempty"`
	TargetCassandraVPCType  string `json:"targetCassandraVpcType,omitempty"`
	TargetKafkaCDCID        string `json:"targetKafkaCdcId,omitempty"`
	TargetKafkaVPCType      string `json:"targetKafkaVpcType,omitempty"`
	TargetOpenSearchCDCID   string `json:"targetOpenSearchCdcId,omitempty"`
	TargetOpenSearchVPCType string `json:"targetOpenSearchVpcType,omitempty"`
	CadenceNodeCount        int    `json:"cadenceNodeCount"`
}

type CadenceDataCentre struct {
	DataCentre           `json:",inline"`
	CadenceBundleOptions `json:",inline"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	Cluster               `json:",inline"`
	DataCentres           []*CadenceDataCentre `json:"dataCentres,omitempty"`
	ConcurrentResizes     int                  `json:"concurrentResizes,omitempty"`
	NotifySupportContacts bool                 `json:"notifySupportContacts,omitempty"`
	Description           string               `json:"description,omitempty"`
	ProvisioningType      string               `json:"provisioningType"`

	// Packaged Cassandra
	CassandraNodeSize                       string `json:"cassandraNodeSize,omitempty"`
	CassandraNodesNumber                    int    `json:"cassandraNodesNumber,omitempty"`
	CassandraReplicationFactor              int    `json:"cassandraReplicationFactor,omitempty"`
	CassandraNetwork                        string `json:"cassandraNetwork,omitempty"`
	CassandraPrivateIPBroadcastForDiscovery bool   `json:"cassandraPrivateIPBroadcastForDiscovery,omitempty"`
	CassandraPasswordAndUserAuth            bool   `json:"cassandraPasswordAndUserAuth,omitempty"`

	UseAdvancedVisibility bool `json:"useAdvancedVisibility,omitempty"`
	// Packaged Kafka
	KafkaNodeSize          string `json:"kafkaNodeSize,omitempty"`
	KafkaNodesNumber       int    `json:"kafkaNodesNumber,omitempty"`
	KafkaNetwork           string `json:"kafkaNetwork"`
	KafkaReplicationFactor int    `json:"kafkaReplicationFactor,omitempty"`
	KafkaPartitionsNumber  int    `json:"kafkaPartitionsNumber,omitempty"`
	// Packaged OpenSearch
	OpensearchNodeSize                  string `json:"opensearchNodeSize,omitempty"`
	OpensearchReplicationFactor         int    `json:"opensearchReplicationFactor,omitempty"`
	OpenSearchNodesPerReplicationFactor int    `json:"openSearchNodesPerReplicationFactor,omitempty"`
	OpenSearchNetwork                   string `json:"openSearchNetwork,omitempty"`

	UseCadenceWebAuth           bool   `json:"useCadenceWebAuth,omitempty"`
	ArchivalS3URI               string `json:"archivalS3Uri,omitempty"`
	ArchivalS3Region            string `json:"archivalS3Region,omitempty"`
	AWSAccessKeySecretNamespace string `json:"awsAccessKeySecretNamespace,omitempty"`
	AWSAccessKeySecretName      string `json:"awsAccessKeySecretName,omitempty"`
	ClientEncryption            bool   `json:"clientEncryption,omitempty"`
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

func (cs *CadenceSpec) ToInstAPIv1(ctx *context.Context, k8sClient client.Client, logger *logr.Logger) (*models.CadenceClusterAPIv1, error) {
	cadenceBundles, err := cs.bundlesToInstAPIv1(ctx, k8sClient, logger)
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

func (cs *CadenceSpec) bundlesToInstAPIv1(ctx *context.Context, k8sClient client.Client, logger *logr.Logger) ([]*models.CadenceBundleAPIv1, error) {
	if len(cs.DataCentres) < 1 {
		return nil, models.ZeroDataCentres
	}
	dataCentre := cs.DataCentres[0]

	var AWSAccessKeyID string
	var AWSSecretAccessKey string

	if dataCentre.EnableArchival {
		awsCredsSecret := &v1.Secret{}
		awsSecretNamespacedName := types.NamespacedName{Name: cs.AWSAccessKeySecretName, Namespace: cs.AWSAccessKeySecretNamespace}
		err := k8sClient.Get(*ctx, awsSecretNamespacedName, awsCredsSecret)
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
			ClientEncryption:        cs.ClientEncryption,
			TargetCassandraCDCID:    dataCentre.TargetCassandraCDCID,
			TargetCassandraVPCType:  dataCentre.TargetCassandraVPCType,
			TargetKafkaCDCID:        dataCentre.TargetKafkaCDCID,
			TargetKafkaVPCType:      dataCentre.TargetKafkaVPCType,
			TargetOpenSearchCDCID:   dataCentre.TargetOpenSearchCDCID,
			TargetOpenSearchVPCType: dataCentre.TargetOpenSearchVPCType,
			EnableArchival:          dataCentre.EnableArchival,
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

func (cadence *Cadence) NewClusterMetadataPatch() (*client.Patch, error) {
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

	return &patch, nil
}

func (cs *CadenceSpec) GetUpdatedFields(oldSpec *CadenceSpec) *models.CadenceUpdatedFields {
	updatedFields := new(models.CadenceUpdatedFields)

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
