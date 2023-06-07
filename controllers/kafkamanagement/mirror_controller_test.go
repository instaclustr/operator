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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var newMirrorLatency int32 = 3000

var _ = Describe("Kafka Mirror Controller", func() {
	var (
		mirrorResource v1alpha1.Mirror
		mirrorYAML     v1alpha1.Mirror
		m              = "mirror"
		ns             = "default"
		mirrorNS       = types.NamespacedName{Name: m, Namespace: ns}
		timeout        = time.Second * 40
		interval       = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/mirror_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &mirrorYAML)
	Expect(err).NotTo(HaveOccurred())

	mirrorObjMeta := metav1.ObjectMeta{
		Name:      m,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	mirrorYAML.ObjectMeta = mirrorObjMeta

	ctx := context.Background()

	When("apply a Mirror manifest", func() {
		It("should create a Mirror resources", func() {
			Expect(k8sClient.Create(ctx, &mirrorYAML)).Should(Succeed())
			By("sending a Mirror specification to the Instaclustr API and get an ID of created resource.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mirrorNS, &mirrorResource); err != nil {
					return false
				}

				return mirrorResource.Status.ID != openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a Mirror latency", func() {
		It("should update the Mirror resources", func() {
			Expect(k8sClient.Get(ctx, mirrorNS, &mirrorResource)).Should(Succeed())
			patch := mirrorResource.NewPatch()

			mirrorResource.Spec.TargetLatency = newMirrorLatency
			mirrorResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}

			Expect(k8sClient.Patch(ctx, &mirrorResource, patch)).Should(Succeed())

			By("sending a new Mirror configs request to the Instaclustr API, it" +
				"gets a new data from the InstAPI and update it in k8s Mirror resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mirrorNS, &mirrorResource); err != nil {
					return false
				}

				if mirrorResource.Status.TargetLatency != newMirrorLatency {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, mirrorNS, &mirrorResource)).Should(Succeed())

			mirrorResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}
			Expect(k8sClient.Delete(ctx, &mirrorResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mirrorNS, &mirrorResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}).Should(BeTrue())
		})
	})
})
