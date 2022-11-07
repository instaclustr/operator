package clusterresources

import (
	"github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
)

func isFirewallRuleStatusesEqual(a, b *v1alpha1.ClusterNetworkFirewallRuleStatus) bool {
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