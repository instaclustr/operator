package instaclustr

import (
	"net/http"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
)

type APIv2 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
}

type APIv1 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetPostgreSQLCluster(url, id string) (*clustersv1alpha1.PostgreSQLStatus, error)
}
