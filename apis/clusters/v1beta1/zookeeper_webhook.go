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

var zookeeperlog = logf.Log.WithName("zookeeper-resource")

type zookeeperValidator struct {
	API validation.Validation
}

func (z *Zookeeper) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(z).WithValidator(webhook.CustomValidator(&zookeeperValidator{
		API: api,
	})).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-zookeeper,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=zookeepers,verbs=create;update,versions=v1beta1,name=mzookeeper.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Zookeeper{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (z *Zookeeper) Default() {
	for _, dataCentre := range z.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-zookeeper,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=zookeepers,verbs=create;update,versions=v1beta1,name=vzookeeper.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &zookeeperValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (zv *zookeeperValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	z, ok := obj.(*Zookeeper)
	if !ok {
		return fmt.Errorf("cannot assert object %v to zookeeper", obj.GetObjectKind())
	}

	zookeeperlog.Info("validate create", "name", z.Name)

	err := z.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	appVersions, err := zv.API.ListAppVersions(models.ZookeeperAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.ZookeeperClusterKind, err)
	}

	err = validateAppVersion(appVersions, models.ZookeeperAppType, z.Spec.Version)
	if err != nil {
		return err
	}

	if len(z.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	for _, dc := range z.Spec.DataCentres {
		err = dc.DataCentre.ValidateCreation()
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (zv *zookeeperValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	z, ok := new.(*Zookeeper)
	if !ok {
		return fmt.Errorf("cannot assert object %v to zookeeper", new.GetObjectKind())
	}

	zookeeperlog.Info("validate update", "name", z.Name)

	if z.Status.ID == "" {
		return zv.ValidateCreate(ctx, z)
	}

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if z.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (zv *zookeeperValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	z, ok := obj.(*Zookeeper)
	if !ok {
		return fmt.Errorf("cannot assert object %v to zookeeper", obj.GetObjectKind())
	}

	zookeeperlog.Info("validate delete", "name", z.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
