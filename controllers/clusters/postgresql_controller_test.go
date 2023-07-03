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

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

const newPostgreSQLNodeSize = "PGS-DEV-t4g.medium-30"

var _ = Describe("PostgreSQL Controller", func() {
	var (
		postgresqlResource v1beta1.PostgreSQL
		postgresqlYAML     v1beta1.PostgreSQL
		p                  = "postgresql"
		ns                 = "default"
		postgresqlNS       = types.NamespacedName{Name: p, Namespace: ns}
		timeout            = time.Second * 60
		interval           = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/postgresql_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &postgresqlYAML)
	Expect(err).NotTo(HaveOccurred())

	postgresqlObjMeta := metav1.ObjectMeta{
		Name:      p,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	postgresqlYAML.ObjectMeta = postgresqlObjMeta

	ctx := context.Background()

	When("apply a PostgreSQL manifest", func() {
		It("should create a PostgreSQL resources", func() {
			Expect(k8sClient.Create(ctx, &postgresqlYAML)).Should(Succeed())
			By("sending PostgreSQL specification to the Instaclustr API and get ID of created cluster.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, postgresqlNS, &postgresqlResource); err != nil {
					return false
				}

				return postgresqlResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {
		It("should update a PostgreSQL resources", func() {
			Expect(k8sClient.Get(ctx, postgresqlNS, &postgresqlResource)).Should(Succeed())
			patch := postgresqlResource.NewPatch()

			postgresqlResource.Spec.DataCentres[0].NodeSize = newPostgreSQLNodeSize

			postgresqlResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}
			Expect(k8sClient.Patch(ctx, &postgresqlResource, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s PostgreSQL resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, postgresqlNS, &postgresqlResource); err != nil {
					return false
				}

				if len(postgresqlResource.Status.DataCentres) == 0 || len(postgresqlResource.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return postgresqlResource.Status.DataCentres[0].Nodes[0].Size == newPostgreSQLNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the PostgreSQL resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, postgresqlNS, &postgresqlResource)).Should(Succeed())

			postgresqlResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

			Expect(k8sClient.Delete(ctx, &postgresqlResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, postgresqlNS, &postgresqlResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
