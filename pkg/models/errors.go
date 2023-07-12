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

package models

import "errors"

var (
	ErrZeroDataCentres                       = errors.New("cluster spec doesn't have data centres")
	ErrNetworkOverlaps                       = errors.New("cluster network overlaps")
	ErrImmutableTwoFactorDelete              = errors.New("twoFactorDelete field is immutable")
	ErrImmutableCloudProviderSettings        = errors.New("cloudProviderSettings are immutable")
	ErrImmutableIntraDataCentreReplication   = errors.New("intraDataCentreReplication fields are immutable")
	ErrImmutableInterDataCentreReplication   = errors.New("interDataCentreReplication fields are immutable")
	ErrNotValidPassword                      = errors.New("password must include at least 3 out of 4 of the following: (Uppercase, Lowercase, Number, Special Characters)")
	ErrImmutableDataCentresNumber            = errors.New("data centres number is immutable")
	ErrImmutableSpark                        = errors.New("spark field is immutable")
	ErrImmutableAWSSecurityGroupFirewallRule = errors.New("awsSecurityGroupFirewallRule is immutable")
	ErrImmutableTags                         = errors.New("tags field is immutable")
	ErrTypeAssertion                         = errors.New("unable to assert type")
	ErrImmutableSchemaRegistry               = errors.New("schema registry is immutable")
	ErrImmutableRestProxy                    = errors.New("rest proxy is immutable")
	ErrImmutableKarapaceSchemaRegistry       = errors.New("karapace schema registry is immutable")
	ErrImmutableKarapaceRestProxy            = errors.New("karapace rest proxy is immutable")
	ErrImmutableDedicatedZookeeper           = errors.New("dedicated zookeeper nodes cannot be changed")
	ErrDecreasedDataCentresNumber            = errors.New("data centres number cannot be decreased")
	ErrImmutableDataCentres                  = errors.New("data centres are immutable")
	ErrImmutableTargetCluster                = errors.New("TargetCluster field is immutable")
	ErrImmutableExternalCluster              = errors.New("ExternalCluster field is immutable")
	ErrImmutableManagedCluster               = errors.New("ManagedCluster field is immutable")
	ErrIncorrectStartHour                    = errors.New("startHour must be between 0 and 23 inclusive")
	ErrIncorrectDayOfWeek                    = errors.New("dayOfWeek is invalid")
	ErrTooManyExclusionHours                 = errors.New("cannot add more than 20 hours of exclusion times per week")
	ErrImmutableAWSArchival                  = errors.New("AWSArchival array is immutable")
	ErrImmutableStandardProvisioning         = errors.New("StandardProvisioning array is immutable")
	ErrImmutableSharedProvisioning           = errors.New("SharedProvisioning array is immutable")
	ErrImmutablePackagedProvisioning         = errors.New("PackagedProvisioning array is immutable")
	ErrImmutableAdvancedVisibility           = errors.New("AdvancedVisibility array is immutable")
	ErrImmutablePrivateLink                  = errors.New("PrivateLink array is immutable")
	ErrImmutableNodesNumber                  = errors.New("nodes number is immutable")
	ErrMissingSecretKeys                     = errors.New("the secret is missing the correct keys for the user")
	ErrUserStillExist                        = errors.New("the user is still attached to cluster")
)
