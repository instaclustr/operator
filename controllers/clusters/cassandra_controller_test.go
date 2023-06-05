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

const newCassandraNodeSize = "CAS-DEV-t4g.small-30"

var _ = Describe("Cassandra Controller", func() {
	var (
		cassandraResource v1alpha1.Cassandra
		cassandraYAML     v1alpha1.Cassandra
		c                 = "cassandra"
		ns                = "default"
		cassandraNS       = types.NamespacedName{Name: c, Namespace: ns}
		timeout           = time.Second * 30
		interval          = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/cassandra_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &cassandraYAML)
	Expect(err).NotTo(HaveOccurred())

	cassandraObjMeta := metav1.ObjectMeta{
		Name:      c,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	cassandraYAML.ObjectMeta = cassandraObjMeta

	ctx := context.Background()

	When("apply a cassandra manifest", func() {
		It("should create a cassandra resources", func() {
			Expect(k8sClient.Create(ctx, &cassandraYAML)).Should(Succeed())
			By("sending cassandra specification to the Instaclustr API and get ID of created cluster.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNS, &cassandraResource); err != nil {
					return false
				}

				return cassandraResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {
		It("should update a cassandra resources", func() {
			Expect(k8sClient.Get(ctx, cassandraNS, &cassandraResource)).Should(Succeed())
			patch := cassandraResource.NewPatch()

			cassandraResource.Spec.DataCentres[0].NodeSize = newCassandraNodeSize

			cassandraResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}
			Expect(k8sClient.Patch(ctx, &cassandraResource, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s cassandra resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNS, &cassandraResource); err != nil {
					return false
				}

				if len(cassandraResource.Status.DataCentres) == 0 || len(cassandraResource.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return cassandraResource.Status.DataCentres[0].Nodes[0].Size == newCassandraNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the cassandra resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, cassandraNS, &cassandraResource)).Should(Succeed())

			cassandraResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

			Expect(k8sClient.Delete(ctx, &cassandraResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, cassandraNS, &cassandraResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
