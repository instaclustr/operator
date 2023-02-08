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

var _ = Describe("Successful creation of a Cassandra resource", func() {
	Context("When setting up a Cassandra CRD", func() {
		cassandraSpec := v1alpha1.CassandraSpec{
			Cluster: v1alpha1.Cluster{
				Name:                  "cassandraTest",
				Version:               "3.11.13",
				PCICompliance:         false,
				PrivateNetworkCluster: false,
				SLATier:               "NON_PRODUCTION",
				TwoFactorDelete:       nil,
			},
			LuceneEnabled:       false,
			PasswordAndUserAuth: false,
			Spark: []*v1alpha1.Spark{{
				Version: "2.3.2",
			}},
			DataCentres: []*v1alpha1.CassandraDataCentre{{
				DataCentre: v1alpha1.DataCentre{
					NodesNumber:   3,
					Network:       "172.16.0.0/19",
					NodeSize:      "CAS-DEV-t4g.small-30",
					CloudProvider: "AWS_VPC",
					Name:          "US_EAST_1DC",
					Region:        "US_EAST_1",
				},
				ContinuousBackup:               false,
				PrivateIPBroadcastForDiscovery: false,
				ClientToClusterEncryption:      false,
				ReplicationFactor:              3,
			}},
		}

		ctx := context.Background()
		resource := v1alpha1.Cassandra{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cassandra",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: cassandraSpec,
		}

		It("Should create a Cassandra resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending Cassandra Specification to Instaclustr API v2")
			var cassandraCluster v1alpha1.Cassandra
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "cassandra", Namespace: "default"}, &cassandraCluster); err != nil {
					return false
				}

				return cassandraCluster.Status.ID == openapi.CreatedID
			}).Should(BeTrue())
		})
	})
})
