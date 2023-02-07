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

var kafkalog = logf.Log.WithName("kafka-resource")

func (r *Kafka) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1alpha1-kafka,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkas,verbs=create;update,versions=v1alpha1,name=mkafka.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Kafka{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *Kafka) Default() {
	kafkalog.Info("default", "name", k.Name)

	for _, dataCentre := range k.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-kafka,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkas,verbs=create;update,versions=v1alpha1,name=vkafka.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Kafka{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateCreate() error {
	kafkalog.Info("validate create", "name", k.Name)

	err := k.Spec.Cluster.ValidateCreation(models.KafkaVersions)
	if err != nil {
		return err
	}

	if len(k.Spec.DataCentres) != 1 {
		return models.ErrMultipleDataCentres
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateUpdate(old runtime.Object) error {
	kafkalog.Info("validate update", "name", k.Name)

	oldKafka, ok := old.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to Kafka", old.GetObjectKind())
	}

	err := k.Spec.validateUpdate(&oldKafka.Spec)
	if err != nil {
		return fmt.Errorf("cannot update, error: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *Kafka) ValidateDelete() error {
	kafkalog.Info("validate delete", "name", k.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (ks *KafkaSpec) validateUpdate(old *KafkaSpec) error {
	newImmut := ks.newKafkaImmutableFields()
	oldImmut := old.newKafkaImmutableFields()

	if newImmut.cluster != oldImmut.cluster ||
		newImmut.specificFields != oldImmut.specificFields {
		return fmt.Errorf("immutable fields have been changed, old spec: %+v: new spec: %+v", oldImmut, newImmut)
	}

	if err := validateTwoFactorDelete(ks.TwoFactorDelete, old.TwoFactorDelete); err != nil {
		return err
	}
	if err := validateSchemaRegistryUpdate(ks.SchemaRegistry, old.SchemaRegistry); err != nil {
		return err
	}
	if err := validateKarapaceSchemaRegistryUpdate(ks.KarapaceSchemaRegistry, old.KarapaceSchemaRegistry); err != nil {
		return err
	}
	if err := validateRestProxyUpdate(ks.RestProxy, old.RestProxy); err != nil {
		return err
	}
	if err := validateKarapaceRestProxyUpdate(ks.KarapaceRestProxy, old.KarapaceRestProxy); err != nil {
		return err
	}
	if err := validateZookeeperUpdate(ks.DedicatedZookeeper, old.DedicatedZookeeper); err != nil {
		return err
	}

	return nil
}

func validateSchemaRegistryUpdate(new, old []*SchemaRegistry) error {
	if new == nil && old == nil {
		return nil
	}

	if len(new) != len(old) {
		return models.ErrImmutableSchemaRegistry
	}

	for i := range new {
		if *new[i] != *old[i] {
			return models.ErrImmutableSchemaRegistry
		}
	}

	return nil
}

func validateRestProxyUpdate(new, old []*RestProxy) error {
	if new == nil && old == nil {
		return nil
	}

	if len(new) != len(old) {
		return models.ErrImmutableRestProxy
	}

	for i := range new {
		if *new[i] != *old[i] {
			return models.ErrImmutableRestProxy
		}
	}

	return nil
}

func validateKarapaceSchemaRegistryUpdate(new, old []*KarapaceSchemaRegistry) error {
	if new == nil && old == nil {
		return nil
	}

	if len(new) != len(old) {
		return models.ErrImmutableKarapaceSchemaRegistry
	}

	for i := range new {
		if *new[i] != *old[i] {
			return models.ErrImmutableKarapaceSchemaRegistry
		}
	}

	return nil
}

func validateKarapaceRestProxyUpdate(new, old []*KarapaceRestProxy) error {
	if new == nil && old == nil {
		return nil
	}

	if len(new) != len(old) {
		return models.ErrImmutableKarapaceRestProxy
	}

	for i := range new {
		if *new[i] != *old[i] {
			return models.ErrImmutableKarapaceRestProxy
		}
	}

	return nil
}

func validateZookeeperUpdate(new, old []*DedicatedZookeeper) error {
	if new == nil && old == nil {
		return nil
	}

	if len(new) != len(old) {
		return models.ErrImmutableDedicatedZookeeper
	}

	for i := range new {
		if new[i].NodesNumber != old[i].NodesNumber {
			return models.ErrImmutableZookeeperNodeNumber
		}
	}

	return nil
}

type immutableKafkaFields struct {
	specificFields specificKafkaFields
	cluster        immutableCluster
}

type specificKafkaFields struct {
	replicationFactorNumber   int32
	partitionsNumber          int32
	allowDeleteTopics         bool
	autoCreateTopics          bool
	clientToClusterEncryption bool
}

func (ks *KafkaSpec) newKafkaImmutableFields() *immutableKafkaFields {
	return &immutableKafkaFields{
		specificFields: specificKafkaFields{
			replicationFactorNumber:   ks.ReplicationFactorNumber,
			partitionsNumber:          ks.PartitionsNumber,
			allowDeleteTopics:         ks.AllowDeleteTopics,
			autoCreateTopics:          ks.AutoCreateTopics,
			clientToClusterEncryption: ks.ClientToClusterEncryption,
		},
		cluster: ks.Cluster.newImmutableFields(),
	}
}
