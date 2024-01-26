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
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var kafkaconnectlog = logf.Log.WithName("kafkaconnect-resource")

type kafkaConnectValidator struct {
	API    validation.Validation
	Client client.Client
}

func (r *KafkaConnect) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&kafkaConnectValidator{
		API:    api,
		Client: mgr.GetClient(),
	})).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-kafkaconnect,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=create;update,versions=v1beta1,name=mkafkaconnect.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KafkaConnect{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KafkaConnect) Default() {
	kafkaconnectlog.Info("default", "name", r.Name)

	if r.Spec.Name == "" {
		r.Spec.Name = r.Name
	}

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}

	for _, dataCentre := range r.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-kafkaconnect,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=create;update,versions=v1beta1,name=vkafkaconnect.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &kafkaConnectValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (kcv *kafkaConnectValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	kc, ok := obj.(*KafkaConnect)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka connect", obj.GetObjectKind())
	}

	kafkaconnectlog.Info("validate create", "name", kc.Name)

	err := kc.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	appVersions, err := kcv.API.ListAppVersions(models.KafkaConnectAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.KafkaConnectAppKind, err)
	}

	err = validateAppVersion(appVersions, models.KafkaConnectAppType, kc.Spec.Version)
	if err != nil {
		return err
	}

	if len(kc.Spec.TargetCluster) > 1 {
		return fmt.Errorf("targetCluster array size must be between 0 and 1")
	}

	for _, tc := range kc.Spec.TargetCluster {
		if len(tc.ManagedCluster) > 1 {
			return fmt.Errorf("managedCluster array size must be between 0 and 1")
		}

		if len(tc.ExternalCluster) > 1 {
			return fmt.Errorf("externalCluster array size must be between 0 and 1")
		}
		for _, mc := range tc.ManagedCluster {
			if (mc.TargetKafkaClusterID == "" && mc.ClusterRef == nil) ||
				(mc.TargetKafkaClusterID != "" && mc.ClusterRef != nil) {
				return fmt.Errorf("only one of dataCenter ID and cluster reference fields should be specified")
			}

			if mc.TargetKafkaClusterID != "" {
				clusterIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(mc.TargetKafkaClusterID))
				if !clusterIDMatched || err != nil {
					return fmt.Errorf("cluster ID is a UUID formated string. It must fit the pattern: %s, %v", models.UUIDStringRegExp, err)
				}
			}

			if !validation.Contains(mc.KafkaConnectVPCType, models.KafkaConnectVPCTypes) {
				return fmt.Errorf("kafkaConnect VPC type: %s is unavailable, available VPC types: %v",
					mc.KafkaConnectVPCType, models.KafkaConnectVPCTypes)
			}
		}
	}

	if len(kc.Spec.CustomConnectors) > 1 {
		return fmt.Errorf("customConnectors array size must be between 0 and 1")
	}

	if len(kc.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	for _, dc := range kc.Spec.DataCentres {
		//TODO: add support of multiple DCs for OnPrem clusters
		if len(kc.Spec.DataCentres) > 1 && dc.CloudProvider == models.ONPREMISES {
			return models.ErrOnPremicesWithMultiDC
		}

		err = dc.DataCentre.ValidateCreation()
		if err != nil {
			return err
		}

		err = validateReplicationFactor(models.KafkaConnectReplicationFactors, dc.ReplicationFactor)
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
func (kcv *kafkaConnectValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	kc, ok := new.(*KafkaConnect)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka connect", new.GetObjectKind())
	}

	kafkaconnectlog.Info("validate update", "name", kc.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if kc.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	if kc.Status.ID == "" {
		return kcv.ValidateCreate(ctx, kc)
	}

	oldCluster, ok := old.(*KafkaConnect)
	if !ok {
		return models.ErrTypeAssertion
	}

	err := kc.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	// ensuring if the cluster is ready for the spec updating
	if (kc.Status.CurrentClusterOperationStatus != models.NoOperation || kc.Status.State != models.RunningStatus) && kc.Generation != oldCluster.Generation {
		return models.ErrClusterIsNotReadyToUpdate
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (kcv *kafkaConnectValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	kc, ok := obj.(*KafkaConnect)
	if !ok {
		return fmt.Errorf("cannot assert object %v to kafka connect", obj.GetObjectKind())
	}

	kafkaconnectlog.Info("validate delete", "name", kc.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

type immutableKafkaConnectFields struct {
	immutableCluster
}

type immutableKafkaConnectDCFields struct {
	immutableDC       immutableDC
	ReplicationFactor int
}

func (kc *KafkaConnectSpec) newImmutableFields() *immutableKafkaConnectFields {
	return &immutableKafkaConnectFields{
		immutableCluster: kc.Cluster.newImmutableFields(),
	}
}

func (kc *KafkaConnectSpec) validateUpdate(oldSpec KafkaConnectSpec) error {
	newImmutableFields := kc.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldSpec, kc)
	}

	err := kc.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	err = kc.validateImmutableTargetClusterFieldsUpdate(kc.TargetCluster, oldSpec.TargetCluster)
	if err != nil {
		return err
	}

	err = validateTwoFactorDelete(kc.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	return nil
}

func (kc *KafkaConnectSpec) validateImmutableDataCentresFieldsUpdate(oldSpec KafkaConnectSpec) error {
	if len(kc.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range kc.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		if newDC.NodesNumber < oldDC.NodesNumber {
			return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldDC.NodesNumber)
		}

		if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
		}
	}

	return nil
}

func (kdc *KafkaConnectDataCentre) newImmutableFields() *immutableKafkaConnectDCFields {
	return &immutableKafkaConnectDCFields{
		immutableDC: immutableDC{
			Name:                kdc.Name,
			Region:              kdc.Region,
			CloudProvider:       kdc.CloudProvider,
			ProviderAccountName: kdc.ProviderAccountName,
			Network:             kdc.Network,
		},
		ReplicationFactor: kdc.ReplicationFactor,
	}
}

func (kc *KafkaConnectSpec) validateImmutableTargetClusterFieldsUpdate(new, old []*TargetCluster) error {
	if len(new) == 0 && len(old) == 0 {
		return models.ErrImmutableTargetCluster
	}

	if len(old) != len(new) {
		return models.ErrImmutableTargetCluster
	}

	for _, index := range new {
		for _, elem := range old {
			err := validateImmutableExternalClusterFields(index, elem)
			if err != nil {
				return err
			}

			err = validateImmutableManagedClusterFields(index, elem)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateImmutableExternalClusterFields(new, old *TargetCluster) error {
	for _, index := range new.ExternalCluster {
		for _, elem := range old.ExternalCluster {
			if *index != *elem {
				return models.ErrImmutableExternalCluster
			}
		}
	}
	return nil
}

func validateImmutableManagedClusterFields(new, old *TargetCluster) error {
	for _, nm := range new.ManagedCluster {
		for _, om := range old.ManagedCluster {
			if nm.TargetKafkaClusterID != om.TargetKafkaClusterID ||
				nm.KafkaConnectVPCType != om.KafkaConnectVPCType ||
				(nm.ClusterRef != nil && om.ClusterRef == nil) ||
				(nm.ClusterRef == nil && om.ClusterRef != nil) ||
				(nm.ClusterRef != nil && *nm.ClusterRef != *om.ClusterRef) {
				return models.ErrImmutableManagedCluster
			}
		}
	}
	return nil
}
