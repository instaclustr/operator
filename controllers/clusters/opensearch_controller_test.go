package clusters

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

const newOpenSearchNodeSize = "SRH-DEV-t4g.small-30"

var _ = Describe("OpenSearch Controller", func() {
	var (
		openSearchResource v1alpha1.OpenSearch
		openSearchYAML     v1alpha1.OpenSearch
		o                  = "opensearch"
		ns                 = "default"
		openSearchNS       = types.NamespacedName{Name: o, Namespace: ns}
		timeout            = time.Second * 30
		interval           = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/opensearch_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &openSearchYAML)
	Expect(err).NotTo(HaveOccurred())

	openSearchObjMeta := metav1.ObjectMeta{
		Name:      o,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	openSearchYAML.ObjectMeta = openSearchObjMeta

	ctx := context.Background()

	When("apply a OpenSearch manifest", func() {
		It("should create a OpenSearch resources", func() {
			Expect(k8sClient.Create(ctx, &openSearchYAML)).Should(Succeed())
			By("sending OpenSearch specification to the Instaclustr API and get ID of created cluster.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNS, &openSearchResource); err != nil {
					return false
				}

				return openSearchResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {
		It("should update a OpenSearch resources", func() {
			Expect(k8sClient.Get(ctx, openSearchNS, &openSearchResource)).Should(Succeed())
			patch := openSearchResource.NewPatch()

			openSearchResource.Spec.ClusterManagerNodes[0].NodeSize = newOpenSearchNodeSize

			openSearchResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}
			Expect(k8sClient.Patch(ctx, &openSearchResource, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s OpenSearch resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNS, &openSearchResource); err != nil {
					return false
				}

				return openSearchResource.Spec.ClusterManagerNodes[0].NodeSize == newOpenSearchNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the OpenSearch resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, openSearchNS, &openSearchResource)).Should(Succeed())

			openSearchResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

			Expect(k8sClient.Delete(ctx, &openSearchResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, openSearchNS, &openSearchResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
