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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var newRecourseType = "GROUP"

var _ = Describe("Kafka ACL Controller", func() {
	acl := v1beta1.KafkaACL{}
	aclManifest := v1beta1.KafkaACL{}

	yfile, err := os.ReadFile("datatest/kafkaacl_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &aclManifest)
	Expect(err).NotTo(HaveOccurred())

	aclNamespacedName := types.NamespacedName{Name: aclManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a Kafka ACL manifest", func() {
		It("should create a ACL resources", func() {
			Expect(k8sClient.Create(ctx, &aclManifest)).Should(Succeed())
			By("sending ACL specification to the Instaclustr API and get ID of created resource.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, aclNamespacedName, &acl); err != nil {
					return false
				}

				return acl.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a ACL ", func() {
		It("should update the ACL resources", func() {
			Expect(k8sClient.Get(ctx, aclNamespacedName, &acl)).Should(Succeed())

			patch := acl.NewPatch()
			acl.Spec.ACLs[0].ResourceType = newRecourseType
			Expect(k8sClient.Patch(ctx, &acl, patch)).Should(Succeed())

			By("sending a new ACL configs request to the Instaclustr API, it" +
				"gets a new data from the InstAPI and update it in k8s ACL resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, aclNamespacedName, &acl); err != nil {
					return false
				}

				return acl.GetAnnotations()[models.ResourceStateAnnotation] == models.UpdatedEvent
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the ACL resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, aclNamespacedName, &acl)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &acl)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, aclNamespacedName, &acl)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
