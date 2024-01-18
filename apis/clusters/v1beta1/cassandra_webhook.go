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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

var cassandralog = logf.Log.WithName("cassandra-resource")

type cassandraValidator struct {
	Client client.Client
	API    validation.Validation
}

func (r *Cassandra) SetupWebhookWithManager(mgr ctrl.Manager, api validation.Validation) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).WithValidator(webhook.CustomValidator(&cassandraValidator{
		Client: mgr.GetClient(),
		API:    api,
	})).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/mutate-clusters-instaclustr-com-v1beta1-cassandra,mutating=true,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cassandras,verbs=create;update,versions=v1beta1,name=mcassandra.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-clusters-instaclustr-com-v1beta1-cassandra,mutating=false,failurePolicy=fail,sideEffects=None,groups=clusters.instaclustr.com,resources=cassandras,verbs=create;update,versions=v1beta1,name=vcassandra.kb.io,admissionReviewVersions=v1
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch

var _ webhook.CustomValidator = &cassandraValidator{}
var _ webhook.Defaulter = &Cassandra{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cassandra) Default() {
	cassandralog.Info("default", "name", c.Name)

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

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", obj.GetObjectKind())
	}

	cassandralog.Info("validate create", "name", c.Name)

	if c.Spec.RestoreFrom != nil {
		if c.Spec.RestoreFrom.ClusterID == "" {
			return fmt.Errorf("restore clusterID field is empty")
		} else {
			return nil
		}
	}

	err := c.Spec.Cluster.ValidateCreation()
	if err != nil {
		return err
	}

	contains, err := ContainsKubeVirtAddon(ctx, cv.Client)
	if err != nil {
		return err
	}

	if c.Spec.OnPremisesSpec != nil && c.Spec.OnPremisesSpec.EnableAutomation {
		if !contains {
			return models.ErrKubeVirtAddonNotFound
		}

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

	appVersions, err := cv.API.ListAppVersions(models.CassandraAppKind)
	if err != nil {
		return fmt.Errorf("cannot list versions for kind: %v, err: %w",
			models.CassandraAppKind, err)
	}

	err = validateAppVersion(appVersions, models.CassandraAppType, c.Spec.Version)
	if err != nil {
		return err
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
			err = dc.DataCentre.ValidateOnPremisesCreation()
			if err != nil {
				return err
			}
		} else {
			err = dc.DataCentre.ValidateCreation()
			if err != nil {
				return err
			}
		}

		if !c.Spec.PrivateNetworkCluster && dc.PrivateIPBroadcastForDiscovery {
			return fmt.Errorf("cannot use private ip broadcast for discovery on public network cluster")
		}

		err = validateReplicationFactor(models.CassandraReplicationFactors, dc.ReplicationFactor)
		if err != nil {
			return err
		}

		if ((dc.NodesNumber*dc.ReplicationFactor)/dc.ReplicationFactor)%dc.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", dc.ReplicationFactor)
		}

		err = c.Spec.validateResizeSettings(dc.NodesNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {
	c, ok := new.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", new.GetObjectKind())
	}

	cassandralog.Info("validate update", "name", c.Name)

	// skip validation when we receive cluster specification update from the Instaclustr Console.
	if c.Annotations[models.ExternalChangesAnnotation] == models.True {
		return nil
	}

	oldCluster, ok := old.(*Cassandra)
	if !ok {
		return models.ErrTypeAssertion
	}

	if oldCluster.Spec.BundledUseOnly && c.Generation != oldCluster.Generation {
		return models.ErrBundledUseOnlyResourceUpdateIsNotSupported
	}

	if oldCluster.Spec.RestoreFrom != nil {
		return nil
	}

	if c.Status.ID == "" {
		return cv.ValidateCreate(ctx, c)
	}

	err := c.Spec.validateUpdate(oldCluster.Spec)
	if err != nil {
		return fmt.Errorf("cannot update immutable fields: %v", err)
	}

	// ensuring if the cluster is ready for the spec updating
	if (c.Status.CurrentClusterOperationStatus != models.NoOperation || c.Status.State != models.RunningStatus) && c.Generation != oldCluster.Generation {
		return models.ErrClusterIsNotReadyToUpdate
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cv *cassandraValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	c, ok := obj.(*Cassandra)
	if !ok {
		return fmt.Errorf("cannot assert object %v to cassandra", obj.GetObjectKind())
	}

	cassandralog.Info("validate delete", "name", c.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

type immutableCassandraFields struct {
	specificCassandra
	immutableCluster
}

type specificCassandra struct {
	LuceneEnabled       bool
	PasswordAndUserAuth bool
}

type immutableCassandraDCFields struct {
	immutableDC
	specificCassandraDC
}

type specificCassandraDC struct {
	replicationFactor              int
	continuousBackup               bool
	privateIpBroadcastForDiscovery bool
	clientToClusterEncryption      bool
}

func (cs *CassandraSpec) newImmutableFields() *immutableCassandraFields {
	return &immutableCassandraFields{
		specificCassandra: specificCassandra{
			LuceneEnabled:       cs.LuceneEnabled,
			PasswordAndUserAuth: cs.PasswordAndUserAuth,
		},
		immutableCluster: cs.Cluster.newImmutableFields(),
	}
}

func (cs *CassandraSpec) validateUpdate(oldSpec CassandraSpec) error {
	newImmutableFields := cs.newImmutableFields()
	oldImmutableFields := oldSpec.newImmutableFields()

	if *newImmutableFields != *oldImmutableFields {
		return fmt.Errorf("cannot update immutable fields: old spec: %+v: new spec: %+v", oldImmutableFields, newImmutableFields)
	}

	err := cs.validateDataCentresUpdate(oldSpec)
	if err != nil {
		return err
	}
	err = validateTwoFactorDelete(cs.TwoFactorDelete, oldSpec.TwoFactorDelete)
	if err != nil {
		return err
	}

	for _, dc := range cs.DataCentres {
		err = cs.validateResizeSettings(dc.NodesNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *CassandraSpec) validateDataCentresUpdate(oldSpec CassandraSpec) error {
	if len(cs.DataCentres) < len(oldSpec.DataCentres) {
		return models.ErrDecreasedDataCentresNumber
	}

	toValidate := map[string]*CassandraDataCentre{}
	for _, dc := range oldSpec.DataCentres {
		toValidate[dc.Name] = dc
	}

	for _, newDC := range cs.DataCentres {
		oldDC, ok := toValidate[newDC.Name]
		if !ok {
			if len(cs.DataCentres) == len(oldSpec.DataCentres) {
				return fmt.Errorf("cannot change datacentre name: %v", newDC.Name)
			}

			if err := newDC.ValidateCreation(); err != nil {
				return err
			}

			if !cs.PrivateNetworkCluster && newDC.PrivateIPBroadcastForDiscovery {
				return fmt.Errorf("cannot use private ip broadcast for discovery on public network cluster")
			}

			err := validateReplicationFactor(models.CassandraReplicationFactors, newDC.ReplicationFactor)
			if err != nil {
				return err
			}

			if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
				return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
			}

		}

		newDCImmutableFields := newDC.newImmutableFields()
		oldDCImmutableFields := oldDC.newImmutableFields()

		if *newDCImmutableFields != *oldDCImmutableFields {
			return fmt.Errorf("cannot update immutable data centre fields: new spec: %v: old spec: %v", newDCImmutableFields, oldDCImmutableFields)
		}

		if ((newDC.NodesNumber*newDC.ReplicationFactor)/newDC.ReplicationFactor)%newDC.ReplicationFactor != 0 {
			return fmt.Errorf("number of nodes must be a multiple of replication factor: %v", newDC.ReplicationFactor)
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

		if !oldDC.DebeziumEquals(newDC) {
			return models.ErrDebeziumImmutable
		}

	}

	return nil
}

func (cdc *CassandraDataCentre) newImmutableFields() *immutableCassandraDCFields {
	return &immutableCassandraDCFields{
		immutableDC{
			Name:                cdc.Name,
			Region:              cdc.Region,
			CloudProvider:       cdc.CloudProvider,
			ProviderAccountName: cdc.ProviderAccountName,
			Network:             cdc.Network,
		},
		specificCassandraDC{
			replicationFactor:              cdc.ReplicationFactor,
			continuousBackup:               cdc.ContinuousBackup,
			privateIpBroadcastForDiscovery: cdc.PrivateIPBroadcastForDiscovery,
			clientToClusterEncryption:      cdc.ClientToClusterEncryption,
		},
	}
}

func (c *CassandraSpec) validateResizeSettings(nodeNumber int) error {
	for _, rs := range c.ResizeSettings {
		if rs.Concurrency > nodeNumber {
			return fmt.Errorf("resizeSettings.concurrency cannot be greater than number of nodes: %v", nodeNumber)
		}
	}

	return nil
}
