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

package clusterresources

import (
	"strings"

	k8sCore "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

func areFirewallRuleStatusesEqual(a, b *v1beta1.FirewallRuleStatus) bool {
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

func arePeeringStatusesEqual(a, b *v1beta1.PeeringStatus) bool {
	if a.ID != b.ID ||
		a.Name != b.Name ||
		a.StatusCode != b.StatusCode ||
		a.FailureReason != b.FailureReason {
		return false
	}

	return true
}

func areEncryptionKeyStatusesEqual(a, b *v1beta1.AWSEncryptionKeyStatus) bool {
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

func CheckIfUserExistsOnInstaclustrAPI(username, clusterID, app string, api instaclustr.API) (bool, error) {
	users, err := api.FetchUsers(clusterID, app)
	if err != nil {
		return false, err
	}

	return slices.Contains(users, username), nil
}

func getUserCreds(secret *k8sCore.Secret) (username, password string, err error) {
	password = string(secret.Data[models.Password])
	username = string(secret.Data[models.Username])

	if len(username) == 0 || len(password) == 0 {
		return "", "", models.ErrMissingSecretKeys
	}

	newLineSuffix := "\n"

	if strings.HasSuffix(username, newLineSuffix) {
		username = strings.TrimRight(username, newLineSuffix)
	}

	if strings.HasSuffix(password, newLineSuffix) {
		password = strings.TrimRight(password, newLineSuffix)
	}

	return username, password, nil
}

func subnetsEqual(subnets1, subnets2 []string) bool {
	if len(subnets1) != len(subnets2) {
		return false
	}

	for _, s1 := range subnets1 {
		var equal bool
		for _, s2 := range subnets2 {
			if s1 == s2 {
				equal = true
			}
		}

		if !equal {
			return false
		}
	}

	return true
}
