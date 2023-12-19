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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
)

// log is for logging in this package.
var clusterbackuplog = logf.Log.WithName("clusterbackup-resource")

func (r *ClusterBackup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-clusterresources-instaclustr-com-v1beta1-clusterbackup,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=create;update,versions=v1beta1,name=mclusterbackup.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterBackup{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterBackup) Default() {
	clusterbackuplog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1beta1-clusterbackup,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=clusterbackups,verbs=create;update,versions=v1beta1,name=vclusterbackup.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterBackup{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBackup) ValidateCreate() error {
	clusterbackuplog.Info("validate create", "name", r.Name)

	_, ok := models.ClusterKindsMap[r.Spec.ClusterRef.ClusterKind]
	if !ok {
		return models.ErrUnsupportedBackupClusterKind
	}

	if r.Spec.ClusterRef.Name != "" && r.Spec.ClusterRef.Namespace == "" {
		return models.ErrEmptyNamespace
	}

	if r.Spec.ClusterRef.Namespace != "" && r.Spec.ClusterRef.Name == "" {
		return models.ErrEmptyName
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBackup) ValidateUpdate(old runtime.Object) error {
	clusterbackuplog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterBackup) ValidateDelete() error {
	clusterbackuplog.Info("validate delete", "name", r.Name)

	return nil
}
