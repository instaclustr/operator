package instaclustr

import (
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/APIv1/models"
	"net/http"
)

type APIv2 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
}

type APIv1 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetPostgreSQLCluster(url, id string) (*modelsv1.PgStatus, error)
}
