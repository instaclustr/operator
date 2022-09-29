package instaclustr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetPostgreSQLClusterAPIv1(url, id string) (*PostgreSQLStatusAPIv1, error) {
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

	var clusterStatus *PostgreSQLStatusAPIv1
	err = json.Unmarshal(body, clusterStatus)
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}
