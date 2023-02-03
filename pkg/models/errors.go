package models

import "errors"

var (
	ErrZeroDataCentres                     = errors.New("cluster spec doesn't have data centres")
	ErrEmptyAdvancedVisibility             = errors.New("advanced visibility fields are empty")
	ErrNetworkOverlaps                     = errors.New("cluster network overlaps")
	ErrImmutableTwoFactorDelete            = errors.New("twoFactorDelete field is immutable")
	ErrImmutableCloudProviderSettings      = errors.New("cloudProviderSettings are immutable")
	ErrImmutableIntraDataCentreReplication = errors.New("intraDataCentreReplication fields are immutable")
	ErrImmutableInterDataCentreReplication = errors.New("interDataCentreReplication fields are immutable")
	ErrNotValidPassword                    = errors.New("password must include at least 3 out of 4 of the following: (Uppercase, Lowercase, Number, Special Characters)")
	ErrImmutableDataCentresNumber          = errors.New("data centres number is immutable")
	ErrImmutableSpark                      = errors.New("spark field is immutable")
	ErrImmutableTags                       = errors.New("tags field is immutable")
	ErrTypeAssertion                       = errors.New("unable to assert type")
)
