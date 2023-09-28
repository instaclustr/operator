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
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/controllers/tests"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

const newKafkaConnectNodeNumbers = 6

var _ = Describe("Kafka Connect Controller", func() {
	kafkaConnect := v1beta1.KafkaConnect{}
	kafkaConnectManifest := v1beta1.KafkaConnect{}

	yfile, err := os.ReadFile("datatest/kafkaconnect_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &kafkaConnectManifest)
	Expect(err).NotTo(HaveOccurred())

	kafkaConnectNamespacedName := types.NamespacedName{Name: kafkaConnectManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	clusterID := kafkaConnectManifest.Spec.Name + openapi.CreatedID
	When("apply a Kafka Connect manifest", func() {
		It("should create a Kafka Connect resources", func() {
			Expect(k8sClient.Create(ctx, &kafkaConnectManifest)).Should(Succeed())
			done := tests.NewChannelWithTimeout(timeout)

			By("sending Kafka Connect specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaConnectNamespacedName, &kafkaConnect); err != nil {
					return false
				}

				return kafkaConnect.Status.ID == clusterID
			}).Should(BeTrue())

			<-done

			By("creating secret with the default user credentials")
			secret := kafkaConnect.NewDefaultUserSecret("", "")
			secretNamespacedName := types.NamespacedName{
				Namespace: secret.Namespace,
				Name:      secret.Name,
			}

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, secretNamespacedName, secret); err != nil {
					return false
				}

				return string(secret.Data[models.Username]) == kafkaConnect.Status.ID+"_username" &&
					string(secret.Data[models.Password]) == kafkaConnect.Status.ID+"_password"
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {
		It("should update a Kafka Connect resources", func() {
			Expect(k8sClient.Get(ctx, kafkaConnectNamespacedName, &kafkaConnect)).Should(Succeed())

			patch := kafkaConnect.NewPatch()
			kafkaConnect.Spec.DataCentres[0].NodesNumber = newKafkaConnectNodeNumbers
			Expect(k8sClient.Patch(ctx, &kafkaConnect, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s Kafka resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaConnectNamespacedName, &kafkaConnect); err != nil {
					return false
				}

				if len(kafkaConnect.Status.DataCentres) == 0 {
					return false
				}

				return kafkaConnect.Status.DataCentres[0].NodeNumber == newKafkaConnectNodeNumbers
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka Connect resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, kafkaConnectNamespacedName, &kafkaConnect)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &kafkaConnect)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, kafkaConnectNamespacedName, &kafkaConnect)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("Deleting the Kafka Connect resource by avoiding operator", func() {
		It("should try to get the cluster details and receive StatusNotFound", func() {
			kafkaConnectManifest2 := kafkaConnectManifest.DeepCopy()
			kafkaConnectManifest2.Name += "-test-external-delete"
			kafkaConnectManifest2.ResourceVersion = ""

			kafkaConnect2 := v1beta1.KafkaConnect{}
			kafkaConnect2NamespacedName := types.NamespacedName{
				Namespace: kafkaConnectManifest2.Namespace,
				Name:      kafkaConnectManifest2.Name,
			}

			Expect(k8sClient.Create(ctx, kafkaConnectManifest2)).Should(Succeed())

			doneCh := tests.NewChannelWithTimeout(timeout)

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaConnect2NamespacedName, &kafkaConnect2); err != nil {
					return false
				}

				if kafkaConnect2.Status.State != models.RunningStatus {
					return false
				}

				doneCh <- struct{}{}

				return true
			}, timeout, interval).Should(BeTrue())

			<-doneCh

			By("using all possible ways other than k8s to delete a resource, the k8s operator scheduler should notice the changes and set the status of the deleted resource accordingly")
			Expect(instaClient.DeleteCluster(kafkaConnect2.Status.ID, instaclustr.KafkaConnectEndpoint)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, kafkaConnect2NamespacedName, &kafkaConnect2)
				if err != nil {
					return false
				}

				return kafkaConnect2.Status.State == models.DeletedStatus
			}, timeout, interval).Should(BeTrue())
		})
	})
})
