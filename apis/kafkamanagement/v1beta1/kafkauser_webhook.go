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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
)

// log is for logging in this package.
var kafkauserlog = logf.Log.WithName("kafkauser-resource")

func (r *KafkaUser) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kafkamanagement-instaclustr-com-v1beta1-kafkauser,mutating=true,failurePolicy=fail,sideEffects=None,groups=kafkamanagement.instaclustr.com,resources=kafkausers,verbs=create;update,versions=v1beta1,name=mkafkauser.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KafkaUser{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KafkaUser) Default() {
	kafkauserlog.Info("default", "name", r.Name)

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kafkamanagement-instaclustr-com-v1beta1-kafkauser,mutating=false,failurePolicy=fail,sideEffects=None,groups=kafkamanagement.instaclustr.com,resources=kafkausers,verbs=create;update,versions=v1beta1,name=vkafkauser.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KafkaUser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ku *KafkaUser) ValidateCreate() error {
	kafkauserlog.Info("validate create", "name", ku.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(ku.Spec)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ku *KafkaUser) ValidateUpdate(old runtime.Object) error {
	kafkauserlog.Info("validate update", "name", ku.Name)

	oldUser := old.(*KafkaUser)
	if *ku.Spec.SecretRef != *oldUser.Spec.SecretRef {
		return models.ErrImmutableSecretRef
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ku *KafkaUser) ValidateDelete() error {
	kafkauserlog.Info("validate delete", "name", ku.Name)

	return nil
}
