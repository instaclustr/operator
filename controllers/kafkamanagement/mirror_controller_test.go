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
	"github.com/instaclustr/operator/controllers/tests"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
)

var newMirrorLatency int32 = 3000

var _ = Describe("Kafka Mirror Controller", func() {
	mirror := v1beta1.Mirror{}
	mirrorManifest := v1beta1.Mirror{}

	yfile, err := os.ReadFile("datatest/mirror_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &mirrorManifest)
	Expect(err).NotTo(HaveOccurred())

	mirrorNamespacedName := types.NamespacedName{Name: mirrorManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a Mirror manifest", func() {
		It("should create a Mirror resources", func() {
			Expect(k8sClient.Create(ctx, &mirrorManifest)).Should(Succeed())
			By("sending a Mirror specification to the Instaclustr API and get an ID of created resource.")
			done := tests.NewChannelWithTimeout(timeout)
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mirrorNamespacedName, &mirror); err != nil {
					return false
				}

				return mirror.Status.ID == openapi.CreatedID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a Mirror latency", func() {
		It("should update the Mirror resources", func() {
			Expect(k8sClient.Get(ctx, mirrorNamespacedName, &mirror)).Should(Succeed())

			patch := mirror.NewPatch()
			mirror.Spec.TargetLatency = newMirrorLatency
			Expect(k8sClient.Patch(ctx, &mirror, patch)).Should(Succeed())

			By("sending a new Mirror configs request to the Instaclustr API, it" +
				"gets a new data from the InstAPI and update it in k8s Mirror resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mirrorNamespacedName, &mirror); err != nil {
					return false
				}

				if mirror.Status.TargetLatency != newMirrorLatency {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, mirrorNamespacedName, &mirror)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &mirror)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mirrorNamespacedName, &mirror)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}).Should(BeTrue())
		})
	})
})
