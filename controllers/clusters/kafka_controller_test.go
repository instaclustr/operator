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

package clusters

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/controllers/tests"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
)

const newKafkaNodeSize = "KFK-DEV-t4g.medium-80"

var _ = Describe("Kafka Controller", func() {
	kafka := v1beta1.Kafka{}
	kafkaManifest := v1beta1.Kafka{}

	yfile, err := os.ReadFile("datatest/kafka_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &kafkaManifest)
	Expect(err).NotTo(HaveOccurred())

	kafkaNamespacedName := types.NamespacedName{Name: kafkaManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a Kafka manifest", func() {
		It("should create a Kafka resources", func() {
			Expect(k8sClient.Create(ctx, &kafkaManifest)).Should(Succeed())
			done := tests.NewChannelWithTimeout(timeout)

			By("sending Kafka specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNamespacedName, &kafka); err != nil {
					return false
				}

				return kafka.Status.ID == openapi.CreatedID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a node size", func() {
		It("should update a Kafka resources", func() {
			Expect(k8sClient.Get(ctx, kafkaNamespacedName, &kafka)).Should(Succeed())

			patch := kafka.NewPatch()
			kafka.Spec.DataCentres[0].NodeSize = newKafkaNodeSize
			Expect(k8sClient.Patch(ctx, &kafka, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s Kafka resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNamespacedName, &kafka); err != nil {
					return false
				}

				if len(kafka.Status.DataCentres) == 0 || len(kafka.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return kafka.Status.DataCentres[0].Nodes[0].Size == newKafkaNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, kafkaNamespacedName, &kafka)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &kafka)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, kafkaNamespacedName, &kafka)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
