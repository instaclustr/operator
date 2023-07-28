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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var opensearchuserlog = logf.Log.WithName("opensearchuser-resource")

func (u *OpenSearchUser) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(u).
		Complete()
}

//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-opensearchuser,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=opensearchusers,verbs=create;update,versions=v1beta1,name=vopensearchuser.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OpenSearchUser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (u *OpenSearchUser) ValidateCreate() error {
	opensearchuserlog.Info("validate create", "name", u.Name)

	if u.Spec.SecretRef.Name == "" || u.Spec.SecretRef.Namespace == "" {
		return fmt.Errorf("secretRef.name and secretRef.namespace should not be empty")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (u *OpenSearchUser) ValidateUpdate(old runtime.Object) error {
	opensearchuserlog.Info("validate update", "name", u.Name)

	oldUser := old.(*OpenSearchUser)
	if *u.Spec.SecretRef != *oldUser.Spec.SecretRef {
		return fmt.Errorf("spec.secretRef field is immutable")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (u *OpenSearchUser) ValidateDelete() error {
	return nil
}
