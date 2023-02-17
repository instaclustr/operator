package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/instaclustr/operator/pkg/models"
)

type VPCPeeringSpec struct {
	DataCentreID string   `json:"cdcId"`
	PeerSubnets  []string `json:"peerSubnets"`
}

type PeeringStatus struct {
	ID            string `json:"id,omitempty"`
	StatusCode    string `json:"statusCode,omitempty"`
	Name          string `json:"name,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type FirewallRuleSpec struct {
	ClusterID string `json:"clusterId"`
	Type      string `json:"type"`
}

type FirewallRuleStatus struct {
	ID             string `json:"id,omitempty"`
	DeferredReason string `json:"deferredReason,omitempty"`
	Status         string `json:"status,omitempty"`
}

type immutablePeeringFields struct {
	DataCentreID string
}

func IsClusterBeingDeleted(deletionTimestamp *metav1.Time, twoFactorDeleteLen int, deletionConfirmedAnnotation string) bool {
	if deletionTimestamp != nil {
		if twoFactorDeleteLen == 0 {
			return true
		}

		return deletionConfirmedAnnotation == models.True
	}
	return false
}
