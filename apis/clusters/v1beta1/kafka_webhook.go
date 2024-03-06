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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
	"github.com/instaclustr/operator/pkg/utils/slices"
	"github.com/instaclustr/operator/pkg/validation"
)

var kafkalog = logf.Log.WithName("kafka-resource")

type kafkaValidator struct {
	API    validation.Validation
	Client client.Client
}

func (r *Kafka) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&kafkaValidator{
		API:    api,
		Client: mgr.GetClient(),
	})).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-kafka,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkas,verbs=create;update,versions=v1beta1,name=mkafka.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Kafka{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *Kafka) Default() {
	kafkalog.Info("default", "name", k.Name)

	if k.Spec.Name == "" {
		k.Spec.Name = k.Name
	}

	if k.GetAnnotations() == nil {
		k.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-kafka,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkas,verbs=create;update,versions=v1beta1,name=vkafka.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &kafkaValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (kv *kafkaValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	k, ok := obj.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka", obj.GetObjectKind())
	}

	kafkalog.Info("validate create", "name", k.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(k.Spec)
	if err != nil {
		return err
	}

	err = k.Spec.GenericClusterSpec.ValidateCreation()
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

	for _, dc := range k.Spec.DataCentres {
		//TODO: add support of multiple DCs for OnPrem clusters
		if len(k.Spec.DataCentres) > 1 && dc.CloudProvider == models.ONPREMISES {
			return models.ErrOnPremisesWithMultiDC
		}
		err = dc.GenericDataCentreSpec.validateCreation()
		if err != nil {
			return err
		}

		err = validateReplicationFactor(models.KafkaReplicationFactors, k.Spec.ReplicationFactor)
		if err != nil {
			return err
		}

		if ((dc.NodesNumber*k.Spec.ReplicationFactor)/k.Spec.ReplicationFactor)%k.Spec.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", k.Spec.ReplicationFactor)
		}

		if !k.Spec.PrivateNetwork && dc.PrivateLink != nil {
			return models.ErrPrivateLinkOnlyWithPrivateNetworkCluster
		}

		if dc.CloudProvider != models.AWSVPC && dc.PrivateLink != nil {
			return models.ErrPrivateLinkSupportedOnlyForAWS
		}
	}

	for _, rs := range k.Spec.ResizeSettings {
		err = validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (kv *kafkaValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	k, ok := new.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka", new.GetObjectKind())
	}

	if k.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
		return nil
	}

	// skip validation when handle external changes from Instaclustr
	if k.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	if k.Status.ID == "" {
		return kv.ValidateCreate(ctx, k)
	}

	kafkalog.Info("validate update", "name", k.Name)

	oldKafka, ok := old.(*Kafka)
	if !ok {
		return fmt.Errorf("cannot assert object %v to Kafka", old.GetObjectKind())
	}

	err := k.Spec.validateUpdate(&oldKafka.Spec)
	if err != nil {
		return fmt.Errorf("cannot update, error: %v", err)
	}

	for _, rs := range k.Spec.ResizeSettings {
		err := validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	if IsClusterNotReadyForSpecUpdate(k.Status.CurrentClusterOperationStatus, k.Status.State, k.Generation, oldKafka.Generation) {
		return models.ErrClusterIsNotReadyToUpdate
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

	if len(ks.DataCentres) != len(old.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for _, dc := range ks.DataCentres {
		if ((dc.NodesNumber*ks.ReplicationFactor)/ks.ReplicationFactor)%ks.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", ks.ReplicationFactor)
		}
	}

	err := ks.validateImmutableDataCentresFieldsUpdate(old)
	if err != nil {
		return err
	}

	if err := validateTwoFactorDelete(ks.TwoFactorDelete, old.TwoFactorDelete); err != nil {
		return err
	}

	if !slices.EqualsPtr(ks.SchemaRegistry, old.SchemaRegistry) {
		return models.ErrImmutableSchemaRegistry
	}
	if !slices.EqualsPtr(ks.KarapaceSchemaRegistry, old.KarapaceSchemaRegistry) {
		return models.ErrImmutableKarapaceSchemaRegistry
	}
	if !slices.EqualsPtr(ks.RestProxy, old.RestProxy) {
		return models.ErrImmutableRestProxy
	}
	if !slices.EqualsPtr(ks.Kraft, old.Kraft) {
		return models.ErrImmutableKraft
	}
	if !slices.EqualsPtr(ks.KarapaceRestProxy, old.KarapaceRestProxy) {
		return models.ErrImmutableKarapaceRestProxy
	}
	if ok := validateZookeeperUpdate(ks.DedicatedZookeeper, old.DedicatedZookeeper); !ok {
		return models.ErrImmutableDedicatedZookeeper
	}

	return nil
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

func isPrivateLinkValid(new, old []*PrivateLink) bool {
	if new == nil && old == nil {
		return true
	}

	if len(new) != len(old) {
		return false
	}

	for i := range new {
		if new[i].AdvertisedHostname != old[i].AdvertisedHostname {
			return false
		}
	}
	return true
}

func (ks *KafkaSpec) validateImmutableDataCentresFieldsUpdate(oldSpec *KafkaSpec) error {
	if len(ks.DataCentres) < len(oldSpec.DataCentres) {
		return models.ErrDecreasedDataCentresNumber
	}

	toValidate := map[string]*KafkaDataCentre{}
	for _, dc := range oldSpec.DataCentres {
		toValidate[dc.Name] = dc
	}

	for _, newDC := range ks.DataCentres {
		oldDC, ok := toValidate[newDC.Name]
		if !ok {
			return fmt.Errorf("cannot change datacentre name: %v", newDC.Name)
		}

		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		if newDC.NodesNumber < oldDC.NodesNumber {
			return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldDC.NodesNumber)
		}

		err := newDC.validateImmutableCloudProviderSettingsUpdate(&oldDC.GenericDataCentreSpec)
		if err != nil {
			return err
		}

		err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		if ok = isPrivateLinkValid(newDC.PrivateLink, oldDC.PrivateLink); !ok {
			return fmt.Errorf("advertisedHostname field cannot be changed")
		}
	}

	return nil
}

type immutableKafkaFields struct {
	specificFields specificKafkaFields
	cluster        immutableCluster
}

type specificKafkaFields struct {
	replicationFactor         int
	partitionsNumber          int
	allowDeleteTopics         bool
	autoCreateTopics          bool
	clientToClusterEncryption bool
	bundledUseOnly            bool
	privateNetworkCluster     bool
	clientBrokerAuthWithMtls  bool
	pciCompliance             bool
}

func (ks *KafkaSpec) newKafkaImmutableFields() *immutableKafkaFields {
	return &immutableKafkaFields{
		specificFields: specificKafkaFields{
			replicationFactor:         ks.ReplicationFactor,
			partitionsNumber:          ks.PartitionsNumber,
			allowDeleteTopics:         ks.AllowDeleteTopics,
			autoCreateTopics:          ks.AutoCreateTopics,
			clientToClusterEncryption: ks.ClientToClusterEncryption,
			bundledUseOnly:            ks.BundledUseOnly,
			privateNetworkCluster:     ks.PrivateNetwork,
			clientBrokerAuthWithMtls:  ks.ClientBrokerAuthWithMTLS,
			pciCompliance:             ks.PCICompliance,
		},
		cluster: ks.GenericClusterSpec.immutableFields(),
	}
}

func (ksdc *KafkaDataCentre) newImmutableFields() *immutableKakfaDCFields {
	return &immutableKakfaDCFields{
		immutableDC{
			Name:                ksdc.Name,
			Region:              ksdc.Region,
			CloudProvider:       ksdc.CloudProvider,
			ProviderAccountName: ksdc.ProviderAccountName,
			Network:             ksdc.Network,
		},
	}
}

type immutableKakfaDCFields struct {
	immutableDC
}
