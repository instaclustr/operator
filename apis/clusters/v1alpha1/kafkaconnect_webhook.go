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
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

// log is for logging in this package.
var kafkaconnectlog = logf.Log.WithName("kafkaconnect-resource")

func (r *KafkaConnect) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1alpha1-kafkaconnect,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=create;update,versions=v1alpha1,name=mkafkaconnect.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KafkaConnect{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KafkaConnect) Default() {
	kafkaconnectlog.Info("default", "name", r.Name)

	for _, dataCentre := range r.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-kafkaconnect,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=create;update,versions=v1alpha1,name=vkafkaconnect.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KafkaConnect{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KafkaConnect) ValidateCreate() error {
	kafkaconnectlog.Info("validate create", "name", k.Name)

	err := k.Spec.Cluster.ValidateCreation(models.KafkaConnectVersions)
	if err != nil {
		return err
	}

	if len(k.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	if len(k.Spec.TargetCluster) > 1 {
		return fmt.Errorf("targetCluster array size must be between 0 and 1")
	}

	for _, tc := range k.Spec.TargetCluster {
		if len(tc.ManagedCluster) > 1 {
			return fmt.Errorf("managedCluster array size must be between 0 and 1")
		}

		if len(tc.ExternalCluster) > 1 {
			return fmt.Errorf("externalCluster array size must be between 0 and 1")
		}
		for _, mc := range tc.ManagedCluster {
			clusterIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(mc.TargetKafkaClusterID))
			if !clusterIDMatched || err != nil {
				return fmt.Errorf("cluster ID is a UUID formated string. It must fit the pattern: %s, %v", models.UUIDStringRegExp, err)
			}

			if !validation.Contains(mc.KafkaConnectVPCType, models.KafkaConnectVPCTypes) {
				return fmt.Errorf("kafkaConnect VPC type: %s is unavailable, available VPC types: %v",
					mc.KafkaConnectVPCType, models.KafkaConnectVPCTypes)
			}
		}
	}

	if len(k.Spec.CustomConnectors) > 1 {
		return fmt.Errorf("customConnectors array size must be between 0 and 1")
	}

	for _, dc := range k.Spec.DataCentres {
		err := dc.ValidateCreation()
		if err != nil {
			return err
		}

		if ((dc.NodesNumber*dc.ReplicationFactor)/dc.ReplicationFactor)%dc.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", dc.ReplicationFactor)
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KafkaConnect) ValidateUpdate(old runtime.Object) error {
	kafkaconnectlog.Info("validate update", "name", k.Name)

	if k.DeletionTimestamp != nil &&
		(len(k.Spec.TwoFactorDelete) != 0 &&
			k.Annotations[models.DeletionConfirmed] != models.True) {
		return nil
	}

	if k.Status.ID == "" {
		return k.ValidateCreate()
	}

	oldCluster, ok := old.(*KafkaConnect)
	if !ok {
		return models.ErrTypeAssertion
	}

	err := k.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KafkaConnect) ValidateDelete() error {
	kafkaconnectlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
