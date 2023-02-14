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

var awsvpcpeeringlog = logf.Log.WithName("awsvpcpeering-resource")

func (aws *AWSVPCPeering) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(aws).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1alpha1-awsvpcpeering,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsvpcpeerings,verbs=create;update,versions=v1alpha1,name=vawsvpcpeering.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSVPCPeering{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSVPCPeering) ValidateCreate() error {
	awsvpcpeeringlog.Info("validate create", "name", aws.Name)

	if aws.Spec.PeerAWSAccountID == "" {
		return fmt.Errorf("peerAwsAccountId is empty")
	}
	if aws.Spec.PeerRegion == "" {
		return fmt.Errorf("peerRegion is empty")
	}
	if aws.Spec.DataCentreID == "" {
		return fmt.Errorf("cdcId is empty")
	}
	if aws.Spec.PeerSubnets == nil {
		return fmt.Errorf("peerSubnets is empty")
	}

	err := aws.Spec.Validate(models.AWSRegions)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSVPCPeering) ValidateUpdate(old runtime.Object) error {
	awsvpcpeeringlog.Info("validate update", "name", aws.Name)

	err := aws.Spec.ValidateUpdate(old.(*AWSVPCPeering).Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (aws *AWSVPCPeering) ValidateDelete() error {
	awsvpcpeeringlog.Info("validate delete", "name", aws.Name)

	return nil
}
