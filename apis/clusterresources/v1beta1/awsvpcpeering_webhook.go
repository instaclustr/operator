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
)

var awsvpcpeeringlog = logf.Log.WithName("awsvpcpeering-resource")

func (r *AWSVPCPeering) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-awsvpcpeering,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsvpcpeerings,verbs=create;update,versions=v1beta1,name=mawsvpcpeering.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AWSVPCPeering{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AWSVPCPeering) Default() {
	awsvpcpeeringlog.Info("default", "name", r.Name)

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-awsvpcpeering,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsvpcpeerings,verbs=create;update,versions=v1beta1,name=vawsvpcpeering.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSVPCPeering{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSVPCPeering) ValidateCreate() error {
	awsvpcpeeringlog.Info("validate create", "name", r.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(r.Spec)
	if err != nil {
		return err
	}

	if r.Spec.PeerAWSAccountID == "" {
		return fmt.Errorf("peer AWS Account ID is empty")
	}

	if r.Spec.PeerRegion == "" {
		return fmt.Errorf("peer AWS Account Region is empty")
	}

	if (r.Spec.DataCentreID == "" && r.Spec.ClusterRef == nil) ||
		(r.Spec.DataCentreID != "" && r.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: dataCentreId, clusterRef")
	}

	if r.Spec.PeerSubnets == nil {
		return fmt.Errorf("peer Subnets list is empty")
	}

	err = r.Spec.Validate(models.AWSRegions)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSVPCPeering) ValidateUpdate(old runtime.Object) error {
	awsvpcpeeringlog.Info("validate update", "name", r.Name)

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	err := r.Spec.ValidateUpdate(old.(*AWSVPCPeering).Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSVPCPeering) ValidateDelete() error {
	awsvpcpeeringlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
