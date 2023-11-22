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
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-version"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/controllers/clusterresources"
	"github.com/instaclustr/operator/pkg/models"
)

const CannotHandleUserEvent = "Cannot handle resource event. Reason: %v"

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
		a.CurrentClusterOperationStatus != b.CurrentClusterOperationStatus ||
		!areDataCentresEqual(a.DataCentres, b.DataCentres) ||
		!areDataCentreOptionsEqual(a.Options, b.Options) ||
		!b.PrivateLinkStatusesEqual(a) {
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
		if a[i].ID != b[i].ID {
			continue
		}

		if a[i].Status != b[i].Status ||
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
		if a[i].ID != b[i].ID {
			continue
		}

		if a[i].Size != b[i].Size ||
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

func getSortedAppVersions(versions []*models.AppVersions, appType string) []*version.Version {
	for _, apps := range versions {
		if apps.Application == appType {
			newVersions := make([]*version.Version, len(apps.Versions))
			for i, versionStr := range apps.Versions {
				v, _ := version.NewVersion(versionStr)
				newVersions[i] = v
			}

			sort.Sort(version.Collection(newVersions))

			return newVersions
		}
	}

	return nil
}

func createSpecDifferenceMessage(k8sSpec, iSpec any) (string, error) {
	k8sData, err := json.Marshal(k8sSpec)
	if err != nil {
		return "", err
	}

	iData, err := json.Marshal(iSpec)
	if err != nil {
		return "", err
	}

	msg := "There are external changes on the Instaclustr console. Please reconcile the specification manually. "
	specDifference := fmt.Sprintf("k8s spec: %s; data from instaclustr: %s", k8sData, iData)

	return msg + specDifference, nil
}

func isClusterResourceRefExists(ref *v1beta1.ClusterResourceRef, compareRefs []*v1beta1.ClusterResourceRef) bool {
	var exist bool
	for _, compareRef := range compareRefs {
		if *ref == *compareRef {
			exist = true
			break
		}
	}

	if exist {
		return exist
	}
	return false
}

var msgDeleteClusterWithTwoFactorDelete = "Please confirm cluster deletion via email or phone. " +
	"If you have canceled a cluster deletion and want to put the cluster on deletion again, " +
	"remove \"triggered\" from Instaclustr.com/clusterDeletion annotation."

var msgExternalChanges = "The k8s specification is different from Instaclustr Console. " +
	"Update operations are blocked. Please check operator logs and edit the cluster spec manually, " +
	"so that it would corresponds to the data from Instaclustr."

var msgSpecStillNoMatch = "k8s resource specification still doesn't match with data on the Instaclustr Console. Double check the difference."

// deleteDefaultUserSecret deletes the secret with default user credentials.
// It ignores NotFound error.
func deleteDefaultUserSecret(
	ctx context.Context,
	client client.Client,
	clusterNamespacedName types.NamespacedName,
) error {
	l := log.FromContext(ctx)

	l.Info("Deleting default user secret...",
		"resource namespaced name", clusterNamespacedName,
	)

	secret := &v1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, clusterNamespacedName.Name),
		Namespace: clusterNamespacedName.Namespace,
	}, secret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("The secret for the given resource is not found, skipping...",
				"resource namespaced name", clusterNamespacedName,
			)
			return nil
		}

		return err
	}

	return client.Delete(ctx, secret)
}

// Object is a general representation of any object the operator works with
type Object interface {
	client.Object
	NewPatch() client.Patch
}

func HandleCreateResource(
	r client.Client,
	ctx context.Context,
	l logr.Logger,
	kind string,
	ref *v1beta1.ClusterResourceRef,
	clusterID string,
	CDCs []*v1beta1.DataCentreStatus,
) error {
	req := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	var resource clusterresources.Object
	var isCDC bool
	cdcID := CDCs[0].ID

	switch kind {
	case models.ClusterbackupRef:
		resource = &clusterresourcesv1beta1.ClusterBackup{}
	case models.ClusterNetworkFirewallRuleRef:
		resource = &clusterresourcesv1beta1.ClusterNetworkFirewallRule{}
	case models.AWSVPCPeeringRef:
		resource = &clusterresourcesv1beta1.AWSVPCPeering{}
		isCDC = true
	case models.AWSSecurityGroupFirewallRuleRef:
		resource = &clusterresourcesv1beta1.AWSSecurityGroupFirewallRule{}
	case models.ExclusionWindowRef:
		resource = &clusterresourcesv1beta1.ExclusionWindow{}
	case models.GCPVPCPeeringRef:
		resource = &clusterresourcesv1beta1.GCPVPCPeering{}
		isCDC = true
	case models.AzureVNetPeeringRef:
		resource = &clusterresourcesv1beta1.AzureVNetPeering{}
		isCDC = true
	default:
		l.Info("Provided reference to resource that is not supported", "kind", kind)
		return nil
	}

	if isCDC && ref.DataCentreName != "" {
		for _, cdc := range CDCs {
			if cdc.Name == ref.DataCentreName {
				cdcID = cdc.ID
				break
			}
		}
	}
	err := r.Get(ctx, req, resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Provided resource is not found", "request", req)
			return err
		}
		l.Error(err, "Cannot get cluster resource", "request", req)
		return err
	}

	patch := resource.NewPatch()

	if isCDC {
		resource.AttachToCluster(cdcID)
	} else {
		resource.AttachToCluster(clusterID)
	}

	err = r.Status().Patch(ctx, resource, patch)
	if err != nil {
		return err
	}

	l.Info("PostgreSQL clusterresource was patched",
		"Reference", ref,
		"Resource Kind", kind,
		"Event", models.CreatingEvent,
	)

	return nil
}

func HandleDeleteResource(
	r client.Client,
	ctx context.Context,
	l logr.Logger,
	kind string,
	ref *v1beta1.ClusterResourceRef,
) error {
	req := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	var resource clusterresources.Object

	switch kind {
	case models.ClusterNetworkFirewallRuleRef:
		resource = &clusterresourcesv1beta1.ClusterNetworkFirewallRule{}
	case models.AWSVPCPeeringRef:
		resource = &clusterresourcesv1beta1.AWSVPCPeering{}
	case models.AWSSecurityGroupFirewallRuleRef:
		resource = &clusterresourcesv1beta1.AWSSecurityGroupFirewallRule{}
	case models.ExclusionWindowRef:
		resource = &clusterresourcesv1beta1.ExclusionWindow{}
	case models.GCPVPCPeeringRef:
		resource = &clusterresourcesv1beta1.GCPVPCPeering{}
	case models.AzureVNetPeeringRef:
		resource = &clusterresourcesv1beta1.AzureVNetPeering{}
	default:
		l.Info("Provided reference to resource that is not support deletion", "kind", kind)
		return nil
	}

	err := r.Get(ctx, req, resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Error(err, "Cannot get a cluster resource. The resource is not found", "request", req)
			return err
		}
		l.Error(err, "Cannot get cluster resource", "request", req)
		return err
	}

	patch := resource.NewPatch()

	resource.DetachFromCluster()

	err = r.Status().Patch(ctx, resource, patch)
	if err != nil {
		return err
	}

	l.Info("PostgreSQL clusterresource was updated",
		"Reference", ref,
		"Resource Kind", kind,
		"Event", models.DeletingEvent,
	)

	return nil
}

func HandleResourceEvent(
	r client.Client,
	resourceKind string,
	oldRefs, newRefs []*v1beta1.ClusterResourceRef,
	clusterID string,
	CDCs []*v1beta1.DataCentreStatus,
) error {
	ctx := context.TODO()
	l := log.FromContext(ctx)

	for _, ref := range newRefs {
		exist := isClusterResourceRefExists(ref, oldRefs)
		if exist {
			continue
		}
		err := HandleCreateResource(r, ctx, l, resourceKind, ref, clusterID, CDCs)
		if err != nil {
			l.Error(err, "Cannot create clusterresource", "resource kind", resourceKind, "namespace and name", ref)

			return err
		}
		oldRefs = append(oldRefs, ref)
	}
	for _, oldRef := range oldRefs {
		exist := isClusterResourceRefExists(oldRef, newRefs)
		if exist {
			continue
		}
		err := HandleDeleteResource(r, ctx, l, resourceKind, oldRef)
		if err != nil {
			l.Error(err, "Cannot delete clusterresource", "resource kind", resourceKind, "namespace and name", oldRef)

			return err
		}
	}
	return nil
}
