package kafkamanagement

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

	"github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var newTopicConfig = map[string]string{
	"min.insync.replicas": "2",
	"retention.ms":        "603800000",
}

var _ = Describe("Kafka Topic Controller", func() {
	var (
		topicResource v1alpha1.Topic
		topicYAML     v1alpha1.Topic
		t             = "topic"
		ns            = "default"
		topicNS       = types.NamespacedName{Name: t, Namespace: ns}
		timeout       = time.Second * 20
		interval      = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/topic_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &topicYAML)
	Expect(err).NotTo(HaveOccurred())

	topicObjMeta := metav1.ObjectMeta{
		Name:      t,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	topicYAML.ObjectMeta = topicObjMeta

	ctx := context.Background()

	When("apply a Kafka topic manifest", func() {
		It("should create a topic resources", func() {
			Expect(k8sClient.Create(ctx, &topicYAML)).Should(Succeed())
			By("sending topic specification to the Instaclustr API and get ID of created resource.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, topicNS, &topicResource); err != nil {
					return false
				}

				return topicResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a topic configs", func() {
		It("should update the topic resources", func() {
			Expect(k8sClient.Get(ctx, topicNS, &topicResource)).Should(Succeed())
			patch := topicResource.NewPatch()

			topicResource.Spec.TopicConfigs = newTopicConfig
			topicResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}

			Expect(k8sClient.Patch(ctx, &topicResource, patch)).Should(Succeed())

			By("sending a new topic configs request to the Instaclustr API, it" +
				"gets a new data from the InstAPI and update it in k8s Topic resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, topicNS, &topicResource); err != nil {
					return false
				}

				for newK, newV := range newTopicConfig {
					if oldV, ok := topicResource.Status.TopicConfigs[newK]; !ok || oldV != newV {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, topicNS, &topicResource)).Should(Succeed())

			topicResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

			Expect(k8sClient.Delete(ctx, &topicResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, topicNS, &topicResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}).Should(BeTrue())
		})
	})
})
