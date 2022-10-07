package instaclustr

import (
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"net/http"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
)

type API interface {
	DoRequest(url string, method string, data []byte) (*http.Response, error)
	CreateCluster(url string, clusterSpec any) (string, error)
	GetClusterStatus(id, clusterEndpoint string) (*v1alpha1.ClusterStatus, error)
	GetCassandraDCs(id, clusterEndpoint string) (*modelsv2.CassandraDCs, error)
}
