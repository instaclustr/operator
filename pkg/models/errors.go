package models

import "errors"

var (
	ErrZeroDataCentres                       = errors.New("cluster spec doesn't have data centres")
	ErrMultipleDataCentres                   = errors.New("cluster spec cannot have more than one data centre")
	ErrEmptyAdvancedVisibility               = errors.New("advanced visibility fields are empty")
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
	ErrImmutableZookeeperNodeNumber          = errors.New("zookeeper node number is immutable")
	ErrImmutableDedicatedZookeeper           = errors.New("additional dedicated nodes cannot be added")
)
