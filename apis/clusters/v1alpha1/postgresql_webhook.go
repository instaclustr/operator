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

// log is for logging in this package.
var postgresqllog = logf.Log.WithName("postgresql-resource")

func (r *PostgreSQL) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-postgresql,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=postgresqls,verbs=create;update,versions=v1alpha1,name=vpostgresql.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &PostgreSQL{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (pg *PostgreSQL) ValidateCreate() error {
	postgresqllog.Info("validate create", "name", pg.Name)

	if pg.Spec.PgRestoreFrom != nil {
		if pg.Spec.PgRestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := pg.Spec.Cluster.Validate(models.PostgreSQLVersions)
	if err != nil {
		return err
	}

	if len(pg.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	for _, dc := range pg.Spec.DataCentres {
		err = dc.DataCentre.Validate()
		if err != nil {
			return err
		}

		err = dc.ValidatePGBouncer()
		if err != nil {
			return err
		}

		if len(pg.Spec.DataCentres) > 1 &&
			len(dc.InterDataCentreReplication) == 0 {
			return fmt.Errorf("interDataCentreReplication field is required when more than 1 data centre")
		}

		if len(dc.IntraDataCentreReplication) != 1 {
			return fmt.Errorf("intraDataCentreReplication required to have 1 item")
		}
		if !Contains(dc.IntraDataCentreReplication[0].ReplicationMode, models.ReplicationModes) {
			return fmt.Errorf("replicationMode '%s' is unavailable, available values: %v",
				dc.IntraDataCentreReplication[0].ReplicationMode,
				models.ReplicationModes)
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgreSQL) ValidateUpdate(old runtime.Object) error {
	postgresqllog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PostgreSQL) ValidateDelete() error {
	postgresqllog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
