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

	k8sCore "k8s.io/api/core/v1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
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
	RestoredClusterName string `json:"restoredClusterName,omitempty"`

	// An optional list of cluster data centres for which custom VPC settings will be used.
	CDCConfigs []*RestoreCDCConfig `json:"cdsConfigs,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.
	PointInTime int64 `json:"pointInTime,omitempty"`
}

// PgSpec defines the desired state of PostgreSQL
// +kubebuilder:validation:XValidation:rule="has(self.extensions) == has(oldSelf.extensions)",message="extensions cannot be changed after it is set"
type PgSpec struct {
	PgRestoreFrom         *PgRestoreFrom `json:"pgRestoreFrom,omitempty"`
	Cluster               `json:",inline"`
	OnPremisesSpec        *OnPremisesSpec   `json:"onPremisesSpec,omitempty"`
	DataCentres           []*PgDataCentre   `json:"dataCentres,omitempty"`
	ClusterConfigurations map[string]string `json:"clusterConfigurations,omitempty"`
	SynchronousModeStrict bool              `json:"synchronousModeStrict,omitempty"`
	UserRefs              []*Reference      `json:"userRefs,omitempty"`
	//+kubebuilder:validate:MaxItems:=1
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty"`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="extensions cannot be changed after it is set"
	Extensions PgExtensions `json:"extensions,omitempty"`
}

// PgStatus defines the observed state of PostgreSQL
type PgStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

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

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}

func (pg *PostgreSQL) GetJobID(jobName string) string {
	return pg.Kind + "/" + client.ObjectKeyFromObject(pg).String() + "/" + jobName
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
			ClusterRef: &clusterresourcesv1beta1.ClusterRef{
				Name:        pg.Name,
				Namespace:   pg.Namespace,
				ClusterKind: models.PgClusterKind,
			},
		},
	}
}

func (pg *PostgreSQL) RestoreInfoToInstAPI(restoreData *PgRestoreFrom) any {
	iRestore := struct {
		CDCConfigs          []*RestoreCDCConfig `json:"cdcConfigs,omitempty"`
		PointInTime         int64               `json:"pointInTime,omitempty"`
		ClusterID           string              `json:"clusterId,omitempty"`
		RestoredClusterName string              `json:"restoredClusterName,omitempty"`
	}{
		CDCConfigs:          restoreData.CDCConfigs,
		PointInTime:         restoreData.PointInTime,
		ClusterID:           restoreData.ClusterID,
		RestoredClusterName: restoreData.RestoredClusterName,
	}

	return iRestore
}

func (pg *PostgreSQL) GetDataCentreID(cdcName string) string {
	if cdcName == "" {
		return pg.Status.DataCentres[0].ID
	}
	for _, cdc := range pg.Status.DataCentres {
		if cdc.Name == cdcName {
			return cdc.ID
		}
	}
	return ""
}
func (pg *PostgreSQL) GetClusterID() string {
	return pg.Status.ID
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
		Description:           pgs.Description,
		PrivateNetworkCluster: pgs.PrivateNetworkCluster,
		SLATier:               pgs.SLATier,
		TwoFactorDelete:       pgs.TwoFactorDeletesToInstAPI(),
		Extensions:            pgs.Extensions.ToInstAPI(),
	}
}

func (pgs *PgSpec) DCsToInstAPI() (iDCs []*models.PGDataCentre) {
	for _, dc := range pgs.DataCentres {
		iDCs = append(iDCs, dc.ToInstAPI())
	}
	return
}

func (pgs *PgSpec) ToClusterUpdate() *models.PGClusterUpdate {
	return &models.PGClusterUpdate{
		DataCentres:    pgs.DCsToInstAPI(),
		ResizeSettings: resizeSettingsToInstAPI(pgs.ResizeSettings),
	}
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
	return pgs.Cluster.IsEqual(iPG.Cluster) &&
		pgs.SynchronousModeStrict == iPG.SynchronousModeStrict &&
		pgs.AreDCsEqual(iPG.DataCentres)
}

func (pgs *PgSpec) AreDCsEqual(iDCs []*PgDataCentre) bool {
	if len(pgs.DataCentres) != len(iDCs) {
		return false
	}

	for i, iDC := range iDCs {
		dc := pgs.DataCentres[i]

		if iDC.Name != dc.Name {
			continue
		}

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
				models.ClusterIDLabel:     pg.Status.ID,
			},
		},
		StringData: map[string]string{
			models.Username: models.DefaultPgUsernameValue,
			models.Password: defaultUserPassword,
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
			Description:           iPg.Description,
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

func GetDefaultPgUserSecret(
	ctx context.Context,
	name string,
	ns string,
	k8sClient client.Client,
) (*k8sCore.Secret, error) {
	userSecret := &k8sCore.Secret{}
	userSecretNamespacedName := types.NamespacedName{
		Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, name),
		Namespace: ns,
	}
	err := k8sClient.Get(ctx, userSecretNamespacedName, userSecret)
	if err != nil {
		return nil, err
	}

	return userSecret, nil
}

func (pg *PostgreSQL) GetExposePorts() []k8scorev1.ServicePort {
	var exposePorts []k8scorev1.ServicePort
	if !pg.Spec.PrivateNetworkCluster {
		exposePorts = []k8scorev1.ServicePort{
			{
				Name: models.PostgreSQLDB,
				Port: models.Port5432,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: models.Port5432,
				},
			},
		}
	}
	return exposePorts
}

func (pg *PostgreSQL) GetHeadlessPorts() []k8scorev1.ServicePort {
	headlessPorts := []k8scorev1.ServicePort{
		{
			Name: models.PostgreSQLDB,
			Port: models.Port5432,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: models.Port5432,
			},
		},
	}
	return headlessPorts
}

// PgExtension defines desired state of an extension
type PgExtension struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// PgExtensions defines desired state of extensions
type PgExtensions []PgExtension

func (p PgExtensions) ToInstAPI() []*models.PGExtension {
	iExtensions := make([]*models.PGExtension, 0, len(p))

	for _, ext := range p {
		iExtensions = append(iExtensions, &models.PGExtension{
			Name:    ext.Name,
			Enabled: ext.Enabled,
		})
	}

	return iExtensions
}
