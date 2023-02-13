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

const newKafkaNodeSize = "KFK-DEV-t4g.medium-80"

var _ = Describe("Kafka Controller", func() {
	var (
		kafkaResource v1alpha1.Kafka
		kafkaYAML     v1alpha1.Kafka
		k             = "kafka"
		ns            = "default"
		kafkaNS       = types.NamespacedName{Name: k, Namespace: ns}
		timeout       = time.Second * 15
		interval      = time.Second * 2
	)

	yfile, err := os.ReadFile("datatest/kafka_v1alpha1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &kafkaYAML)
	Expect(err).NotTo(HaveOccurred())

	kafkaObjMeta := metav1.ObjectMeta{
		Name:      k,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	kafkaYAML.ObjectMeta = kafkaObjMeta

	ctx := context.Background()

	When("apply a Kafka manifest", func() {
		It("should create a Kafka resources", func() {
			Expect(k8sClient.Create(ctx, &kafkaYAML)).Should(Succeed())
			By("sending Kafka specification to the Instaclustr API and get ID of created cluster.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNS, &kafkaResource); err != nil {
					return false
				}

				return kafkaResource.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {
		It("should update a Kafka resources", func() {
			Expect(k8sClient.Get(ctx, kafkaNS, &kafkaResource)).Should(Succeed())
			patch := kafkaResource.NewPatch()

			kafkaResource.Spec.DataCentres[0].NodeSize = newKafkaNodeSize

			kafkaResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}
			Expect(k8sClient.Patch(ctx, &kafkaResource, patch)).Should(Succeed())

			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s Kafka resource")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNS, &kafkaResource); err != nil {
					return false
				}

				if len(kafkaResource.Status.DataCentres) == 0 || len(kafkaResource.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return kafkaResource.Status.DataCentres[0].Nodes[0].Size == newKafkaNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, kafkaNS, &kafkaResource)).Should(Succeed())

			kafkaResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}

			Expect(k8sClient.Delete(ctx, &kafkaResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, kafkaNS, &kafkaResource)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
