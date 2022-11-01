package instaclustr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	clusterresourcesv1alpha1 "github.com/instaclustr/operator/apis/clusterresources/v1alpha1"
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	apiv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	apiv2convertors "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	"github.com/instaclustr/operator/pkg/models"
)

type Client struct {
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
) *Client {
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{},
	}
	return &Client{
		username:       username,
		key:            key,
		serverHostname: serverHostname,
		httpClient:     httpClient,
	}
}

func (c *Client) DoRequest(url string, method string, data []byte) (*http.Response, error) {
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

func (c *Client) CreateCluster(url string, cluster any) (string, error) {

	jsonDataCreate, err := json.Marshal(cluster)
	if err != nil {
		return "", err
	}

	url = c.serverHostname + url
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

func (c *Client) GetClusterStatus(id, clusterEndpoint string) (*v1alpha1.ClusterStatus, error) {
	url := c.serverHostname + clusterEndpoint + id
	if clusterEndpoint == ClustersEndpointV1 {
		url += TerraformDescription
	}

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	var clusterStatus *v1alpha1.ClusterStatus
	if clusterEndpoint == ClustersEndpointV1 {
		if resp.StatusCode != http.StatusAccepted {
			return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
		}

		clusterStatus, err = apiv1.ClusterStatusFromInstAPI(body)
	} else {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
		}

		clusterStatus, err = apiv2convertors.ClusterStatusFromInstAPI(body)
	}
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}

func (c *Client) UpdateNodeSize(
	clusterEndpoint,
	clusterID string,
	resizedDataCentre *v1alpha1.ResizedDataCentre,
	concurrentResizes int,
	notifySupportContacts bool,
	nodePurpose string,
) error {
	var url string
	if clusterEndpoint == ClustersEndpointV1 {
		url = fmt.Sprintf(ClustersResizeEndpoint, c.serverHostname, clusterID, resizedDataCentre.DataCentreID)
	}

	resizeRequest := &models.ResizeRequest{
		NewNodeSize:           resizedDataCentre.NewNodeSize,
		ConcurrentResizes:     concurrentResizes,
		NotifySupportContacts: notifySupportContacts,
		NodePurpose:           nodePurpose,
	}

	resizePayload, err := json.Marshal(resizeRequest)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPost, resizePayload)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		return StatusPreconditionFailed
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetActiveDataCentreResizeOperations(clusterID, dataCentreID string) ([]*models.DataCentreResizeOperations, error) {
	var dcResizeOperations []*models.DataCentreResizeOperations

	url := fmt.Sprintf(ClustersResizeEndpoint+"?%s", c.serverHostname, clusterID, dataCentreID, ActiveOnly)

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

	err = json.Unmarshal(body, &dcResizeOperations)
	if err != nil {
		return nil, err
	}

	return dcResizeOperations, nil
}

func (c *Client) GetClusterConfigurations(clusterEndpoint, clusterID, bundle string) (map[string]string, error) {
	var instClusterConfigurations []*models.ClusterConfigurations

	url := c.serverHostname + clusterEndpoint + clusterID + "/" + bundle + ClusterConfigurationsEndpoint

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

	err = json.Unmarshal(body, &instClusterConfigurations)
	if err != nil {
		return nil, err
	}

	var clusterConfigurations = make(map[string]string)
	if len(instClusterConfigurations) > 0 {
		for _, clusterConfiguration := range instClusterConfigurations {
			clusterConfigurations[clusterConfiguration.ParameterName] = clusterConfiguration.ParameterValue
		}
	}

	return clusterConfigurations, err
}

func (c *Client) UpdateClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName, paramValue string) error {
	clusterConfigurationsToUpdate := &models.ClusterConfigurations{
		ParameterName:  paramName,
		ParameterValue: paramValue,
	}

	url := c.serverHostname + clusterEndpoint + clusterID + "/" + bundle + ClusterConfigurationsEndpoint

	clusterConfigurationsPayload, err := json.Marshal(clusterConfigurationsToUpdate)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPut, clusterConfigurationsPayload)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) ResetClusterConfiguration(clusterEndpoint, clusterID, bundle, paramName string) error {
	url := c.serverHostname + clusterEndpoint + clusterID + "/" + bundle + ClusterConfigurationsEndpoint + ClusterConfigurationsParameterEndpoint + paramName

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1alpha1.TwoFactorDelete) error {
	url := c.serverHostname + clusterEndpoint + clusterID

	clusterModifyRequest := &models.ClusterModifyRequest{
		Description: description,
	}

	if twoFactorDelete != nil {
		clusterModifyRequest.TwoFactorDelete = &models.TwoFactorDelete{
			DeleteVerifyEmail: twoFactorDelete.Email,
			DeleteVerifyPhone: twoFactorDelete.Phone,
		}
	}

	clusterModifyPayload, err := json.Marshal(clusterModifyRequest)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPut, clusterModifyPayload)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateCluster(id, clusterEndpoint string, InstaDCs any) error {
	url := c.serverHostname + clusterEndpoint + id
	data, err := json.Marshal(InstaDCs)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPut, data)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ClusterIsNotReadyToResize
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) DeleteCluster(id, clusterEndpoint string) error {
	url := c.serverHostname + clusterEndpoint + id

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) AddDataCentre(id, clusterEndpoint string, dataCentre any) error {
	url := c.serverHostname + clusterEndpoint + id + AddDataCentresEndpoint

	dataCentrePayload, err := json.Marshal(dataCentre)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPost, dataCentrePayload)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetAWSPeeringStatus(peerID,
	peeringEndpoint string,
) (*clusterresourcesv1alpha1.AWSVPCPeeringStatus, error) {
	url := c.serverHostname + peeringEndpoint + peerID

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var peeringStatus clusterresourcesv1alpha1.AWSVPCPeeringStatus
	err = json.Unmarshal(body, &peeringStatus)
	if err != nil {
		return nil, err
	}

	return &peeringStatus, nil
}

func (c *Client) CreateAWSPeering(url string,
	AWSSpec *clusterresourcesv1alpha1.AWSVPCPeeringSpec,
) (*clusterresourcesv1alpha1.AWSVPCPeeringStatus, error) {

	jsonDataCreate, err := json.Marshal(AWSSpec)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
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

	var creationResponse *clusterresourcesv1alpha1.AWSVPCPeeringStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) UpdateAWSPeering(peerID,
	peeringEndpoint string,
	peerSpec *clusterresourcesv1alpha1.AWSVPCPeeringSpec,
) error {
	url := c.serverHostname + peeringEndpoint + peerID

	data, err := json.Marshal(peerSpec)
	if err != nil {
		return err
	}
	resp, err := c.DoRequest(url, http.MethodPut, data)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) DeleteAWSPeering(peerID, peeringEndpoint string) error {
	url := c.serverHostname + peeringEndpoint + peerID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetFirewallRuleStatus(
	firewallRuleId string,
) (*clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus, error) {
	url := c.serverHostname + FirewallRuleEndpoint + firewallRuleId

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var firewallRuleStatus clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus
	err = json.Unmarshal(body, &firewallRuleStatus)
	if err != nil {
		return nil, err
	}

	return &firewallRuleStatus, nil
}

func (c *Client) CreateFirewallRule(
	url string,
	firewallRuleSpec *clusterresourcesv1alpha1.ClusterNetworkFirewallRuleSpec,
) (*clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus, error) {
	jsonFirewallRule, err := json.Marshal(firewallRuleSpec)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonFirewallRule)
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

	var creationResponse *clusterresourcesv1alpha1.ClusterNetworkFirewallRuleStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}
