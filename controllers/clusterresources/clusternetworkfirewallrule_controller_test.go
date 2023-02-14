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

var _ = Describe("Successful creation of a ClusterNetworkFirewallRule resource", func() {
	Context("When setting up a ClusterNetworkFirewallRule CRD", func() {
		clusterNetworkFirewallRuleSpec := v1alpha1.ClusterNetworkFirewallRuleSpec{
			FirewallRuleSpec: v1alpha1.FirewallRuleSpec{
				ClusterID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				Type:      "SECURITY",
			},
			Network: "191.54.123.1/24",
		}

		ctx := context.Background()
		resource := v1alpha1.ClusterNetworkFirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "clusternetworkfwrule",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: clusterNetworkFirewallRuleSpec,
		}

		It("Should create a ClusterNetworkFirewallRule resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending ClusterNetworkFirewallRule Specification to Instaclustr API v2")
			var clusterNetworkFirewallRule v1alpha1.ClusterNetworkFirewallRule
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "clusternetworkfwrule", Namespace: "default"}, &clusterNetworkFirewallRule); err != nil {
					return false
				}

				return clusterNetworkFirewallRule.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
