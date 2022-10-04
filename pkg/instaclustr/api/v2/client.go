package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	clustersv2alpha1 "github.com/instaclustr/operator/apis/clusters/v2alpha1"
	"io"
	"net/http"
)

type Client struct {
	Username        string
	Key             string
	ServerHostname  string
	HttpClient      *http.Client
	OperatorVersion string
}

func (c Client) DoRequest(url string, method string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Instaclustr-Source", c.OperatorVersion)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c Client) CreateCluster(url string, cluster any) (string, error) {
	jsonDataCreate, err := json.Marshal(cluster)
	if err != nil {
		return "", err
	}

	url = c.ServerHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse struct {
		ID string `json:"id"`
	}

	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return "", err
	}

	return creationResponse.ID, nil
}

func (c Client) GetCassandraClusterStatus(id string) (*clustersv2alpha1.CassandraStatus, error) {
	url := c.ServerHostname + CassandraEndpoint + id
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
