package models

type KafkaUserAPIv2 struct {
	Username           string                 `json:"username,omitempty"`
	Password           string                 `json:"password,omitempty"`
	Options            *KafkaUserOptionsAPIv2 `json:"options"`
	ClusterID          string                 `json:"clusterId"`
	InitialPermissions string                 `json:"initialPermissions"`
}

type KafkaUserOptionsAPIv2 struct {
	OverrideExistingUser bool   `json:"overrideExistingUser,omitempty"`
	SASLSCRAMMechanism   string `json:"saslScramMechanism"`
}
