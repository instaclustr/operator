package models

const (
	Username = "Username"
	Password = "Password"
)

type KafkaUser struct {
	Username           string            `json:"username,omitempty"`
	Password           string            `json:"password,omitempty"`
	Options            *KafkaUserOptions `json:"options"`
	ClusterID          string            `json:"clusterId"`
	InitialPermissions string            `json:"initialPermissions"`
}

type KafkaUserOptions struct {
	OverrideExistingUser bool   `json:"overrideExistingUser,omitempty"`
	SASLSCRAMMechanism   string `json:"saslScramMechanism"`
}
