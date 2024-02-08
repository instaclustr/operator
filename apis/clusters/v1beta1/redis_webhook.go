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
	"github.com/instaclustr/operator/pkg/validation"
)

var redislog = logf.Log.WithName("redis-resource")

type redisValidator struct {
	API    validation.Validation
	Client client.Client
}

func (r *Redis) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&redisValidator{
		API:    api,
		Client: mgr.GetClient(),
	})).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-redis,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=redis,verbs=create;update,versions=v1beta1,name=mredis.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Redis{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Redis) Default() {
	redislog.Info("default", "name", r.Name)

	if r.Spec.Name == "" {
		r.Spec.Name = r.Name
		redislog.Info("default values are set", "name", r.Name)
	}

	if r.GetAnnotations() == nil {
		r.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}

	for _, dataCentre := range r.Spec.DataCentres {
		if dataCentre.MasterNodes != 0 {
			if dataCentre.ReplicationFactor > 0 {
				dataCentre.ReplicaNodes = dataCentre.MasterNodes * dataCentre.ReplicationFactor
			} else {
				dataCentre.ReplicationFactor = dataCentre.ReplicaNodes / dataCentre.MasterNodes
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-redis,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=redis,verbs=create;update,versions=v1beta1,name=vredis.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &redisValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (rv *redisValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	r, ok := obj.(*Redis)
	if !ok {
		return fmt.Errorf("cannot assert object %v to redis", obj.GetObjectKind())
	}

	redislog.Info("validate create", "name", r.Name)

	if r.Spec.RestoreFrom != nil {
		if r.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := r.Spec.GenericClusterSpec.ValidateCreation()
	if err != nil {
		return err
	}

	contains, err := ContainsKubeVirtAddon(ctx, rv.Client)
	if err != nil {
		return err
	}

	if r.Spec.OnPremisesSpec != nil && r.Spec.OnPremisesSpec.EnableAutomation {
		if !contains {
			return models.ErrKubeVirtAddonNotFound
		}
		err = r.Spec.OnPremisesSpec.ValidateCreation()
		if err != nil {
			return err
		}
		if r.Spec.PrivateNetwork {
			err = r.Spec.OnPremisesSpec.ValidateSSHGatewayCreation()
			if err != nil {
				return err
			}
		}
	}

	err = r.Spec.ValidatePrivateLink()
	if err != nil {
		return err
	}

	appVersions, err := rv.API.ListAppVersions(models.RedisAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.RedisAppKind, err)
	}

	err = validateAppVersion(appVersions, models.RedisAppType, r.Spec.Version)
	if err != nil {
		return err
	}

	if len(r.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	if len(r.Spec.DataCentres) > 1 {
		if r.Spec.OnPremisesSpec != nil {
			return models.ErrOnPremicesWithMultiDC
		}
		return models.ErrCreateClusterWithMultiDC
	}

	for _, dc := range r.Spec.DataCentres {
		if r.Spec.OnPremisesSpec != nil {
			err = dc.GenericDataCentreSpec.ValidateOnPremisesCreation()
			if err != nil {
				return err
			}
		} else {
			err = dc.ValidateCreate()
			if err != nil {
				return err
			}
		}
	}

	for _, rs := range r.Spec.ResizeSettings {
		err = validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (rv *redisValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	r, ok := new.(*Redis)
	if !ok {
		return fmt.Errorf("cannot assert object %v to redis", new.GetObjectKind())
	}

	redislog.Info("validate update", "name", r.Name)

	if r.Annotations[models.ResourceStateAnnotation] == models.CreatingEvent {
		return nil
	}

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if r.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldRedis, ok := old.(*Redis)
	if !ok {
		return models.ErrTypeAssertion
	}

	if r.Status.ID == "" {
		return rv.ValidateCreate(ctx, r)
	}

	if oldRedis.Spec.RestoreFrom != nil {
		return nil
	}

	err := r.Spec.ValidateUpdate(oldRedis.Spec)
	if err != nil {
		return fmt.Errorf("update validation error: %v", err)
	}

	for _, rs := range r.Spec.ResizeSettings {
		err := validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	// ensuring if the cluster is ready for the spec updating
	if (r.Status.CurrentClusterOperationStatus != models.NoOperation || r.Status.State != models.RunningStatus) && r.Generation != oldRedis.Generation {
		return models.ErrClusterIsNotReadyToUpdate
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (rv *redisValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	r, ok := obj.(*Redis)
	if !ok {
		return fmt.Errorf("cannot assert object %v to redis", obj.GetObjectKind())
	}

	redislog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

type immutableRedisFields struct {
	specificRedisFields
	immutableCluster
}

type specificRedisFields struct {
	ClientEncryption    bool
	PasswordAndUserAuth bool
}

type immutableRedisDCFields struct {
	immutableDC
	ReplicationFactor int
}

func (rs *RedisSpec) ValidateUpdate(oldSpec RedisSpec) error {
	newImmutableFields := rs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable spec fields: old spec: %+v: new spec: %+v",
			oldImmutableFields, newImmutableFields)
	}

	err := validateTwoFactorDelete(rs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = rs.validateDCsUpdate(oldSpec)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RedisSpec) validateDCsUpdate(oldSpec RedisSpec) error {
	// contains old DCs which should be validated
	toValidate := map[string]*RedisDataCentre{}
	for _, dc := range oldSpec.DataCentres {
		toValidate[dc.Name] = dc
	}

	for _, newDC := range rs.DataCentres {
		oldDC, ok := toValidate[newDC.Name]
		if !ok {
			// validating creation of the new dc
			if err := newDC.ValidateCreate(); err != nil {
				return err
			}
			continue
		}

		// validating updating of the DC
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		err := validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		err = newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		if newDC.MasterNodes < oldDC.MasterNodes {
			return fmt.Errorf("deleting nodes is not supported. Master nodes number must be greater than: %v", oldDC.MasterNodes)
		}

		if newDC.ReplicaNodes < oldDC.ReplicaNodes {
			return fmt.Errorf("deleting nodes is not supported. Number of replica nodes must be greater than: %v", oldDC.ReplicaNodes)
		}

		err = validatePrivateLinkUpdate(newDC.PrivateLink, oldDC.PrivateLink)
		if err != nil {
			return err
		}

		// deleting validated DCs from the map to ensure if it is validated
		delete(toValidate, oldDC.Name)
	}

	// ensuring if all old DCs were validated
	if len(toValidate) > 0 {
		return models.ErrUnsupportedDeletingDC
	}

	return nil
}

func (rs *RedisSpec) newImmutableFields() *immutableRedisFields {
	return &immutableRedisFields{
		specificRedisFields: specificRedisFields{
			ClientEncryption:    rs.ClientEncryption,
			PasswordAndUserAuth: rs.PasswordAndUserAuth,
		},
		immutableCluster: rs.GenericClusterSpec.immutableFields(),
	}
}

func (rdc *RedisDataCentre) newImmutableFields() *immutableRedisDCFields {
	return &immutableRedisDCFields{
		immutableDC: immutableDC{
			Name:                rdc.Name,
			Region:              rdc.Region,
			CloudProvider:       rdc.CloudProvider,
			ProviderAccountName: rdc.ProviderAccountName,
			Network:             rdc.Network,
		},
		ReplicationFactor: rdc.ReplicationFactor,
	}
}

func (rdc *RedisDataCentre) ValidateUpdate() error {
	err := rdc.ValidateNodesNumber()
	if err != nil {
		return err
	}

	return nil
}

func (rdc *RedisDataCentre) ValidateCreate() error {
	err := rdc.GenericDataCentreSpec.validateCreation()
	if err != nil {
		return err
	}

	err = rdc.ValidateNodesNumber()
	if err != nil {
		return err
	}

	err = rdc.ValidatePrivateLink()
	if err != nil {
		return err
	}

	return nil
}

func (rdc *RedisDataCentre) ValidateNodesNumber() error {
	if rdc.ReplicaNodes < 0 || rdc.ReplicaNodes > 100 {
		return fmt.Errorf("replica nodes number should not be less than 0 or more than 100")
	}

	if rdc.MasterNodes < 3 || rdc.MasterNodes > 100 {
		return fmt.Errorf("master nodes should not be less than 3 or more than 100")
	}

	return nil
}

func (rdc *RedisDataCentre) ValidatePrivateLink() error {
	if rdc.CloudProvider != models.AWSVPC && len(rdc.PrivateLink) > 0 {
		return models.ErrPrivateLinkSupportedOnlyForAWS
	}

	return nil
}

func (rs *RedisSpec) ValidatePrivateLink() error {
	if len(rs.DataCentres) > 1 &&
		(rs.DataCentres[0].PrivateLink != nil || rs.DataCentres[1].PrivateLink != nil) {
		return models.ErrPrivateLinkSupportedOnlyForSingleDC
	}

	return nil
}
