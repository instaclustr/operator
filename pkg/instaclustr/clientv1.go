package instaclustr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	clustersv1alpha1 "github.com/instaclustr/operator/apis/clusters/v1alpha1"
)

func (c *Client) GetPostgreSQLCluster(url, id string) (*clustersv1alpha1.PostgreSQLStatus, error) {
	url = c.serverHostname + url + id
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var clusterStatus *clustersv1alpha1.PostgreSQLStatus
	err = json.Unmarshal(body, clusterStatus)
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}
