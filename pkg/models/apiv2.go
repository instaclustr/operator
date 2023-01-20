package models

const (
	NoOperation = "NO_OPERATION"

	DefaultAccountName = "INSTACLUSTR"
)

type NodeStatusV2 struct {
	ID             string   `json:"id"`
	Rack           string   `json:"rack"`
	NodeSize       string   `json:"nodeSize"`
	PublicAddress  string   `json:"publicAddress"`
	PrivateAddress string   `json:"privateAddress"`
	Status         string   `json:"status"`
	NodeRoles      []string `json:"nodeRoles"`
}
