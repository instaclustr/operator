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

var _ = Describe("Successful creation of a AWSVPCPeering resource", func() {
	Context("When setting up a AWSVPCPeering CRD", func() {
		awsVPCPeeringSpec := v1alpha1.AWSVPCPeeringSpec{
			VPCPeeringSpec: v1alpha1.VPCPeeringSpec{
				DataCentreID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				PeerSubnets:  []string{"172.31.0.0/16", "192.168.0.0/16"},
			},
			PeerAWSAccountID: "152668027680",
			PeerVPCID:        "vpc-87241ae1",
			PeerRegion:       "US_EAST_1",
		}

		ctx := context.Background()
		resource := v1alpha1.AWSVPCPeering{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "awsvpcpeering",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: awsVPCPeeringSpec,
		}

		It("Should create a AWSVPCPeering resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending AWSVPCPeering Specification to Instaclustr API v2")
			var awsVPCPeering v1alpha1.AWSVPCPeering
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "awsvpcpeering", Namespace: "default"}, &awsVPCPeering); err != nil {
					return false
				}

				return awsVPCPeering.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
