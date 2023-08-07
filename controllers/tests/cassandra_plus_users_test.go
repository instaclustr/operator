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

package tests

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/strings/slices"

	clusterresource "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Basic Cassandra User controller + Basic Cassandra cluster controllers flow", func() {
	var (
		ns = "default"

		user         clusterresource.CassandraUser
		userManifest clusterresource.CassandraUser

		secret         v1.Secret
		secretManifest v1.Secret

		cassandra         v1beta1.Cassandra
		cassandraManifest v1beta1.Cassandra

		timeout  = time.Second * 5
		interval = time.Second * 2
	)

	ctx := context.Background()

	cassandraUserYAML, err := os.ReadFile("datatest/clusterresources_v1beta1_cassandrauser.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(cassandraUserYAML, &userManifest)
	Expect(err).NotTo(HaveOccurred())

	cassandraUserNS := types.NamespacedName{Name: userManifest.ObjectMeta.Name, Namespace: ns}

	secretYAML, err := os.ReadFile("datatest/secret.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(secretYAML, &secretManifest)
	Expect(err).NotTo(HaveOccurred())

	secretNS := types.NamespacedName{Name: secretManifest.ObjectMeta.Name, Namespace: ns}

	cassandraYAML, err := os.ReadFile("datatest/clusters_v1beta1_cassandra.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(cassandraYAML, &cassandraManifest)
	Expect(err).NotTo(HaveOccurred())

	cassandraNS := types.NamespacedName{Name: cassandraManifest.ObjectMeta.Name, Namespace: ns}

	When("apply a secret and a cassandra user manifests", func() {
		It("should create both resources and they've got to have a link with themselves through finalizer", func() {
			Expect(k8sClient.Create(ctx, &secretManifest)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &userManifest)).Should(Succeed())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraUserNS, &user); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, secretNS, &secret); err != nil {
					return false
				}

				if user.Finalizers == nil {
					return false
				}

				uniqFinalizer := user.GetDeletionFinalizer()

				return slices.Contains(user.Finalizers, uniqFinalizer) || slices.Contains(secret.Finalizers, uniqFinalizer)
			}).Should(BeTrue())
		})
	})

	When("apply a cassandra manifest", func() {
		cassandraManifest.Annotations = map[string]string{models.ResourceStateAnnotation: models.CreatingEvent}
		It("should create a cassandra resource", func() {
			Expect(k8sClient.Create(ctx, &cassandraManifest)).Should(Succeed())

			By("sending Cassandra specification to the Instaclustr API and get ID of a created cluster")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNS, &cassandra); err != nil {
					return false
				}

				return cassandra.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("add the user to a Cassandra UserReference", func() {
		newUsers := []*v1beta1.UserReference{{
			Namespace: userManifest.Namespace,
			Name:      userManifest.Name,
		}}
		cassandraNS := types.NamespacedName{Name: cassandraManifest.ObjectMeta.Name, Namespace: ns}
		userNS := types.NamespacedName{Name: userManifest.ObjectMeta.Name, Namespace: ns}

		It("should create the user for the cluster", func() {

			Expect(k8sClient.Get(ctx, cassandraNS, &cassandra)).Should(Succeed())

			patch := cassandra.NewPatch()
			cassandra.Spec.UserRefs = newUsers

			Expect(k8sClient.Patch(ctx, &cassandra, patch)).Should(Succeed())

			By("going to Cassandra(cluster) controller predicate and put user entity to creation state. " +
				"Finally creates the user for the corresponded cluster")

			clusterID := cassandra.Status.ID
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNS, &cassandra); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, userNS, &user); err != nil {
					return false
				}

				if state, exist := user.Status.ClustersEvents[clusterID]; exist && state != models.Created {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("remove the user from the Cassandra UserReference", func() {
		It("should delete the user for the cluster", func() {
			Expect(k8sClient.Get(ctx, cassandraNS, &cassandra)).Should(Succeed())

			patch := cassandra.NewPatch()
			cassandra.Spec.UserRefs = []*v1beta1.UserReference{}

			Expect(k8sClient.Patch(ctx, &cassandra, patch)).Should(Succeed())

			By("going to Cassandra(cluster) controller predicate and put user entity to deletion state. " +
				"Finally deletes the user for the corresponded cluster")

			clusterID := cassandra.Status.ID
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, cassandraNS, &cassandra); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, cassandraUserNS, &user); err != nil {
					return false
				}

				if _, exist := user.Status.ClustersEvents[clusterID]; exist {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
