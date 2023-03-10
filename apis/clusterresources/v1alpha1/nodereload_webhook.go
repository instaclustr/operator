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
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
)

var nodereloadlog = logf.Log.WithName("nodereload-resource")

func (nr *NodeReload) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(nr).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusterresources-instaclustr-com-v1alpha1-nodereload,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusterresources.instaclustr.com,resources=nodereloads,verbs=create;update,versions=v1alpha1,name=vnodereload.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NodeReload{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (nr *NodeReload) ValidateCreate() error {
	nodereloadlog.Info("validate create", "name", nr.Name)

	if nr.Spec.Nodes == nil {
		return fmt.Errorf("nodes list is empty")
	}

	for _, node := range nr.Spec.Nodes {
		nodeIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(node.ID))
		if !nodeIDMatched || err != nil {
			return fmt.Errorf("node ID is a UUID formated string. It must fit the pattern: %s. %v", models.UUIDStringRegExp, err)
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (nr *NodeReload) ValidateUpdate(old runtime.Object) error {
	nodereloadlog.Info("validate update", "name", nr.Name)

	if nr.Status.NodeInProgress.ID == "" {
		return nr.ValidateCreate()
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (nr *NodeReload) ValidateDelete() error {
	nodereloadlog.Info("validate delete", "name", nr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
