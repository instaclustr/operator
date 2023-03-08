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

// log is for logging in this package.
var cadencelog = logf.Log.WithName("cadence-resource")

func (r *Cadence) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1alpha1-cadence,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cadences,verbs=create;update,versions=v1alpha1,name=mcadence.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Cadence{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cadence) Default() {
	cadencelog.Info("default", "name", c.Name)

	for _, dataCentre := range c.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-cadence,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cadences,verbs=create;update,versions=v1alpha1,name=vcadence.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Cadence{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *Cadence) ValidateCreate() error {
	cadencelog.Info("validate create", "name", c.Name)

	err := c.Spec.Cluster.ValidateCreation(models.CadenceVersions)
	if err != nil {
		return err
	}

	if len(c.Spec.StandardProvisioning)+len(c.Spec.SharedProvisioning)+len(c.Spec.PackagedProvisioning) == 0 {
		return fmt.Errorf("one of StandardProvisioning, SharedProvisioning or PackagedProvisioning arrays must not be empty")
	}

	if len(c.Spec.StandardProvisioning)+len(c.Spec.SharedProvisioning)+len(c.Spec.PackagedProvisioning) > 1 {
		return fmt.Errorf("only one of StandardProvisioning, SharedProvisioning or PackagedProvisioning arrays must not be empty")
	}

	if len(c.Spec.AWSArchival) > 1 {
		return fmt.Errorf("AWSArchival array size must be between 0 and 1")
	}

	if len(c.Spec.StandardProvisioning) > 1 {
		return fmt.Errorf("StandardProvisioning array size must be between 0 and 1")
	}

	if len(c.Spec.SharedProvisioning) > 1 {
		return fmt.Errorf("SharedProvisioning array size must be between 0 and 1")
	}

	if len(c.Spec.PackagedProvisioning) > 1 {
		return fmt.Errorf("PackagedProvisioning array size must be between 0 and 1")
	}

	for _, awsArchival := range c.Spec.AWSArchival {
		err := awsArchival.validate()
		if err != nil {
			return err
		}
	}

	for _, sp := range c.Spec.StandardProvisioning {
		if len(sp.AdvancedVisibility) > 1 {
			return fmt.Errorf("AdvancedVisibility array size must be between 0 and 1")
		}

		err := sp.validate()
		if err != nil {
			return err
		}
	}

	for _, pp := range c.Spec.PackagedProvisioning {
		if (pp.UseAdvancedVisibility && pp.BundledKafkaSpec == nil) || (pp.UseAdvancedVisibility && pp.BundledOpenSearchSpec == nil) ||
			(pp.UseAdvancedVisibility && pp.BundledKafkaSpec == nil && pp.BundledOpenSearchSpec == nil) {
			return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must not be empty because UseAdvancedVisibility is set to true")
		}

		if pp.BundledKafkaSpec != nil {
			err := pp.BundledKafkaSpec.validate()
			if err != nil {
				return err
			}
		}

		if pp.BundledCassandraSpec != nil {
			err = pp.BundledCassandraSpec.validate()
			if err != nil {
				return err
			}
		}

		if pp.BundledOpenSearchSpec != nil {
			err = pp.BundledOpenSearchSpec.validate()
			if err != nil {
				return err
			}
		}
	}

	if len(c.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	for _, dc := range c.Spec.DataCentres {
		err := dc.DataCentre.ValidateCreation()
		if err != nil {
			return err
		}

	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *Cadence) ValidateUpdate(old runtime.Object) error {
	cadencelog.Info("validate update", "name", c.Name)

	oldCluster, ok := old.(*Cadence)
	if !ok {
		return models.ErrTypeAssertion
	}

	err := c.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cadence) ValidateDelete() error {
	cadencelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
