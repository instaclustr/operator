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

var _ = Describe("Successful creation of a GCP VPC Peering resource", func() {
	Context("When setting up a GCP VPC Peering CRD", func() {
		gcpVPCPeeringSpec := v1beta1.GCPVPCPeeringSpec{
			VPCPeeringSpec: v1beta1.VPCPeeringSpec{
				DataCentreID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				PeerSubnets:  []string{"172.31.0.0/16", "192.168.0.0/16"},
			},
			PeerProjectID:      "pid-132313",
			PeerVPCNetworkName: "vpc-123123123",
		}

		ctx := context.Background()
		resource := v1beta1.GCPVPCPeering{
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

			patch := resource.NewPatch()
			resource.Status.CDCID = "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2"
			resource.Status.ResourceState = models.CreatingEvent
			Expect(k8sClient.Status().Patch(ctx, &resource, patch)).Should(Succeed())

			By("Sending GCP VPC Peering Specification to Instaclustr API v2")
			var gcpVNetPeering v1beta1.GCPVPCPeering
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "gcpvpcpeering", Namespace: "default"}, &gcpVNetPeering); err != nil {
					return false
				}

				return gcpVNetPeering.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
