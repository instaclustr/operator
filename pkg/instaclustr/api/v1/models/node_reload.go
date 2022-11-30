package models

type NodeReload struct {
	Bundle string `json:"bundle"`
	NodeID string `json:"nodeId"`
}

type NodeReloadStatusAPIv1 struct {
	Operations []*Operation `json:"operations"`
}

type Operation struct {
	TimeCreated  int64  `json:"timeCreated"`
	TimeModified int64  `json:"timeModified"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}
