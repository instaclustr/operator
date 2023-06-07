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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

const zookeeperNodeSizeFromYAML = "zookeeper-developer-t3.small-20"

var _ = Describe("Zookeeper Controller", func() {
	var (
		zookeeperResource v1alpha1.Zookeeper
		zookeeperYAML     v1alpha1.Zookeeper
		zook              = "zookeeper"
		ns                = "default"
		zookeeperNS       = types.NamespacedName{Name: zook, Namespace: ns}
		timeout           = time.Second * 40
		interval          = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/zookeeper_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &zookeeperYAML)
	Expect(err).NotTo(HaveOccurred())

	zookeeperObjMeta := metav1.ObjectMeta{
		Name:      zook,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	zookeeperYAML.ObjectMeta = zookeeperObjMeta

	ctx := context.Background()

	When("apply a zookeeper manifest", func() {
		It("should create a zookeeper resources", func() {
			Expect(k8sClient.Create(ctx, &zookeeperYAML)).Should(Succeed())
			By("sending zookeeper specification to the Instaclustr API and get ID of created cluster.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, zookeeperNS, &zookeeperResource); err != nil {
					return false
				}

				return zookeeperResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})

		When("zookeeper is created, the status job get new data from the InstAPI.", func() {
			It("updates a k8s zookeeper status", func() {
				Expect(k8sClient.Get(ctx, zookeeperNS, &zookeeperResource)).Should(Succeed())

				Eventually(func() bool {
					if err := k8sClient.Get(ctx, zookeeperNS, &zookeeperResource); err != nil {
						return false
					}

					if len(zookeeperResource.Status.DataCentres) == 0 || len(zookeeperResource.Status.DataCentres[0].Nodes) == 0 {
						return false
					}

					return zookeeperResource.Status.DataCentres[0].Nodes[0].Size == zookeeperNodeSizeFromYAML
				}, timeout, interval).Should(BeTrue())
			})
		})

		When("delete the zookeeper resource", func() {
			It("should send delete request to the Instaclustr API", func() {
				Expect(k8sClient.Get(ctx, zookeeperNS, &zookeeperResource)).Should(Succeed())

				zookeeperResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

				Expect(k8sClient.Delete(ctx, &zookeeperResource)).Should(Succeed())

				By("sending delete request to Instaclustr API")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, zookeeperNS, &zookeeperResource)
					if err != nil && !k8serrors.IsNotFound(err) {
						return false
					}

					return true
				}, timeout, interval).Should(BeTrue())
			})
		})
	})
})
