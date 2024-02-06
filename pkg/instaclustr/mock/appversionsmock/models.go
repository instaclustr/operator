package appversionsmock

import (
	"net/http"

	"github.com/instaclustr/operator/pkg/models"
)

type mockClient struct {
	*http.Client
}

func NewInstAPI() *mockClient {
	return &mockClient{}
}

func (c *mockClient) ListAppVersions(app string) ([]*models.AppVersions, error) {
	versions := []string{"1.0.0"}

	switch app {
	case models.OpenSearchAppKind:
		return []*models.AppVersions{{Application: models.OpenSearchAppType, Versions: versions}}, nil
	case models.KafkaAppKind:
		return []*models.AppVersions{{Application: models.KafkaAppType, Versions: versions}}, nil
	}

	return []*models.AppVersions{{Application: app, Versions: versions}}, nil
}
