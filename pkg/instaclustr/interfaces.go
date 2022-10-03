package instaclustr

import (
	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/apiv1"
	"net/http"
)

type APIv2 interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetCassandraClusterStatus(id string) (*clustersv2alpha1.CassandraStatus, error)
}

type APIv1 interface {
	GetPostgreSQLCluster(url, id string) (*apiv1.PgStatus, error)
}
