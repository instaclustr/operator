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

const newPostgreSQLNodeSize = "PGS-DEV-t4g.medium-30"

var _ = Describe("PostgreSQL Controller", func() {
	postgresql := v1beta1.PostgreSQL{}
	postgresqlManifest := v1beta1.PostgreSQL{}

	yfile, err := os.ReadFile("datatest/postgresql_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &postgresqlManifest)
	Expect(err).NotTo(HaveOccurred())

	postgresqlNamespacedName := types.NamespacedName{Name: postgresqlManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a PostgreSQL manifest", func() {
		It("should create a PostgreSQL resources", func() {
			Expect(k8sClient.Create(ctx, &postgresqlManifest)).Should(Succeed())
			done := tests.NewChannelWithTimeout(timeout)

			By("sending PostgreSQL specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, postgresqlNamespacedName, &postgresql); err != nil {
					return false
				}

				return postgresql.Status.ID == openapi.CreatedID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a node size", func() {
		It("should update a PostgreSQL resources", func() {
			Expect(k8sClient.Get(ctx, postgresqlNamespacedName, &postgresql)).Should(Succeed())

			patch := postgresql.NewPatch()
			postgresql.Spec.DataCentres[0].NodeSize = newPostgreSQLNodeSize
			Expect(k8sClient.Patch(ctx, &postgresql, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s PostgreSQL resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, postgresqlNamespacedName, &postgresql); err != nil {
					return false
				}

				if len(postgresql.Status.DataCentres) == 0 || len(postgresql.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return postgresql.Status.DataCentres[0].Nodes[0].Size == newPostgreSQLNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the PostgreSQL resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, postgresqlNamespacedName, &postgresql)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &postgresql)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, postgresqlNamespacedName, &postgresql)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("Deleting the PostgreSQL resource by avoiding operator", func() {
		It("should try to get the cluster details and receive StatusNotFound", func() {
			postgresqlManifest2 := postgresqlManifest.DeepCopy()
			postgresqlManifest2.Name += "-test-external-delete"
			postgresqlManifest2.ResourceVersion = ""

			pg2 := v1beta1.PostgreSQL{}
			pg2NamespacedName := types.NamespacedName{
				Namespace: postgresqlManifest2.Namespace,
				Name:      postgresqlManifest2.Name,
			}

			Expect(k8sClient.Create(ctx, postgresqlManifest2)).Should(Succeed())

			doneCh := tests.NewChannelWithTimeout(timeout)

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, pg2NamespacedName, &pg2); err != nil {
					return false
				}

				if pg2.Status.State != models.RunningStatus {
					return false
				}

				doneCh <- struct{}{}

				return true
			}, timeout, interval).Should(BeTrue())

			<-doneCh

			By("testing by deleting the cluster by Instaclutr API client")
			Expect(instaClient.DeleteCluster(pg2.Status.ID, instaclustr.PGSQLEndpoint)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, pg2NamespacedName, &pg2)
				if err != nil {
					return false
				}

				return pg2.Status.State == models.DeletedStatus
			}, timeout, interval).Should(BeTrue())
		})
	})

})
