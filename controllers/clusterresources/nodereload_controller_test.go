package clusterresources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
)

var _ = Describe("Successful creation of NodeReload resource", func() {
	var (
		ctx = context.Background()
	)

	When("apply NodeReload manifest", func() {
		It("should create the NodeReload resource and successfully reload all the PostgreSQL nodes", func() {
			manifest := &v1beta1.NodeReload{
				TypeMeta: metav1.TypeMeta{
					Kind: "NodeReload",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodereload-resource-test",
					Namespace: metav1.NamespaceDefault,
				},
				Spec: v1beta1.NodeReloadSpec{
					Nodes: []*v1beta1.Node{
						{ID: "mock-node-id-1"},
						{ID: "mock-node-id-2"},
					},
				},
			}

			Expect(k8sClient.Create(ctx, manifest)).Should(Succeed())

			key := client.ObjectKeyFromObject(manifest)
			nrs := &v1beta1.NodeReload{}

			By("sending API calls to Instaclustr to trigger PostgreSQL node reload")

			Eventually(func() ([]*v1beta1.Node, error) {
				if err := k8sClient.Get(ctx, key, nrs); err != nil {
					return nil, err
				}

				return nrs.Status.CompletedNodes, nil
			}, timeout, interval).Should(Equal(nrs.Spec.Nodes))
		})
	})

	When("apply NodeReload manifest with unknown nodeID", func() {
		It("should create the NodeReload resource and successfully reload only the first node", func() {
			manifest := &v1beta1.NodeReload{
				TypeMeta: metav1.TypeMeta{
					Kind: "NodeReload",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodereload-resource-test-wrong-1",
					Namespace: metav1.NamespaceDefault,
				},
				Spec: v1beta1.NodeReloadSpec{
					Nodes: []*v1beta1.Node{
						{ID: "mock-node-id-1"},
						{ID: "mock-node-id-2-wrong"},
					},
				},
			}

			Expect(k8sClient.Create(ctx, manifest)).Should(Succeed())

			key := client.ObjectKeyFromObject(manifest)
			nrs := &v1beta1.NodeReload{}

			type result struct {
				CompletedNodes []*v1beta1.Node
				FailedNodes    []*v1beta1.Node
			}

			Eventually(func() (*result, error) {
				if err := k8sClient.Get(ctx, key, nrs); err != nil {
					return nil, err
				}

				return &result{
					CompletedNodes: nrs.Status.CompletedNodes,
					FailedNodes:    nrs.Status.FailedNodes,
				}, nil
			}, timeout, interval).Should(Equal(&result{
				CompletedNodes: manifest.Spec.Nodes[:1],
				FailedNodes:    manifest.Spec.Nodes[1:],
			}))
		})
	})

	When("apply NodeReload manifest with only unknown nodeIDs", func() {
		It("should create the NodeReload resource and fail to reload all the nodes", func() {
			manifest := &v1beta1.NodeReload{
				TypeMeta: metav1.TypeMeta{
					Kind: "NodeReload",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodereload-resource-test-wrong-2",
					Namespace: metav1.NamespaceDefault,
				},
				Spec: v1beta1.NodeReloadSpec{
					Nodes: []*v1beta1.Node{
						{ID: "mock-node-id-1-wrong"},
						{ID: "mock-node-id-2-wrong"},
					},
				},
			}

			Expect(k8sClient.Create(ctx, manifest)).Should(Succeed())

			key := client.ObjectKeyFromObject(manifest)
			nrs := &v1beta1.NodeReload{}

			Eventually(func() ([]*v1beta1.Node, error) {
				if err := k8sClient.Get(ctx, key, nrs); err != nil {
					return nil, err
				}

				return nrs.Status.FailedNodes, nil
			}, timeout, interval).Should(Equal(nrs.Spec.Nodes))
		})
	})

})
