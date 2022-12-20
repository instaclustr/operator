package clusterresources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
)

var _ = Describe("Successful creation of a NodeReload resource", func() {
	Context("When setting up a NodeReload CRD", func() {
		nodeReloadSpec := v1alpha1.NodeReloadSpec{
			Nodes: []*v1alpha1.Node{{
				NodeID: "node-13124134",
				Bundle: "CASSANDRA",
			}},
		}

		ctx := context.Background()
		resource := v1alpha1.NodeReload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nodereload",
				Namespace: "default",
			},
			Spec: nodeReloadSpec,
		}

		It("Should create a NodeReload resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending NodeReload Specification to Instaclustr API v2")
			var nodeReload v1alpha1.NodeReload
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "nodereload", Namespace: "default"}, &nodeReload); err != nil {
					return false
				}

				return &nodeReload.Status != nil
			}).Should(BeTrue())
		})
	})
})
