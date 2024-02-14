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
var topiclog = logf.Log.WithName("topic-resource")

func (r *Topic) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kafkamanagement-instaclustr-com-v1beta1-topic,mutating=true,failurePolicy=fail,sideEffects=None,groups=kafkamanagement.instaclustr.com,resources=topics,verbs=create;update,versions=v1beta1,name=mtopic.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Topic{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Topic) Default() {
	topiclog.Info("default", "name", r.Name)

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

//+kubebuilder:webhook:path=/validate-kafkamanagement-instaclustr-com-v1beta1-topic,mutating=false,failurePolicy=fail,sideEffects=None,groups=kafkamanagement.instaclustr.com,resources=topics,verbs=create;update,versions=v1beta1,name=vtopic.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Topic{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Topic) ValidateCreate() error {
	topiclog.Info("validate create", "name", r.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(r.Spec)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Topic) ValidateUpdate(old runtime.Object) error {
	topiclog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Topic) ValidateDelete() error {
	topiclog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
