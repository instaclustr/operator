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

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
	"github.com/instaclustr/operator/pkg/validation"
)

var clusternetworkfirewallrulelog = logf.Log.WithName("clusternetworkfirewallrule-resource")

func (fr *ClusterNetworkFirewallRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(fr).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-clusternetworkfirewallrule,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules,verbs=create;update,versions=v1beta1,name=mclusternetworkfirewallrule.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterNetworkFirewallRule{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterNetworkFirewallRule) Default() {
	clusternetworkfirewallrulelog.Info("default", "name", r.Name)

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-clusternetworkfirewallrule,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=clusternetworkfirewallrules,verbs=create;update,versions=v1beta1,name=vclusternetworkfirewallrule.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterNetworkFirewallRule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (fr *ClusterNetworkFirewallRule) ValidateCreate() error {
	clusternetworkfirewallrulelog.Info("validate create", "name", fr.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(fr.Spec)
	if err != nil {
		return err
	}

	if !validation.Contains(fr.Spec.Type, models.BundleTypes) {
		return fmt.Errorf("type  %s is unavailable, available types: %v",
			fr.Spec.Type, models.BundleTypes)
	}

	if (fr.Spec.ClusterID == "" && fr.Spec.ClusterRef == nil) ||
		(fr.Spec.ClusterID != "" && fr.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: clusterId, clusterRef")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (fr *ClusterNetworkFirewallRule) ValidateUpdate(old runtime.Object) error {
	clusternetworkfirewallrulelog.Info("validate update", "name", fr.Name)

	if fr.Status.ID == "" {
		return fr.ValidateCreate()
	}

	oldRule, ok := old.(*ClusterNetworkFirewallRule)
	if !ok {
		return models.ErrTypeAssertion
	}

	if fr.DeletionTimestamp == nil &&
		fr.Generation != oldRule.Generation {
		return fmt.Errorf("cluster network firewall rule resource is immutable")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (fr *ClusterNetworkFirewallRule) ValidateDelete() error {
	clusternetworkfirewallrulelog.Info("validate delete", "name", fr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
