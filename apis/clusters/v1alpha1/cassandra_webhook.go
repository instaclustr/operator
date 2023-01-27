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

var cassandralog = logf.Log.WithName("cassandra-resource")

func (r *Cassandra) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-cassandra,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cassandras,verbs=create;update,versions=v1alpha1,name=vcassandra.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Cassandra{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *Cassandra) ValidateCreate() error {
	cassandralog.Info("validate create", "name", c.Name)

	if c.Spec.RestoreFrom != nil {
		if c.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := c.Spec.Cluster.Validate(models.CassandraVersions)
	if err != nil {
		return err
	}

	if len(c.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	for _, dc := range c.Spec.DataCentres {
		err := dc.DataCentre.Validate()
		if err != nil {
			return err
		}

		if !c.Spec.PrivateNetworkCluster && dc.PrivateIPBroadcastForDiscovery {
			return fmt.Errorf("cannot use private ip broadcast for discovery on public network cluster")
		}
	}

	if len(c.Spec.Spark) > 1 {
		return fmt.Errorf("spark should not have more than 1 item")
	}
	if len(c.Spec.Spark) == 1 {
		if !Contains(c.Spec.Spark[0].Version, models.SparkVersions) {
			return fmt.Errorf("cluster spark version %s is unavailable, available versions: %v",
				c.Spec.Spark[0].Version, models.SparkVersions)
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *Cassandra) ValidateUpdate(old runtime.Object) error {
	cassandralog.Info("validate update", "name", c.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (c *Cassandra) ValidateDelete() error {
	cassandralog.Info("validate delete", "name", c.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}