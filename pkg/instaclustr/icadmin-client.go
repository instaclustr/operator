package instaclustr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"io"
	"net/http"
	"time"
)

type IcadminClient struct {
	icadminUsername string
	icadminKey      string
	serverHostname  string
	httpClient      *http.Client
}

func NewIcadminClient(
	icadminUsername,
	icadminKey,
	serverHostname string,
	timeout time.Duration,
) *IcadminClient {
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{},
	}
	return &IcadminClient{
		icadminUsername: icadminUsername,
		icadminKey:      icadminKey,
		serverHostname:  serverHostname,
		httpClient:      httpClient,
	}
}

func (c *IcadminClient) DoRequest(url string, method string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.icadminUsername, c.icadminKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Instaclustr-Source", OperatorVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *IcadminClient) GetIgnitionScript(nodeID string) (string, error) {
	url := fmt.Sprintf(IgnitionScriptEndpoint, c.serverHostname, nodeID)
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	holder := struct {
		Script string `json:"script"`
	}{}

	err = json.Unmarshal(body, &holder)
	if err != nil {
		return "", err
	}

	return holder.Script, nil
}

func (c *IcadminClient) GetGateways(cdcID string) ([]*v1beta1.Gateway, error) {
	url := fmt.Sprintf(GatewayEndpoint, c.serverHostname, cdcID)
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	holder := struct {
		Gateways []*v1beta1.Gateway `json:"gateways"`
	}{}

	err = json.Unmarshal(body, &holder)
	if err != nil {
		return nil, err
	}

	return holder.Gateways, nil
}

func (c *IcadminClient) GetOnPremisesNodes(clusterID string) ([]*v1beta1.OnPremiseNode, error) {
	url := fmt.Sprintf(NodesEndpoint, c.serverHostname, clusterID)
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	holder := struct {
		Nodes []*v1beta1.OnPremiseNode `json:"nodes"`
	}{}

	err = json.Unmarshal(body, &holder)
	if err != nil {
		return nil, err
	}

	return holder.Nodes, nil
}

func (c *IcadminClient) SetPrivateGatewayIP(gatewayID, ip string) error {
	url := fmt.Sprintf(GatewayPrivateIPEndpoint, c.serverHostname, gatewayID, ip)

	resp, err := c.DoRequest(url, http.MethodPut, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *IcadminClient) SetPublicGatewayIP(gatewayID, ip string) error {
	url := fmt.Sprintf(GatewayPublicIPEndpoint, c.serverHostname, gatewayID, ip)

	resp, err := c.DoRequest(url, http.MethodPut, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *IcadminClient) SetNodeIPs(nodeID string, request *v1beta1.OnPremiseNode) error {
	url := fmt.Sprintf(NodeIPsEndpoint, c.serverHostname, nodeID)

	holder := struct {
		Updates *v1beta1.OnPremiseNode `json:"updates"`
	}{
		Updates: request,
	}

	data, err := json.Marshal(holder)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPut, data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}
