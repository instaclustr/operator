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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Successful creation of a Azure VNet Peering resource", func() {
	Context("When setting up a Azure VNet Peering CRD", func() {
		azureVNetPeeringSpec := v1beta1.AzureVNetPeeringSpec{
			VPCPeeringSpec: v1beta1.VPCPeeringSpec{
				PeerSubnets: []string{"172.31.0.0/16", "192.168.0.0/16"},
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
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: azureVNetPeeringSpec,
		}

		It("Should create a Azure VNet Peering resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			patch := resource.NewPatch()
			resource.Status.CDCID = "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2"
			resource.Status.ResourceState = models.CreatingEvent
			Expect(k8sClient.Status().Patch(ctx, &resource, patch)).Should(Succeed())

			By("Sending Azure VNet Peering Specification to Instaclustr API v2")
			var azureVNetPeering v1beta1.AzureVNetPeering
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "azurevnetpeering", Namespace: "default"}, &azureVNetPeering); err != nil {
					return false
				}

				return azureVNetPeering.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
