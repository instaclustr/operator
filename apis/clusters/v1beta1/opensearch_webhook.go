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
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var opensearchlog = logf.Log.WithName("opensearch-resource")

type openSearchValidator struct {
	API validation.Validation
}

func (r *OpenSearch) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&openSearchValidator{
		API: api,
	})).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-opensearch,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=opensearches,verbs=create;update,versions=v1beta1,name=vopensearch.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-opensearch,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=opensearches,verbs=create;update,versions=v1beta1,name=mopensearch.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &openSearchValidator{}
var _ webhook.Defaulter = &OpenSearch{}

func (os *OpenSearch) Default() {
	for _, dataCentre := range os.Spec.DataCentres {
		setDefaultValues(dataCentre)

		if os.GetAnnotations() == nil {
			os.SetAnnotations(map[string]string{
				models.ResourceStateAnnotation: "",
			})
		}

		if dataCentre.Name == "" {
			dataCentre.Name = dataCentre.Region
		}
	}

	opensearchlog.Info("default values are set", "name", os.Name)
}

func setDefaultValues(dc *OpenSearchDataCentre) {
	if dc.ProviderAccountName == "" {
		dc.ProviderAccountName = models.DefaultAccountName
	}
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (osv *openSearchValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	os, ok := obj.(*OpenSearch)
	if !ok {
		return fmt.Errorf("cannot assert object %v to openSearch", obj.GetObjectKind())
	}

	opensearchlog.Info("validate create", "name", os.Name)

	if os.Spec.RestoreFrom != nil {
		if os.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := os.Spec.ValidateCreation()
	if err != nil {
		return err
	}

	if len(os.Spec.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	err = os.Spec.validateDedicatedManager()
	if err != nil {
		return err
	}

	for _, dc := range os.Spec.DataCentres {
		err = validateCreation(dc)
		if err != nil {
			return err
		}

		err = dc.validateDataNode(os.Spec.DataNodes)
		if err != nil {
			return err
		}

		err = validateReplicationFactor(models.OpenSearchReplicationFactors, dc.ReplicationFactor)
		if err != nil {
			return err
		}
	}

	appVersions, err := osv.API.ListAppVersions(models.OpenSearchAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.OpenSearchAppKind, err)
	}

	err = validateAppVersion(appVersions, models.OpenSearchAppType, os.Spec.Version)
	if err != nil {
		return err
	}

	err = os.Spec.validateResizeSettings(len(os.Spec.ClusterManagerNodes))
	if err != nil {
		return err
	}

	for _, node := range os.Spec.DataNodes {
		err = os.Spec.validateResizeSettings(node.NodesNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateCreation(dc *OpenSearchDataCentre) error {
	if !validation.Contains(dc.CloudProvider, models.CloudProviders) {
		return fmt.Errorf("cloud provider %s is unavailable for data centre: %s, available values: %v",
			dc.CloudProvider, dc.Name, models.CloudProviders)
	}

	switch dc.CloudProvider {
	case "AWS_VPC":
		if !validation.Contains(dc.Region, models.AWSRegions) {
			return fmt.Errorf("AWS Region: %s is unavailable, available regions: %v",
				dc.Region, models.AWSRegions)
		}
	case "AZURE_AZ":
		if !validation.Contains(dc.Region, models.AzureRegions) {
			return fmt.Errorf("azure Region: %s is unavailable, available regions: %v",
				dc.Region, models.AzureRegions)
		}
	case "GCP":
		if !validation.Contains(dc.Region, models.GCPRegions) {
			return fmt.Errorf("GCP Region: %s is unavailable, available regions: %v",
				dc.Region, models.GCPRegions)
		}
	}

	if dc.ProviderAccountName == models.DefaultAccountName && len(dc.CloudProviderSettings) != 0 {
		return fmt.Errorf("cloud provider settings can be used only with RIYOA accounts")
	}

	if len(dc.CloudProviderSettings) > 1 {
		return fmt.Errorf("cloud provider settings should not have more than 1 item")
	}

	for _, cp := range dc.CloudProviderSettings {
		err := cp.ValidateCreation()
		if err != nil {
			return err
		}
	}

	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(dc.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", dc.Network, err)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (osv *openSearchValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	os, ok := new.(*OpenSearch)
	if !ok {
		return fmt.Errorf("cannot assert object %v to openSearch", new.GetObjectKind())
	}

	opensearchlog.Info("validate update", "name", os.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if os.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster := old.(*OpenSearch)
	if oldCluster.Spec.RestoreFrom != nil {
		return nil
	}

	if os.Status.ID == "" {
		return osv.ValidateCreate(ctx, os)
	}

	err := os.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (osv *openSearchValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	os, ok := obj.(*OpenSearch)
	if !ok {
		return fmt.Errorf("cannot assert object %v to openSearch", obj.GetObjectKind())
	}

	opensearchlog.Info("validate delete", "name", os.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

type immutableOpenSearchFields struct {
	specificFields specificOpenSearchFields
	cluster        immutableCluster
}

type specificOpenSearchFields struct {
	ICUPlugin                bool
	AsynchronousSearchPlugin bool
	KNNPlugin                bool
	ReportingPlugin          bool
	SQLPlugin                bool
	NotificationsPlugin      bool
	AnomalyDetectionPlugin   bool
	LoadBalancer             bool
	IndexManagementPlugin    bool
	AlertingPlugin           bool
	BundledUseOnly           bool
}

func (oss *OpenSearchSpec) newImmutableFields() *immutableOpenSearchFields {
	return &immutableOpenSearchFields{
		specificFields: specificOpenSearchFields{
			ICUPlugin:                oss.ICUPlugin,
			AsynchronousSearchPlugin: oss.AsynchronousSearchPlugin,
			KNNPlugin:                oss.KNNPlugin,
			ReportingPlugin:          oss.ReportingPlugin,
			SQLPlugin:                oss.SQLPlugin,
			NotificationsPlugin:      oss.NotificationsPlugin,
			AnomalyDetectionPlugin:   oss.AnomalyDetectionPlugin,
			LoadBalancer:             oss.LoadBalancer,
			IndexManagementPlugin:    oss.IndexManagementPlugin,
			AlertingPlugin:           oss.AlertingPlugin,
			BundledUseOnly:           oss.BundledUseOnly,
		},
		cluster: oss.Cluster.newImmutableFields(),
	}
}

type immutableOpenSearchDCFields struct {
	immutableDC
	specificOpenSearchDC
}

type specificOpenSearchDC struct {
	PrivateLink       bool
	ReplicationFactor int
}

func (oss *OpenSearchDataCentre) newImmutableFields() *immutableOpenSearchDCFields {
	return &immutableOpenSearchDCFields{
		immutableDC{
			Name:                oss.Name,
			Region:              oss.Region,
			CloudProvider:       oss.CloudProvider,
			ProviderAccountName: oss.ProviderAccountName,
			Network:             oss.Network,
		},
		specificOpenSearchDC{
			PrivateLink:       oss.PrivateLink,
			ReplicationFactor: oss.ReplicationFactor,
		},
	}
}

func (oss *OpenSearchSpec) validateDedicatedManager() error {
	for _, node := range oss.ClusterManagerNodes {
		if node.DedicatedManager && oss.DataNodes == nil {
			return fmt.Errorf("cluster with dedicated manager nodes must have data nodes")
		}
	}

	return nil
}

func (oss *OpenSearchSpec) validateUpdate(oldSpec OpenSearchSpec) error {
	newImmutableFields := oss.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldImmutableFields, newImmutableFields)
	}

	err := oss.validateImmutableDataCentresUpdate(oldSpec.DataCentres)
	if err != nil {
		return err
	}
	err = validateTwoFactorDelete(oss.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}
	err = validateDataNode(oss.DataNodes, oldSpec.DataNodes)
	if err != nil {
		return err
	}

	return nil
}

func (oss *OpenSearchSpec) validateImmutableDataCentresUpdate(oldDCs []*OpenSearchDataCentre) error {
	newDCs := oss.DataCentres
	if len(newDCs) != len(oldDCs) {
		return models.ErrImmutableDataCentresNumber
	}

	for _, newDC := range newDCs {
		for _, oldDC := range oldDCs {
			if oldDC.Name == newDC.Name {
				newDCImmutableFields := newDC.newImmutableFields()
				oldDCImmutableFields := oldDC.newImmutableFields()

				if *newDCImmutableFields != *oldDCImmutableFields {
					return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
				}

				err := validateImmutableCloudProviderSettingsUpdate(newDC.CloudProviderSettings, oldDC.CloudProviderSettings)
				if err != nil {
					return err
				}

				err = newDC.validateDataNode(oss.DataNodes)
				if err != nil {
					return err
				}

				err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (dc *OpenSearchDataCentre) validateDataNode(nodes []*OpenSearchDataNodes) error {
	for _, node := range nodes {
		if node.NodesNumber%dc.ReplicationFactor != 0 {
			return fmt.Errorf("number of data nodes must be a multiple of replication factor: %v", dc.ReplicationFactor)
		}
	}

	return nil
}

func validateImmutableCloudProviderSettingsUpdate(newSettings, oldSettings []*CloudProviderSettings) error {
	if len(oldSettings) != len(newSettings) {
		return models.ErrImmutableCloudProviderSettings
	}

	for i := range newSettings {
		if *newSettings[i] != *oldSettings[i] {
			return models.ErrImmutableCloudProviderSettings
		}
	}

	return nil
}

func validateDataNode(newNodes, oldNodes []*OpenSearchDataNodes) error {
	for i := range oldNodes {
		if oldNodes[i].NodesNumber > newNodes[i].NodesNumber {
			return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldNodes[i].NodesNumber)
		}
	}

	return nil
}
