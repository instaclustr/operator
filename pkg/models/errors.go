package models

import "errors"

// errors for operator
var (
	ZeroDataCentres            = errors.New("cluster spec doesn't have data centres")
	ErrEmptyAdvancedVisibility = errors.New("advanced visibility fields are empty")
	NetworkOverlaps            = errors.New("cluster network overlaps")
)
