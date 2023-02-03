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
	"context"
	"fmt"
	"strconv"
	"unicode"

	k8sCore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

type PgDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL options
	ClientEncryption           bool                          `json:"clientEncryption"`
	PostgresqlNodeCount        int32                         `json:"postgresqlNodeCount"`
	InterDataCentreReplication []*InterDataCentreReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*IntraDataCentreReplication `json:"intraDataCentreReplication"`
	// PGBouncer options
	PGBouncerVersion string `json:"pgBouncerVersion,omitempty"`
	PoolMode         string `json:"poolMode,omitempty"`
}

type InterDataCentreReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type IntraDataCentreReplication struct {
	ReplicationMode string `json:"replicationMode"`
}

type PgRestoreFrom struct {
	ClusterID           string              `json:"clusterId"`
	ClusterNameOverride string              `json:"clusterNameOverride,omitempty"`
	CDCInfos            []*PgRestoreCDCInfo `json:"cdcInfos,omitempty"`
	PointInTime         int64               `json:"pointInTime,omitempty"`
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
	ClusterStatus         `json:",inline"`
	DefaultUserSecretName string `json:"defaultUserSecretName,omitempty"`
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
	ClientEncryption    bool
	PostgresqlNodeCount int32
	PGBouncerVersion    string
	PoolMode            string
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

func (pg *PostgreSQL) NewBackupSpec(startTimestamp int) *clusterresourcesv1alpha1.ClusterBackup {
	return &clusterresourcesv1alpha1.ClusterBackup{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterBackupKind,
			APIVersion: models.ClusterresourcesV1alpha1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        models.PgBackupPrefix + pg.Status.ID + "-" + strconv.Itoa(startTimestamp),
			Namespace:   pg.Namespace,
			Annotations: map[string]string{models.StartTimestampAnnotation: strconv.Itoa(startTimestamp)},
			Labels:      map[string]string{models.ClusterIDLabel: pg.Status.ID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1alpha1.ClusterBackupSpec{
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
	instTFD := pgs.TwoFactorDeletesToInstAPI()

	instDCs := pgs.DataCentresToInstAPI()
	return &models.PGCluster{
		Name:                  pgs.Name,
		PostgreSQLVersion:     pgs.Version,
		DataCentres:           instDCs,
		SynchronousModeStrict: pgs.SynchronousModeStrict,
		PrivateNetworkCluster: pgs.PrivateNetworkCluster,
		SLATier:               pgs.SLATier,
		TwoFactorDelete:       instTFD,
	}
}

func (pdc *PgDataCentre) ToInstAPI() *models.PGDataCentre {
	instDC := &models.PGDataCentre{
		DataCentre: modelsv2.DataCentre{
			Name:                pdc.Name,
			Network:             pdc.Network,
			NodeSize:            pdc.NodeSize,
			NumberOfNodes:       pdc.PostgresqlNodeCount,
			CloudProvider:       pdc.CloudProvider,
			Region:              pdc.Region,
			ProviderAccountName: pdc.ProviderAccountName,
		},
		PGBouncer:                 []*models.PGBouncer{},
		ClientToClusterEncryption: pdc.ClientEncryption,
	}

	pdc.CloudProviderSettingsToInstAPI(&instDC.DataCentre)

	pdc.TagsToInstAPI(&instDC.DataCentre)

	pdc.InterDCReplicationToInstAPI(instDC)

	pdc.IntraDCReplicationToInstAPI(instDC)

	if pdc.PGBouncerVersion != "" {
		pdc.PGBouncerToInstAPI(instDC)
	}

	return instDC
}

func (pdc *PgDataCentre) InterDCReplicationToInstAPI(instDC *models.PGDataCentre) {
	var instReplication []*models.PGInterDCReplication
	for _, interDCReplication := range pdc.InterDataCentreReplication {
		instReplication = append(instReplication, &models.PGInterDCReplication{
			IsPrimaryDataCentre: interDCReplication.IsPrimaryDataCentre,
		})
	}

	instDC.InterDataCentreReplication = instReplication
}

func (pdc *PgDataCentre) IntraDCReplicationToInstAPI(instDC *models.PGDataCentre) {
	var instReplication []*models.PGIntraDCReplication
	for _, interDCReplication := range pdc.IntraDataCentreReplication {
		instReplication = append(instReplication, &models.PGIntraDCReplication{
			ReplicationMode: interDCReplication.ReplicationMode,
		})
	}

	instDC.IntraDataCentreReplication = instReplication
}

func (pdc *PgDataCentre) PGBouncerToInstAPI(instDC *models.PGDataCentre) {
	instDC.PGBouncer = []*models.PGBouncer{
		{
			PGBouncerVersion: pdc.PGBouncerVersion,
			PoolMode:         pdc.PoolMode,
		},
	}
}

func (pgs *PgSpec) AreSpecsEqual(instCluster *models.PGStatus) bool {
	if pgs.Name != instCluster.Name ||
		pgs.Version != instCluster.PostgreSQLVersion ||
		pgs.PCICompliance != instCluster.PCIComplianceMode ||
		pgs.PrivateNetworkCluster != instCluster.PrivateNetworkCluster ||
		pgs.SLATier != instCluster.SLATier ||
		!pgs.IsTwoFactorDeleteEqual(instCluster.TwoFactorDelete) ||
		pgs.SynchronousModeStrict != instCluster.SynchronousModeStrict ||
		!pgs.AreDCsEqual(instCluster.DataCentres) {
		return false
	}

	return true
}

func (pgs *PgSpec) AreDCsEqual(instDCs []*models.PGDataCentre) bool {
	if len(pgs.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range pgs.DataCentres {
			if instDC.Name == k8sDC.Name {
				if k8sDC.PostgresqlNodeCount != instDC.NumberOfNodes ||
					k8sDC.Region != instDC.Region ||
					k8sDC.CloudProvider != instDC.CloudProvider ||
					k8sDC.ProviderAccountName != instDC.ProviderAccountName ||
					!k8sDC.AreCloudProviderSettingsEqual(instDC.AWSSettings, instDC.GCPSettings, instDC.AzureSettings) ||
					k8sDC.Network != instDC.Network ||
					k8sDC.NodeSize != instDC.NodeSize ||
					!k8sDC.AreTagsEqual(instDC.Tags) ||
					k8sDC.ClientEncryption != instDC.ClientToClusterEncryption ||
					!k8sDC.AreInterDCReplicationsEqual(instDC.InterDataCentreReplication) ||
					!k8sDC.AreIntraDCReplicationsEqual(instDC.IntraDataCentreReplication) ||
					!k8sDC.ArePGBouncersEqual(instDC.PGBouncer) {
					return false
				}

				break
			}
		}
	}

	return true
}

// SetDefaultValues should be implemented using validation webhook
// TODO https://github.com/instaclustr/operator/issues/219
func (pgs *PgSpec) SetDefaultValues() {
	for _, dataCentre := range pgs.DataCentres {
		if dataCentre.ProviderAccountName == "" {
			dataCentre.ProviderAccountName = models.DefaultAccountName
		}

		if len(dataCentre.CloudProviderSettings) == 0 {
			dataCentre.CloudProviderSettings = append(dataCentre.CloudProviderSettings, &CloudProviderSettings{
				DiskEncryptionKey:      "",
				ResourceGroup:          "",
				CustomVirtualNetworkID: "",
			})
		}
	}
}

func (pgs *PgSpec) DataCentresToInstAPI() []*models.PGDataCentre {
	var instDCs []*models.PGDataCentre
	for _, k8sDC := range pgs.DataCentres {
		instDCs = append(instDCs, k8sDC.ToInstAPI())
	}
	return instDCs
}

func (pgs *PgSpec) SetSpecFromInstAPI(instCluster *models.PGStatus) {
	pgs.Name = instCluster.Name
	pgs.Version = instCluster.PostgreSQLVersion
	pgs.PCICompliance = instCluster.PCIComplianceMode
	pgs.PrivateNetworkCluster = instCluster.PrivateNetworkCluster
	pgs.SLATier = instCluster.SLATier
	pgs.SetTwoFactorDeletesFromInstAPI(instCluster.TwoFactorDelete)
	pgs.SynchronousModeStrict = instCluster.SynchronousModeStrict
	pgs.SetDCsFromInstAPI(instCluster.DataCentres)
}

func (pgs *PgSpec) SetDCsFromInstAPI(instDCs []*models.PGDataCentre) {
	var k8sDCs []*PgDataCentre
	for _, instDC := range instDCs {
		dc := &PgDataCentre{}
		dc.SetDCFromInstAPI(instDC)

		k8sDCs = append(k8sDCs, dc)
	}

	pgs.DataCentres = k8sDCs
}

func (pdc *PgDataCentre) SetDCFromInstAPI(instDC *models.PGDataCentre) {
	pdc.Name = instDC.Name
	pdc.Region = instDC.Region
	pdc.CloudProvider = instDC.CloudProvider
	pdc.ProviderAccountName = instDC.ProviderAccountName
	pdc.Network = instDC.Network
	pdc.NodeSize = instDC.NodeSize
	pdc.SetTagsFromInstAPI(instDC.Tags)
	pdc.SetCloudProviderSettingsFromInstAPI(&instDC.DataCentre)
	pdc.ClientEncryption = instDC.ClientToClusterEncryption
	pdc.PostgresqlNodeCount = instDC.NumberOfNodes
	pdc.SetInterDCReplicationsFromInstAPI(instDC.InterDataCentreReplication)
	pdc.SetIntraDCReplicationsFromInstAPI(instDC.IntraDataCentreReplication)
	pdc.SetPGBouncerFromInstAPI(instDC.PGBouncer)
}

func (pdc *PgDataCentre) SetInterDCReplicationsFromInstAPI(instIRs []*models.PGInterDCReplication) {
	k8sIRs := []*InterDataCentreReplication{}
	for _, instIR := range instIRs {
		k8sIR := &InterDataCentreReplication{}
		k8sIR.SetInterDCReplicationFromInstAPI(instIR)
		k8sIRs = append(k8sIRs, k8sIR)
	}
	pdc.InterDataCentreReplication = k8sIRs
}

func (ir *InterDataCentreReplication) SetInterDCReplicationFromInstAPI(instIR *models.PGInterDCReplication) {
	ir.IsPrimaryDataCentre = instIR.IsPrimaryDataCentre
}

func (pdc *PgDataCentre) SetIntraDCReplicationsFromInstAPI(instIRs []*models.PGIntraDCReplication) {
	k8sIRs := []*IntraDataCentreReplication{}
	for _, instIR := range instIRs {
		k8sIR := &IntraDataCentreReplication{}
		k8sIR.SetIntraDCReplicationFromInstAPI(instIR)
		k8sIRs = append(k8sIRs, k8sIR)
	}
	pdc.IntraDataCentreReplication = k8sIRs
}

func (ir *IntraDataCentreReplication) SetIntraDCReplicationFromInstAPI(instIR *models.PGIntraDCReplication) {
	ir.ReplicationMode = instIR.ReplicationMode
}

func (pdc *PgDataCentre) SetPGBouncerFromInstAPI(instPGBs []*models.PGBouncer) {
	for _, instPGB := range instPGBs {
		pdc.PGBouncerVersion = instPGB.PGBouncerVersion
		pdc.PoolMode = instPGB.PoolMode
	}
}

func (ps *PgStatus) SetStatusFromInstAPI(instCluster *models.PGStatus) {
	ps.ID = instCluster.ID
	ps.Status = instCluster.Status
	ps.CurrentClusterOperationStatus = instCluster.CurrentClusterOperationStatus
	ps.TwoFactorDeleteEnabled = len(instCluster.TwoFactorDelete) > 0
	ps.SetDCsFromInstAPI(instCluster.DataCentres)
}

func (ps *PgStatus) SetDCsFromInstAPI(instDCs []*models.PGDataCentre) {
	k8sDCs := []*DataCentreStatus{}
	for _, instDC := range instDCs {
		k8sDC := &DataCentreStatus{
			ID:         instDC.ID,
			Status:     instDC.Status,
			NodeNumber: instDC.NumberOfNodes,
		}

		k8sDC.SetNodesStatusFromInstAPI(instDC.Nodes)
		k8sDCs = append(k8sDCs, k8sDC)
	}
	ps.DataCentres = k8sDCs
}

func (ps *PgStatus) AreStatusesEqual(instStatus *models.PGStatus) bool {
	if ps.ID != instStatus.ID ||
		ps.Status != instStatus.Status ||
		ps.CurrentClusterOperationStatus != instStatus.CurrentClusterOperationStatus ||
		!ps.AreDCsEqual(instStatus.DataCentres) {
		return false
	}

	return true
}

func (ps *PgStatus) AreDCsEqual(instDCs []*models.PGDataCentre) bool {
	if len(ps.DataCentres) != len(instDCs) {
		return false
	}

	for _, instDC := range instDCs {
		for _, k8sDC := range ps.DataCentres {
			if instDC.ID == k8sDC.ID {
				if instDC.Status != k8sDC.Status ||
					instDC.NumberOfNodes != k8sDC.NodeNumber ||
					!k8sDC.AreNodesEqual(instDC.Nodes) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (pdc *PgDataCentre) AreInterDCReplicationsEqual(instIRs []*models.PGInterDCReplication) bool {
	if len(pdc.InterDataCentreReplication) != len(instIRs) {
		return false
	}

	for i, instIR := range instIRs {
		if instIR.IsPrimaryDataCentre != pdc.InterDataCentreReplication[i].IsPrimaryDataCentre {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) AreIntraDCReplicationsEqual(instIRs []*models.PGIntraDCReplication) bool {
	if len(pdc.IntraDataCentreReplication) != len(instIRs) {
		return false
	}

	for i, instIR := range instIRs {
		if instIR.ReplicationMode != pdc.IntraDataCentreReplication[i].ReplicationMode {
			return false
		}
	}

	return true
}

func (pdc *PgDataCentre) ArePGBouncersEqual(instPGBs []*models.PGBouncer) bool {
	if pdc.PoolMode != "" &&
		pdc.PGBouncerVersion != "" &&
		len(instPGBs) == 0 {
		return false
	}

	for _, instPGB := range instPGBs {
		if instPGB.PoolMode != pdc.PoolMode ||
			instPGB.PGBouncerVersion != pdc.PGBouncerVersion {
			return false
		}
	}

	return true
}

func (pg *PostgreSQL) GetUserPassword(secret *k8sCore.Secret) string {
	password := secret.Data[models.DefaultUserPassword]
	if len(password) == 0 {
		return ""
	}

	return string(password[:len(password)-1])
}

func (pg *PostgreSQL) GetUserSecret(ctx context.Context, k8sClient client.Client) (*k8sCore.Secret, error) {
	userSecret := &k8sCore.Secret{}
	userSecretNamespacedName := types.NamespacedName{
		Name:      pg.Status.DefaultUserSecretName,
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

func (pg *PostgreSQL) NewUserSecret() *k8sCore.Secret {
	return &k8sCore.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.SecretKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      models.DefaultUserSecretPrefix + pg.Name,
			Namespace: pg.Namespace,
			Labels:    map[string]string{models.ControlledByLabel: pg.Name},
		},
		StringData: map[string]string{models.DefaultUserPassword: ""},
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
	if pdc.PGBouncerVersion == "" {
		if pdc.PoolMode != "" {
			return fmt.Errorf("poolMode field is filled. Fill PGBouncerVersion field to enable PGBouncer")
		}
	} else {
		if !Contains(pdc.PGBouncerVersion, models.PGBouncerVersions) {
			return fmt.Errorf("pgBouncerVersion '%s' is unavailable, available versions: %v",
				pdc.PGBouncerVersion,
				models.PGBouncerVersions)
		}
		if pdc.PoolMode != "" &&
			!Contains(pdc.PoolMode, models.PoolModes) {
			return fmt.Errorf("poolMode '%s' is unavailable, available poolModes: %v",
				pdc.PoolMode,
				models.PoolModes)
		}
	}

	return nil
}

func (pgs *PgSpec) ValidateImmutableFieldsUpdate(oldSpec PgSpec) error {
	newImmutableFields := pgs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if newImmutableFields != oldImmutableFields {
		return fmt.Errorf("cannot update immutable spec fields: old spec: %+v: new spec: %+v", oldSpec, pgs)
	}

	err := validateTwoFactorDelete(pgs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = pgs.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	return nil
}

func (pgs *PgSpec) validateImmutableDataCentresFieldsUpdate(oldSpec PgSpec) error {
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
			ClientEncryption:    pdc.ClientEncryption,
			PostgresqlNodeCount: pdc.PostgresqlNodeCount,
			PGBouncerVersion:    pdc.PGBouncerVersion,
			PoolMode:            pdc.PoolMode,
		},
	}
}
