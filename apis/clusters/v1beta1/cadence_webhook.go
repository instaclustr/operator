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

var cadencelog = logf.Log.WithName("cadence-resource")

type cadenceValidator struct {
	API validation.Validation
}

func (c *Cadence) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).WithValidator(webhook.CustomValidator(&cadenceValidator{
		API: api,
	})).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-cadence,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cadences,verbs=create;update,versions=v1beta1,name=mcadence.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Cadence{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cadence) Default() {
	cadencelog.Info("default", "name", c.Name)

	if c.Spec.Name == "" {
		c.Spec.Name = c.Name
	}

	if c.GetAnnotations() == nil {
		c.SetAnnotations(map[string]string{
			models.ResourceStateAnnotation: "",
		})
	}

	for _, dataCentre := range c.Spec.DataCentres {
		dataCentre.SetDefaultValues()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-cadence,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cadences,verbs=create;update,versions=v1beta1,name=vcadence.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &cadenceValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cv *cadenceValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cadence)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cadence", obj.GetObjectKind())
	}

	cadencelog.Info("validate create", "name", c.Name)

	err := c.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	if c.Spec.OnPremisesSpec != nil {
		err = c.Spec.OnPremisesSpec.ValidateCreation()
		if err != nil {
			return err
		}
		if c.Spec.PrivateNetworkCluster {
			err = c.Spec.OnPremisesSpec.ValidateSSHGatewayCreation()
			if err != nil {
				return err
			}
		}
	}

	appVersions, err := cv.API.ListAppVersions(models.CadenceAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.CadenceAppKind, err)
	}

	err = validateAppVersion(appVersions, models.CadenceAppType, c.Spec.Version)
	if err != nil {
		return err
	}

	if len(c.Spec.StandardProvisioning)+len(c.Spec.SharedProvisioning)+len(c.Spec.PackagedProvisioning) == 0 {
		return fmt.Errorf("one of StandardProvisioning, SharedProvisioning or PackagedProvisioning arrays must not be empty")
	}

	if len(c.Spec.StandardProvisioning)+len(c.Spec.SharedProvisioning)+len(c.Spec.PackagedProvisioning) > 1 {
		return fmt.Errorf("only one of StandardProvisioning, SharedProvisioning or PackagedProvisioning arrays must not be empty")
	}

	if len(c.Spec.AWSArchival) > 1 {
		return fmt.Errorf("AWSArchival array size must be between 0 and 1")
	}

	if len(c.Spec.StandardProvisioning) > 1 {
		return fmt.Errorf("StandardProvisioning array size must be between 0 and 1")
	}

	if len(c.Spec.SharedProvisioning) > 1 {
		return fmt.Errorf("SharedProvisioning array size must be between 0 and 1")
	}

	if len(c.Spec.PackagedProvisioning) > 1 {
		return fmt.Errorf("PackagedProvisioning array size must be between 0 and 1")
	}

	for _, awsArchival := range c.Spec.AWSArchival {
		err := awsArchival.validate()
		if err != nil {
			return err
		}
	}

	for _, sp := range c.Spec.StandardProvisioning {
		if len(sp.AdvancedVisibility) > 1 {
			return fmt.Errorf("AdvancedVisibility array size must be between 0 and 1")
		}

		err := sp.validate()
		if err != nil {
			return err
		}
	}

	for _, pp := range c.Spec.PackagedProvisioning {
		if (pp.UseAdvancedVisibility && pp.BundledKafkaSpec == nil) || (pp.UseAdvancedVisibility && pp.BundledOpenSearchSpec == nil) {
			return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must not be empty because UseAdvancedVisibility is set to true")
		}

		if pp.BundledKafkaSpec != nil {
			err := pp.BundledKafkaSpec.validate()
			if err != nil {
				return err
			}
		}

		if pp.BundledCassandraSpec != nil {
			err = pp.BundledCassandraSpec.validate()
			if err != nil {
				return err
			}
		}

		if pp.BundledOpenSearchSpec != nil {
			err = pp.BundledOpenSearchSpec.validate()
			if err != nil {
				return err
			}
		}
	}

	if len(c.Spec.TargetPrimaryCadence) > 1 {
		return fmt.Errorf("targetPrimaryCadence array must consist of <= 1 elements")
	}

	if len(c.Spec.DataCentres) == 0 {
		return fmt.Errorf("data centres field is empty")
	}

	//TODO: add support of multiple DCs for OnPrem clusters
	if len(c.Spec.DataCentres) > 1 && c.Spec.OnPremisesSpec != nil {
		return fmt.Errorf("on-premises cluster can be provisioned with only one data centre")
	}

	for _, dc := range c.Spec.DataCentres {
		if c.Spec.OnPremisesSpec != nil {
			err := dc.DataCentre.ValidateOnPremisesCreation()
			if err != nil {
				return err
			}
		} else {
			err := dc.DataCentre.ValidateCreation()
			if err != nil {
				return err
			}
		}

		if !c.Spec.PrivateNetworkCluster && dc.PrivateLink != nil {
			return models.ErrPrivateLinkOnlyWithPrivateNetworkCluster
		}

		if dc.CloudProvider != models.AWSVPC && dc.PrivateLink != nil {
			return models.ErrPrivateLinkSupportedOnlyForAWS
		}
	}

	for _, rs := range c.Spec.ResizeSettings {
		err := validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cv *cadenceValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	c, ok := new.(*Cadence)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cadence", new.GetObjectKind())
	}

	cadencelog.Info("validate update", "name", c.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if c.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster, ok := old.(*Cadence)
	if !ok {
		return models.ErrTypeAssertion
	}

	err := c.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	for _, rs := range c.Spec.ResizeSettings {
		err := validateSingleConcurrentResize(rs.Concurrency)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cv *cadenceValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cadence)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cadence", obj.GetObjectKind())
	}

	cadencelog.Info("validate delete", "name", c.Name)

	return nil
}

type immutableCadenceFields struct {
	immutableCluster
	UseCadenceWebAuth bool
	UseHTTPAPI        bool
}

type immutableCadenceDCFields struct {
	immutableDC      immutableDC
	ClientEncryption bool
}

func (cs *CadenceSpec) newImmutableFields() *immutableCadenceFields {
	return &immutableCadenceFields{
		immutableCluster:  cs.Cluster.newImmutableFields(),
		UseCadenceWebAuth: cs.UseCadenceWebAuth,
		UseHTTPAPI:        cs.UseHTTPAPI,
	}
}

func (cs *CadenceSpec) validateUpdate(oldSpec CadenceSpec) error {
	newImmutableFields := cs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldSpec, cs)
	}

	err := cs.validateImmutableDataCentresFieldsUpdate(oldSpec)
	if err != nil {
		return err
	}

	err = validateTwoFactorDelete(cs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	err = cs.validateAWSArchival(oldSpec.AWSArchival)
	if err != nil {
		return err
	}

	err = cs.validateStandardProvisioning(oldSpec.StandardProvisioning)
	if err != nil {
		return err
	}

	err = cs.validateSharedProvisioning(oldSpec.SharedProvisioning)
	if err != nil {
		return err
	}

	err = cs.validatePackagedProvisioning(oldSpec.PackagedProvisioning)
	if err != nil {
		return err
	}

	err = cs.validateTargetsPrimaryCadence(&oldSpec)
	if err != nil {
		return err
	}

	return nil
}

func (cs *CadenceSpec) validateAWSArchival(old []*AWSArchival) error {
	if len(cs.AWSArchival) != len(old) {
		return models.ErrImmutableAWSArchival
	}

	for i, arch := range cs.AWSArchival {
		if *arch != *old[i] {
			return models.ErrImmutableAWSArchival
		}
	}

	return nil
}

func (cs *CadenceSpec) validateStandardProvisioning(old []*StandardProvisioning) error {
	if len(cs.StandardProvisioning) != len(old) {
		return models.ErrImmutableStandardProvisioning
	}

	for i, sp := range cs.StandardProvisioning {
		if *sp.TargetCassandra != *old[i].TargetCassandra {
			return models.ErrImmutableStandardProvisioning
		}

		err := sp.validateAdvancedVisibility(sp.AdvancedVisibility, old[i].AdvancedVisibility)
		if err != nil {
			return err
		}

	}

	return nil
}

func (sp *StandardProvisioning) validateAdvancedVisibility(new, old []*AdvancedVisibility) error {
	if len(old) != len(new) {
		return models.ErrImmutableAdvancedVisibility
	}

	for i, av := range new {
		if *old[i].TargetKafka != *av.TargetKafka ||
			*old[i].TargetOpenSearch != *av.TargetOpenSearch {
			return models.ErrImmutableAdvancedVisibility
		}
	}

	return nil
}

func (cs *CadenceSpec) validateSharedProvisioning(old []*SharedProvisioning) error {
	if len(cs.SharedProvisioning) != len(old) {
		return models.ErrImmutableSharedProvisioning
	}

	for i, sp := range cs.SharedProvisioning {
		if *sp != *old[i] {
			return models.ErrImmutableSharedProvisioning
		}
	}

	return nil
}

func (cs *CadenceSpec) validatePackagedProvisioning(old []*PackagedProvisioning) error {
	if len(cs.PackagedProvisioning) != len(old) {
		return models.ErrImmutablePackagedProvisioning
	}

	for i, pp := range cs.PackagedProvisioning {
		if pp.UseAdvancedVisibility {
			if pp.BundledKafkaSpec == nil || pp.BundledOpenSearchSpec == nil {
				return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must not be empty because UseAdvancedVisibility is set to true")
			}
			if *pp.BundledKafkaSpec != *old[i].BundledKafkaSpec {
				return models.ErrImmutablePackagedProvisioning
			}
			if *pp.BundledOpenSearchSpec != *old[i].BundledOpenSearchSpec {
				return models.ErrImmutablePackagedProvisioning
			}
		} else {
			if pp.BundledKafkaSpec != nil || pp.BundledOpenSearchSpec != nil {
				return fmt.Errorf("BundledKafkaSpec and BundledOpenSearchSpec structs must be empty because UseAdvancedVisibility is set to false")
			}
		}

		if *pp.BundledCassandraSpec != *old[i].BundledCassandraSpec ||
			pp.UseAdvancedVisibility != old[i].UseAdvancedVisibility {
			return models.ErrImmutablePackagedProvisioning
		}
	}

	return nil
}

func (cs *CadenceSpec) validateTargetsPrimaryCadence(old *CadenceSpec) error {
	if len(cs.TargetPrimaryCadence) != len(old.TargetPrimaryCadence) {
		return fmt.Errorf("targetPrimaryCadence is immutable")
	}

	for _, oldTarget := range old.TargetPrimaryCadence {
		for _, newTarget := range cs.TargetPrimaryCadence {
			if *oldTarget != *newTarget {
				return fmt.Errorf("targetPrimaryCadence is immutable")
			}
		}
	}

	return nil
}

func (cdc *CadenceDataCentre) newImmutableFields() *immutableCadenceDCFields {
	return &immutableCadenceDCFields{
		immutableDC: immutableDC{
			Name:                cdc.Name,
			Region:              cdc.Region,
			CloudProvider:       cdc.CloudProvider,
			ProviderAccountName: cdc.ProviderAccountName,
			Network:             cdc.Network,
		},
		ClientEncryption: cdc.ClientEncryption,
	}
}

func (cs *CadenceSpec) validateImmutableDataCentresFieldsUpdate(oldSpec CadenceSpec) error {
	if len(cs.DataCentres) != len(oldSpec.DataCentres) {
		return models.ErrImmutableDataCentresNumber
	}

	for i, newDC := range cs.DataCentres {
		oldDC := oldSpec.DataCentres[i]
		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		if newDC.NodesNumber < oldDC.NodesNumber {
			return fmt.Errorf("deleting nodes is not supported. Number of nodes must be greater than: %v", oldDC.NodesNumber)
		}

		err := newDC.validateImmutableCloudProviderSettingsUpdate(oldDC.CloudProviderSettings)
		if err != nil {
			return err
		}

		err = validateTagsUpdate(newDC.Tags, oldDC.Tags)
		if err != nil {
			return err
		}

		err = newDC.validatePrivateLink(oldDC.PrivateLink)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cdc *CadenceDataCentre) validatePrivateLink(old []*PrivateLink) error {
	if len(cdc.PrivateLink) != len(old) {
		return models.ErrImmutablePrivateLink
	}

	for i, pl := range cdc.PrivateLink {
		if *pl != *old[i] {
			return models.ErrImmutablePrivateLink
		}
	}

	return nil
}

func (aws *AWSArchival) validate() error {
	if !validation.Contains(aws.ArchivalS3Region, models.AWSRegions) {
		return fmt.Errorf("AWS Region: %s is unavailable, available regions: %v",
			aws.ArchivalS3Region, models.AWSRegions)
	}

	s3URIMatched, err := regexp.Match(models.S3URIRegExp, []byte(aws.ArchivalS3URI))
	if !s3URIMatched || err != nil {
		return fmt.Errorf("the provided S3 URI: %s is not correct and must fit the pattern: %s. %v",
			aws.ArchivalS3URI, models.S3URIRegExp, err)
	}

	return nil
}

func (sp *StandardProvisioning) validate() error {
	for _, av := range sp.AdvancedVisibility {
		if !validation.Contains(av.TargetOpenSearch.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target OpenSearch Dependency VPC Type: %s is unavailable, available options: %v",
				av.TargetOpenSearch.DependencyVPCType, models.DependencyVPCs)
		}

		osDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(av.TargetOpenSearch.DependencyCDCID))
		if !osDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target OpenSearch Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				av.TargetOpenSearch.DependencyCDCID, models.UUIDStringRegExp, err)
		}

		if !validation.Contains(av.TargetKafka.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target Kafka Dependency VPC Type: %s is unavailable, available options: %v",
				av.TargetKafka.DependencyVPCType, models.DependencyVPCs)
		}

		kafkaDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(av.TargetKafka.DependencyCDCID))
		if !kafkaDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target Kafka Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				av.TargetKafka.DependencyCDCID, models.UUIDStringRegExp, err)
		}

		if !validation.Contains(sp.TargetCassandra.DependencyVPCType, models.DependencyVPCs) {
			return fmt.Errorf("target Kafka Dependency VPC Type: %s is unavailable, available options: %v",
				sp.TargetCassandra.DependencyVPCType, models.DependencyVPCs)
		}

		cassandraDependencyCDCIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(sp.TargetCassandra.DependencyCDCID))
		if !cassandraDependencyCDCIDMatched || err != nil {
			return fmt.Errorf("target Kafka Dependency CDCID: %s is not a UUID formated string. It must fit the pattern: %s. %v",
				sp.TargetCassandra.DependencyCDCID, models.UUIDStringRegExp, err)
		}
	}

	return nil
}

func (b *BundledKafkaSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(b.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", b.Network, err)
	}

	err = validateReplicationFactor(models.KafkaReplicationFactors, b.ReplicationFactor)
	if err != nil {
		return err
	}

	if ((b.NodesNumber*b.ReplicationFactor)/b.ReplicationFactor)%b.ReplicationFactor != 0 {
		return fmt.Errorf("kafka: number of nodes must be a multiple of replication factor: %v", b.ReplicationFactor)
	}

	return nil
}

func (c *BundledCassandraSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(c.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", c.Network, err)
	}

	err = validateReplicationFactor(models.CassandraReplicationFactors, c.ReplicationFactor)
	if err != nil {
		return err
	}

	if ((c.NodesNumber*c.ReplicationFactor)/c.ReplicationFactor)%c.ReplicationFactor != 0 {
		return fmt.Errorf("cassandra: number of nodes must be a multiple of replication factor: %v", c.ReplicationFactor)
	}

	return nil
}

func (o *BundledOpenSearchSpec) validate() error {
	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(o.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %v", o.Network, err)
	}

	err = validateReplicationFactor(models.OpenSearchReplicationFactors, o.ReplicationFactor)
	if err != nil {
		return err
	}

	return nil
}
