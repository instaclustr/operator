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
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-version"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/pkg/instaclustr"
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

type objectDiff struct {
	Field            string `json:"field"`
	K8sValue         any    `json:"k8sValue"`
	InstaclustrValue any    `json:"instaclustrValue"`
}

func createSpecDifferenceMessage[T any](k8sSpec, iSpec T) (string, error) {
	diffs, err := dcomparison.StructsDiff(models.SpecPath, k8sSpec, iSpec)
	if err != nil {
		return "", fmt.Errorf("failed to create spec difference message, err: %w", err)
	}

	objectDiffs := make([]objectDiff, 0, len(diffs))
	for _, diff := range diffs {
		objectDiffs = append(objectDiffs, objectDiff{
			Field:            diff.Field,
			K8sValue:         diff.Value1,
			InstaclustrValue: diff.Value2,
		})
	}

	b, err := json.Marshal(objectDiffs)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s Diffs: %s", models.ExternalChangesBaseMessage, b), nil
}

func newDefaultUserSecret(username, password, name, namespace string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.SecretKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(models.DefaultUserSecretNameTemplate, models.DefaultUserSecretPrefix, name),
			Namespace: namespace,
			Labels: map[string]string{
				models.ControlledByLabel:  name,
				models.DefaultSecretLabel: "true",
			},
		},
		StringData: map[string]string{
			models.Username: username,
			models.Password: password,
		},
	}
}

var msgDeleteClusterWithTwoFactorDelete = "Please confirm cluster deletion via email or phone. " +
	"If you have canceled a cluster deletion and want to put the cluster on deletion again, " +
	"remove \"triggered\" from Instaclustr.com/clusterDeletion annotation."

var msgExternalChanges = "The k8s specification is different from Instaclustr Console. " +
	"Update operations are blocked. Please check operator logs and edit the cluster spec manually, " +
	"so that it would corresponds to the data from Instaclustr."

// Object is a general representation of any object the operator works with
type Object interface {
	client.Object
	NewPatch() client.Patch
}

type ExternalChanger[T any] interface {
	Object
	GetSpec() T
	IsSpecEqual(T) bool
	GetClusterID() string
}

func handleExternalChanges[T any](
	r record.EventRecorder,
	c client.Client,
	resource, iResource ExternalChanger[T],
	l logr.Logger,
) (reconcile.Result, error) {
	patch := resource.NewPatch()

	if !resource.IsSpecEqual(iResource.GetSpec()) {
		resource.GetAnnotations()[models.ExternalChangesAnnotation] = models.True

		l.Info(msgExternalChanges,
			"specification of k8s resource", resource.GetSpec(),
			"data from Instaclustr ", iResource.GetSpec())

		msgDiffSpecs, err := createSpecDifferenceMessage(resource.GetSpec(), iResource.GetSpec())
		if err != nil {
			l.Error(err, "Cannot create specification difference message",
				"instaclustr data", iResource.GetSpec(), "k8s resource spec", resource.GetSpec())
			return models.ExitReconcile, nil
		}

		r.Eventf(resource, models.Warning, models.ExternalChanges, msgDiffSpecs)
	} else {
		resource.GetAnnotations()[models.ExternalChangesAnnotation] = ""
		resource.GetAnnotations()[models.ResourceStateAnnotation] = models.UpdatedEvent

		l.Info("External changes have been reconciled", "cluster ID", resource.GetClusterID())
		r.Event(resource, models.Normal, models.ExternalChanges, "External changes have been reconciled")
	}

	err := c.Patch(context.Background(), resource, patch)
	if err != nil {
		l.Error(err, "Cannot patch cluster resource",
			"cluster name", resource.GetName(), "cluster ID", resource.GetClusterID())

		r.Eventf(resource, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return reconcile.Result{}, err
	}

	return models.ExitReconcile, nil
}

func reconcileExternalChanges(c client.Client, r record.EventRecorder, obj Object) error {
	patch := obj.NewPatch()
	obj.GetAnnotations()[models.ResourceStateAnnotation] = ""
	err := c.Patch(context.Background(), obj, patch)
	if err != nil {
		return fmt.Errorf("failed to automaticly handle external changes, err: %w", err)
	}

	r.Event(obj, models.Normal, models.ExternalChanges,
		"External changes were automatically reconciled",
	)

	return nil
}

func calculateAvailableNetworks(cadenceNetwork string) ([]string, error) {
	clustersCIDRs := []string{cadenceNetwork}
	for i := 0; i <= 3; i++ {
		newCIDR, err := incrementCIDR(clustersCIDRs[i])
		if err != nil {
			return nil, err
		}
		clustersCIDRs = append(clustersCIDRs, newCIDR)
	}

	return clustersCIDRs, nil
}

func incrementCIDR(cidr string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	ipParts := strings.Split(ip.String(), ".")
	secondOctet, err := strconv.Atoi(ipParts[1])
	if err != nil {
		return "", err
	}

	secondOctet++
	ipParts[1] = strconv.Itoa(secondOctet)
	prefixLength, _ := ipnet.Mask.Size()

	incrementedIP := strings.Join(ipParts, ".")
	return fmt.Sprintf("%s/%d", incrementedIP, prefixLength), nil
}

func getClusterIDByName(api instaclustr.API, appType string, name string) (string, error) {
	clusters, err := api.ListClustersByName(name)
	if err != nil {
		return "", fmt.Errorf("failed to list clusters by name, err: %w", err)
	}

	if len(clusters) == 0 {
		return "", nil
	}

	if clusters[0].Application != appType {
		return "", fmt.Errorf("the cluster %s already exists, but it has other application type %s", name, clusters[0].Application)
	}

	return clusters[0].ID, nil
}
