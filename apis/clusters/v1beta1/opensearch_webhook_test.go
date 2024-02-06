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
	openSearchManifest := OpenSearch{}
	testOpenSearchManifest := OpenSearch{}

	It("Reading kafka manifest", func() {
		yfile, err := os.ReadFile("../../../controllers/clusters/datatest/opensearch_v1beta1.yaml")
		Expect(err).Should(Succeed())

		err = yaml.Unmarshal(yfile, &openSearchManifest)
		Expect(err).Should(Succeed())
		openSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{{Email: "emailTEST", Phone: "phoneTEST"}}
		testOpenSearchManifest = openSearchManifest
	})

	ctx := context.Background()

	When("apply an OpenSearch manifest", func() {
		It("should test OpenSearch creation flow", func() {
			testOpenSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{openSearchManifest.Spec.TwoFactorDelete[0], openSearchManifest.Spec.TwoFactorDelete[0]}
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.TwoFactorDelete = openSearchManifest.Spec.TwoFactorDelete

			testOpenSearchManifest.Spec.SLATier = "some SLATier that is not supported"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.SLATier = openSearchManifest.Spec.SLATier

			testOpenSearchManifest.Spec.DataCentres = []*OpenSearchDataCentre{}
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres = openSearchManifest.Spec.DataCentres

			prevDataNode := openSearchManifest.Spec.DataNodes[0]
			testOpenSearchManifest.Spec.DataNodes = []*OpenSearchDataNodes{}
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataNodes = []*OpenSearchDataNodes{prevDataNode}

			prevDC := *openSearchManifest.Spec.DataCentres[0]
			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "test"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())

			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "AWS_VPC"
			testOpenSearchManifest.Spec.DataCentres[0].Region = "test"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())

			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "AZURE_AZ"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())

			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "GCP"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0] = &prevDC

			prevStringValue := openSearchManifest.Spec.DataCentres[0].ProviderAccountName
			testOpenSearchManifest.Spec.DataCentres[0].ProviderAccountName = models.DefaultAccountName
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].ProviderAccountName = prevStringValue

			providerSettings := openSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0]
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings = []*CloudProviderSettings{providerSettings, providerSettings}
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings = []*CloudProviderSettings{providerSettings}

			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].ResourceGroup = "test"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())

			prevStringValue = openSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey = ""
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].CustomVirtualNetworkID = "test"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].ResourceGroup = ""
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].CustomVirtualNetworkID = ""
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey = prevStringValue

			prevStringValue = openSearchManifest.Spec.DataCentres[0].Network
			testOpenSearchManifest.Spec.DataCentres[0].Network = "test/test"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Network = prevStringValue

			testOpenSearchManifest.Spec.DataNodes[0].NodesNumber++
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataNodes[0].NodesNumber--

			testOpenSearchManifest.Spec.DataCentres[0].NumberOfRacks++
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].NumberOfRacks--

			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "GCP"
			testOpenSearchManifest.Spec.DataCentres[0].PrivateLink = true
			testOpenSearchManifest.Spec.DataCentres[0].Region = "us-east1"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "AWS_VPC"
			testOpenSearchManifest.Spec.DataCentres[0].Region = "US_EAST_1"

			testOpenSearchManifest.Spec.PrivateNetwork = !testOpenSearchManifest.Spec.PrivateNetwork
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].PrivateLink = false
			testOpenSearchManifest.Spec.PrivateNetwork = !testOpenSearchManifest.Spec.PrivateNetwork

			testOpenSearchManifest.Spec.ResizeSettings[0].Concurrency++
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ResizeSettings[0].Concurrency--

			prevManagedNode := openSearchManifest.Spec.ClusterManagerNodes[0]
			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{prevManagedNode, prevManagedNode, prevManagedNode, prevManagedNode}
			testOpenSearchManifest.Spec.ResizeSettings[0].Concurrency += 3
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ResizeSettings[0].Concurrency -= 3
			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{prevManagedNode}

			testOpenSearchManifest.Spec.Version += ".1"
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.Version = openSearchManifest.Spec.Version

		})
	})

	When("updating an OpenSearch manifest", func() {
		It("should test OpenSearch update flow", func() {
			testOpenSearchManifest.Status.State = models.RunningStatus
			Expect(k8sClient.Create(ctx, &testOpenSearchManifest)).Should(Succeed())

			patch := testOpenSearchManifest.NewPatch()
			testOpenSearchManifest.Status.ID = models.CreatedEvent
			Expect(k8sClient.Status().Patch(ctx, &testOpenSearchManifest, patch)).Should(Succeed())

			testOpenSearchManifest.Spec.Name += "newValue"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.Name = openSearchManifest.Spec.Name

			testOpenSearchManifest.Spec.BundledUseOnly = !testOpenSearchManifest.Spec.BundledUseOnly
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.BundledUseOnly = !testOpenSearchManifest.Spec.BundledUseOnly

			testOpenSearchManifest.Spec.ICUPlugin = !testOpenSearchManifest.Spec.ICUPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ICUPlugin = !testOpenSearchManifest.Spec.ICUPlugin

			testOpenSearchManifest.Spec.AsynchronousSearchPlugin = !testOpenSearchManifest.Spec.AsynchronousSearchPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.AsynchronousSearchPlugin = !testOpenSearchManifest.Spec.AsynchronousSearchPlugin

			testOpenSearchManifest.Spec.KNNPlugin = !testOpenSearchManifest.Spec.KNNPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.KNNPlugin = !testOpenSearchManifest.Spec.KNNPlugin

			testOpenSearchManifest.Spec.ReportingPlugin = !testOpenSearchManifest.Spec.ReportingPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ReportingPlugin = !testOpenSearchManifest.Spec.ReportingPlugin

			testOpenSearchManifest.Spec.SQLPlugin = !testOpenSearchManifest.Spec.SQLPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.SQLPlugin = !testOpenSearchManifest.Spec.SQLPlugin

			testOpenSearchManifest.Spec.NotificationsPlugin = !testOpenSearchManifest.Spec.NotificationsPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.NotificationsPlugin = !testOpenSearchManifest.Spec.NotificationsPlugin

			testOpenSearchManifest.Spec.AnomalyDetectionPlugin = !testOpenSearchManifest.Spec.AnomalyDetectionPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.AnomalyDetectionPlugin = !testOpenSearchManifest.Spec.AnomalyDetectionPlugin

			testOpenSearchManifest.Spec.LoadBalancer = !testOpenSearchManifest.Spec.LoadBalancer
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.LoadBalancer = !testOpenSearchManifest.Spec.LoadBalancer

			testOpenSearchManifest.Spec.IndexManagementPlugin = !testOpenSearchManifest.Spec.IndexManagementPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.IndexManagementPlugin = !testOpenSearchManifest.Spec.IndexManagementPlugin

			testOpenSearchManifest.Spec.AlertingPlugin = !testOpenSearchManifest.Spec.AlertingPlugin
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.AlertingPlugin = !testOpenSearchManifest.Spec.AlertingPlugin

			testOpenSearchManifest.Spec.PCICompliance = !testOpenSearchManifest.Spec.PCICompliance
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.PCICompliance = !testOpenSearchManifest.Spec.PCICompliance

			testOpenSearchManifest.Spec.PrivateNetwork = !testOpenSearchManifest.Spec.PrivateNetwork
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.PrivateNetwork = !testOpenSearchManifest.Spec.PrivateNetwork

			prevStringValue := openSearchManifest.Spec.SLATier
			testOpenSearchManifest.Spec.SLATier = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.SLATier = prevStringValue

			prevTwoFactorDelete := testOpenSearchManifest.Spec.TwoFactorDelete
			testOpenSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{prevTwoFactorDelete[0], prevTwoFactorDelete[0]}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{prevTwoFactorDelete[0]}

			testOpenSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.TwoFactorDelete = []*TwoFactorDelete{prevTwoFactorDelete[0]}

			prevStringValue = openSearchManifest.Spec.TwoFactorDelete[0].Email
			testOpenSearchManifest.Spec.TwoFactorDelete[0].Email = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.TwoFactorDelete[0].Email = prevStringValue

			prevIngestNodes := testOpenSearchManifest.Spec.IngestNodes
			testOpenSearchManifest.Spec.IngestNodes = []*OpenSearchIngestNodes{prevIngestNodes[0], prevIngestNodes[0]}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.IngestNodes = []*OpenSearchIngestNodes{prevIngestNodes[0]}

			testOpenSearchManifest.Spec.IngestNodes = []*OpenSearchIngestNodes{}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.IngestNodes = []*OpenSearchIngestNodes{prevIngestNodes[0]}

			prevStringValue = openSearchManifest.Spec.IngestNodes[0].NodeSize
			testOpenSearchManifest.Spec.IngestNodes[0].NodeSize = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.IngestNodes[0].NodeSize = prevStringValue

			prevClusterManagedNodes := testOpenSearchManifest.Spec.ClusterManagerNodes
			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{prevClusterManagedNodes[0], prevClusterManagedNodes[0]}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{prevClusterManagedNodes[0]}

			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ClusterManagerNodes = []*ClusterManagerNodes{prevClusterManagedNodes[0]}

			prevStringValue = openSearchManifest.Spec.ClusterManagerNodes[0].NodeSize
			testOpenSearchManifest.Spec.ClusterManagerNodes[0].NodeSize = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.ClusterManagerNodes[0].NodeSize = prevStringValue

			By("changing datacentres fields")

			prevDCs := openSearchManifest.Spec.DataCentres
			testOpenSearchManifest.Spec.DataCentres = []*OpenSearchDataCentre{prevDCs[0], prevDCs[0]}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres = []*OpenSearchDataCentre{prevDCs[0]}

			prevStringValue = prevDCs[0].Name
			testOpenSearchManifest.Spec.DataCentres[0].Name = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Name = prevStringValue

			prevStringValue = prevDCs[0].Region
			testOpenSearchManifest.Spec.DataCentres[0].Region = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Region = prevStringValue

			prevStringValue = prevDCs[0].CloudProvider
			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProvider = prevStringValue

			prevStringValue = prevDCs[0].ProviderAccountName
			testOpenSearchManifest.Spec.DataCentres[0].ProviderAccountName = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].ProviderAccountName = prevStringValue

			prevStringValue = prevDCs[0].Network
			testOpenSearchManifest.Spec.DataCentres[0].Network = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Network = prevStringValue

			testOpenSearchManifest.Spec.DataCentres[0].PrivateLink = !testOpenSearchManifest.Spec.DataCentres[0].PrivateLink
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].PrivateLink = !testOpenSearchManifest.Spec.DataCentres[0].PrivateLink

			testOpenSearchManifest.Spec.DataCentres[0].NumberOfRacks += 1
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].NumberOfRacks -= 1

			prevCloudProviderSettings := openSearchManifest.Spec.DataCentres[0].CloudProviderSettings
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings = []*CloudProviderSettings{prevCloudProviderSettings[0], prevCloudProviderSettings[0]}
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings = []*CloudProviderSettings{prevCloudProviderSettings[0]}

			prevStringValue = openSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].DiskEncryptionKey = prevStringValue

			prevStringValue = openSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].ResourceGroup
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].ResourceGroup = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].ResourceGroup = prevStringValue

			prevStringValue = openSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].CustomVirtualNetworkID
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].CustomVirtualNetworkID = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].CloudProviderSettings[0].CustomVirtualNetworkID = prevStringValue

			testOpenSearchManifest.Spec.DataCentres[0].Tags["test"] = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			delete(testOpenSearchManifest.Spec.DataCentres[0].Tags, "test")

			delete(testOpenSearchManifest.Spec.DataCentres[0].Tags, "tag")
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Tags["tag"] = "oneTag"

			testOpenSearchManifest.Spec.DataCentres[0].Tags["tag"] = "test"
			Expect(k8sClient.Patch(ctx, &testOpenSearchManifest, patch)).ShouldNot(Succeed())
			testOpenSearchManifest.Spec.DataCentres[0].Tags["tag"] = "oneTag"

		})

	})
})
