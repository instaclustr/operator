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
