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
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
)

// log is for logging in this package.

var opensearchegressruleslog = logf.Log.WithName("opensearchegressrules-resource")

var destinationTypes = []string{"SLACK", "WEBHOOK", "CUSTOM_WEBHOOK", "CHIME"}
var sourcePlugins = []string{"NOTIFICATIONS", "ALERTING"}

func (r *OpenSearchEgressRules) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-opensearchegressrules,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=opensearchegressrules,verbs=create;update,versions=v1beta1,name=vopensearchegressrules.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OpenSearchEgressRules{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OpenSearchEgressRules) ValidateCreate() error {
	opensearchegressruleslog.Info("validate create", "name", r.Name)
	matched, err := regexp.MatchString(models.OpenSearchBindingIDPattern, r.Spec.OpenSearchBindingID)
	if err != nil {
		return fmt.Errorf("can`t match openSearchBindingId to pattern: %s, error: %w", models.OpenSearchBindingIDPattern, err)
	}
	if !matched {
		return fmt.Errorf("mismatching openSearchBindingId to pattern: %s", models.OpenSearchBindingIDPattern)
	}

	if !slices.Contains(sourcePlugins, r.Spec.Source) {
		return fmt.Errorf("the source should be equal to one of the options: %q , got: %q", sourcePlugins, r.Spec.Source)
	}

	if !slices.Contains(destinationTypes, r.Spec.Type) {
		return fmt.Errorf("the type should be equal to one of the options: %q , got: %q", destinationTypes, r.Spec.Type)
	}

	if (r.Spec.ClusterID == "" && r.Spec.ClusterRef == nil) ||
		(r.Spec.ClusterID != "" && r.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: clusterId, clusterRef")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OpenSearchEgressRules) ValidateUpdate(old runtime.Object) error {
	opensearchegressruleslog.Info("validate update", "name", r.Name)

	oldRules := old.(*OpenSearchEgressRules)

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	if r.Spec != oldRules.Spec && r.Generation != oldRules.Generation {
		return models.ErrImmutableSpec
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OpenSearchEgressRules) ValidateDelete() error {
	opensearchegressruleslog.Info("validate delete", "name", r.Name)

	return nil
}
