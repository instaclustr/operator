/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusters

import (
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
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

func convertAPIv2ConfigToMap(instConfigs []*models.ConfigurationProperties) map[string]string {
	newConfigs := map[string]string{}
	for _, instConfig := range instConfigs {
		newConfigs[instConfig.Name] = instConfig.Value
	}
	return newConfigs
}
func areStatusesEqual(a, b *v1beta1.ClusterStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil ||
		a.ID != b.ID ||
		a.State != b.State ||
		a.CDCID != b.CDCID ||
		a.TwoFactorDeleteEnabled != b.TwoFactorDeleteEnabled ||
		!areDataCentresEqual(a.DataCentres, b.DataCentres) ||
		!areDataCentreOptionsEqual(a.Options, b.Options) {
		return false
	}

	return true
}

func areDataCentreOptionsEqual(a, b *v1beta1.Options) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func areDataCentresEqual(a, b []*v1beta1.DataCentreStatus) bool {
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

func isDataCentreNodesEqual(a, b []*v1beta1.Node) bool {
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

func isClusterActive(clusterID string, activeClusters []*models.ActiveClusters) bool {
	for _, activeCluster := range activeClusters {
		for _, cluster := range activeCluster.Clusters {
			if cluster.ID == clusterID {
				return true
			}
		}
	}

	return false
}

var msgDeleteClusterWithTwoFactorDelete = "Please confirm cluster deletion via email or phone. " +
	"If you have canceled a cluster deletion and want to put the cluster on deletion again, " +
	"remove \"triggered\" from Instaclustr.com/clusterDeletion annotation."

var msgExternalChanges = "The k8s specification is different from Instaclustr. Please reconcile the specs manually, " +
	"add \"instaclustr.com/allowSpecAmend: \"true\" \" annotation to be able to change the k8s resource spec."

var msgSpecStillNoMatch = "k8s specification still don't match with Instaclustr Console. Double check the difference."
