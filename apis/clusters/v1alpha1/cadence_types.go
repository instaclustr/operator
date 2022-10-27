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
	UseAdvancedVisibility       bool   `json:"useAdvancedVisibility,omitempty"`
	UseCadenceWebAuth           bool   `json:"useCadenceWebAuth,omitempty"`
	EnableArchival              bool   `json:"enableArchival,omitempty"`
	ClientEncryption            bool   `json:"clientEncryption,omitempty"`
	TargetCassandraCDCID        string `json:"targetCassandraCdcId,omitempty"`
	TargetCassandraVPCType      string `json:"targetCassandraVpcType,omitempty"`
	TargetKafkaCDCID            string `json:"targetKafkaCdcId,omitempty"`
	TargetKafkaVPCType          string `json:"targetKafkaVpcType,omitempty"`
	TargetOpenSearchCDCID       string `json:"targetOpenSearchCdcId,omitempty"`
	TargetOpenSearchVPCType     string `json:"targetOpenSearchVpcType,omitempty"`
	ArchivalS3URI               string `json:"archivalS3Uri,omitempty"`
	ArchivalS3Region            string `json:"archivalS3Region,omitempty"`
	AWSAccessKeySecretNamespace string `json:"awsAccessKeySecretNamespace,omitempty"`
	AWSAccessKeySecretName      string `json:"awsAccessKeySecretName,omitempty"`
	CadenceNodeCount            int    `json:"cadenceNodeCount"`
	ProvisioningType            string `json:"provisioningType"`
}

type CadenceDataCentre struct {
	DataCentre           `json:",inline"`
	CadenceBundleOptions `json:",inline"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	Cluster     `json:",inline"`
	DataCentres []*CadenceDataCentre `json:"dataCentres,omitempty"`
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

func (spec *CadenceSpec) ToInstAPIv1(ctx *context.Context, k8sClient client.Client) (*models.CadenceClusterAPIv1, error) {
	cadenceBundles, err := spec.bundlesToInstAPIv1(ctx, k8sClient)
	if err != nil {
		return nil, err
	}

	return &models.CadenceClusterAPIv1{
		Cluster: models.Cluster{
			ClusterName:           spec.Name,
			NodeSize:              spec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: spec.PrivateNetworkCluster,
			SLATier:               spec.SLATier,
			Provider:              spec.DataCentres[0].providerToInstAPIv1(),
			TwoFactorDelete:       spec.twoFactorDeleteToInstAPIv1(),
			RackAllocation:        spec.rackAllocationToInstAPIv1(),
			DataCentre:            spec.DataCentres[0].Region,
			DataCentreCustomName:  spec.DataCentres[0].Name,
			ClusterNetwork:        spec.DataCentres[0].Network,
		},
		Bundles: cadenceBundles,
	}, nil
}

func (spec *CadenceSpec) bundlesToInstAPIv1(ctx *context.Context, k8sClient client.Client) ([]*models.CadenceBundleAPIv1, error) {
	dataCentre := spec.DataCentres[0]

	var AWSAccessKeyID string
	var AWSSecretAccessKey string

	if dataCentre.EnableArchival {
		var awsCredsSecret *v1.Secret
		awsSecretNamespacedName := types.NamespacedName{Name: dataCentre.AWSAccessKeySecretName, Namespace: dataCentre.AWSAccessKeySecretNamespace}
		err := k8sClient.Get(*ctx, awsSecretNamespacedName, awsCredsSecret)
		if err != nil {
			return nil, err
		}

		AWSAccessKeyID = awsCredsSecret.StringData[models.AWSAccessKeyID]
		AWSSecretAccessKey = awsCredsSecret.StringData[models.AWSSecretAccessKey]
	}

	cadenceBundle := &models.CadenceBundleAPIv1{
		Bundle: models.Bundle{
			Bundle:  models.Cadence,
			Version: spec.Version,
		},
		Options: &models.CadenceBundleOptionsAPIv1{
			UseAdvancedVisibility:   dataCentre.UseAdvancedVisibility,
			UseCadenceWebAuth:       dataCentre.UseCadenceWebAuth,
			ClientEncryption:        dataCentre.ClientEncryption,
			TargetCassandraCDCID:    dataCentre.TargetCassandraCDCID,
			TargetCassandraVPCType:  dataCentre.TargetCassandraVPCType,
			TargetKafkaCDCID:        dataCentre.TargetKafkaCDCID,
			TargetKafkaVPCType:      dataCentre.TargetKafkaVPCType,
			TargetOpenSearchCDCID:   dataCentre.TargetOpenSearchCDCID,
			TargetOpenSearchVPCType: dataCentre.TargetOpenSearchVPCType,
			EnableArchival:          dataCentre.EnableArchival,
			ArchivalS3URI:           dataCentre.ArchivalS3URI,
			ArchivalS3Region:        dataCentre.ArchivalS3Region,
			AWSAccessKeyID:          AWSAccessKeyID,
			AWSSecretAccessKey:      AWSSecretAccessKey,
			CadenceNodeCount:        dataCentre.CadenceNodeCount,
			ProvisioningType:        dataCentre.ProvisioningType,
		},
	}
	cadenceBundles := []*models.CadenceBundleAPIv1{cadenceBundle}

	return cadenceBundles, nil
}

func (spec *CadenceSpec) rackAllocationToInstAPIv1() *models.RackAllocation {
	return &models.RackAllocation{
		NodesPerRack:  spec.DataCentres[0].NodesNumber,
		NumberOfRacks: spec.DataCentres[0].RacksNumber,
	}
}

func (spec *CadenceSpec) twoFactorDeleteToInstAPIv1() *models.TwoFactorDelete {
	if len(spec.TwoFactorDelete) < 1 {
		return nil
	}

	return &models.TwoFactorDelete{
		DeleteVerifyEmail: spec.TwoFactorDelete[0].Email,
		DeleteVerifyPhone: spec.TwoFactorDelete[0].Phone,
	}
}

func (cadence *Cadence) NewClusterMetadataPatch(
	ctx *context.Context,
	logger *logr.Logger,
) (*client.Patch, error) {
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
