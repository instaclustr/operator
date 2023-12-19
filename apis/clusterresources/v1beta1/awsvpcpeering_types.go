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
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

// AWSVPCPeeringSpec defines the desired state of AWSVPCPeering
type AWSVPCPeeringSpec struct {
	PeeringSpec      `json:",inline"`
	PeerAWSAccountID string `json:"peerAwsAccountId"`
	PeerVPCID        string `json:"peerVpcId"`
	PeerRegion       string `json:"peerRegion,omitempty"`
}

// AWSVPCPeeringStatus defines the observed state of AWSVPCPeering
type AWSVPCPeeringStatus struct {
	PeeringStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="StatusCode",type="string",JSONPath=".status.statusCode"

// AWSVPCPeering is the Schema for the awsvpcpeerings API
type AWSVPCPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSVPCPeeringSpec   `json:"spec,omitempty"`
	Status AWSVPCPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSVPCPeeringList contains a list of AWSVPCPeering
type AWSVPCPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSVPCPeering `json:"items"`
}

func (aws *AWSVPCPeering) GetJobID(jobName string) string {
	return aws.Kind + "/" + client.ObjectKeyFromObject(aws).String() + "/" + jobName
}

func (aws *AWSVPCPeering) NewPatch() client.Patch {
	old := aws.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&AWSVPCPeering{}, &AWSVPCPeeringList{})
}

type immutableAWSVPCPeeringFields struct {
	specificFields specificAWSVPCPeeringFields
	peering        immutablePeeringFields
}

type specificAWSVPCPeeringFields struct {
	peerAWSAccountID string
	peerRegion       string
}

func (aws *AWSVPCPeeringSpec) newImmutableFields() *immutableAWSVPCPeeringFields {
	return &immutableAWSVPCPeeringFields{
		specificAWSVPCPeeringFields{
			peerAWSAccountID: aws.PeerAWSAccountID,
			peerRegion:       aws.PeerRegion,
		},
		immutablePeeringFields{
			DataCentreID: aws.DataCentreID,
		},
	}
}

func (aws *AWSVPCPeeringSpec) ValidateUpdate(oldSpec AWSVPCPeeringSpec) error {
	newImmutableFields := aws.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	err := aws.Validate(models.AWSRegions)
	if err != nil {
		return err
	}

	if newImmutableFields.peering != oldImmutableFields.peering ||
		newImmutableFields.specificFields != oldImmutableFields.specificFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldSpec, aws)
	}

	return nil
}

func (aws *AWSVPCPeeringSpec) Validate(availableRegions []string) error {
	peerAWSAccountIDMatched, err := regexp.Match(models.PeerAWSAccountIDRegExp, []byte(aws.PeerAWSAccountID))
	if !peerAWSAccountIDMatched || err != nil {
		return fmt.Errorf("AWS Account ID to peer should contain 12-digit number, that uniquely identifies an AWS account and fit pattern: %s. %v", models.PeerAWSAccountIDRegExp, err)
	}

	peerAWSVPCIDMatched, err := regexp.Match(models.PeerVPCIDRegExp, []byte(aws.PeerVPCID))
	if !peerAWSVPCIDMatched || err != nil {
		return fmt.Errorf("VPC ID must begin with 'vpc-' and fit pattern: %s. %v", models.PeerVPCIDRegExp, err)
	}

	if aws.DataCentreID != "" {
		dataCentreIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(aws.DataCentreID))
		if !dataCentreIDMatched || err != nil {
			return fmt.Errorf("data centre ID is a UUID formated string. It must fit the pattern: %s. %v", models.UUIDStringRegExp, err)
		}
	}

	if !validation.Contains(aws.PeerRegion, availableRegions) {
		return fmt.Errorf("AWS Region to peer: %s is unavailable, available regions: %v",
			aws.PeerRegion, availableRegions)
	}

	for _, subnet := range aws.PeerSubnets {
		peerSubnetMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(subnet))
		if !peerSubnetMatched || err != nil {
			return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", subnet, err)
		}
	}

	return nil
}
