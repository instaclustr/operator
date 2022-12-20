package clusters

import (
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/models"
)

// confirmDeletion confirms if resource is deleting and set appropriate annotation.
func confirmDeletion(obj client.Object) bool {
	annots := obj.GetAnnotations()

	if obj.GetDeletionTimestamp() != nil && annots[models.ClusterDeletionAnnotation] != models.Triggered {
		annots[models.ResourceStateAnnotation] = models.DeletingEvent
		return true
	}

	return false
}

func isStatusesEqual(a, b *clustersv1alpha1.ClusterStatus) bool {
	if a.ID != b.ID ||
		a.Status != b.Status ||
		a.CDCID != b.CDCID ||
		a.TwoFactorDeleteEnabled != b.TwoFactorDeleteEnabled ||
		!isDataCentreEqual(a.DataCentres, b.DataCentres) ||
		!isDataCentreOptionsEqual(a.Options, b.Options) {
		return false
	}

	return true
}

func isDataCentreOptionsEqual(a, b *clustersv1alpha1.Options) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func isDataCentreEqual(a, b []*clustersv1alpha1.DataCentreStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID ||
			a[i].Status != b[i].Status ||
			a[i].NodeNumber != b[i].NodeNumber ||
			a[i].EncryptionKeyID != b[i].EncryptionKeyID {
			return false
		}

		if !isDataCentreNodesEqual(a[i].Nodes, b[i].Nodes) {
			return false
		}
	}

	return true
}

func isDataCentreNodesEqual(a, b []*clustersv1alpha1.Node) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID ||
			a[i].Size != b[i].Size ||
			a[i].PublicAddress != b[i].PublicAddress ||
			a[i].PrivateAddress != b[i].PrivateAddress ||
			a[i].Status != b[i].Status ||
			!slices.Equal(a[i].Roles, b[i].Roles) ||
			a[i].Rack != b[i].Rack {
			return false
		}
	}

	return true
}
