package v1alpha1

import (
	"fmt"
	"regexp"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func (c *Cluster) ValidateCreation(availableVersions []string) error {
	clusterNameMatched, err := regexp.Match(models.ClusterNameRegExp, []byte(c.Name))
	if !clusterNameMatched || err != nil {
		return fmt.Errorf("name should have lenght from 3 to 32 symbols and fit pattern: %s",
			models.ClusterNameRegExp)
	}

	if len(c.TwoFactorDelete) > 1 {
		return fmt.Errorf("twoFactorDelete should not have more than 1 item")
	}
	if !validation.Contains(c.Version, availableVersions) {
		return fmt.Errorf("version %s is unavailable, available values: %v",
			c.Version, availableVersions)
	}
	if !validation.Contains(c.SLATier, models.SLATiers) {
		return fmt.Errorf("slaTier %s is unavailable, available values: %v",
			c.SLATier, models.SLATiers)
	}

	return nil
}

func (dc *DataCentre) ValidateCreation() error {
	if !validation.Contains(dc.CloudProvider, models.CloudProviders) {
		return fmt.Errorf("cloudProvider %s is unavailable for data centre: %s, available values: %v",
			dc.CloudProvider, dc.Name, models.CloudProviders)
	}

	switch dc.CloudProvider {
	case "AWS_VPC":
		if !validation.Contains(dc.Region, models.AWSRegions) {
			return fmt.Errorf("AWS Region: %s is unavailable, available regions: %v",
				dc.Region, models.AWSRegions)
		}
	case "AZURE", "AZURE_AZ":
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

	networkMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(dc.Network))
	if !networkMatched || err != nil {
		return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts. All bits in the host part of the CIDR must be 0. Suffix must be between 16-28. %s", dc.Network, err.Error())
	}

	if len(dc.CloudProviderSettings) > 1 {
		return fmt.Errorf("cloudProviderSettings should not have more than 1 item")
	}

	if len(dc.CloudProviderSettings) == 1 {
		err := dc.CloudProviderSettings[0].ValidateCreation()
		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *DataCentre) validateImmutableCloudProviderSettingsUpdate(oldSettings []*CloudProviderSettings) error {
	if len(oldSettings) != len(dc.CloudProviderSettings) {
		return models.ErrImmutableCloudProviderSettings
	}

	for i, newProviderSettings := range dc.CloudProviderSettings {
		if *newProviderSettings != *oldSettings[i] {
			return models.ErrImmutableCloudProviderSettings
		}
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

func validateTwoFactorDelete(new, old []*TwoFactorDelete) error {
	if len(new) != 0 && len(old) == 0 {
		return nil
	}
	if len(old) != len(new) {
		return models.ErrImmutableTwoFactorDelete
	}
	if len(old) != 0 &&
		*old[0] != *new[0] {
		return models.ErrImmutableTwoFactorDelete
	}

	return nil
}

func validateSpark(new, old []*Spark) error {
	if len(old) != len(new) {
		return models.ErrImmutableSpark
	}
	if len(old) != 0 &&
		*old[0] != *new[0] {
		return models.ErrImmutableSpark
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
