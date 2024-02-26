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
	"strconv"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type PgDataCentre struct {
	GenericDataCentreSpec `json:",inline"`
	// PostgreSQL options
	ClientEncryption bool   `json:"clientEncryption"`
	NodeSize         string `json:"nodeSize"`
	NumberOfNodes    int    `json:"numberOfNodes"`

	//+kubebuilder:Validation:MaxItems:=1
	InterDataCentreReplication []*InterDataCentreReplication `json:"interDataCentreReplication,omitempty"`
	//+kubebuilder:Validation:MaxItems:=1
	IntraDataCentreReplication []*IntraDataCentreReplication `json:"intraDataCentreReplication"`
	//+kubebuilder:Validation:MaxItems:=1
	PGBouncer []*PgBouncer `json:"pgBouncer,omitempty"`
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
type PgSpec struct {
	GenericClusterSpec `json:",inline"`

	SynchronousModeStrict bool              `json:"synchronousModeStrict,omitempty"`
	ClusterConfigurations map[string]string `json:"clusterConfigurations,omitempty"`
	PgRestoreFrom         *PgRestoreFrom    `json:"pgRestoreFrom,omitempty" dcomparisonSkip:"true"`
	//+kubebuilder:Validation:MinItems:=1
	//+kubebuilder:Validation:MaxItems=2
	DataCentres []*PgDataCentre `json:"dataCentres,omitempty"`
	UserRefs    []*Reference    `json:"userRefs,omitempty" dcomparisonSkip:"true"`
	//+kubebuilder:validate:MaxItems:=1
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty" dcomparisonSkip:"true"`
	Extensions     PgExtensions      `json:"extensions,omitempty"`
}

// PgStatus defines the observed state of PostgreSQL
type PgStatus struct {
	GenericStatus `json:",inline"`

	DataCentres []*PgDataCentreStatus `json:"dataCentres"`

	DefaultUserSecretRef *Reference `json:"userRefs,omitempty"`
}

type PgDataCentreStatus struct {
	GenericDataCentreStatus `json:",inline"`

	NumberOfNodes int     `json:"numberOfNodes"`
	Nodes         []*Node `json:"nodes"`
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
	if old.Annotations == nil {
		old.Annotations = make(map[string]string)
	}
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
		GenericClusterFields:  pgs.GenericClusterSpec.ToInstAPI(),
		PostgreSQLVersion:     pgs.Version,
		DataCentres:           pgs.DCsToInstAPI(),
		SynchronousModeStrict: pgs.SynchronousModeStrict,
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
		GenericDataCentreFields:    pdc.GenericDataCentreSpec.ToInstAPI(),
		ClientToClusterEncryption:  pdc.ClientEncryption,
		PGBouncer:                  pdc.PGBouncerToInstAPI(),
		InterDataCentreReplication: pdc.InterDCReplicationToInstAPI(),
		IntraDataCentreReplication: pdc.IntraDCReplicationToInstAPI(),
		NodeSize:                   pdc.NodeSize,
		NumberOfNodes:              pdc.NumberOfNodes,
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

func (c *PostgreSQL) GetSpec() PgSpec { return c.Spec }

func (c *PostgreSQL) IsSpecEqual(spec PgSpec) bool {
	return c.Spec.IsEqual(spec)
}

func (pgs *PgSpec) IsEqual(iPG PgSpec) bool {
	return pgs.GenericClusterSpec.Equals(&iPG.GenericClusterSpec) &&
		pgs.SynchronousModeStrict == iPG.SynchronousModeStrict &&
		pgs.DCsEqual(iPG.DataCentres) &&
		slices.EqualsUnordered(pgs.Extensions, iPG.Extensions)
}

func (pgs *PgSpec) DCsEqual(instaModels []*PgDataCentre) bool {
	if len(pgs.DataCentres) != len(instaModels) {
		return false
	}

	m := map[string]*PgDataCentre{}
	for _, dc := range pgs.DataCentres {
		m[dc.Name] = dc
	}

	for _, iDC := range instaModels {
		dc, ok := m[iDC.Name]
		if !ok {
			return false
		}

		if !dc.Equals(iDC) {
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

func (pg *PostgreSQL) NewUserSecret(defaultUserPassword string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
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

func (pg *PostgreSQL) FromInstAPI(instaModel *models.PGCluster) {
	pg.Spec.FromInstAPI(instaModel)
	pg.Status.FromInstAPI(instaModel)
}

func (pgs *PgSpec) FromInstAPI(instaModel *models.PGCluster) {
	pgs.GenericClusterSpec.FromInstAPI(&instaModel.GenericClusterFields, instaModel.PostgreSQLVersion)
	pgs.SynchronousModeStrict = instaModel.SynchronousModeStrict

	pgs.DCsFromInstAPI(instaModel.DataCentres)
}

func (pgs *PgSpec) DCsFromInstAPI(instaModels []*models.PGDataCentre) {
	dcs := make([]*PgDataCentre, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &PgDataCentre{}
		dc.FromInstAPI(instaModel)
		dcs[i] = dc
	}
	pgs.DataCentres = dcs
}

func (pgs *PgStatus) FromInstAPI(instaModel *models.PGCluster) {
	pgs.GenericStatus.FromInstAPI(&instaModel.GenericClusterFields)
	pgs.DCsFromInstAPI(instaModel.DataCentres)
}

func (pgs *PgStatus) DCsFromInstAPI(instaModels []*models.PGDataCentre) {
	dcs := make([]*PgDataCentreStatus, len(instaModels))
	for i, instaModel := range instaModels {
		dc := &PgDataCentreStatus{}
		dc.FromInstAPI(instaModel)
		dcs[i] = dc
	}
	pgs.DataCentres = dcs
}

func GetDefaultPgUserSecret(
	ctx context.Context,
	name string,
	ns string,
	k8sClient client.Client,
) (*k8scorev1.Secret, error) {
	userSecret := &k8scorev1.Secret{}
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
	if !pg.Spec.PrivateNetwork {
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

func (pdc *PgDataCentre) Equals(o *PgDataCentre) bool {
	return pdc.GenericDataCentreSpec.Equals(&o.GenericDataCentreSpec) &&
		pdc.ClientEncryption == o.ClientEncryption &&
		pdc.NumberOfNodes == o.NumberOfNodes &&
		pdc.NodeSize == o.NodeSize &&
		slices.EqualsPtr(pdc.InterDataCentreReplication, o.InterDataCentreReplication) &&
		slices.EqualsPtr(pdc.IntraDataCentreReplication, o.IntraDataCentreReplication) &&
		slices.EqualsPtr(pdc.PGBouncer, o.PGBouncer)
}

func (pdc *PgDataCentre) FromInstAPI(instaModel *models.PGDataCentre) {
	pdc.GenericDataCentreSpec.FromInstAPI(&instaModel.GenericDataCentreFields)

	pdc.ClientEncryption = instaModel.ClientToClusterEncryption
	pdc.NodeSize = instaModel.NodeSize
	pdc.NumberOfNodes = instaModel.NumberOfNodes

	pdc.InterReplicationsFromInstAPI(instaModel.InterDataCentreReplication)
	pdc.IntraReplicationsFromInstAPI(instaModel.IntraDataCentreReplication)
	pdc.PGBouncerFromInstAPI(instaModel.PGBouncer)
}

func (pdc *PgDataCentre) InterReplicationsFromInstAPI(instaModels []*models.PGInterDCReplication) {
	pdc.InterDataCentreReplication = make([]*InterDataCentreReplication, len(instaModels))
	for i, instaModel := range instaModels {
		pdc.InterDataCentreReplication[i] = &InterDataCentreReplication{
			IsPrimaryDataCentre: instaModel.IsPrimaryDataCentre,
		}
	}
}

func (pdc *PgDataCentre) IntraReplicationsFromInstAPI(instaModels []*models.PGIntraDCReplication) {
	pdc.IntraDataCentreReplication = make([]*IntraDataCentreReplication, len(instaModels))
	for i, instaModel := range instaModels {
		pdc.IntraDataCentreReplication[i] = &IntraDataCentreReplication{
			ReplicationMode: instaModel.ReplicationMode,
		}
	}
}

func (pdc *PgDataCentre) PGBouncerFromInstAPI(instaModels []*models.PGBouncer) {
	pdc.PGBouncer = make([]*PgBouncer, len(instaModels))
	for i, instaModel := range instaModels {
		pdc.PGBouncer[i] = &PgBouncer{
			PGBouncerVersion: instaModel.PGBouncerVersion,
			PoolMode:         instaModel.PoolMode,
		}
	}
}

func (pgs *PgSpec) ClusterConfigurationsFromInstAPI(instaModels []*models.ClusterConfigurations) {
	pgs.ClusterConfigurations = make(map[string]string, len(instaModels))
	for _, instaModel := range instaModels {
		pgs.ClusterConfigurations[instaModel.ParameterName] = instaModel.ParameterValue
	}
}

func (s *PgDataCentreStatus) FromInstAPI(instaModel *models.PGDataCentre) {
	s.GenericDataCentreStatus.FromInstAPI(&instaModel.GenericDataCentreFields)
	s.Nodes = nodesFromInstAPI(instaModel.Nodes)
	s.NumberOfNodes = instaModel.NumberOfNodes
}

func (s *PgDataCentreStatus) Equals(o *PgDataCentreStatus) bool {
	return s.GenericDataCentreStatus.Equals(&o.GenericDataCentreStatus) &&
		s.NumberOfNodes == o.NumberOfNodes &&
		nodesEqual(s.Nodes, o.Nodes)
}

func (pgs *PgStatus) Equals(o *PgStatus) bool {
	return pgs.GenericStatus.Equals(&o.GenericStatus) &&
		pgs.DCsEqual(o.DataCentres)
}

func (pgs *PgStatus) DCsEqual(o []*PgDataCentreStatus) bool {
	if len(pgs.DataCentres) != len(o) {
		return false
	}

	m := map[string]*PgDataCentreStatus{}
	for _, dc := range pgs.DataCentres {
		m[dc.ID] = dc
	}

	for _, iDC := range o {
		dc, ok := m[iDC.ID]
		if !ok {
			return false
		}

		if !dc.Equals(iDC) {
			return false
		}
	}

	return true
}
