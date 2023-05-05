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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var kafkalog = logf.Log.WithName("kafka-resource")

type kafkaValidator struct {
	API validation.Validation
}

func (r *Kafka) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&kafkaValidator{
		API: api,
	})).
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

var _ webhook.CustomValidator = &kafkaValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (kv *kafkaValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	k, ok := obj.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka", obj.GetObjectKind())
	}

	kafkalog.Info("validate create", "name", k.Name)

	err := k.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	appVersions, err := kv.API.ListAppVersions(models.KafkaAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.KafkaAppKind, err)
	}

	err = validateAppVersion(appVersions, models.KafkaAppType, k.Spec.Version)
	if err != nil {
		return err
	}

	if len(k.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (kv *kafkaValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	k, ok := new.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka", new.GetObjectKind())
	}

	kafkalog.Info("validate update", "name", k.Name)

	if k.Status.ID == "" {
		return kv.ValidateCreate(ctx, k)
	}

	// skip validation when handle external changes from Instaclustr
	if k.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

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
func (kv *kafkaValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	k, ok := obj.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka", obj.GetObjectKind())
	}

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

	if !isKafkaAddonsEqual[SchemaRegistry](ks.SchemaRegistry, old.SchemaRegistry) {
		return models.ErrImmutableSchemaRegistry
	}
	if !isKafkaAddonsEqual[KarapaceSchemaRegistry](ks.KarapaceSchemaRegistry, old.KarapaceSchemaRegistry) {
		return models.ErrImmutableKarapaceSchemaRegistry
	}
	if !isKafkaAddonsEqual[RestProxy](ks.RestProxy, old.RestProxy) {
		return models.ErrImmutableRestProxy
	}
	if !isKafkaAddonsEqual[KarapaceRestProxy](ks.KarapaceRestProxy, old.KarapaceRestProxy) {
		return models.ErrImmutableKarapaceRestProxy
	}
	if ok := validateZookeeperUpdate(ks.DedicatedZookeeper, old.DedicatedZookeeper); !ok {
		return models.ErrImmutableDedicatedZookeeper
	}

	return nil
}

func isKafkaAddonsEqual[T KafkaAddons](new, old []*T) bool {
	if new == nil && old == nil {
		return true
	}

	if len(new) != len(old) {
		return false
	}

	for i := range new {
		if *new[i] != *old[i] {
			return false
		}
	}

	return true
}

func validateZookeeperUpdate(new, old []*DedicatedZookeeper) bool {
	if new == nil && old == nil {
		return true
	}

	if len(new) != len(old) {
		return false
	}

	for i := range new {
		if new[i].NodesNumber != old[i].NodesNumber {
			return false
		}
	}

	return true
}

type immutableKafkaFields struct {
	specificFields specificKafkaFields
	cluster        immutableCluster
}

type specificKafkaFields struct {
	replicationFactorNumber   int
	partitionsNumber          int
	allowDeleteTopics         bool
	autoCreateTopics          bool
	clientToClusterEncryption bool
	bundledUseOnly            bool
}

func (ks *KafkaSpec) newKafkaImmutableFields() *immutableKafkaFields {
	return &immutableKafkaFields{
		specificFields: specificKafkaFields{
			replicationFactorNumber:   ks.ReplicationFactorNumber,
			partitionsNumber:          ks.PartitionsNumber,
			allowDeleteTopics:         ks.AllowDeleteTopics,
			autoCreateTopics:          ks.AutoCreateTopics,
			clientToClusterEncryption: ks.ClientToClusterEncryption,
			bundledUseOnly:            ks.BundledUseOnly,
		},
		cluster: ks.Cluster.newImmutableFields(),
	}
}
