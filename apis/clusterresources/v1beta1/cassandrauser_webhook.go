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
var cassandrauserlog = logf.Log.WithName("cassandrauser-resource")

func (r *CassandraUser) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-cassandrauser,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=cassandrausers,verbs=create;update,versions=v1beta1,name=vcassandrauser.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CassandraUser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (u *CassandraUser) ValidateCreate() error {
	cassandrauserlog.Info("validate create", "name", u.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(u.Spec)
	if err != nil {
		return err
	}

	if u.Spec.SecretRef.Name == "" || u.Spec.SecretRef.Namespace == "" {
		return models.ErrEmptySecretRef
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (u *CassandraUser) ValidateUpdate(old runtime.Object) error {
	cassandrauserlog.Info("validate update", "name", u.Name)

	oldUser := old.(*CassandraUser)
	if *u.Spec.SecretRef != *oldUser.Spec.SecretRef {
		return models.ErrImmutableSecretRef
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CassandraUser) ValidateDelete() error {
	cassandrauserlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
