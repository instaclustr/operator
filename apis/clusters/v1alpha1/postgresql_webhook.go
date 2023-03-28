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
	"github.com/instaclustr/operator/pkg/validation"
)

var postgresqllog = logf.Log.WithName("postgresql-resource")

func (r *PostgreSQL) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1alpha1-postgresql,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=postgresqls,verbs=create;update,versions=v1alpha1,name=mpostgresql.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1alpha1-postgresql,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=postgresqls,verbs=create;update,versions=v1alpha1,name=vpostgresql.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &PostgreSQL{}
var _ webhook.Defaulter = &PostgreSQL{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (pg *PostgreSQL) Default() {
	for _, dataCentre := range pg.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

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

	err := pg.Spec.Cluster.ValidateCreation(models.PostgreSQLVersions)
	if err != nil {
		return err
	}

	if len(pg.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	for _, dc := range pg.Spec.DataCentres {
		err = dc.DataCentre.ValidateCreation()
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
		if !validation.Contains(dc.IntraDataCentreReplication[0].ReplicationMode, models.ReplicationModes) {
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

	if r.DeletionTimestamp != nil &&
		(len(r.Spec.TwoFactorDelete) != 0 &&
			r.Annotations[models.DeletionConfirmed] != models.True) {
		return nil
	}

	oldCluster, ok := old.(*PostgreSQL)
	if !ok {
		return models.ErrTypeAssertion
	}

	if oldCluster.Spec.PgRestoreFrom != nil {
		return nil
	}

	if r.Status.ID == "" {
		return r.ValidateCreate()
	}

	err := r.Spec.ValidateImmutableFieldsUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("immutable fields validation error: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PostgreSQL) ValidateDelete() error {
	postgresqllog.Info("validate delete", "name", r.Name)

	return nil
}
