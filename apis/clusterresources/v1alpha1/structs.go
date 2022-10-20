package v1alpha1

import "encoding/json"

type VPCPeeringSpec struct {
	DataCentreID string   `json:"cdcId"`
	PeerSubnets  []string `json:"peerSubnets"`
}

type VPCPeeringStatus struct {
	ID         string `json:"id"`
	StatusCode string `json:"statusCode"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}
