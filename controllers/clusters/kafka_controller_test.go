package clusters

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Kafka Controller", func() {
	var (
		kafkaResource v1alpha1.Kafka
		k             = "kafka"
		ns            = "default"
		kafkaNS       = types.NamespacedName{Name: k, Namespace: ns}
		timeout       = time.Second * 30
		interval      = time.Second * 2
	)

	kafkaSpec := v1alpha1.KafkaSpec{
		Cluster: v1alpha1.Cluster{
			Name:                  "kafkaTest",
			Version:               "2.7.1",
			PCICompliance:         false,
			PrivateNetworkCluster: false,
			SLATier:               "NON_PRODUCTION",
			TwoFactorDelete:       nil,
		},
		SchemaRegistry:            nil,
		ReplicationFactorNumber:   3,
		PartitionsNumber:          3,
		RestProxy:                 nil,
		AllowDeleteTopics:         false,
		AutoCreateTopics:          false,
		ClientToClusterEncryption: false,
		DataCentres: []*v1alpha1.KafkaDataCentre{{
			NodesNumber:   3,
			Network:       "192.168.0.0/18",
			NodeSize:      "KFK-DEV-t4g.small-5",
			CloudProvider: "AWS_VPC",
			Name:          "US_WEST_2_DC",
			Region:        "US_WEST_2",
		}},
		DedicatedZookeeper:                nil,
		ClientBrokerAuthWithMTLS:          false,
		ClientAuthBrokerWithoutEncryption: false,
		ClientAuthBrokerWithEncryption:    false,
		KarapaceRestProxy:                 nil,
		KarapaceSchemaRegistry:            nil,
	}

	kafkaObjMeta := metav1.ObjectMeta{
		Name:      k,
		Namespace: ns,
		Annotations: map[string]string{
			models.ResourceStateAnnotation: models.CreatingEvent,
		},
	}

	ctx := context.Background()

	kafkaManifest := v1alpha1.Kafka{
		ObjectMeta: kafkaObjMeta,
		Spec:       kafkaSpec,
	}

	When("apply a Kafka manifest", func() {
		It("should create a Kafka resources", func() {
			Expect(k8sClient.Create(ctx, &kafkaManifest)).Should(Succeed())
			By("sending Kafka specification to Inst API v2 and get cluster ID.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNS, &kafkaResource); err != nil {
					return false
				}

				return kafkaResource.Status.ID == mock.CreatedID
			}).Should(BeTrue())
		})
	})

	When("changing a node size", func() {

		newDC := []*v1alpha1.KafkaDataCentre{{
			NodesNumber:   3,
			Network:       "192.168.0.0/18",
			NodeSize:      mock.NewKafkaNodeSize,
			CloudProvider: "AWS_VPC",
			Name:          "US_WEST_2_DC",
			Region:        "US_WEST_2",
		}}

		It("should update a Kafka resources", func() {
			Expect(k8sClient.Get(ctx, kafkaNS, &kafkaResource)).Should(Succeed())

			patch := kafkaResource.NewPatch()
			kafkaResource.Spec.DataCentres = newDC
			kafkaResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.UpdatingEvent}

			Expect(k8sClient.Patch(ctx, &kafkaResource, patch)).Should(Succeed())

			By("sending resize request to Inst API v2.")

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, kafkaNS, &kafkaResource); err != nil {
					return false
				}

				if kafkaResource.Status.ID != mock.UpdatedID {
					return false
				}

				return kafkaResource.Status.DataCentres[0].Nodes[0].Size == mock.NewKafkaNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the Kafka resource", func() {
		It("should send deletion request", func() {
			Expect(k8sClient.Get(ctx, kafkaNS, &kafkaResource)).Should(Succeed())

			kafkaResource.Annotations = map[string]string{models.ResourceStateAnnotation: models.DeletingEvent}
			kafkaResource.Status.ID = mock.DeleteID

			Expect(k8sClient.Delete(ctx, &kafkaResource)).Should(Succeed())

			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, kafkaNS, &kafkaResource)
				if k8serrors.IsNotFound(err) {
					return true
				}
				if err != nil {
					return false
				}

				return false
			}, timeout, interval).Should(BeTrue())
		})
	})
})
