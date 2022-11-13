package convertors

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
)

type topicConfigs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type topicStatus struct {
	ID           string          `json:"id"`
	TopicConfigs []*topicConfigs `json:"configs"`
}

type UpdateTopicConfigs struct {
	TopicConfigs []*topicConfigs `json:"configs"`
}

func TopicStatusFromInstAPI(body []byte) (v1alpha1.TopicStatus, error) {
	var status topicStatus
	err := json.Unmarshal(body, &status)
	if err != nil {
		return v1alpha1.TopicStatus{}, err
	}

	StatusCRD := v1alpha1.TopicStatus{
		ID:           status.ID,
		TopicConfigs: topicConfigsToCRD(status.TopicConfigs),
	}

	return StatusCRD, nil
}

func topicConfigsToCRD(c []*topicConfigs) map[string]string {
	if c == nil {
		return nil
	}

	crdConfigs := make(map[string]string)

	for _, config := range c {
		crdConfigs[config.Key] = config.Value
	}

	return crdConfigs
}

func TopicConfigsUpdateToInstAPI(crdTopic map[string]string) *UpdateTopicConfigs {
	var instaTags UpdateTopicConfigs

	for k, v := range crdTopic {
		instaTags.TopicConfigs = append(instaTags.TopicConfigs, &topicConfigs{
			Key:   k,
			Value: v,
		})
	}

	return &instaTags
}
