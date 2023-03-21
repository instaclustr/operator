package clusterresources

import (
	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
)

func areFirewallRuleStatusesEqual(a, b *v1alpha1.FirewallRuleStatus) bool {
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

func arePeeringStatusesEqual(a, b *v1alpha1.PeeringStatus) bool {
	if a.ID != b.ID ||
		a.Name != b.Name ||
		a.StatusCode != b.StatusCode ||
		a.FailureReason != b.FailureReason {
		return false
	}

	return true
}

func areEncryptionKeyStatusesEqual(a, b *v1alpha1.AWSEncryptionKeyStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil ||
		b == nil ||
		a.ID != b.ID ||
		a.InUse != b.InUse {
		return false
	}

	return true
}
