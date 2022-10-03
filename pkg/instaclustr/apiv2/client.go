package apiv2

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
)

type ClientV2 struct {
	Client *instaclustr.Client
}

func NewClientV2(client *instaclustr.Client) *ClientV2 {
	return &ClientV2{
		Client: client,
	}
}

func (c *ClientV2) GetCassandraClusterStatus(id string) (*clustersv2alpha1.CassandraStatus, error) {

	url := c.Client.ServerHostname + instaclustr.CassandraEndpoint + id
	resp, err := c.Client.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var clusterStatus *clustersv2alpha1.CassandraStatus
	err = json.Unmarshal(body, &clusterStatus)
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}
