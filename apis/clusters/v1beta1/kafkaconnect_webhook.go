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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var kafkaconnectlog = logf.Log.WithName("kafkaconnect-resource")

type kafkaConnectValidator struct {
	API validation.Validation
}

func (r *KafkaConnect) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&kafkaConnectValidator{
		API: api,
	})).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-kafkaconnect,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=kafkaconnects,verbs=create;update,versions=v1beta1,name=mkafkaconnect.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KafkaConnect{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KafkaConnect) Default() {
	kafkaconnectlog.Info("default", "name", r.Name)

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

	if len(kc.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
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

	if len(kc.Spec.CustomConnectors) > 1 {
		return fmt.Errorf("customConnectors array size must be between 0 and 1")
	}

	for _, dc := range kc.Spec.DataCentres {
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
