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

const newCassandraNodeSize = "CAS-DEV-t4g.small-30"

var _ = Describe("Cassandra Controller", func() {
	cassandra := v1beta1.Cassandra{}
	cassandraManifest := v1beta1.Cassandra{}

	yfile, err := os.ReadFile("datatest/cassandra_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &cassandraManifest)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	clusterID := cassandraManifest.Spec.Name + openapi.CreatedID
	cassandraNamespacedName := types.NamespacedName{Name: cassandraManifest.ObjectMeta.Name, Namespace: defaultNS}

	When("apply a cassandra manifest", func() {
		It("should create a cassandra resources", func() {
			Expect(k8sClient.Create(ctx, &cassandraManifest)).Should(Succeed())

			done := tests.NewChannelWithTimeout(timeout)

			By("sending cassandra specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNamespacedName, &cassandra); err != nil {
					return false
				}

				return cassandra.Status.ID == clusterID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a node size", func() {
		It("should update a cassandra resources", func() {
			Expect(k8sClient.Get(ctx, cassandraNamespacedName, &cassandra)).Should(Succeed())

			patch := cassandra.NewPatch()
			cassandra.Spec.DataCentres[0].NodeSize = newCassandraNodeSize
			Expect(k8sClient.Patch(ctx, &cassandra, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s cassandra resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNamespacedName, &cassandra); err != nil {
					return false
				}

				if len(cassandra.Status.DataCentres) == 0 || len(cassandra.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return cassandra.Status.DataCentres[0].Nodes[0].Size == newCassandraNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the cassandra resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, cassandraNamespacedName, &cassandra)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &cassandra)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, cassandraNamespacedName, &cassandra)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
