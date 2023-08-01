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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
)

var awsencryptionkeylog = logf.Log.WithName("awsencryptionkey-resource")

func (aws *AWSEncryptionKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(aws).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-awsencryptionkey,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsencryptionkeys,verbs=create;update,versions=v1beta1,name=mawsencryptionkey.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AWSEncryptionKey{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (aws *AWSEncryptionKey) Default() {
	awsencryptionkeylog.Info("default", "name", aws.Name)

	if aws.GetAnnotations() == nil {
		aws.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-awsencryptionkey,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsencryptionkeys,verbs=create;update,versions=v1beta1,name=vawsencryptionkey.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSEncryptionKey{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSEncryptionKey) ValidateCreate() error {
	awsencryptionkeylog.Info("validate create", "name", aws.Name)

	aliasMatched, err := regexp.Match(models.EncryptionKeyAliasRegExp, []byte(aws.Spec.Alias))
	if !aliasMatched || err != nil {
		return fmt.Errorf("AWS Encryption key alias must fit the pattern: %s, %v", models.EncryptionKeyAliasRegExp, err)
	}

	if len(aws.Spec.Alias) < 1 || len(aws.Spec.Alias) > 30 {
		return fmt.Errorf("AWS Encryption key alias must have 1 to 30 characters: %v", err)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSEncryptionKey) ValidateUpdate(old runtime.Object) error {
	awsencryptionkeylog.Info("validate update", "name", aws.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSEncryptionKey) ValidateDelete() error {
	awsencryptionkeylog.Info("validate delete", "name", aws.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
