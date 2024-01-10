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

import (
	"errors"
)

var (
	ErrNotEmptyCSRs                               = errors.New("certificate creation allowed only if the user is created on the specific cluster")
	ErrEmptyCertGeneratingFields                  = errors.New("the fields for generating certificate signing request are empty")
	ErrZeroDataCentres                            = errors.New("cluster spec doesn't have data centres")
	ErrMoreThanOneKraft                           = errors.New("cluster spec does not support more than one kraft")
	ErrMoreThanThreeControllerNodeCount           = errors.New("kraft does not support more than three controller nodes")
	ErrNetworkOverlaps                            = errors.New("cluster network overlaps")
	ErrImmutableTwoFactorDelete                   = errors.New("twoFactorDelete field is immutable")
	ErrImmutableCloudProviderSettings             = errors.New("cloudProviderSettings are immutable")
	ErrImmutableIntraDataCentreReplication        = errors.New("intraDataCentreReplication fields are immutable")
	ErrImmutableInterDataCentreReplication        = errors.New("interDataCentreReplication fields are immutable")
	ErrImmutableDataCentresNumber                 = errors.New("data centres number is immutable")
	ErrImmutableAWSSecurityGroupFirewallRule      = errors.New("awsSecurityGroupFirewallRule is immutable")
	ErrImmutableTags                              = errors.New("tags field is immutable")
	ErrTypeAssertion                              = errors.New("unable to assert type")
	ErrImmutableSchemaRegistry                    = errors.New("schema registry is immutable")
	ErrImmutableRestProxy                         = errors.New("rest proxy is immutable")
	ErrImmutableKraft                             = errors.New("kraft is immutable")
	ErrImmutableKarapaceSchemaRegistry            = errors.New("karapace schema registry is immutable")
	ErrImmutableKarapaceRestProxy                 = errors.New("karapace rest proxy is immutable")
	ErrImmutableDedicatedZookeeper                = errors.New("dedicated zookeeper nodes cannot be changed")
	ErrDecreasedDataCentresNumber                 = errors.New("data centres number cannot be decreased")
	ErrImmutableTargetCluster                     = errors.New("TargetCluster field is immutable")
	ErrImmutableExternalCluster                   = errors.New("ExternalCluster field is immutable")
	ErrImmutableManagedCluster                    = errors.New("ManagedCluster field is immutable")
	ErrIncorrectDayOfWeek                         = errors.New("dayOfWeek field is invalid")
	ErrImmutableAWSArchival                       = errors.New("AWSArchival array is immutable")
	ErrImmutableStandardProvisioning              = errors.New("StandardProvisioning array is immutable")
	ErrImmutableSharedProvisioning                = errors.New("SharedProvisioning array is immutable")
	ErrImmutablePackagedProvisioning              = errors.New("PackagedProvisioning array is immutable")
	ErrImmutableAdvancedVisibility                = errors.New("AdvancedVisibility array is immutable")
	ErrImmutablePrivateLink                       = errors.New("PrivateLink array is immutable")
	ErrImmutableNodesNumber                       = errors.New("nodes number is immutable")
	ErrImmutableSecretRef                         = errors.New("secret reference is immutable")
	ErrEmptySecretRef                             = errors.New("secretRef.name and secretRef.namespace should not be empty")
	ErrMissingSecretKeys                          = errors.New("the secret is missing the correct keys for the user")
	ErrUserStillExist                             = errors.New("the user is still attached to the cluster. If you want to delete the user, remove the user from the cluster specification first")
	ErrOnlyOneEntityTwoFactorDelete               = errors.New("currently only one entity of two factor delete can be filled")
	ErrPrivateLinkOnlyWithPrivateNetworkCluster   = errors.New("private link is available only for private network clusters")
	ErrPrivateLinkSupportedOnlyForSingleDC        = errors.New("private link is only supported for a single data centre")
	ErrPrivateLinkSupportedOnlyForAWS             = errors.New("private link is supported only for an AWS cloud provider")
	ErrImmutableSpec                              = errors.New("resource specification is immutable")
	ErrUnsupportedBackupClusterKind               = errors.New("backups for provided cluster kind are not supported")
	ErrUnsupportedClusterKind                     = errors.New("provided cluster kind is not supported")
	ErrExposeServiceNotCreatedYet                 = errors.New("expose service is not created yet")
	ErrExposeServiceEndpointsNotCreatedYet        = errors.New("expose service endpoints is not created yet")
	ErrOnlySingleConcurrentResizeAvailable        = errors.New("only single concurrent resize is allowed")
	ErrBundledUseOnlyResourceUpdateIsNotSupported = errors.New("updating of bundled use resource is not supported")
	ErrDebeziumImmutable                          = errors.New("debezium array is immutable")
	ErrEmptyNamespace                             = errors.New("namespace field is empty")
	ErrEmptyName                                  = errors.New("name field is empty")
	ErrCreateClusterWithMultiDC                   = errors.New("multiple data center is still not supported. Please create a cluster with one data centre and add a second one when the cluster is in the running state")
	ErrOnPremicesWithMultiDC                      = errors.New("on-premises cluster can be provisioned with only one data centre")
	ErrUnsupportedDeletingDC                      = errors.New("deleting data centre is not supported")
)
