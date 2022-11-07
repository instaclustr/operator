package models

import "errors"

var (
	ZeroDataCentres = errors.New("cluster spec doesn't have data centres")
	NetworkOverlaps = errors.New("cluster network overlaps")
)
