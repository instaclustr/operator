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
	"unicode"

	k8sCore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/requiredfieldsvalidator"
	"github.com/instaclustr/operator/pkg/utils/slices"
	"github.com/instaclustr/operator/pkg/validation"
)

var postgresqllog = logf.Log.WithName("postgresql-resource")

type pgValidator struct {
	API       validation.Validation
	K8sClient client.Client
}

func (r *PostgreSQL) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&pgValidator{
		API:       api,
		K8sClient: mgr.GetClient(),
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
	postgresqllog.Info("default", "name", pg.Name)

	if pg.Spec.Name == "" {
		pg.Spec.Name = pg.Name
	}

	if pg.GetAnnotations() == nil {
		pg.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (pgv *pgValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	pg, ok := obj.(*PostgreSQL)
	if !ok {
		return fmt.Errorf("cannot assert object %v to postgreSQL", obj.GetObjectKind())
	}

	postgresqllog.Info("validate create", "name", pg.Name)

	err := requiredfieldsvalidator.ValidateRequiredFields(pg.Spec)
	if err != nil {
		return err
	}

	if pg.Spec.PgRestoreFrom != nil {
		if pg.Spec.PgRestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err = pg.Spec.GenericClusterSpec.ValidateCreation()
	if err != nil {
		return err
	}

	if pg.Spec.UserRefs != nil {
		err = pgv.validatePostgreSQLUsers(ctx, pg)
		if err != nil {
			return err
		}
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
		//TODO: add support of multiple DCs for OnPrem clusters
		if len(pg.Spec.DataCentres) > 1 && dc.CloudProvider == models.ONPREMISES {
			return models.ErrOnPremisesWithMultiDC
		}

		err = dc.GenericDataCentreSpec.validateCreation()
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

	for _, rs := range pg.Spec.ResizeSettings {
		err = validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
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

	if pg.Annotations[models.ResourceStateAnnotation] == models.SyncingEvent {
		return nil
	}

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

	if pg.Spec.UserRefs != nil {
		err := pgv.validatePostgreSQLUsers(ctx, pg)
		if err != nil {
			return err
		}
	}

	err := pg.Spec.ValidateImmutableFieldsUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("immutable fields validation error: %v", err)
	}

	for _, rs := range pg.Spec.ResizeSettings {
		err := validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	// ensuring if the cluster is ready for the spec updating
	if (pg.Status.CurrentClusterOperationStatus != models.NoOperation || pg.Status.State != models.RunningStatus) && pg.Generation != oldCluster.Generation {
		return models.ErrClusterIsNotReadyToUpdate
	}

	return nil
}

func (pgv *pgValidator) validatePostgreSQLUsers(ctx context.Context, pg *PostgreSQL) error {
	nodeList := &k8sCore.NodeList{}

	err := pgv.K8sClient.List(ctx, nodeList, &client.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot list node list, error: %v", err)
	}

	var externalIPExists bool
	for _, node := range nodeList.Items {
		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == k8sCore.NodeExternalIP {
				externalIPExists = true
				break
			}
		}
	}

	if !externalIPExists || pg.Spec.PrivateNetwork {
		return fmt.Errorf("cannot create PostgreSQL user, if your cluster is private or has no external ips " +
			"you need to configure peering and remove user references from cluster specification")
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

type immutablePostgreSQLFields struct {
	specificPostgreSQLFields
	immutableCluster
}

type specificPostgreSQLFields struct {
	SynchronousModeStrict bool
}

type immutablePostgreSQLDCFields struct {
	immutableDC
	specificPostgreSQLDC
}

type specificPostgreSQLDC struct {
	ClientEncryption bool
	NumberOfNodes    int
}

func (pg *PostgreSQL) ValidateDefaultUserPassword(password string) bool {
	categories := map[string]bool{}
	for _, symbol := range password {
		switch {
		case unicode.IsNumber(symbol):
			categories["number"] = true
		case unicode.IsUpper(symbol):
			categories["upper"] = true
		case unicode.IsLower(symbol):
			categories["lower"] = true
		case unicode.IsPunct(symbol) || unicode.IsSymbol(symbol):
			categories["special"] = true
		}
	}

	return len(categories) > 2
}

func (pdc *PgDataCentre) ValidatePGBouncer() error {
	for _, pgb := range pdc.PGBouncer {
		if !validation.Contains(pgb.PGBouncerVersion, models.PGBouncerVersions) {
			return fmt.Errorf("pgBouncerVersion '%s' is unavailable, available versions: %v",
				pgb.PGBouncerVersion,
				models.PGBouncerVersions)
		}

		if !validation.Contains(pgb.PoolMode, models.PoolModes) {
			return fmt.Errorf("poolMode '%s' is unavailable, available poolModes: %v",
				pgb.PoolMode,
				models.PoolModes)
		}
	}

	return nil
}

func (pgs *PgSpec) ValidateImmutableFieldsUpdate(oldSpec PgSpec) error {
	newImmutableFields := pgs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable spec fields: old spec: %+v: new spec: %+v", oldSpec, pgs)
	}

	err := validateTwoFactorDelete(pgs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = pgs.validateImmutableDCsFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	if !slices.Equals(pgs.Extensions, oldSpec.Extensions) {
		return fmt.Errorf("spec.extensions are immutable")
	}

	return nil
}

func (pgs *PgSpec) validateImmutableDCsFieldsUpdate(oldSpec PgSpec) error {
	if len(pgs.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range pgs.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		err := validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		err = newDC.validateImmutableCloudProviderSettingsUpdate(&oldDC.GenericDataCentreSpec)
		if err != nil {
			return err
		}

		err = newDC.validateIntraDCImmutableFields(oldDC.IntraDataCentreReplication)
		if err != nil {
			return err
		}

		err = newDC.validateInterDCImmutableFields(oldDC.InterDataCentreReplication)
		if err != nil {
			return err
		}

		if newDC.NodesNumber != oldDC.NodesNumber {
			return models.ErrImmutableNodesNumber
		}

	}

	return nil
}

func (pdc *PgDataCentre) validateInterDCImmutableFields(oldInterDC []*InterDataCentreReplication) error {
	if len(pdc.InterDataCentreReplication) != len(oldInterDC) {
		return models.ErrImmutableInterDataCentreReplication
	}

	for i, newInterDC := range pdc.InterDataCentreReplication {
		if *newInterDC != *oldInterDC[i] {
			return models.ErrImmutableInterDataCentreReplication
		}
	}

	return nil
}

func (pdc *PgDataCentre) validateIntraDCImmutableFields(oldIntraDC []*IntraDataCentreReplication) error {
	if len(pdc.IntraDataCentreReplication) != len(oldIntraDC) {
		return models.ErrImmutableIntraDataCentreReplication
	}

	for i, newIntraDC := range pdc.IntraDataCentreReplication {
		if *newIntraDC != *oldIntraDC[i] {
			return models.ErrImmutableIntraDataCentreReplication
		}
	}

	return nil
}

func (pgs *PgSpec) newImmutableFields() *immutablePostgreSQLFields {
	return &immutablePostgreSQLFields{
		specificPostgreSQLFields: specificPostgreSQLFields{
			SynchronousModeStrict: pgs.SynchronousModeStrict,
		},
		immutableCluster: pgs.GenericClusterSpec.immutableFields(),
	}
}

func (pdc *PgDataCentre) newImmutableFields() *immutablePostgreSQLDCFields {
	return &immutablePostgreSQLDCFields{
		immutableDC: immutableDC{
			Name:                pdc.Name,
			Region:              pdc.Region,
			CloudProvider:       pdc.CloudProvider,
			ProviderAccountName: pdc.ProviderAccountName,
			Network:             pdc.Network,
		},
		specificPostgreSQLDC: specificPostgreSQLDC{
			ClientEncryption: pdc.ClientEncryption,
			NumberOfNodes:    pdc.NodesNumber,
		},
	}
}
