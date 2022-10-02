package instaclustr

import (
	"net/http"

	"github.com/instaclustr/operator/pkg/instaclustr/models"
)

type APIv2 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
}

type APIv1 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetPostgreSQLCluster(url, id string) (*models.PgStatusAPIv1, error)
}
