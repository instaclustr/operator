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

var _ = Describe("Successful creation of a AWS Security Group Firewall Rule resource", func() {
	Context("When setting up a AWS Security Group Firewall Rule CRD", func() {
		awsSGFirewallRuleSpec := v1alpha1.AWSSecurityGroupFirewallRuleSpec{
			FirewallRuleSpec: v1alpha1.FirewallRuleSpec{
				ClusterID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				Type:      "SECURITY",
			},
			SecurityGroupID: "sg-1434412",
		}

		ctx := context.Background()
		resource := v1alpha1.AWSSecurityGroupFirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "awssgfwrule",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: awsSGFirewallRuleSpec,
		}

		It("Should create a AWS Security Group Firewall Rule resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending AWS Security Group Firewall Rule Specification to Instaclustr API v2")
			var awsSGFirewallRule v1alpha1.AWSSecurityGroupFirewallRule
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "awssgfwrule", Namespace: "default"}, &awsSGFirewallRule); err != nil {
					return false
				}

				return awsSGFirewallRule.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
