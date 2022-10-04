package instaclustr

import (
	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	v1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1"
	"net/http"
)

type ClientInterface interface {
	V1() V1Interface
	V2() V2Interface
}

type V1Interface interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetPostgreSQLCluster(url, id string) (*v1.PgStatus, error)
}

type V2Interface interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetCassandraClusterStatus(id string) (*clustersv2alpha1.CassandraStatus, error)
}
