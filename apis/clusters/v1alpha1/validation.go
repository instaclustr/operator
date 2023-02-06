package v1alpha1

import (
	"fmt"
	"regexp"

	"github.com/instaclustr/operator/pkg/models"
)

func Contains(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (c *Cluster) ValidateCreation(availableVersions []string) error {
	clusterNameMatched, err := regexp.Match(models.ClusterNameRegExp, []byte(c.Name))
	if !clusterNameMatched || err != nil {
		return fmt.Errorf("cluster name should have lenght from 3 to 32 symbols and fit pattern: %s",
			models.ClusterNameRegExp)
	}

	if len(c.TwoFactorDelete) > 1 {
		return fmt.Errorf("two factor delete should not have more than 1 item")
	}
	if !Contains(c.Version, availableVersions) {
		return fmt.Errorf("cluster version %s is unavailable, available versions: %v",
			c.Version, availableVersions)
	}
	if !Contains(c.SLATier, models.SLATiers) {
		return fmt.Errorf("cluster SLATier %s is unavailable, available values: %v",
			c.SLATier, models.SLATiers)
	}

	return nil
}

func (dc *DataCentre) ValidateCreation() error {
	if !Contains(dc.CloudProvider, models.CloudProviders) {
		return fmt.Errorf("cloud provider %s is unavailable for data centre: %s, available values: %v",
			dc.CloudProvider, dc.Name, models.CloudProviders)
	}
	if len(dc.CloudProviderSettings) > 1 {
		return fmt.Errorf("cloud provider settings should not have more than 1 item")
	}
	if len(dc.CloudProviderSettings) == 1 {
		err := dc.CloudProviderSettings[0].ValidateCreation()
		if err != nil {
			return err
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
