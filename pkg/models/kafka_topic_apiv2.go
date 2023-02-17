package models

type TopicConfigs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TopicStatus struct {
	ID           string          `json:"id"`
	TopicConfigs []*TopicConfigs `json:"configs"`
}

type UpdateTopicConfigs struct {
	TopicConfigs []*TopicConfigs `json:"configs"`
}
