package apiv1

import (
	"encoding/json"
	"fmt"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"io"
	"net/http"
)

type ClientAPIv1 struct {
	Client *instaclustr.Client
}

func NewClientV1(
	client *instaclustr.Client,
) *ClientAPIv1 {
	return &ClientAPIv1{
		Client: client,
	}
}

func (c *ClientAPIv1) GetPostgreSQLCluster(url, id string) (*PgStatus, error) {
	url = c.Client.ServerHostname + url + id
	resp, err := c.Client.DoRequest(url, http.MethodGet, nil)
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

	var clusterStatus *PgStatus
	err = json.Unmarshal(body, &clusterStatus)
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}
