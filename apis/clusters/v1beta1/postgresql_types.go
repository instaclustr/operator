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
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"

	k8sCore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

type PgDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL options
	ClientEncryption           bool                          `json:"clientEncryption"`
	InterDataCentreReplication []*InterDataCentreReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*IntraDataCentreReplication `json:"intraDataCentreReplication"`
	PGBouncer                  []*PgBouncer                  `json:"pgBouncer,omitempty"`
}

type PgBouncer struct {
	PGBouncerVersion string `json:"pgBouncerVersion"`
	PoolMode         string `json:"poolMode"`
}

type InterDataCentreReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type IntraDataCentreReplication struct {
	ReplicationMode string `json:"replicationMode"`
}

type PgRestoreFrom struct {
	// Original cluster ID. Backup from that cluster will be used for restore
	ClusterID string `json:"clusterId"`

	// The display name of the restored cluster.
	ClusterNameOverride string `json:"clusterNameOverride,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCInfos []*PgRestoreCDCInfo `json:"cdcInfos,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// The cluster network for this cluster to be restored to.
	ClusterNetwork string `json:"clusterNetwork,omitempty"`
}

type PgRestoreCDCInfo struct {
	CDCID            string `json:"cdcId,omitempty"`
	RestoreToSameVPC bool   `json:"restoreToSameVpc,omitempty"`
	CustomVPCID      string `json:"customVpcId,omitempty"`
	CustomVPCNetwork string `json:"customVpcNetwork,omitempty"`
}

// PgSpec defines the desired state of PostgreSQL
type PgSpec struct {
	PgRestoreFrom         *PgRestoreFrom `json:"pgRestoreFrom,omitempty"`
	Cluster               `json:",inline"`
	DataCentres           []*PgDataCentre   `json:"dataCentres,omitempty"`
	ClusterConfigurations map[string]string `json:"clusterConfigurations,omitempty"`
	Description           string            `json:"description,omitempty"`
	SynchronousModeStrict bool              `json:"synchronousModeStrict,omitempty"`
}

// PgStatus defines the observed state of PostgreSQL
type PgStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgSpec   `json:"spec,omitempty"`
	Status PgStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
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
}

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}

func (pg *PostgreSQL) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(pg).String() + "/" + jobName
}

func (pg *PostgreSQL) NewPatch() client.Patch {
	old := pg.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (pg *PostgreSQL) NewBackupSpec(startTimestamp int) *clusterresourcesv1beta1.ClusterBackup {
	return &clusterresourcesv1beta1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1beta1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.PgBackupPrefix + pg.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   pg.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: pg.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1beta1.ClusterBackupSpec{
			ClusterID:   pg.Status.ID,
			ClusterKind: models.PgClusterKind,
		},
	}
}

func (pgs *PgSpec) HasRestore() bool {
	if pgs.PgRestoreFrom != nil && pgs.PgRestoreFrom.ClusterID != "" {
		return true
	}

	return false
}

func (pgs *PgSpec) ToInstAPI() *models.PGCluster {
	return &models.PGCluster{
		Name:                  pgs.Name,
		PostgreSQLVersion:     pgs.Version,
		DataCentres:           pgs.DCsToInstAPI(),
		SynchronousModeStrict: pgs.SynchronousModeStrict,
		PrivateNetworkCluster: pgs.PrivateNetworkCluster,
		SLATier:               pgs.SLATier,
		TwoFactorDelete:       pgs.TwoFactorDeletesToInstAPI(),
	}
}

func (pgs *PgSpec) DCsToInstAPI() (iDCs []*models.PGDataCentre) {
	for _, dc := range pgs.DataCentres {
		iDCs = append(iDCs, dc.ToInstAPI())
	}
	return
}

func (pdc *PgDataCentre) ToInstAPI() *models.PGDataCentre {
	return &models.PGDataCentre{
		DataCentre:                 pdc.DataCentre.ToInstAPI(),
		PGBouncer:                  pdc.PGBouncerToInstAPI(),
		ClientToClusterEncryption:  pdc.ClientEncryption,
		InterDataCentreReplication: pdc.InterDCReplicationToInstAPI(),
		IntraDataCentreReplication: pdc.IntraDCReplicationToInstAPI(),
	}
}

func (pdc *PgDataCentre) InterDCReplicationToInstAPI() (iIRs []*models.PGInterDCReplication) {
	for _, interDCReplication := range pdc.InterDataCentreReplication {
		iIRs = append(iIRs, &models.PGInterDCReplication{
			IsPrimaryDataCentre: interDCReplication.IsPrimaryDataCentre,
		})
	}
	return
}

func (pdc *PgDataCentre) IntraDCReplicationToInstAPI() (iIRs []*models.PGIntraDCReplication) {
	for _, interDCReplication := range pdc.IntraDataCentreReplication {
		iIRs = append(iIRs, &models.PGIntraDCReplication{
			ReplicationMode: interDCReplication.ReplicationMode,
		})
	}
	return
}

func (pdc *PgDataCentre) PGBouncerToInstAPI() (iPgB []*models.PGBouncer) {
	for _, pgb := range pdc.PGBouncer {
		iPgB = append(iPgB, &models.PGBouncer{
			PGBouncerVersion: pgb.PGBouncerVersion,
			PoolMode:         pgb.PoolMode,
		})
	}

	return
}

func (pgs *PgSpec) IsEqual(iPG PgSpec) bool {
	return pgs.Cluster.IsEqual(iPG.Cluster) && pgs.SynchronousModeStrict == iPG.SynchronousModeStrict
}

func (pgs *PgSpec) AreDCsEqual(iDCs []*PgDataCentre) bool {
	if len(pgs.DataCentres) != len(iDCs) {
		return false
	}

	for i, iDC := range iDCs {
		dc := pgs.DataCentres[i]
		if !dc.IsEqual(iDC.DataCentre) ||
			dc.ClientEncryption != iDC.ClientEncryption ||
			!dc.AreInterDCReplicationsEqual(iDC.InterDataCentreReplication) ||
			!dc.AreIntraDCReplicationsEqual(iDC.IntraDataCentreReplication) ||
			!dc.ArePGBouncersEqual(iDC.PGBouncer) {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) AreInterDCReplicationsEqual(iIRs []*InterDataCentreReplication) bool {
	if len(pdc.InterDataCentreReplication) != len(iIRs) {
		return false
	}

	for i, iIR := range iIRs {
		if iIR.IsPrimaryDataCentre != pdc.InterDataCentreReplication[i].IsPrimaryDataCentre {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) AreIntraDCReplicationsEqual(iIRs []*IntraDataCentreReplication) bool {
	if len(pdc.IntraDataCentreReplication) != len(iIRs) {
		return false
	}

	for i, iIR := range iIRs {
		if iIR.ReplicationMode != pdc.IntraDataCentreReplication[i].ReplicationMode {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) ArePGBouncersEqual(iPGBs []*PgBouncer) bool {
	if len(iPGBs) != len(pdc.PGBouncer) {
		return false
	}

	for i, iPGB := range iPGBs {
		if iPGB.PoolMode != pdc.PGBouncer[i].PoolMode ||
			iPGB.PGBouncerVersion != pdc.PGBouncer[i].PGBouncerVersion {
			return false
		}
	}

	return true
}

func (pg *PostgreSQL) GetUserPassword(secret *k8sCore.Secret) string {
	password := secret.Data[models.Password]
	if len(password) == 0 {
		return ""
	}

	return string(password)
}

func (pg *PostgreSQL) GetUserSecret(ctx context.Context, k8sClient client.Client) (*k8sCore.Secret, error) {
	userSecret := &k8sCore.Secret{}
	userSecretNamespacedName := types.NamespacedName{
		Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, pg.Name),
		Namespace: pg.Namespace,
	}
	err := k8sClient.Get(ctx, userSecretNamespacedName, userSecret)
	if err != nil {
		return nil, err
	}

	return userSecret, nil
}

func (pg *PostgreSQL) GetUserSecretName(ctx context.Context, k8sClient client.Client) (string, error) {
	var err error

	labelsToQuery := fmt.Sprintf("%s=%s", models.ClusterIDLabel, pg.Status.ID)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return "", err
	}

	userSecretList := &k8sCore.SecretList{}
	err = k8sClient.List(ctx, userSecretList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return "", err
	}

	if len(userSecretList.Items) == 0 {
		return "", nil
	}

	return userSecretList.Items[0].Name, nil
}

func (pg *PostgreSQL) NewUserSecret(defaultUserPassword string) *k8sCore.Secret {
	return &k8sCore.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.SecretKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, pg.Name),
			Namespace: pg.Namespace,
			Labels: map[string]string{
				models.ControlledByLabel:  pg.Name,
				models.DefaultSecretLabel: "true",
			},
		},
		StringData: map[string]string{
			models.Username: models.DefaultPgUsernameValue,
			models.Password: defaultUserPassword,
		},
	}
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

		err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		err = newDC.validateIntraDCImmutableFields(oldDC.IntraDataCentreReplication)
		if err != nil {
			return nil
		}

		err = newDC.validateInterDCImmutableFields(oldDC.InterDataCentreReplication)
		if err != nil {
			return nil
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
		immutableCluster: pgs.Cluster.newImmutableFields(),
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
		},
	}
}

func (pg *PostgreSQL) FromInstAPI(iData []byte) (*PostgreSQL, error) {
	iPg := &models.PGCluster{}
	err := json.Unmarshal(iData, iPg)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		TypeMeta:   pg.TypeMeta,
		ObjectMeta: pg.ObjectMeta,
		Spec:       pg.Spec.FromInstAPI(iPg),
		Status:     pg.Status.FromInstAPI(iPg),
	}, nil
}

func (pg *PostgreSQL) DefaultPasswordFromInstAPI(iData []byte) (string, error) {
	type defaultPasswordResponse struct {
		DefaultUserPassword string `json:"defaultUserPassword,omitempty"`
	}

	dpr := &defaultPasswordResponse{}
	err := json.Unmarshal(iData, dpr)
	if err != nil {
		return "", err
	}

	return dpr.DefaultUserPassword, nil
}

func (pgs *PgSpec) FromInstAPI(iPg *models.PGCluster) PgSpec {
	return PgSpec{
		Cluster: Cluster{
			Name:                  iPg.Name,
			Version:               iPg.PostgreSQLVersion,
			PCICompliance:         iPg.PCIComplianceMode,
			PrivateNetworkCluster: iPg.PrivateNetworkCluster,
			SLATier:               iPg.SLATier,
			TwoFactorDelete:       pgs.Cluster.TwoFactorDeleteFromInstAPI(iPg.TwoFactorDelete),
		},
		DataCentres:           pgs.DCsFromInstAPI(iPg.DataCentres),
		SynchronousModeStrict: iPg.SynchronousModeStrict,
	}
}

func (pgs *PgSpec) DCsFromInstAPI(iDCs []*models.PGDataCentre) (dcs []*PgDataCentre) {
	for _, iDC := range iDCs {
		dcs = append(dcs, &PgDataCentre{
			DataCentre:                 pgs.Cluster.DCFromInstAPI(iDC.DataCentre),
			ClientEncryption:           iDC.ClientToClusterEncryption,
			InterDataCentreReplication: pgs.InterDCReplicationsFromInstAPI(iDC.InterDataCentreReplication),
			IntraDataCentreReplication: pgs.IntraDCReplicationsFromInstAPI(iDC.IntraDataCentreReplication),
			PGBouncer:                  pgs.PGBouncerFromInstAPI(iDC.PGBouncer),
		})
	}
	return
}

func (pgs *PgSpec) InterDCReplicationsFromInstAPI(iIRs []*models.PGInterDCReplication) (irs []*InterDataCentreReplication) {
	for _, iIR := range iIRs {
		ir := &InterDataCentreReplication{
			IsPrimaryDataCentre: iIR.IsPrimaryDataCentre,
		}
		irs = append(irs, ir)
	}
	return
}

func (pgs *PgSpec) IntraDCReplicationsFromInstAPI(iIRs []*models.PGIntraDCReplication) (irs []*IntraDataCentreReplication) {
	for _, iIR := range iIRs {
		ir := &IntraDataCentreReplication{
			ReplicationMode: iIR.ReplicationMode,
		}
		irs = append(irs, ir)
	}
	return
}

func (pgs *PgSpec) PGBouncerFromInstAPI(iPgBs []*models.PGBouncer) (pgbs []*PgBouncer) {
	for _, iPgB := range iPgBs {
		pgb := &PgBouncer{
			PGBouncerVersion: iPgB.PGBouncerVersion,
			PoolMode:         iPgB.PoolMode,
		}
		pgbs = append(pgbs, pgb)
	}
	return
}

func (pgs *PgStatus) FromInstAPI(iPg *models.PGCluster) PgStatus {
	return PgStatus{
		ClusterStatus: ClusterStatus{
			ID:                            iPg.ID,
			State:                         iPg.Status,
			DataCentres:                   pgs.DCsFromInstAPI(iPg.DataCentres),
			CurrentClusterOperationStatus: iPg.CurrentClusterOperationStatus,
			MaintenanceEvents:             pgs.MaintenanceEvents,
		},
	}
}

func (pgs *PgStatus) DCsFromInstAPI(iDCs []*models.PGDataCentre) (dcs []*DataCentreStatus) {
	for _, iDC := range iDCs {
		dcs = append(dcs, pgs.ClusterStatus.DCFromInstAPI(iDC.DataCentre))
	}
	return
}
