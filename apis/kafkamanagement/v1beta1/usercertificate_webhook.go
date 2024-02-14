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
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
)

// log is for logging in this package.
var usercertificatelog = logf.Log.WithName("usercertificate-resource")

func (r *UserCertificate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-kafkamanagement-instaclustr-com-v1beta1-usercertificate,mutating=false,failurePolicy=fail,sideEffects=None,groups=kafkamanagement.instaclustr.com,resources=usercertificates,verbs=create;update,versions=v1beta1,name=vusercertificate.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &UserCertificate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cert *UserCertificate) ValidateCreate() error {
	usercertificatelog.Info("validate create", "name", cert.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(cert.Spec)
	if err != nil {
		return err
	}

	if cert.Spec.SecretRef == nil && cert.Spec.CertificateRequestTemplate == nil {
		return errors.New("one of the following fields should be set: spec.secretRef, spec.generateCSR")
	}

	if cert.Spec.SecretRef != nil && cert.Spec.CertificateRequestTemplate != nil {
		return errors.New("only one of the following fields can be set: spec.secretRef, spec.generateCSR")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cert *UserCertificate) ValidateUpdate(old runtime.Object) error {
	usercertificatelog.Info("validate update", "name", cert.Name)

	oldCert := old.(*UserCertificate)

	if oldCert.Spec.SecretRef != nil && cert.Spec.SecretRef == nil {
		return errors.New("spec.secretRef cannot be removed after it is set")
	}

	if oldCert.Spec.CertificateRequestTemplate != nil && cert.Spec.CertificateRequestTemplate == nil {
		return errors.New("spec.certificateRequestTemplate cannot be removed after it is set")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cert *UserCertificate) ValidateDelete() error {
	usercertificatelog.Info("validate delete", "name", cert.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
