package kafkamanagement

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Successful creation of a kafka ACL resource", func() {
	Context("When setting up a kafka ACL CRD", func() {
		kafkaACLSpec := kafkamanagementv1alpha1.KafkaACLSpec{
			ACLs: []kafkamanagementv1alpha1.ACL{
				{
					Principal:      "User:test",
					PermissionType: "ALLOW",
					Host:           "*",
					PatternType:    "PREFIXED",
					ResourceName:   "kafka-topic",
					Operation:      "CREATE",
					ResourceType:   "CLUSTER",
				},
			},
			UserQuery: "test",
			ClusterID: "c1af59c6-ba0e-4cc2-a0f3-65cee17a5f37",
		}

		ctx := context.Background()
		resource := kafkamanagementv1alpha1.KafkaACL{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kafkaacl",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: kafkaACLSpec,
		}

		It("Should create a kafka ACL resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending kafka ACL specification to Instaclustr API v2")
			var kafkaACL kafkamanagementv1alpha1.KafkaACL
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "kafkaacl", Namespace: "default"}, &kafkaACL); err != nil {
					return false
				}

				return kafkaACL.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
