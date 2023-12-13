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

package clusterresources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("AWSEncryptionKey controller", Ordered, func() {
	awsEncryptionKey := &v1beta1.AWSEncryptionKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-encryption-key-test",
			Namespace: metav1.NamespaceDefault,
			Annotations: map[string]string{
				"instaclustr.com/test": "true",
			},
		},
		Spec: v1beta1.AWSEncryptionKeySpec{
			Alias: "test",
			ARN:   "arn:aws:kms:us-east-1:152668027680:key/460b469b-9f0a-4801-854a-e23626957d00",
		},
	}

	key := client.ObjectKeyFromObject(awsEncryptionKey)

	When("Apply AWSEncryptionKey manifest", func() {
		It("Should send API call to create the resource", func() {
			Expect(k8sClient.Create(ctx, awsEncryptionKey)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, awsEncryptionKey)).Should(Succeed())
				g.Expect(awsEncryptionKey.Finalizers).Should(ContainElement(models.DeletionFinalizer))
				g.Expect(awsEncryptionKey.Status.ID).ShouldNot(BeEmpty())
			}, timeout, interval).Should(Succeed())
		})
	})

	When("Delete AWSEncryptionKey k8s resource", func() {
		It("Should send API call to delete the resource and delete resource in k8s", FlakeAttempts(3), func() {
			id := awsEncryptionKey.Status.ID

			By("sending delete request to k8s")
			Expect(k8sClient.Delete(ctx, awsEncryptionKey)).Should(Succeed())

			By("checking if resource exists on the Instaclustr API")
			Eventually(func() error {
				_, err := instaClient.GetEncryptionKeyStatus(id, instaclustr.AWSEncryptionKeyEndpoint)
				return err
			}, timeout, interval).Should(Equal(instaclustr.NotFound))

			By("checking if the resource exists in k8s")
			Eventually(func() bool {
				var resource v1beta1.AWSEncryptionKey
				return errors.IsNotFound(k8sClient.Get(ctx, key, &resource))
			}, timeout, interval).Should(BeTrue())
		})
	})
})
