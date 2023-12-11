package kafkamanagement

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Kafka User Controller", func() {
	secret := &v1.Secret{}
	kafkaUser := &v1beta1.KafkaUser{}

	b, err := os.ReadFile(filepath.Join(".", "datatest", "kafkauser_v1beta1.yaml"))
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(b, kafkaUser)
	Expect(err).NotTo(HaveOccurred())

	b, err = os.ReadFile(filepath.Join(".", "datatest", "kafkauser_v1beta1_secret.yaml"))
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(b, secret)
	Expect(err).NotTo(HaveOccurred())

	const testClusterID = "testClusterID"

	When("Apply kafka user manifest", func() {
		It("creates kafka user resource and waits until status is filled", func() {
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
			Expect(k8sClient.Create(ctx, kafkaUser)).Should(Succeed())

			Expect(kafkaUser.Status.ClustersEvents).Should(BeEmpty())
		})
	})

	When("Add kafka user reference in kafka cluster resource", func() {
		It("should create user for the given cluster", func() {
			patch := kafkaUser.NewPatch()
			kafkaUser.Status.ClustersEvents = map[string]string{
				testClusterID: models.CreatingEvent,
			}

			Expect(k8sClient.Status().Patch(ctx, kafkaUser, patch)).Should(Succeed())

			Expect(kafkaUser.Status.ClustersEvents).ShouldNot(BeEmpty())

			Eventually(func(g Gomega) {
				g.Expect(kafkaUser.Status.ClustersEvents[testClusterID]).Should(Or(
					Equal(models.Created),
					Equal(models.UpdatedEvent),
				))
			}, timeout, interval)
		})
	})

	When("Remove kafka user reference from kafka cluster resource", func() {
		It("should delete user from the given cluster", func() {
			patch := kafkaUser.NewPatch()
			kafkaUser.Status.ClustersEvents = map[string]string{
				testClusterID: models.DeletingEvent,
			}

			Expect(k8sClient.Status().Patch(ctx, kafkaUser, patch)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(kafkaUser.Status.ClustersEvents).Should(BeEmpty())
			}, timeout, interval)
		})
	})
})
