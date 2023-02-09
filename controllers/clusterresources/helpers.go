package clusterresources

import (
	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
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
