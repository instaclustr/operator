/*
Copyright 2023.

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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	mock "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Cadence Controller", func() {
	Context("Standard provisioning of cadence cluster", func() {
		cadenceManifest := &v1beta1.Cadence{}

		b, err := os.ReadFile(filepath.Join(".", "datatest", "cadence_v1beta1.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(yaml.Unmarshal(b, cadenceManifest)).Should(Succeed())

		It("creates the cadence custom resource in k8s", func() {
			Expect(k8sClient.Create(ctx, cadenceManifest)).Should(Succeed())

			key := client.ObjectKeyFromObject(cadenceManifest)
			cadence := &v1beta1.Cadence{}

			expectedClusterID := cadenceManifest.Spec.Name + "-" + mock.CreatedID

			Eventually(func() (string, error) {
				if err := k8sClient.Get(ctx, key, cadence); err != nil {
					return "", err
				}

				return cadence.Status.ID, nil
			}, timeout, interval).Should(Equal(expectedClusterID))
		})

		It("updates node size of the existing cadence custom resource", func() {
			key := client.ObjectKeyFromObject(cadenceManifest)
			cadence := &v1beta1.Cadence{}

			Eventually(func() error {
				return k8sClient.Get(ctx, key, cadence)
			}, timeout, interval).ShouldNot(HaveOccurred())

			const newNodeSize = "cadence-test-node-size-2"

			patch := cadence.NewPatch()
			cadence.Spec.DataCentres[0].NodeSize = newNodeSize
			Expect(k8sClient.Patch(ctx, cadence, patch)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, cadence)).Should(Succeed())

				g.Expect(cadence.Status.DataCentres).ShouldNot(BeEmpty())
				g.Expect(cadence.Status.DataCentres[0].Nodes).ShouldNot(BeEmpty())

				g.Expect(cadence.Status.DataCentres[0].Nodes[0].Size).Should(Equal(newNodeSize))
			}, timeout, interval).Should(Succeed())
		})

		It("deletes the cadence resource in k8s and on Instaclustr", func() {
			key := client.ObjectKeyFromObject(cadenceManifest)
			cadence := &v1beta1.Cadence{}

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, cadence)).Should(Succeed(), "should get the cadence resource")
				g.Expect(cadence.Status.ID).ShouldNot(BeEmpty(), "the cadence id should not be empty")
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, cadence)).Should(Succeed())

			Eventually(func(g Gomega) {
				c := &v1beta1.Cadence{}
				g.Expect(k8sClient.Get(ctx, key, c)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				_, err := instaClient.GetCadence(cadence.Status.ID)
				return err
			}, timeout, interval).Should(Equal(instaclustr.NotFound))
		})
	})

	Context("Packaged provisioning of cadence cluster", func() {
		cadenceManifest := &v1beta1.Cadence{}

		b, err := os.ReadFile(filepath.Join(".", "datatest", "cadence_v1beta1_packaged.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(yaml.Unmarshal(b, cadenceManifest)).Should(Succeed())

		It("creates a cadence and all packaged custom resources in k8s and on Instaclustr", func() {
			Expect(k8sClient.Create(ctx, cadenceManifest)).Should(Succeed())

			cadence := &v1beta1.Cadence{}
			opensearch := &v1beta1.OpenSearch{}
			kafka := &v1beta1.Kafka{}
			cassandra := &v1beta1.Cassandra{}

			key := client.ObjectKeyFromObject(cadenceManifest)

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, cadence)).Should(Succeed())
				g.Expect(cadence.Status.ID).ShouldNot(BeEmpty())
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: metav1.NamespaceDefault,
					Name:      models.OpenSearchChildPrefix + cadence.Name}, opensearch,
				)
			}, timeout, interval).Should(Succeed(), "should get opensearch resource")

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: metav1.NamespaceDefault,
					Name:      models.CassandraChildPrefix + cadence.Name}, cassandra,
				)
			}, timeout, interval).Should(Succeed(), "should get cassandra resource")

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: metav1.NamespaceDefault,
					Name:      models.KafkaChildPrefix + cadence.Name}, kafka,
				)
			}, timeout, interval).Should(Succeed(), "should get kafka resource")
		})

		It("deletes the cadence and bundled resources in k8s and Instaclustr", func() {
			key := client.ObjectKeyFromObject(cadenceManifest)
			cadence := &v1beta1.Cadence{}

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, cadence)).Should(Succeed(), "should get the cadence resource")
				g.Expect(cadence.Status.ID).ShouldNot(BeEmpty(), "the cadence id should not be empty")
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, cadence)).Should(Succeed())

			Eventually(func(g Gomega) {
				c := &v1beta1.Cadence{}
				g.Expect(k8sClient.Get(ctx, key, c)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())
		})
	})
})
