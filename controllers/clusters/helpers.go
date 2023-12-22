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
	"strings"

	"github.com/hashicorp/go-version"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/dcomparison"
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
		var eq bool
		for j := range b {
			if a[i].ID != b[j].ID {
				continue
			}

			if a[i].Size != b[j].Size ||
				a[i].PublicAddress != b[j].PublicAddress ||
				a[i].PrivateAddress != b[j].PrivateAddress ||
				a[i].Status != b[j].Status ||
				!slices.Equal(a[i].Roles, b[j].Roles) ||
				a[i].Rack != b[j].Rack {
				return false
			}
			eq = true
		}
		if !eq {
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

func removeRedundantFieldsFromSpec(k8sSpec any, ignoreFields ...string) ([]byte, error) {
	k8sSpecJson, err := json.Marshal(k8sSpec)
	if err != nil {
		return nil, err
	}

	if len(ignoreFields) == 0 {
		return k8sSpecJson, nil
	}

	k8sSpecMap := map[string]any{}
	err = json.Unmarshal(k8sSpecJson, &k8sSpecMap)

	if err != nil {
		return nil, err
	}

	for _, field := range ignoreFields {
		delete(k8sSpecMap, field)
	}

	k8sSpecJson, err = json.Marshal(k8sSpecMap)
	if err != nil {
		return nil, err
	}
	return k8sSpecJson, nil
}

func createSpecDifferenceMessage(k8sSpec, iSpec any) (string, error) {
	k8sData, err := removeRedundantFieldsFromSpec(k8sSpec, "userRefs")
	if err != nil {
		return "", err
	}

	iData, err := json.Marshal(iSpec)
	if err != nil {
		return "", err
	}

	var k8sSpecMap map[string]any
	err = json.Unmarshal(k8sData, &k8sSpecMap)
	if err != nil {
		return "", err
	}

	var iSpecMap map[string]any
	err = json.Unmarshal(iData, &iSpecMap)
	if err != nil {
		return "", err
	}

	diffs := dcomparison.MapsDiff(models.SpecPath, k8sSpecMap, iSpecMap)

	return fmt.Sprintf("%s Diffs: %s", models.ExternalChangesBaseMessage, prepareDiffMessage(diffs)), nil
}

func prepareDiffMessage(diffs dcomparison.ObjectDiffs) string {
	var diffMessages []string
	for _, diff := range diffs {
		diffMessages = append(diffMessages, fmt.Sprintf(
			"{field: %s, k8sValue: %v, instaclustrValue: %v}",
			diff.Field,
			diff.Value1,
			diff.Value2,
		))
	}

	return strings.Join(diffMessages, ", ")
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
