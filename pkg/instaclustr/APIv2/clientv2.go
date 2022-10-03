package APIv2

import (
	"encoding/json"
	"fmt"
	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/APIv2/models"
	"io"
	"net/http"
)

func (c *instaclustr.Client) GetCassandraClusterStatus(id string) (*clustersv2alpha1.CassandraStatus, error) {

	url := c.serverHostname + modelsv2.CassandraEndpoint + id
	resp, err := c.DoRequest(url, http.MethodGet, nil)
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
