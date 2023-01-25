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

var redislog = logf.Log.WithName("redis-resource")

func (r *Redis) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-redis,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=redis,verbs=create;update,versions=v1alpha1,name=vredis.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Redis{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateCreate() error {
	redislog.Info("validate create", "name", r.Name)

	if r.Spec.RestoreFrom != nil {
		if r.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := r.Spec.Cluster.ValidateCreation(models.RedisVersions)
	if err != nil {
		return err
	}

	if len(r.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	for _, dc := range r.Spec.DataCentres {
		err := dc.DataCentre.ValidateCreation()
		if err != nil {
			return err
		}

		if dc.ReplicaNodes < 0 || dc.ReplicaNodes > 100 {
			return fmt.Errorf("replica nodes should not be less than 0 or more than 100")
		}

		if dc.MasterNodes < 3 || dc.MasterNodes > 100 {
			return fmt.Errorf("master nodes should not be less than 3 or more than 100")
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateUpdate(old runtime.Object) error {
	redislog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Redis) ValidateDelete() error {
	redislog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
