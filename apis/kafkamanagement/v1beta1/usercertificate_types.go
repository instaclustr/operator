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
	"crypto/x509/pkix"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserCertificateSpec defines the desired state of UserCertificateSpec
type UserCertificateSpec struct {
	// SecretRef references to the secret which stores pre-generated certificate request.
	// +kubebuilder:validation:XValidation:message="Cannot be changed after it is set",rule="self == oldSelf"
	SecretRef *FromSecret `json:"secretRef,omitempty"`

	// UserRef references to the KafkaUser resource to whom a certificate will be created.
	// +kubebuilder:validation:XValidation:message="Cannot be changed after it is set",rule="self == oldSelf"
	UserRef Reference `json:"userRef"`

	// ClusterRef references to the Kafka resource to whom a certificate will be created.
	// +kubebuilder:validation:XValidation:message="Cannot be changed after it is set",rule="self == oldSelf"
	ClusterRef Reference `json:"clusterRef"`

	// ValidPeriod is amount of month until a signed certificate is expired.
	// +kubebuilder:validation:Min=3
	// +kubebuilder:validation:Max=120
	// +kubebuilder:validation:XValidation:message="Cannot be changed after it is set",rule="self == oldSelf"
	ValidPeriod int `json:"validPeriod"`

	// CertificateRequestTemplate is a template for generating a CSR.
	// +kubebuilder:validation:XValidation:message="Cannot be changed after it is set",rule="self == oldSelf"
	CertificateRequestTemplate *CSRTemplate `json:"certificateRequestTemplate,omitempty"`
}

type CSRTemplate struct {
	Country            string `json:"country"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizationalUnit"`
}

func (c *CSRTemplate) ToSubject() pkix.Name {
	return pkix.Name{
		Country:            []string{c.Country},
		Organization:       []string{c.Organization},
		OrganizationalUnit: []string{c.OrganizationalUnit},
	}
}

type FromSecret struct {
	Reference `json:",inline"`
	Key       string `json:"key"`
}

type Reference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (r *Reference) AsNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

// UserCertificateStatus defines the observed state of UserCertificateStatus
type UserCertificateStatus struct {
	// CertID is a unique identifier of a certificate on Instaclustr.
	CertID string `json:"certId,omitempty"`

	// ExpiryDate is a date when a signed certificate is expired.
	ExpiryDate string `json:"expiryDate,omitempty"`

	// SignedCertSecretRef references to a secret which stores signed cert.
	SignedCertSecretRef Reference `json:"signedCertSecretRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UserCertificate is the Schema for the usercertificates API
type UserCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserCertificateSpec   `json:"spec,omitempty"`
	Status UserCertificateStatus `json:"status,omitempty"`
}

func (csr *UserCertificate) NewPatch() client.Patch {
	return client.MergeFrom(csr.DeepCopy())
}

//+kubebuilder:object:root=true

// UserCertificateList contains a list of UserCertificate
type UserCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserCertificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserCertificate{}, &UserCertificateList{})
}
