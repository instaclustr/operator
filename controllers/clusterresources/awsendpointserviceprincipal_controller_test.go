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

var _ = Describe("AWSEndpointServicePrincipal controller", Ordered, func() {
	awsEndpointServicePrincipal := &v1beta1.AWSEndpointServicePrincipal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-endpoint-service-principal-test",
			Namespace: metav1.NamespaceDefault,
			Annotations: map[string]string{
				"instaclustr.com/test": "true",
			},
		},
		Spec: v1beta1.AWSEndpointServicePrincipalSpec{
			ClusterDataCenterID: "test-cluster-dc-id",
			PrincipalARN:        "arn:aws:iam::152668027680:role/aws-principal-test-1",
		},
	}

	key := client.ObjectKeyFromObject(awsEndpointServicePrincipal)

	When("Apply awsEndpointServicePrincipal manifest", func() {
		It("Should send API call to create the resource", func() {
			Expect(k8sClient.Create(ctx, awsEndpointServicePrincipal)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, awsEndpointServicePrincipal)).Should(Succeed())
				g.Expect(awsEndpointServicePrincipal.Finalizers).Should(ContainElement(models.DeletionFinalizer))
				g.Expect(awsEndpointServicePrincipal.Status.ID).ShouldNot(BeEmpty())
			}, timeout, interval).Should(Succeed())
		})
	})

	When("Delete AWSEndpointServicePrincipal k8s resource", func() {
		It("Should send API call to delete the resource and delete resource in k8s", FlakeAttempts(3), func() {
			id := awsEndpointServicePrincipal.Status.ID

			By("sending delete request to k8s")
			Expect(k8sClient.Delete(ctx, awsEndpointServicePrincipal)).Should(Succeed())

			By("checking if resource exists on the Instaclustr API")
			Eventually(func() error {
				_, err := instaClient.GetAWSEndpointServicePrincipal(id)
				return err
			}, timeout, interval).Should(Equal(instaclustr.NotFound))

			By("checking if the resource exists in k8s")
			Eventually(func() bool {
				var resource v1beta1.AWSEndpointServicePrincipal
				return errors.IsNotFound(k8sClient.Get(ctx, key, &resource))
			}, timeout, interval).Should(BeTrue())
		})
	})
})
