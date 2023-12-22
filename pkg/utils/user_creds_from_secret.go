/*
Copyright 2023.

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

package utils

import (
	"strings"

	k8sCore "k8s.io/api/core/v1"

	"github.com/instaclustr/operator/pkg/models"
)

func GetUserCreds(secret *k8sCore.Secret) (username, password string, err error) {
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
