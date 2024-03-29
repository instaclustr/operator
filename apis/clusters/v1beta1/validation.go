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
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
	"github.com/instaclustr/operator/pkg/validation"
)

func (ops *OnPremisesSpec) ValidateCreation() error {
	if ops.StorageClassName == "" || ops.DataDiskSize == "" || ops.OSDiskSize == "" || ops.NodeCPU == 0 ||
		ops.NodeMemory == "" || ops.OSImageURL == "" || ops.CloudInitScriptRef == nil {
		return fmt.Errorf("all on-premises spec fields except sshGatewayCPU and sshGatewayMemory if " +
			"it is not private cluster must not be empty")
	}

	osDiskSizeMatched, err := regexp.Match(models.StorageRegExp, []byte(ops.OSDiskSize))
	if !osDiskSizeMatched || err != nil {
		return fmt.Errorf("disk size field for node OS must fit pattern: %s",
			models.StorageRegExp)
	}

	dataDiskSizeMatched, err := regexp.Match(models.StorageRegExp, []byte(ops.DataDiskSize))
	if !dataDiskSizeMatched || err != nil {
		return fmt.Errorf("disk size field for storring cluster data must fit pattern: %s",
			models.StorageRegExp)
	}

	nodeMemoryMatched, err := regexp.Match(models.MemoryRegExp, []byte(ops.DataDiskSize))
	if !nodeMemoryMatched || err != nil {
		return fmt.Errorf("node memory field must fit pattern: %s",
			models.MemoryRegExp)
	}

	return nil
}

func (ops *OnPremisesSpec) ValidateSSHGatewayCreation() error {
	if ops.SSHGatewayCPU == 0 || ops.SSHGatewayMemory == "" {
		return fmt.Errorf("fields SSHGatewayCPU and SSHGatewayMemory must not be empty")
	}
	sshGatewayMemoryMatched, err := regexp.Match(models.MemoryRegExp, []byte(ops.DataDiskSize))
	if !sshGatewayMemoryMatched || err != nil {
		return fmt.Errorf("ssh gateway memory field must fit pattern: %s",
			models.MemoryRegExp)
	}
	return nil
}

func (cps *CloudProviderSettings) ValidateCreation() error {
	if (cps.ResourceGroup != "" && cps.DiskEncryptionKey != "") ||
		(cps.ResourceGroup != "" && cps.CustomVirtualNetworkID != "") {
		return fmt.Errorf("cluster should have cloud provider settings only for 1 cloud provider")
	}

	return nil
}

func validateReplicationFactor(availableReplicationFactors []int, rf int) error {
	for _, availableRf := range availableReplicationFactors {
		if availableRf == rf {
			return nil
		}
	}

	return fmt.Errorf("replication factor must be one of %v",
		availableReplicationFactors)
}

func validateAppVersion(
	versions []*models.AppVersions,
	appType string,
	version string) error {
	for _, appVersions := range versions {
		if appVersions.Application == appType {
			if !validation.Contains(version, appVersions.Versions) {
				return fmt.Errorf("%s version %s is unavailable, available versions: %v",
					appType, version, appVersions.Versions)
			}
		}
	}

	return nil
}

func validateTwoFactorDelete(new, old []*TwoFactorDelete) error {
	if len(old) == 0 && len(new) == 0 ||
		len(old) == 0 && len(new) == 1 {
		return nil
	}

	if len(new) > 1 {
		return models.ErrOnlyOneEntityTwoFactorDelete
	}

	if len(old) != len(new) {
		return models.ErrImmutableTwoFactorDelete
	}

	if *old[0] != *new[0] {
		return models.ErrImmutableTwoFactorDelete
	}

	return nil
}

func validateIngestNodes(new, old []*OpenSearchIngestNodes) error {
	if len(old) != len(new) {
		return models.ErrImmutableIngestNodes
	}

	if len(old) > 0 && *old[0] != *new[0] {
		return models.ErrImmutableIngestNodes
	}

	return nil
}

func validateClusterManagedNodes(new, old []*ClusterManagerNodes) error {
	if len(old) != len(new) {
		return models.ErrImmutableClusterManagedNodes
	}

	if len(old) > 0 && *old[0] != *new[0] {
		return models.ErrImmutableClusterManagedNodes
	}

	return nil
}

func validateTagsUpdate(new, old map[string]string) error {
	if len(old) != len(new) {
		return models.ErrImmutableTags
	}

	for newKey, newValue := range new {
		if oldValue, ok := old[newKey]; !ok || newValue != oldValue {
			return models.ErrImmutableTags
		}
	}

	return nil
}

func validatePrivateLinkUpdate(new, old []*PrivateLink) error {
	if len(old) != len(new) {
		return models.ErrImmutablePrivateLink
	}

	for i, oldPrivateLink := range old {
		if *oldPrivateLink != *new[i] {
			return models.ErrImmutablePrivateLink
		}
	}

	return nil
}

func validateSingleConcurrentResize(concurrentResizes int) error {
	if concurrentResizes > 1 {
		return models.ErrOnlySingleConcurrentResizeAvailable
	}

	return nil
}

//nolint:unused
func ContainsKubeVirtAddon(ctx context.Context, client client.Client) (bool, error) {
	namespaces := &k8scorev1.NamespaceList{}
	err := client.List(ctx, namespaces)
	if err != nil {
		return false, err
	}

	for _, namespace := range namespaces.Items {
		if strings.Contains(namespace.Name, models.KubeVirt) {
			return true, nil
		}
	}

	deployments := &k8sappsv1.DeploymentList{}
	err = client.List(ctx, deployments)
	if err != nil {
		return false, err
	}

	for _, deployment := range deployments.Items {
		if containsKubeVirtLabels(deployment.Labels) || containsKubeVirtLabels(deployment.Annotations) {
			return true, nil
		}
	}

	return false, nil
}

func containsKubeVirtLabels(l map[string]string) bool {
	for key, value := range l {
		if strings.Contains(key, models.KubeVirt) || strings.Contains(value, models.KubeVirt) {
			return true
		}
	}

	return false
}

func (s *GenericClusterSpec) immutableFields() immutableCluster {
	return immutableCluster{
		Name:                  s.Name,
		Version:               s.Version,
		PrivateNetworkCluster: s.PrivateNetwork,
		SLATier:               s.SLATier,
	}
}

var (
	clusterNameRegExp = regexp.MustCompile(models.ClusterNameRegExp)
	peerSubnetsRegExp = regexp.MustCompile(models.PeerSubnetsRegExp)
)

func (s *GenericClusterSpec) ValidateCreation() error {
	if !clusterNameRegExp.Match([]byte(s.Name)) {
		return fmt.Errorf("cluster name should have lenght from 3 to 32 symbols and fit pattern: %s",
			models.ClusterNameRegExp)
	}

	if len(s.TwoFactorDelete) > 1 {
		return fmt.Errorf("two factor delete should not have more than 1 item")
	}

	if !validation.Contains(s.SLATier, models.SLATiers) {
		return fmt.Errorf("cluster SLATier %s is unavailable, available values: %v",
			s.SLATier, models.SLATiers)
	}

	return nil
}

func (s *GenericDataCentreSpec) immutableFields() immutableDC {
	return immutableDC{
		Name:                s.Name,
		Region:              s.Region,
		CloudProvider:       s.CloudProvider,
		ProviderAccountName: s.ProviderAccountName,
		Network:             s.Network,
	}
}

func (s *GenericDataCentreSpec) validateCreation() error {
	if !validation.Contains(s.CloudProvider, models.CloudProviders) {
		return fmt.Errorf("cloud provider %s is unavailable for data centre: %s, available values: %v",
			s.CloudProvider, s.Name, models.CloudProviders)
	}

	switch s.CloudProvider {
	case models.AWSVPC:
		if !validation.Contains(s.Region, models.AWSRegions) {
			return fmt.Errorf("AWS Region: %s is unavailable, available regions: %v",
				s.Region, models.AWSRegions)
		}
	case models.AZUREAZ:
		if !validation.Contains(s.Region, models.AzureRegions) {
			return fmt.Errorf("azure Region: %s is unavailable, available regions: %v",
				s.Region, models.AzureRegions)
		}
	case models.GCP:
		if !validation.Contains(s.Region, models.GCPRegions) {
			return fmt.Errorf("GCP Region: %s is unavailable, available regions: %v",
				s.Region, models.GCPRegions)
		}
	case models.ONPREMISES:
		if s.Region != models.CLIENTDC {
			return fmt.Errorf("ONPREMISES Region: %s is unavailable, available regions: %v",
				s.Region, models.CLIENTDC)
		}
	}

	if s.ProviderAccountName == models.DefaultAccountName && s.hasCloudProviderSettings() {
		return fmt.Errorf("cloud provider settings can be used only with RIYOA accounts")
	}

	err := s.validateCloudProviderSettings()
	if err != nil {
		return err
	}

	if !peerSubnetsRegExp.Match([]byte(s.Network)) {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts and form the Private IP address. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28", s.Network)
	}

	return nil
}

func (s *GenericDataCentreSpec) ValidateOnPremisesCreation() error {
	if s.CloudProvider != models.ONPREMISES {
		return fmt.Errorf("cloud provider %s is unavailable for data centre: %s, available value: %s",
			s.CloudProvider, s.Name, models.ONPREMISES)
	}

	if s.Region != models.CLIENTDC {
		return fmt.Errorf("region %s is unavailable for data centre: %s, available value: %s",
			s.Region, s.Name, models.CLIENTDC)
	}

	return nil
}

func (s *GenericDataCentreSpec) validateImmutableCloudProviderSettingsUpdate(old *GenericDataCentreSpec) error {
	if !slices.EqualsPtr(s.AWSSettings, old.AWSSettings) {
		return models.ErrImmutableCloudProviderSettings
	}

	if !slices.EqualsPtr(s.GCPSettings, old.GCPSettings) {
		return models.ErrImmutableCloudProviderSettings
	}

	if !slices.EqualsPtr(s.AzureSettings, old.AzureSettings) {
		return models.ErrImmutableCloudProviderSettings
	}

	return nil
}

func (s *GenericDataCentreSpec) validateCloudProviderSettings() error {
	if sum := len(s.AWSSettings) + len(s.AzureSettings) + len(s.GCPSettings); sum > 1 {
		return errors.New("only one of [awsSettings, gcpSettings, azureSettings] should be set")
	}

	return nil
}

func (s *GenericDataCentreSpec) hasCloudProviderSettings() bool {
	return s.AWSSettings != nil || s.GCPSettings != nil && s.AzureSettings != nil
}

func validateNetwork(network string) error {
	ip, _, err := net.ParseCIDR(network)
	if err != nil {
		return err
	}

	ipParts := strings.Split(ip.String(), ".")
	secondOctet, err := strconv.Atoi(ipParts[1])
	if err != nil {
		return err
	}
	if secondOctet > 251 || secondOctet < 1 {
		return models.ErrInvalidCIDR
	}
	return nil
}

func IsClusterNotReadyForSpecUpdate(operation, state string, oldGen, newGen int64) bool {
	return (operation != models.NoOperation || state != models.RunningStatus) && newGen != oldGen
}
