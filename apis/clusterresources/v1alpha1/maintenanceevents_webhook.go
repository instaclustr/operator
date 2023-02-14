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
)

var maintenanceeventslog = logf.Log.WithName("maintenanceevents-resource")

func (r *MaintenanceEvents) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1alpha1-maintenanceevents,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=maintenanceevents,verbs=create;update,versions=v1alpha1,name=vmaintenanceevents.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MaintenanceEvents{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (me *MaintenanceEvents) ValidateCreate() error {
	maintenanceeventslog.Info("validate create", "name", me.Name)

	if err := me.Spec.ValidateMaintenanceEventsReschedules(); err != nil {
		return fmt.Errorf("maintenanceEvents reschedules validation failed: %v", err)
	}

	if err := me.Spec.ValidateExclusionWindows(); err != nil {
		return fmt.Errorf("exclusion windows validation failed: %v", err)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (me *MaintenanceEvents) ValidateUpdate(old runtime.Object) error {
	maintenanceeventslog.Info("validate update", "name", me.Name)

	if me.DeletionTimestamp != nil {
		return nil
	}

	if err := me.Spec.ValidateMaintenanceEventsReschedules(); err != nil {
		return fmt.Errorf("maintenanceEvents reschedules validation failed: %v", err)
	}

	if err := me.Spec.ValidateExclusionWindows(); err != nil {
		return fmt.Errorf("exclusionWindows validation failed: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (me *MaintenanceEvents) ValidateDelete() error {
	maintenanceeventslog.Info("validate delete", "name", me.Name)

	return nil
}
