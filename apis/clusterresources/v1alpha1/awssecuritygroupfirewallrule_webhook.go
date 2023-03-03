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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var awssecuritygroupfirewallrulelog = logf.Log.WithName("awssecuritygroupfirewallrule-resource")

func (r *AWSSecurityGroupFirewallRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1alpha1-awssecuritygroupfirewallrule,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awssecuritygroupfirewallrules,verbs=create;update,versions=v1alpha1,name=vawssecuritygroupfirewallrule.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSSecurityGroupFirewallRule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecurityGroupFirewallRule) ValidateCreate() error {
	awssecuritygroupfirewallrulelog.Info("validate create", "name", r.Name)

	if !validation.Contains(r.Spec.Type, models.BundleTypes) {
		return fmt.Errorf("type %s is unavailable, available types: %v",
			r.Spec.Type, models.BundleTypes)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecurityGroupFirewallRule) ValidateUpdate(old runtime.Object) error {
	awssecuritygroupfirewallrulelog.Info("validate update", "name", r.Name)

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	oldRule, ok := old.(*AWSSecurityGroupFirewallRule)
	if !ok {
		return fmt.Errorf("cannot assert object %v to AWSSecurityGroupFirewallRule", old.GetObjectKind())
	}

	if r.DeletionTimestamp == nil &&
		r.Generation != oldRule.Generation {
		return models.ErrImmutableAWSSecurityGroupFirewallRule
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecurityGroupFirewallRule) ValidateDelete() error {
	awssecuritygroupfirewallrulelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
