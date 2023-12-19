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

package clusterresources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("AzureVNetPeering controller", Ordered, func() {
	azureVNetPeeringSpec := v1beta1.AzureVNetPeeringSpec{
		PeeringSpec: v1beta1.PeeringSpec{
			DataCentreID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
			PeerSubnets:  []string{"172.31.0.0/16", "192.168.0.0/16"},
		},
		PeerResourceGroup:      "rg-1231212",
		PeerSubscriptionID:     "sg-123321",
		PeerADObjectID:         "ad-132313",
		PeerVirtualNetworkName: "vnet-123123123",
	}

	ctx := context.Background()
	resource := v1beta1.AzureVNetPeering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "azurevnetpeering",
			Namespace: "default",
			Annotations: map[string]string{
				"instaclustr.com/test": "true",
			},
		},
		Spec: azureVNetPeeringSpec,
	}

	key := client.ObjectKeyFromObject(&resource)

	Context("When setting up a Azure VNet Peering CRD", func() {
		It("Should create a Azure VNet Peering resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending Azure VNet Peering Specification to Instaclustr API v2")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, &resource); err != nil {
					return false
				}

				if !controllerutil.ContainsFinalizer(&resource, models.DeletionFinalizer) {
					return false
				}

				return resource.Status.ID != ""
			}).Should(BeTrue())
		})
	})

	Context("When deleting AzureVNetPeering k8s resource", func() {
		It("Should delete the resource in k8s and API", FlakeAttempts(3), func() {
			id := resource.Status.ID

			By("sending delete request to k8s")
			Expect(k8sClient.Delete(ctx, &resource)).Should(Succeed())

			By("checking if resource exists on the Instaclustr API")
			Eventually(func() error {
				_, err := instaClient.GetPeeringStatus(id, instaclustr.AzurePeeringEndpoint)
				return err
			}, timeout, interval).Should(Equal(instaclustr.NotFound))

			By("checking if the resource exists in k8s")
			Eventually(func() bool {
				var resource v1beta1.AzureVNetPeering
				return errors.IsNotFound(k8sClient.Get(ctx, key, &resource))
			}, timeout, interval).Should(BeTrue())
		})
	})
})
