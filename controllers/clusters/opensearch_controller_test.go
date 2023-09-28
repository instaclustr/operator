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
	"github.com/instaclustr/operator/pkg/instaclustr"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

const newOpenSearchNodeSize = "SRH-DEV-t4g.small-30"

var _ = Describe("OpenSearch Controller", func() {
	openSearch := v1beta1.OpenSearch{}
	openSearchManifest := v1beta1.OpenSearch{}

	yfile, err := os.ReadFile("datatest/opensearch_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &openSearchManifest)
	Expect(err).NotTo(HaveOccurred())

	openSearchNamespacedName := types.NamespacedName{Name: openSearchManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()
	clusterID := openSearchManifest.Spec.Name + openapi.CreatedID

	When("apply a OpenSearch manifest", func() {
		It("should create a OpenSearch resources", func() {
			Expect(k8sClient.Create(ctx, &openSearchManifest)).Should(Succeed())
			done := tests.NewChannelWithTimeout(timeout)

			By("sending OpenSearch specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNamespacedName, &openSearch); err != nil {
					return false
				}

				return openSearch.Status.ID == clusterID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a node size", func() {
		It("should update a OpenSearch resources", func() {
			Expect(k8sClient.Get(ctx, openSearchNamespacedName, &openSearch)).Should(Succeed())

			patch := openSearch.NewPatch()
			openSearch.Spec.ClusterManagerNodes[0].NodeSize = newOpenSearchNodeSize
			Expect(k8sClient.Patch(ctx, &openSearch, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s OpenSearch resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNamespacedName, &openSearch); err != nil {
					return false
				}

				return openSearch.Spec.ClusterManagerNodes[0].NodeSize == newOpenSearchNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the OpenSearch resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, openSearchNamespacedName, &openSearch)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &openSearch)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, openSearchNamespacedName, &openSearch)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("Deleting the Kafka Connect resource by avoiding operator", func() {
		It("should try to get the cluster details and receive StatusNotFound", func() {
			openSearchManifest2 := openSearchManifest.DeepCopy()
			openSearchManifest2.Name += "-test-external-delete"
			openSearchManifest2.ResourceVersion = ""

			openSearch2 := v1beta1.OpenSearch{}
			openSearch2NamespacedName := types.NamespacedName{
				Namespace: openSearchManifest2.Namespace,
				Name:      openSearchManifest2.Name,
			}

			Expect(k8sClient.Create(ctx, openSearchManifest2)).Should(Succeed())

			doneCh := tests.NewChannelWithTimeout(timeout)

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearch2NamespacedName, &openSearch2); err != nil {
					return false
				}

				if openSearch2.Status.State != models.RunningStatus {
					return false
				}

				doneCh <- struct{}{}

				return true
			}, timeout, interval).Should(BeTrue())

			<-doneCh

			By("testing by deleting the cluster by Instaclutr API client")
			Expect(instaClient.DeleteCluster(openSearch2.Status.ID, instaclustr.OpenSearchEndpoint)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, openSearch2NamespacedName, &openSearch2)
				if err != nil {
					return false
				}

				return openSearch2.Status.State == models.DeletedStatus
			}, timeout, interval).Should(BeTrue())
		})
	})

})
