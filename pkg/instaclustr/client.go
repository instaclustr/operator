package instaclustr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type client struct {
	username       string
	key            string
	serverHostname string
	httpClient     *http.Client
}

func NewClient(
	username string,
	key string,
	serverHostname string,
	timeout time.Duration,
) *client {
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{},
	}
	return &client{
		username:       username,
		key:            key,
		serverHostname: serverHostname,
		httpClient:     httpClient,
	}
}

func (c *client) DoRequest(url string, method string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Instaclustr-Source", OperatorVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *client) CreateCluster(url string, cluster any) (string, error) {

	jsonDataCreate, err := json.Marshal(cluster)
	if err != nil {
		return "", err
	}

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
