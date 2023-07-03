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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var cassandralog = logf.Log.WithName("cassandra-resource")

type cassandraValidator struct {
	API validation.Validation
}

func (r *Cassandra) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&cassandraValidator{
		API: api,
	})).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-cassandra,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cassandras,verbs=create;update,versions=v1beta1,name=vcassandra.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-cassandra,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cassandras,verbs=create;update,versions=v1beta1,name=vcassandra.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &cassandraValidator{}
var _ webhook.Defaulter = &Cassandra{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cassandra) Default() {
	for _, dataCentre := range c.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", obj.GetObjectKind())
	}

	cassandralog.Info("validate create", "name", c.Name)

	if c.Spec.RestoreFrom != nil {
		if c.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := c.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	if len(c.Spec.Spark) > 1 {
		return fmt.Errorf("spark should not have more than 1 item")
	}

	appVersions, err := cv.API.ListAppVersions(models.CassandraAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.CassandraAppKind, err)
	}

	err = validateAppVersion(appVersions, models.CassandraAppType, c.Spec.Version)
	if err != nil {
		return err
	}

	for _, spark := range c.Spec.Spark {
		err = validateAppVersion(appVersions, models.SparkAppType, spark.Version)
		if err != nil {
			return err
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

		if !c.Spec.PrivateNetworkCluster && dc.PrivateIPBroadcastForDiscovery {
			return fmt.Errorf("cannot use private ip broadcast for discovery on public network cluster")
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	c, ok := new.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", new.GetObjectKind())
	}

	cassandralog.Info("validate update", "name", c.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if c.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster, ok := old.(*Cassandra)
	if !ok {
		return models.ErrTypeAssertion
	}

	if oldCluster.Spec.RestoreFrom != nil {
		return nil
	}

	if c.Status.ID == "" {
		return cv.ValidateCreate(ctx, c)
	}

	err := c.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", obj.GetObjectKind())
	}

	cassandralog.Info("validate delete", "name", c.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
