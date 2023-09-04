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

package kafkamanagement

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
)

var newTopicConfig = map[string]string{
	"min.insync.replicas": "2",
	"retention.ms":        "603800000",
}

var _ = Describe("Kafka Topic Controller", func() {
	topic := v1beta1.Topic{}
	topicManifest := v1beta1.Topic{}

	yfile, err := os.ReadFile("datatest/topic_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &topicManifest)
	Expect(err).NotTo(HaveOccurred())

	topicNamespacedName := types.NamespacedName{Name: topicManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a Kafka topic manifest", func() {
		It("should create a topic resources", func() {
			Expect(k8sClient.Create(ctx, &topicManifest)).Should(Succeed())
			By("sending topic specification to the Instaclustr API and get ID of created resource.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, topicNamespacedName, &topic); err != nil {
					return false
				}

				return topic.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a topic configs", func() {
		It("should update the topic resources", func() {
			Expect(k8sClient.Get(ctx, topicNamespacedName, &topic)).Should(Succeed())

			patch := topic.NewPatch()
			topic.Spec.TopicConfigs = newTopicConfig
			Expect(k8sClient.Patch(ctx, &topic, patch)).Should(Succeed())
			By("sending a new topic configs request to the Instaclustr API, it" +
				"gets a new data from the InstAPI and update it in k8s Topic resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, topicNamespacedName, &topic); err != nil {
					return false
				}

				for newK, newV := range newTopicConfig {
					if oldV, ok := topic.Status.TopicConfigs[newK]; !ok || oldV != newV {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, topicNamespacedName, &topic)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &topic)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, topicNamespacedName, &topic)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}).Should(BeTrue())
		})
	})
})
