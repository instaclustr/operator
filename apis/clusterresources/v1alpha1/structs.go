package v1alpha1

type VPCPeeringSpec struct {
	ClusterDataCentreID       string   `json:"ClusterDataCentreId"`
	PeerSubnets               []string `json:"peerSubnets,omitempty"`
	AddNetworkToFirewallRules bool     `json:"addNetworkToFirewallRules,omitempty"`
}

type VPCPeeringStatus struct {
	ID         string `json:"id"`
	StatusCode string `json:"statusCode"`
}
