package models

import "errors"

// errors for operator
var (
	ZeroDataCentres            = errors.New("cluster spec doesn't have data centres")
	ErrEmptyAdvancedVisibility = errors.New("advanced visibility fields are empty")
	NetworkOverlaps            = errors.New("cluster network overlaps")
	NotValidPassword           = errors.New("password must include at least 3 out of 4 of the following: (Uppercase, Lowercase, Number, Special Characters)")
)
