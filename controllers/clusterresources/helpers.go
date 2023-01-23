package clusterresources

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

func isFirewallRuleStatusesEqual(a, b *v1alpha1.FirewallRuleStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil ||
		b == nil ||
		a.ID != b.ID ||
		a.Status != b.Status ||
		a.DeferredReason != b.DeferredReason {
		return false
	}

	return true
}

func isPeeringStatusesEqual(a, b *v1alpha1.PeeringStatus) bool {
	if a.ID != b.ID ||
		a.Name != b.Name ||
		a.StatusCode != b.StatusCode ||
		a.FailureReason != b.FailureReason {
		return false
	}

	return true
}

func isMaintenanceEventStatusesEqual(a, b []*v1alpha1.MaintenanceEvent) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].EventID != b[i].EventID ||
			a[i].ClusterID != b[i].ClusterID ||
			a[i].ExpectedServiceDisruption != b[i].ExpectedServiceDisruption ||
			a[i].Description != b[i].Description ||
			a[i].ScheduledStartTime != b[i].ScheduledStartTime ||
			a[i].ScheduledEndTime != b[i].ScheduledEndTime ||
			a[i].ActualStartTime != b[i].ActualStartTime {
			return false
		}
	}
	return true
}

// confirmDeletion confirms if resource is deleting and set appropriate annotation.
func confirmDeletion(obj client.Object) {
	if obj.GetDeletionTimestamp() != nil {
		obj.GetAnnotations()[models.ResourceStateAnnotation] = models.DeletingEvent
		return
	}
}
