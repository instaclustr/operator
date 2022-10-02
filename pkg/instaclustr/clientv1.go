package instaclustr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/instaclustr/operator/pkg/instaclustr/models"
)

func (c *Client) GetPostgreSQLCluster(url, id string) (*models.PgStatusAPIv1, error) {
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

	var clusterStatus models.PgStatusAPIv1
	err = json.Unmarshal(body, &clusterStatus)
	if err != nil {
		return nil, err
	}

	return &clusterStatus, nil
}
