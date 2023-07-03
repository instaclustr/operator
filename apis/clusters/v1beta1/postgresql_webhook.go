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

var postgresqllog = logf.Log.WithName("postgresql-resource")

type pgValidator struct {
	API validation.Validation
}

func (r *PostgreSQL) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&pgValidator{
		API: api,
	})).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-postgresql,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=postgresqls,verbs=create;update,versions=v1beta1,name=mpostgresql.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-postgresql,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=postgresqls,verbs=create;update,versions=v1beta1,name=vpostgresql.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &pgValidator{}
var _ webhook.Defaulter = &PostgreSQL{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (pg *PostgreSQL) Default() {
	for _, dataCentre := range pg.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (pgv *pgValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	pg, ok := obj.(*PostgreSQL)
	if !ok {
		return fmt.Errorf("cannot assert object %v to postgreSQL", obj.GetObjectKind())
	}

	postgresqllog.Info("validate create", "name", pg.Name)

	if pg.Spec.PgRestoreFrom != nil {
		if pg.Spec.PgRestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := pg.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	appVersions, err := pgv.API.ListAppVersions(models.PgAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.PgAppKind, err)
	}

	err = validateAppVersion(appVersions, models.PgAppType, pg.Spec.Version)
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
func (pgv *pgValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	pg, ok := new.(*PostgreSQL)
	if !ok {
		return fmt.Errorf("cannot assert object %v to postgreSQL", new.GetObjectKind())
	}

	postgresqllog.Info("validate update", "name", pg.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if pg.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster, ok := old.(*PostgreSQL)
	if !ok {
		return models.ErrTypeAssertion
	}

	if oldCluster.Spec.PgRestoreFrom != nil {
		return nil
	}

	if pg.Status.ID == "" {
		return pgv.ValidateCreate(ctx, pg)
	}

	err := pg.Spec.ValidateImmutableFieldsUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("immutable fields validation error: %v", err)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (pgv *pgValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	pg, ok := obj.(*PostgreSQL)
	if !ok {
		return fmt.Errorf("cannot assert object %v to postgreSQL", obj.GetObjectKind())
	}

	postgresqllog.Info("validate delete", "name", pg.Name)

	return nil
}
