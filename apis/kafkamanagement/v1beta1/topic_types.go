/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

// TopicSpec defines the desired state of Topic
type TopicSpec struct {
	Partitions        int32             `json:"partitions"`
	ReplicationFactor int32             `json:"replicationFactor"`
	TopicName         string            `json:"topic"`
	ClusterID         string            `json:"clusterId"`
	TopicConfigs      map[string]string `json:"configs,omitempty"`
}

// TopicStatus defines the observed state of Topic
type TopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ID for the Kafka topic. The value of this property has the form: [cluster-id]_[kafka-topic]
	ID string `json:"id"`

	TopicConfigs map[string]string `json:"configs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"

// Topic is the Schema for the topics API
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec,omitempty"`
	Status TopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

func (t *Topic) NewPatch() client.Patch {
	old := t.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&Topic{}, &TopicList{})
}

func (ts *TopicStatus) FromInstAPI(body []byte) (TopicStatus, error) {
	var status models.TopicStatus
	err := json.Unmarshal(body, &status)
	if err != nil {
		return TopicStatus{}, err
	}

	StatusCRD := TopicStatus{
		ID:           status.ID,
		TopicConfigs: ts.topicConfigsToCRD(status.TopicConfigs),
	}

	return StatusCRD, nil
}

func (ts *TopicStatus) topicConfigsToCRD(c []*models.TopicConfigs) map[string]string {
	if c == nil {
		return nil
	}

	crdConfigs := make(map[string]string)

	for _, config := range c {
		crdConfigs[config.Key] = config.Value
	}

	return crdConfigs
}

func (ts *TopicSpec) TopicConfigsUpdateToInstAPI() *models.UpdateTopicConfigs {
	iTags := &models.UpdateTopicConfigs{}
	for k, v := range ts.TopicConfigs {
		iTags.TopicConfigs = append(iTags.TopicConfigs, &models.TopicConfigs{
			Key:   k,
			Value: v,
		})
	}

	return iTags
}
