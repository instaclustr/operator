package clusterresources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Successful creation of a Azure VNet Peering resource", func() {
	Context("When setting up a Azure VNet Peering CRD", func() {
		azureVNetPeeringSpec := v1alpha1.AzureVNetPeeringSpec{
			VPCPeeringSpec: v1alpha1.VPCPeeringSpec{
				DataCentreID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				PeerSubnets:  []string{"172.31.0.0/16", "192.168.0.0/16"},
			},
			PeerResourceGroup:      "rg-1231212",
			PeerSubscriptionID:     "sg-123321",
			PeerADObjectID:         "ad-132313",
			PeerVirtualNetworkName: "vnet-123123123",
		}

		ctx := context.Background()
		resource := v1alpha1.AzureVNetPeering{
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

			By("Sending Azure VNet Peering Specification to Instaclustr API v2")
			var azureVNetPeering v1alpha1.AzureVNetPeering
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "azurevnetpeering", Namespace: "default"}, &azureVNetPeering); err != nil {
					return false
				}

				return azureVNetPeering.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
