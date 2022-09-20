package clusters

type Node struct {
	NodeID         string   `json:"id,omitempty"`
	NodeSize       string   `json:"nodeSize,omitempty"`
	NodeStatus     string   `json:"status,omitempty"`
	NodeRole       []string `json:"nodeRoles,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber,omitempty"`
	ConfirmationEmail       string `json:"confirmationEmail"`
}

type Tag struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}
