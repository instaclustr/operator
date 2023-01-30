package clusters

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Successful creation of a Kafka resource", func() {
	Context("When setting up a Kafka CRD", func() {
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
				NodeSize:      "r5.large-500-gp2",
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

		ctx := context.Background()
		resource := v1alpha1.Kafka{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kafka",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: kafkaSpec,
		}

		It("Should create a Kafka resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending Kafka Specification to Instaclustr API v2")
			var kafkaCluster v1alpha1.Kafka
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "kafka", Namespace: "default"}, &kafkaCluster); err != nil {
					return false
				}

				return kafkaCluster.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})
})
