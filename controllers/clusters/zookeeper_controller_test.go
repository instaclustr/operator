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
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
)

const zookeeperNodeSizeFromYAML = "zookeeper-developer-t3.small-20"

var _ = Describe("Zookeeper Controller", func() {
	zookeeper := v1beta1.Zookeeper{}
	zookeeperManifest := v1beta1.Zookeeper{}

	yfile, err := os.ReadFile("datatest/zookeeper_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &zookeeperManifest)
	Expect(err).NotTo(HaveOccurred())

	zookeeperNamespacedName := types.NamespacedName{Name: zookeeperManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a zookeeper manifest", func() {
		It("should create a zookeeper resources", func() {
			Expect(k8sClient.Create(ctx, &zookeeperManifest)).Should(Succeed())
			By("sending zookeeper specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, zookeeperNamespacedName, &zookeeper); err != nil {
					return false
				}

				return zookeeper.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})

		When("zookeeper is created, the status job get new data from the InstAPI.", func() {
			It("updates a k8s zookeeper status", func() {
				Expect(k8sClient.Get(ctx, zookeeperNamespacedName, &zookeeper)).Should(Succeed())

				Eventually(func() bool {
					if err := k8sClient.Get(ctx, zookeeperNamespacedName, &zookeeper); err != nil {
						return false
					}

					if len(zookeeper.Status.DataCentres) == 0 || len(zookeeper.Status.DataCentres[0].Nodes) == 0 {
						return false
					}

					return zookeeper.Status.DataCentres[0].Nodes[0].Size == zookeeperNodeSizeFromYAML
				}, timeout, interval).Should(BeTrue())
			})
		})

		When("delete the zookeeper resource", func() {
			It("should send delete request to the Instaclustr API", func() {
				Expect(k8sClient.Get(ctx, zookeeperNamespacedName, &zookeeper)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, &zookeeper)).Should(Succeed())
				By("sending delete request to Instaclustr API")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, zookeeperNamespacedName, &zookeeper)
					if err != nil && !k8serrors.IsNotFound(err) {
						return false
					}

					return k8serrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			})
		})
	})
})
