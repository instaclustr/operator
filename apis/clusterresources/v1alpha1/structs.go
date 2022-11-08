package v1alpha1

import "encoding/json"

type VPCPeeringSpec struct {
	DataCentreID string   `json:"cdcId"`
	PeerSubnets  []string `json:"peerSubnets"`
}

type PeeringStatus struct {
	ID            string `json:"id,omitempty"`
	StatusCode    string `json:"statusCode,omitempty"`
	Name          string `json:"name,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type FirewallRuleSpec struct {
	ClusterID string `json:"clusterId"`
	Type      string `json:"type"`
}

type FirewallRuleStatus struct {
	ID             string `json:"id,omitempty"`
	DeferredReason string `json:"deferredReason,omitempty"`
	Status         string `json:"status,omitempty"`
}
