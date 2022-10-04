package models

type ClusterStatus struct {
	ID string `json:"id,omitempty"`

	// Status shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	Status      string              `json:"status,omitempty"`
	DataCentres []*DataCentreStatus `json:"dataCentres,omitempty"`
}

type DataCentreStatus struct {
	ID            string        `json:"id"`
	Status        string        `json:"status"`
	Nodes         []*NodeStatus `json:"nodes"`
	NumberOfNodes int32         `json:"numberOfNodes"`
}

type NodeStatus struct {
	ID             string   `json:"id"`
	Rack           string   `json:"rack"`
	NodeSize       string   `json:"nodeSize"`
	PublicAddress  string   `json:"publicAddress"`
	PrivateAddress string   `json:"privateAddress"`
	Status         string   `json:"status"`
	NodeRoles      []string `json:"nodeRoles"`
}
