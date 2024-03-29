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

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
)

var gcpvpcpeeringlog = logf.Log.WithName("gcpvpcpeering-resource")

func (r *GCPVPCPeering) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-gcpvpcpeering,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings,verbs=create;update,versions=v1beta1,name=mgcpvpcpeering.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &GCPVPCPeering{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GCPVPCPeering) Default() {
	gcpvpcpeeringlog.Info("default", "name", r.Name)

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-gcpvpcpeering,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=gcpvpcpeerings,verbs=create;update,versions=v1beta1,name=vgcpvpcpeering.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GCPVPCPeering{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GCPVPCPeering) ValidateCreate() error {
	gcpvpcpeeringlog.Info("validate create", "name", r.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(r.Spec)
	if err != nil {
		return err
	}

	if r.Spec.PeerVPCNetworkName == "" {
		return fmt.Errorf("peer VPC Network Name is empty")
	}

	if r.Spec.PeerProjectID == "" {
		return fmt.Errorf("peer Project ID is empty")
	}

	if (r.Spec.DataCentreID == "" && r.Spec.ClusterRef == nil) ||
		(r.Spec.DataCentreID != "" && r.Spec.ClusterRef != nil) {
		return fmt.Errorf("only one of the following fields should be specified: dataCentreId, clusterRef")
	}

	if r.Spec.PeerSubnets == nil {
		return fmt.Errorf("peer Subnets list is empty")
	}

	err = r.Spec.Validate()
	if err != nil {
		return err
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GCPVPCPeering) ValidateUpdate(old runtime.Object) error {
	gcpvpcpeeringlog.Info("validate update", "name", r.Name)

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GCPVPCPeering) ValidateDelete() error {
	gcpvpcpeeringlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (gcp *GCPVPCPeeringSpec) Validate() error {
	dataCentreIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(gcp.DataCentreID))
	if err != nil {
		return err
	}
	if !dataCentreIDMatched {
		return fmt.Errorf("data centre ID is a UUID formated string. It must fit the pattern: %s", models.UUIDStringRegExp)
	}

	for _, subnet := range gcp.PeerSubnets {
		peerSubnetMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(subnet))
		if err != nil {
			return err
		}
		if !peerSubnetMatched {
			return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28", subnet)
		}
	}

	return nil
}
