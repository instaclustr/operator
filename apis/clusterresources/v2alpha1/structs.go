package v2alpha1

type VPCPeeringSpec struct {
	ClusterDataCentreID string   `json:"cdcId"`
	PeerSubnets         []string `json:"peerSubnets"`
}

type VPCPeeringStatus struct {
	ID         string `json:"id"`
	StatusCode string `json:"statusCode"`
}
