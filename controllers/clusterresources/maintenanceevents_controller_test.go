package clusterresources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Successful creation of a AWS Security Group Firewall Rule resource", func() {
	Context("When setting up a AWS Security Group Firewall Rule CRD", func() {
		maintenanceEventSpec := v1alpha1.MaintenanceEventsSpec{
			ClusterID:          "375e4d1c-2f77-4d02-a6f2-1af617ff2ab2",
			EventID:            "event-1312312",
			DayOfWeek:          "SUNDAY",
			StartHour:          12,
			DurationInHours:    4,
			ScheduledStartTime: "12-12-12",
		}

		ctx := context.Background()
		resource := v1alpha1.MaintenanceEvents{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "me",
				Namespace: "default",
				Annotations: map[string]string{
					models.ResourceStateAnnotation: models.CreatingEvent,
				},
			},
			Spec: maintenanceEventSpec,
		}

		It("Should create a AWS Security Group Firewall Rule resources", func() {
			Expect(k8sClient.Create(ctx, &resource)).Should(Succeed())

			By("Sending AWS Security Group Firewall Rule Specification to Instaclustr API v2")
			var maintenanceEvent v1alpha1.MaintenanceEvents
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: "me", Namespace: "default"}, &maintenanceEvent); err != nil {
					return false
				}

				return maintenanceEvent.Status.ID == mock.StatusID
			}).Should(BeTrue())
		})
	})
})
