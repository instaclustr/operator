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
	"github.com/instaclustr/operator/pkg/validation"
)

// log is for logging in this package.
var exclusionwindowlog = logf.Log.WithName("exclusionwindow-resource")

func (r *ExclusionWindow) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-exclusionwindow,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=exclusionwindows,verbs=create;update,versions=v1beta1,name=mexclusionwindow.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ExclusionWindow{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ExclusionWindow) Default() {
	exclusionwindowlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-exclusionwindow,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=exclusionwindows,verbs=create;update,versions=v1beta1,name=vexclusionwindow.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ExclusionWindow{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ExclusionWindow) ValidateCreate() error {
	exclusionwindowlog.Info("validate create", "name", r.Name)

	if (r.Spec.ClusterID == "" && r.Spec.ClusterRef == nil) ||
		(r.Spec.ClusterID != "" && r.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: clusterId, clusterRef")
	}

	if !validation.Contains(r.Spec.DayOfWeek, models.DaysOfWeek) {
		return fmt.Errorf("%v, available values: %v",
			models.ErrIncorrectDayOfWeek, models.DaysOfWeek)
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ExclusionWindow) ValidateUpdate(old runtime.Object) error {
	exclusionwindowlog.Info("validate update", "name", r.Name)
	oldWindow := old.(*ExclusionWindow)

	if !r.Spec.validateUpdate(oldWindow.Spec) {
		return models.ErrImmutableSpec
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ExclusionWindow) ValidateDelete() error {
	exclusionwindowlog.Info("validate delete", "name", r.Name)
	return nil
}

func (e *ExclusionWindowSpec) validateUpdate(old ExclusionWindowSpec) bool {
	if e.DayOfWeek != old.DayOfWeek ||
		e.ClusterID != old.ClusterID ||
		e.DurationInHours != old.DurationInHours ||
		e.StartHour != old.StartHour ||
		(e.ClusterRef != nil && old.ClusterRef == nil) ||
		(e.ClusterRef == nil && old.ClusterRef != nil) ||
		(e.ClusterRef != nil && *e.ClusterRef != *old.ClusterRef) {
		return false
	}

	return true
}
