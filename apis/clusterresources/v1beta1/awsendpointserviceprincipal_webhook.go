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

	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
)

// log is for logging in this package.
var awsendpointserviceprincipallog = logf.Log.WithName("awsendpointserviceprincipal-resource")

func (r *AWSEndpointServicePrincipal) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-awsendpointserviceprincipal,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=awsendpointserviceprincipals,verbs=create;update,versions=v1beta1,name=vawsendpointserviceprincipal.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSEndpointServicePrincipal{}

var principalArnPattern, _ = regexp.Compile(`^arn:aws:iam::[0-9]{12}:(root$|user\/[\w+=,.@-]+|role\/[\w+=,.@-]+)$`)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSEndpointServicePrincipal) ValidateCreate() error {
	awsendpointserviceprincipallog.Info("validate create", "name", r.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(r.Spec)
	if err != nil {
		return err
	}

	if (r.Spec.ClusterDataCenterID == "" && r.Spec.ClusterRef == nil) ||
		(r.Spec.ClusterDataCenterID != "" && r.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: dataCentreId, clusterRef")
	}

	if !principalArnPattern.MatchString(r.Spec.PrincipalARN) {
		return fmt.Errorf("spec.principalArn doesn't match following pattern \"^arn:aws:iam::[0-9]{12}:(root$|user\\/[\\w+=,.@-]+|role\\/[\\w+=,.@-]+)$\"")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSEndpointServicePrincipal) ValidateUpdate(old runtime.Object) error {
	awsendpointserviceprincipallog.Info("validate update", "name", r.Name)

	oldResource := old.(*AWSEndpointServicePrincipal)

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	if r.Spec != oldResource.Spec {
		return fmt.Errorf("all fields in the spec are immutable")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSEndpointServicePrincipal) ValidateDelete() error {
	awsendpointserviceprincipallog.Info("validate delete", "name", r.Name)

	return nil
}
