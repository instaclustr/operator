package validation

import "github.com/instaclustr/operator/pkg/models"

type Validation interface {
	ListAppVersions(app string) ([]*models.AppVersions, error)
}
