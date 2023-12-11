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

var _ = Describe("Successful creation of a AWS Security Group Firewall Rule resource", func() {
	Context("When setting up a AWS Security Group Firewall Rule CRD", func() {
		awsSGFirewallRuleSpec := v1beta1.AWSSecurityGroupFirewallRuleSpec{
			FirewallRuleSpec: v1beta1.FirewallRuleSpec{
				Type: "SECURITY",
			},
			SecurityGroupID: "sg-1434412",
		}

		ctx := context.Background()
		resource := v1beta1.AWSSecurityGroupFirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "awssgfwrule",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: awsSGFirewallRuleSpec,
			Status: v1beta1.AWSSecurityGroupFirewallRuleStatus{
				FirewallRuleStatus: v1beta1.FirewallRuleStatus{
					ClusterID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
				},
			},
		}

		It("Should create a AWS Security Group Firewall Rule resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			patch := resource.NewPatch()
			resource.Status.ClusterID = "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2"
			resource.Status.ResourceState = models.CreatingEvent
			Expect(k8sClient.Status().Patch(ctx, &resource, patch)).Should(Succeed())

			By("Sending AWS Security Group Firewall Rule Specification to Instaclustr API v2")
			var awsSGFirewallRule v1beta1.AWSSecurityGroupFirewallRule
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "awssgfwrule", Namespace: "default"}, &awsSGFirewallRule); err != nil {
					return false
				}

				return awsSGFirewallRule.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
