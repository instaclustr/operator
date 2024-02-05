package models

const (
	OperatorUpgradeChecker = "operatorUpgradeChecker"
)

type DockerTagsResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}
