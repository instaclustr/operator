package APIv1

import (
	"encoding/json"
	"fmt"
	"github.com/instaclustr/operator/pkg/instaclustr"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/APIv1/models"
	"io"
	"net/http"
)

func (c *instaclustr.Client) GetPostgreSQLCluster(url, id string) (*modelsv1.PgStatus, error) {
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

	var clusterStatus modelsv1.PgStatus
	err = json.Unmarshal(body, &clusterStatus)
	if err != nil {
		return nil, err
	}

	return &clusterStatus, nil
}
