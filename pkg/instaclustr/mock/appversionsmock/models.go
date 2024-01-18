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
	return []*models.AppVersions{{Application: app, Versions: []string{"1.0.0"}}}, nil
}
