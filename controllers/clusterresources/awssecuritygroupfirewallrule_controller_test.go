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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("AWSSecurityGroupFirewallRule controller", Ordered, func() {
	awsSGFirewallRuleSpec := v1beta1.AWSSecurityGroupFirewallRuleSpec{
		FirewallRuleSpec: v1beta1.FirewallRuleSpec{
			ClusterID: "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
			Type:      "SECURITY",
		},
		SecurityGroupID: "sg-1434412",
	}
	resource := v1beta1.AWSSecurityGroupFirewallRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "awssgfwrule",
			Namespace: "default",
			Annotations: map[string]string{
				"instaclustr.com/test": "true",
			},
		},
		Spec: awsSGFirewallRuleSpec,
	}

	key := types.NamespacedName{Name: resource.Name, Namespace: resource.Namespace}

	Context("When setting up a AWS Security Group Firewall Rule CRD", func() {
		It("Should create a AWS Security Group Firewall Rule resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending AWS Security Group Firewall Rule Specification to Instaclustr API v2")
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

	Context("When deleting AWSSecurityGroupFirewallRule k8s resource", func() {
		It("Should delete the resource in k8s and API", FlakeAttempts(3), func() {
			id := resource.Status.ID

			By("sending delete request to k8s")
			Expect(k8sClient.Delete(ctx, &resource)).Should(Succeed())

			By("checking if resource exists on the Instaclustr API")
			Eventually(func() error {
				_, err := instaClient.GetFirewallRuleStatus(id, instaclustr.AWSSecurityGroupFirewallRuleEndpoint)
				return err
			}, timeout, interval).Should(Equal(instaclustr.NotFound))

			By("checking if the resource exists in k8s")
			Eventually(func() bool {
				var resource v1beta1.AWSSecurityGroupFirewallRule
				return errors.IsNotFound(k8sClient.Get(ctx, key, &resource))
			}, timeout, interval).Should(BeTrue())
		})
	})
})
