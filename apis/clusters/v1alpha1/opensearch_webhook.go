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
)

var opensearchlog = logf.Log.WithName("opensearch-resource")

func (r *OpenSearch) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-opensearch,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=opensearches,verbs=create;update,versions=v1alpha1,name=vopensearch.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1alpha1-opensearch,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=opensearches,verbs=create;update,versions=v1alpha1,name=mopensearch.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OpenSearch{}
var _ webhook.Defaulter = &OpenSearch{}

func (os *OpenSearch) Default() {
	if os.Spec.ConcurrentResizes == 0 {
		os.Spec.ConcurrentResizes = 1
	}

	for _, dataCentre := range os.Spec.DataCentres {
		dataCentre.SetDefaultValues()

		if dataCentre.Name == "" {
			dataCentre.Name = dataCentre.Region
		}
	}

	opensearchlog.Info("default values are set", "name", os.Name)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (os *OpenSearch) ValidateCreate() error {
	opensearchlog.Info("validate create", "name", os.Name)

	if os.Spec.RestoreFrom != nil {
		if os.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := os.Spec.ValidateCreation(models.OpenSearchVersions)
	if err != nil {
		return err
	}

	if len(os.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	for _, dataCentre := range os.Spec.DataCentres {
		err = dataCentre.ValidateCreation()
		if err != nil {
			return err
		}

		if dataCentre.DedicatedMasterNodes &&
			dataCentre.MasterNodeSize == "" {
			return fmt.Errorf("master node size field is empty")
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (os *OpenSearch) ValidateUpdate(old runtime.Object) error {
	opensearchlog.Info("validate update", "name", os.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if os.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster := old.(*OpenSearch)
	if oldCluster.Spec.RestoreFrom != nil {
		return nil
	}

	if os.Status.ID == "" {
		return os.ValidateCreate()
	}

	err := os.Spec.ValidateImmutableFieldsUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("immutable fields validation error: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (os *OpenSearch) ValidateDelete() error {
	opensearchlog.Info("validate delete", "name", os.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
