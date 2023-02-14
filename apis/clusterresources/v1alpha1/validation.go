package v1alpha1

import (
	"fmt"
	"regexp"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func (aws *AWSVPCPeeringSpec) Validate(availableRegions []string) error {
	peerAWSAccountIDMatched, err := regexp.Match(models.PeerAWSAccountIDRegExp, []byte(aws.PeerAWSAccountID))
	if !peerAWSAccountIDMatched || err != nil {
		return fmt.Errorf("peerAwsAccountId should contain 12-digit number, that uniquely identifies an "+
			"AWS account and fit pattern: %s", models.PeerAWSAccountIDRegExp)
	}

	peerAWSVPCIDMatched, err := regexp.Match(models.PeerVPCIDRegExp, []byte(aws.PeerVPCID))
	if !peerAWSVPCIDMatched || err != nil {
		return fmt.Errorf("peerVpcId must begin with 'vpc-' and fit pattern: %s", models.PeerVPCIDRegExp)
	}

	dataCentreIDMatched, err := regexp.Match(models.UUIDStringRegExp, []byte(aws.DataCentreID))
	if !dataCentreIDMatched || err != nil {
		return fmt.Errorf("cdcId is a UUID formated string. It must fit the pattern: %s",
			models.UUIDStringRegExp,
		)
	}

	if !validation.Contains(aws.PeerRegion, availableRegions) {
		return fmt.Errorf("peerRegion %s is unavailable, available values: %v",
			aws.PeerRegion, availableRegions)
	}

	if len(aws.PeerSubnets) > 0 {
		for _, subnet := range aws.PeerSubnets {
			peerSubnetMatched, err := regexp.Match(models.PeerSubnetsRegExp, []byte(subnet))
			if !peerSubnetMatched || err != nil {
				return fmt.Errorf("the provided CIDR: %s must contain four dot separated parts. "+
					"All bits in the host part of the CIDR must be 0. Suffix must be between 16-28", subnet)
			}
		}
	}

	return nil
}
