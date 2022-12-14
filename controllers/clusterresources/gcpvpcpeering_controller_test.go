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

var _ = Describe("Successful creation of a GCP VPC Peering resource", func() {
	Context("When setting up a GCP VPC Peering CRD", func() {
		gcpVPCPeeringSpec := v1alpha1.GCPVPCPeeringSpec{
			VPCPeeringSpec: v1alpha1.VPCPeeringSpec{
				DataCentreID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				PeerSubnets:  []string{"172.31.0.0/16", "192.168.0.0/16"},
			},
			PeerProjectID:      "pid-132313",
			PeerVPCNetworkName: "vpc-123123123",
		}

		ctx := context.Background()
		resource := v1alpha1.GCPVPCPeering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gcpvpcpeering",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: gcpVPCPeeringSpec,
		}

		It("Should create a GCP VPC Peering resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending GCP VPC Peering Specification to Instaclustr API v2")
			var gcpVNetPeering v1alpha1.GCPVPCPeering
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "gcpvpcpeering", Namespace: "default"}, &gcpVNetPeering); err != nil {
					return false
				}

				return gcpVNetPeering.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
