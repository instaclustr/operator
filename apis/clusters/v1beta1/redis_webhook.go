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

var redislog = logf.Log.WithName("redis-resource")

type redisValidator struct {
	API validation.Validation
}

func (r *Redis) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&redisValidator{
		API: api,
	})).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-redis,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=redis,verbs=create;update,versions=v1beta1,name=mredis.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Redis{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Redis) Default() {
	for _, dataCentre := range r.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}

	redislog.Info("default values are set", "name", r.Name)

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

	err := r.Spec.Cluster.ValidateCreation()
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

	for _, dc := range r.Spec.DataCentres {
		err = dc.ValidateCreate()
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
