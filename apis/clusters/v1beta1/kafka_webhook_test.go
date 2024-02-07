package v1beta1

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Kafka Controller", Ordered, func() {
	kafkaManifest := Kafka{}
	testKafkaManifest := Kafka{}

	It("Reading kafka manifest", func() {
		yfile, err := os.ReadFile("../../../controllers/clusters/datatest/kafka_v1beta1.yaml")
		Expect(err).Should(Succeed())

		err = yaml.Unmarshal(yfile, &kafkaManifest)
		Expect(err).Should(Succeed())
		testKafkaManifest = kafkaManifest
	})

	ctx := context.Background()

	When("apply a Kafka manifest", func() {
		It("should test kafka creation flow", func() {

			testKafkaManifest.Spec.Description = "some description"
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Description = kafkaManifest.Spec.Description

			testKafkaManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{kafkaManifest.Spec.TwoFactorDelete[0], kafkaManifest.Spec.TwoFactorDelete[0]}
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).ShouldNot(Succeed())
			testKafkaManifest.Spec.TwoFactorDelete = kafkaManifest.Spec.TwoFactorDelete

			testKafkaManifest.Spec.SLATier = "some SLATier that is not supported"
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).ShouldNot(Succeed())
			testKafkaManifest.Spec.SLATier = kafkaManifest.Spec.SLATier

			testKafkaManifest.Spec.Kraft = []*Kraft{kafkaManifest.Spec.Kraft[0], kafkaManifest.Spec.Kraft[0]}
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Kraft = []*Kraft{testKafkaManifest.Spec.Kraft[0]}

			testKafkaManifest.Spec.Kraft[0].ControllerNodeCount = 4
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Kraft[0].ControllerNodeCount = 3
		})
	})

	When("updating a Kafka manifest", func() {
		It("should test kafka update flow", func() {
			testKafkaManifest.Status.State = models.RunningStatus
			Expect(k8sClient.Create(ctx, &testKafkaManifest)).Should(Succeed())

			patch := testKafkaManifest.NewPatch()
			testKafkaManifest.Status.ID = models.CreatedEvent
			Expect(k8sClient.Status().Patch(ctx, &testKafkaManifest, patch)).Should(Succeed())

			testKafkaManifest.Spec.Name += "newValue"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Name = kafkaManifest.Spec.Name

			testKafkaManifest.Spec.Version += ".1"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Version = kafkaManifest.Spec.Version

			testKafkaManifest.Spec.PCICompliance = !kafkaManifest.Spec.PCICompliance
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.PCICompliance = kafkaManifest.Spec.PCICompliance

			testKafkaManifest.Spec.PrivateNetwork = !kafkaManifest.Spec.PrivateNetwork
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.PrivateNetwork = kafkaManifest.Spec.PrivateNetwork

			testKafkaManifest.Spec.PartitionsNumber++
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.PartitionsNumber = kafkaManifest.Spec.PartitionsNumber

			testKafkaManifest.Spec.AllowDeleteTopics = !kafkaManifest.Spec.AllowDeleteTopics
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.AllowDeleteTopics = kafkaManifest.Spec.AllowDeleteTopics

			testKafkaManifest.Spec.AutoCreateTopics = !kafkaManifest.Spec.AutoCreateTopics
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.AutoCreateTopics = kafkaManifest.Spec.AutoCreateTopics

			testKafkaManifest.Spec.ClientToClusterEncryption = !kafkaManifest.Spec.ClientToClusterEncryption
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.ClientToClusterEncryption = kafkaManifest.Spec.ClientToClusterEncryption

			testKafkaManifest.Spec.BundledUseOnly = !kafkaManifest.Spec.BundledUseOnly
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.BundledUseOnly = kafkaManifest.Spec.BundledUseOnly

			testKafkaManifest.Spec.PrivateNetwork = !kafkaManifest.Spec.PrivateNetwork
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.PrivateNetwork = kafkaManifest.Spec.PrivateNetwork

			testKafkaManifest.Spec.ClientBrokerAuthWithMTLS = !kafkaManifest.Spec.ClientBrokerAuthWithMTLS
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.ClientBrokerAuthWithMTLS = kafkaManifest.Spec.ClientBrokerAuthWithMTLS

			prevSchemaRegistry := kafkaManifest.Spec.SchemaRegistry
			testKafkaManifest.Spec.SchemaRegistry = []*SchemaRegistry{prevSchemaRegistry[0], prevSchemaRegistry[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.SchemaRegistry = prevSchemaRegistry

			prevKarapaceSchemaRegistry := kafkaManifest.Spec.KarapaceSchemaRegistry
			testKafkaManifest.Spec.KarapaceSchemaRegistry = []*KarapaceSchemaRegistry{prevKarapaceSchemaRegistry[0], prevKarapaceSchemaRegistry[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.KarapaceSchemaRegistry = prevKarapaceSchemaRegistry

			prevRestProxy := kafkaManifest.Spec.RestProxy
			testKafkaManifest.Spec.RestProxy = []*RestProxy{prevRestProxy[0], prevRestProxy[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.RestProxy = prevRestProxy

			prevRestProxyVersion := prevRestProxy[0].Version
			testKafkaManifest.Spec.RestProxy[0].Version += ".0"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.RestProxy[0].Version = prevRestProxyVersion

			prevKraft := kafkaManifest.Spec.Kraft
			testKafkaManifest.Spec.Kraft = []*Kraft{prevKraft[0], prevKraft[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Kraft = prevKraft

			prevKraftControllerNodeCount := prevKraft[0].ControllerNodeCount
			testKafkaManifest.Spec.Kraft[0].ControllerNodeCount++
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.Kraft[0].ControllerNodeCount = prevKraftControllerNodeCount

			prevKarapaceRestProxy := kafkaManifest.Spec.KarapaceRestProxy
			testKafkaManifest.Spec.KarapaceRestProxy = []*KarapaceRestProxy{prevKarapaceRestProxy[0], prevKarapaceRestProxy[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.KarapaceRestProxy = prevKarapaceRestProxy

			prevConcurrency := kafkaManifest.Spec.ResizeSettings[0].Concurrency
			testKafkaManifest.Spec.ResizeSettings[0].Concurrency++
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.ResizeSettings[0].Concurrency = prevConcurrency

			prevDedicatedZookeeper := kafkaManifest.Spec.DedicatedZookeeper
			testKafkaManifest.Spec.DedicatedZookeeper = []*DedicatedZookeeper{prevDedicatedZookeeper[0], prevDedicatedZookeeper[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DedicatedZookeeper = prevDedicatedZookeeper

			prevDedicatedZookeeperNodesNumber := kafkaManifest.Spec.DedicatedZookeeper[0].NodesNumber
			testKafkaManifest.Spec.DedicatedZookeeper[0].NodesNumber++
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DedicatedZookeeper[0].NodesNumber = prevDedicatedZookeeperNodesNumber

			testKafkaManifest.Spec.DataCentres = []*KafkaDataCentre{}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres = kafkaManifest.Spec.DataCentres

			By("changing datacentres fields")
			prevIntField := kafkaManifest.Spec.DataCentres[0].NodesNumber
			testKafkaManifest.Spec.DataCentres[0].NodesNumber++
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].NodesNumber = prevIntField

			prevIntField = kafkaManifest.Spec.DataCentres[0].NodesNumber
			testKafkaManifest.Spec.DataCentres[0].NodesNumber = 0
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].NodesNumber = prevIntField

			prevStringField := kafkaManifest.Spec.DataCentres[0].Name
			testKafkaManifest.Spec.DataCentres[0].Name += "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].Name = prevStringField

			prevStringField = kafkaManifest.Spec.DataCentres[0].Region
			testKafkaManifest.Spec.DataCentres[0].Region += "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].Region = prevStringField

			prevStringField = kafkaManifest.Spec.DataCentres[0].CloudProvider
			testKafkaManifest.Spec.DataCentres[0].CloudProvider += "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].CloudProvider = prevStringField

			prevStringField = kafkaManifest.Spec.DataCentres[0].ProviderAccountName
			testKafkaManifest.Spec.DataCentres[0].ProviderAccountName += "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].ProviderAccountName = prevStringField

			prevStringField = kafkaManifest.Spec.DataCentres[0].Network
			testKafkaManifest.Spec.DataCentres[0].Network += "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].Network = prevStringField

			prevCloudProviderSettings := kafkaManifest.Spec.DataCentres[0].CloudProviderSettings
			testKafkaManifest.Spec.DataCentres[0].CloudProviderSettings = []*CloudProviderSettings{prevCloudProviderSettings[0], prevCloudProviderSettings[0]}
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].CloudProviderSettings = prevCloudProviderSettings

			testKafkaManifest.Spec.DataCentres[0].Tags["test"] = "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			delete(testKafkaManifest.Spec.DataCentres[0].Tags, "test")

			delete(testKafkaManifest.Spec.DataCentres[0].Tags, "tag")
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].Tags["tag"] = "oneTag"

			testKafkaManifest.Spec.DataCentres[0].Tags["tag"] = "test"
			Expect(k8sClient.Patch(ctx, &testKafkaManifest, patch)).ShouldNot(Succeed())
			testKafkaManifest.Spec.DataCentres[0].Tags["tag"] = "oneTag"
		})
	})
})
