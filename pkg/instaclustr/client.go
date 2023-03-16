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
	kafkamanagementv1alpha1 "github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
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

func (c *Client) GetOpenSearch(id string) ([]byte, error) {
	url := c.serverHostname + ClustersEndpointV1 + id + TerraformDescription

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

	return body, nil
}

func (c *Client) GetRedis(id string) ([]byte, error) {
	url := c.serverHostname + RedisEndpoint + id

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

	return body, nil
}

func (c *Client) CreateRedisUser(user *models.RedisUser) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	url := c.serverHostname + RedisUserEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
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

	response := &struct {
		ID string `json:"id"`
	}{}

	err = json.Unmarshal(body, response)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func (c *Client) UpdateRedisUser(user *models.RedisUserUpdate) error {
	url := c.serverHostname + RedisUserEndpoint + user.ID
	data, err := json.Marshal(user)
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

func (c *Client) DeleteRedisUser(id string) error {
	url := c.serverHostname + RedisUserEndpoint + id
	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetCassandra(id string) ([]byte, error) {
	url := c.serverHostname + CassandraEndpoint + id

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

	return body, nil
}

func (c *Client) UpdateCassandra(id string, cassandra models.CassandraClusterAPIUpdate) error {
	url := c.serverHostname + CassandraEndpoint + id
	data, err := json.Marshal(cassandra)
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

func (c *Client) GetKafka(id string) ([]byte, error) {
	url := c.serverHostname + KafkaEndpoint + id

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

	return body, nil
}

func (c *Client) GetKafkaConnect(id string) ([]byte, error) {
	url := c.serverHostname + KafkaConnectEndpoint + id

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

	return body, nil
}

func (c *Client) UpdateKafkaConnect(id string, kc models.KafkaConnectAPIUpdate) error {
	url := c.serverHostname + KafkaConnectEndpoint + id
	data, err := json.Marshal(kc)
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

func (c *Client) GetZookeeper(id string) ([]byte, error) {
	url := c.serverHostname + ZookeeperEndpoint + id

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

	return body, nil
}

func (c *Client) UpdateNodeSize(clusterEndpoint string, resizeRequest *models.ResizeRequest) error {
	var url string
	if clusterEndpoint == ClustersEndpointV1 {
		url = fmt.Sprintf(ClustersResizeEndpoint, c.serverHostname, resizeRequest.ClusterID, resizeRequest.DataCentreID)
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
		clusterModifyRequest.TwoFactorDelete = &models.TwoFactorDeleteV1{
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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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

func (c *Client) GetPeeringStatus(peerID,
	peeringEndpoint string,
) (*clusterresourcesv1alpha1.PeeringStatus, error) {
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

	var peeringStatus clusterresourcesv1alpha1.PeeringStatus
	err = json.Unmarshal(body, &peeringStatus)
	if err != nil {
		return nil, err
	}

	return &peeringStatus, nil
}

func (c *Client) CreatePeering(url string, peeringSpec any) (*clusterresourcesv1alpha1.PeeringStatus, error) {

	jsonDataCreate, err := json.Marshal(peeringSpec)
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

	var creationResponse *clusterresourcesv1alpha1.PeeringStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) UpdatePeering(peerID,
	peeringEndpoint string,
	peerSpec any,
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) DeletePeering(peerID, peeringEndpoint string) error {
	url := c.serverHostname + peeringEndpoint + peerID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetFirewallRuleStatus(
	firewallRuleID string,
	firewallRuleEndpoint string,
) (*clusterresourcesv1alpha1.FirewallRuleStatus, error) {
	url := c.serverHostname + firewallRuleEndpoint + firewallRuleID

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

	var firewallRuleStatus *clusterresourcesv1alpha1.FirewallRuleStatus
	err = json.Unmarshal(body, &firewallRuleStatus)
	if err != nil {
		return nil, err
	}

	return firewallRuleStatus, nil
}

func (c *Client) CreateFirewallRule(
	url string,
	firewallRuleSpec any,
) (*clusterresourcesv1alpha1.FirewallRuleStatus, error) {
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

	var creationResponse *clusterresourcesv1alpha1.FirewallRuleStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) DeleteFirewallRule(
	firewallRuleID string,
	firewallRuleEndpoint string,
) error {
	url := c.serverHostname + firewallRuleEndpoint + firewallRuleID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetTopicStatus(id string) ([]byte, error) {
	url := c.serverHostname + KafkaTopicEndpoint + id

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

	return body, nil
}

func (c *Client) CreateKafkaTopic(url string, t *kafkamanagementv1alpha1.Topic) error {
	data, err := json.Marshal(t.Spec)
	if err != nil {
		return err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, data)
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

	t.Status, err = t.Status.FromInstAPI(body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteKafkaTopic(url, id string) error {
	url = c.serverHostname + url + id

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

func (c *Client) UpdateKafkaTopic(url string, t *kafkamanagementv1alpha1.Topic) error {
	data, err := json.Marshal(t.Spec.TopicConfigsUpdateToInstAPI())
	if err != nil {
		return err
	}

	url = c.serverHostname + url

	resp, err := c.DoRequest(url, http.MethodPut, data)
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

	t.Status, err = t.Status.FromInstAPI(body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetKafkaUserStatus(
	kafkaUserID,
	kafkaUserEndpoint string,
) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	url := c.serverHostname + kafkaUserEndpoint + kafkaUserID

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

	var kafkaUserStatus kafkamanagementv1alpha1.KafkaUserStatus
	err = json.Unmarshal(body, &kafkaUserStatus)
	if err != nil {
		return nil, err
	}

	return &kafkaUserStatus, nil
}

func (c *Client) CreateKafkaUser(
	url string,
	kafkaUser *models.KafkaUser,
) (*kafkamanagementv1alpha1.KafkaUserStatus, error) {
	jsonKafkaUser, err := json.Marshal(kafkaUser)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonKafkaUser)
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

	var creationResponse *kafkamanagementv1alpha1.KafkaUserStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) UpdateKafkaUser(
	kafkaUserID string,
	kafkaUserSpec *models.KafkaUser,
) error {
	url := c.serverHostname + KafkaUserEndpoint + kafkaUserID

	data, err := json.Marshal(kafkaUserSpec)
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) DeleteKafkaUser(kafkaUserID, kafkaUserEndpoint string) error {
	url := c.serverHostname + kafkaUserEndpoint + kafkaUserID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateKafkaMirror(m *kafkamanagementv1alpha1.MirrorSpec) (*kafkamanagementv1alpha1.MirrorStatus, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + KafkaMirrorEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
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

	status := &kafkamanagementv1alpha1.MirrorStatus{}

	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (c *Client) GetMirrorStatus(id string) (*kafkamanagementv1alpha1.MirrorStatus, error) {
	url := c.serverHostname + KafkaMirrorEndpoint + id

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

	m := &kafkamanagementv1alpha1.MirrorStatus{}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Client) DeleteKafkaMirror(id string) error {
	url := c.serverHostname + KafkaMirrorEndpoint + id

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

func (c *Client) UpdateKafkaMirror(id string, latency int32) error {
	updateTargetLatency := &struct {
		TargetLatency int32 `json:"targetLatency"`
	}{TargetLatency: latency}

	data, err := json.Marshal(updateTargetLatency)
	if err != nil {
		return err
	}

	url := c.serverHostname + KafkaMirrorEndpoint + id

	resp, err := c.DoRequest(url, http.MethodPut, data)
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

func (c *Client) GetClusterBackups(endpoint, clusterID string) (*models.ClusterBackup, error) {
	url := c.serverHostname + endpoint + clusterID + BackupsEndpoint

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

	clusterBackups := &models.ClusterBackup{}
	err = json.Unmarshal(body, clusterBackups)
	if err != nil {
		return nil, err
	}

	return clusterBackups, nil
}

func (c *Client) TriggerClusterBackup(url, clusterID string) error {
	url = c.serverHostname + url + clusterID + BackupsEndpoint

	resp, err := c.DoRequest(url, http.MethodPost, nil)
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

func (c *Client) CreateExclusionWindow(clusterID string, window clusterresourcesv1alpha1.ExclusionWindowSpec) (string, error) {
	req := &struct {
		ClusterID       string `json:"clusterId"`
		DayOfWeek       string `json:"dayOfWeek"`
		StartHour       int32  `json:"startHour"`
		DurationInHours int32  `json:"durationInHours"`
	}{
		ClusterID:       clusterID,
		DayOfWeek:       window.DayOfWeek,
		StartHour:       window.StartHour,
		DurationInHours: window.DurationInHours,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := c.serverHostname + ExclusionWindowEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	response := &struct {
		ID string `json:"id"`
	}{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func (c *Client) GetMaintenanceEventsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.MaintenanceEventStatus, error) {
	url := c.serverHostname + MaintenanceEventStatusEndpoint + clusterID

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

	statuses := &struct {
		MaintenanceEvents []*clusterresourcesv1alpha1.MaintenanceEventStatus `json:"maintenanceEvents"`
	}{}
	err = json.Unmarshal(body, statuses)
	if err != nil {
		return nil, err
	}

	return statuses.MaintenanceEvents, nil
}

func (c *Client) GetMaintenanceEvents(clusterID string) ([]*v1alpha1.MaintenanceEvent, error) {
	url := c.serverHostname + MaintenanceEventStatusEndpoint + clusterID

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

	status := &struct {
		MaintenanceEvents []*v1alpha1.MaintenanceEvent `json:"maintenanceEvents"`
	}{}
	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, err
	}

	return status.MaintenanceEvents, nil
}

func (c *Client) GetExclusionWindowsStatuses(clusterID string) ([]*clusterresourcesv1alpha1.ExclusionWindowStatus, error) {
	url := c.serverHostname + ExclusionWindowStatusEndpoint + clusterID

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

	status := &struct {
		ExclusionWindows []*clusterresourcesv1alpha1.ExclusionWindowStatus `json:"exclusionWindows"`
	}{}
	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, err
	}

	return status.ExclusionWindows, nil
}

func (c *Client) DeleteExclusionWindow(id string) error {
	url := c.serverHostname + ExclusionWindowEndpoint + id

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateMaintenanceEvent(me clusterresourcesv1alpha1.MaintenanceEventRescheduleSpec) (*clusterresourcesv1alpha1.MaintenanceEventStatus, error) {
	url := c.serverHostname + MaintenanceEventEndpoint + me.ScheduleID

	requestBody := &struct {
		ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	}{
		ScheduledStartTime: me.ScheduledStartTime,
	}

	data, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRequest(url, http.MethodPut, data)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("me code: %d, message: %s", resp.StatusCode, body)
	}

	response := &clusterresourcesv1alpha1.MaintenanceEventStatus{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) CreateNodeReload(nr *clusterresourcesv1alpha1.Node) error {
	url := fmt.Sprintf(NodeReloadEndpoint, c.serverHostname, nr.ID)

	data, err := json.Marshal(nr)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPost, data)
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

func (c *Client) GetNodeReloadStatus(nodeID string) (*models.NodeReloadStatus, error) {
	url := fmt.Sprintf(NodeReloadEndpoint, c.serverHostname, nodeID)

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

	var nodeReload *models.NodeReloadStatus
	err = json.Unmarshal(body, &nodeReload)
	if err != nil {
		return nil, err
	}

	return nodeReload, nil
}

func (c *Client) RestorePgCluster(restoreData *v1alpha1.PgRestoreFrom) (string, error) {
	url := fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	restoreRequest := struct {
		ClusterNameOverride string                       `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.PgRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                        `json:"pointInTime,omitempty"`
		ClusterNetwork      string                       `json:"clusterNetwork,omitempty"`
	}{
		ClusterNameOverride: restoreData.ClusterNameOverride,
		CDCInfos:            restoreData.CDCInfos,
		PointInTime:         restoreData.PointInTime,
		ClusterNetwork:      restoreData.ClusterNetwork,
	}

	jsonData, err := json.Marshal(restoreRequest)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		RestoredCluster string `json:"restoredCluster"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return response.RestoredCluster, nil
}

func (c *Client) RestoreRedisCluster(restoreData *v1alpha1.RedisRestoreFrom) (string, error) {
	url := fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	restoreRequest := struct {
		ClusterNameOverride string                          `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.RedisRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                           `json:"pointInTime,omitempty"`
		IndexNames          string                          `json:"indexNames,omitempty"`
		ClusterNetwork      string                          `json:"clusterNetwork,omitempty"`
	}{
		CDCInfos:            restoreData.CDCInfos,
		ClusterNameOverride: restoreData.ClusterNameOverride,
		PointInTime:         restoreData.PointInTime,
		IndexNames:          restoreData.IndexNames,
		ClusterNetwork:      restoreData.ClusterNetwork,
	}

	jsonData, err := json.Marshal(restoreRequest)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		RestoredCluster string `json:"restoredCluster"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return response.RestoredCluster, nil
}

func (c *Client) RestoreOpenSearchCluster(restoreData *v1alpha1.OpenSearchRestoreFrom) (string, error) {
	url := fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	restoreRequest := struct {
		ClusterNameOverride string                               `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.OpenSearchRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                                `json:"pointInTime,omitempty"`
		IndexNames          string                               `json:"indexNames,omitempty"`
		ClusterNetwork      string                               `json:"clusterNetwork,omitempty"`
	}{
		ClusterNameOverride: restoreData.ClusterNameOverride,
		CDCInfos:            restoreData.CDCInfos,
		PointInTime:         restoreData.PointInTime,
		IndexNames:          restoreData.IndexNames,
		ClusterNetwork:      restoreData.ClusterNetwork,
	}

	jsonData, err := json.Marshal(restoreRequest)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		RestoredCluster string `json:"restoredCluster"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return response.RestoredCluster, nil
}

func (c *Client) CreateKafkaACL(
	url string,
	kafkaACL *kafkamanagementv1alpha1.KafkaACLSpec,
) (*kafkamanagementv1alpha1.KafkaACLStatus, error) {
	jsonKafkaACLUser, err := json.Marshal(kafkaACL)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonKafkaACLUser)
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

	var creationResponse *kafkamanagementv1alpha1.KafkaACLStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) GetKafkaACLStatus(
	kafkaACLID,
	kafkaACLEndpoint string,
) (*kafkamanagementv1alpha1.KafkaACLStatus, error) {
	url := c.serverHostname + kafkaACLEndpoint + kafkaACLID

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

	var kafkaACLStatus kafkamanagementv1alpha1.KafkaACLStatus
	err = json.Unmarshal(body, &kafkaACLStatus)
	if err != nil {
		return nil, err
	}

	return &kafkaACLStatus, nil
}

func (c *Client) DeleteKafkaACL(
	kafkaACLID,
	kafkaACLEndpoint string,
) error {
	url := c.serverHostname + kafkaACLEndpoint + kafkaACLID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateKafkaACL(
	kafkaACLID,
	kafkaACLEndpoint string,
	kafkaACLSpec any,
) error {
	url := c.serverHostname + kafkaACLEndpoint + kafkaACLID

	data, err := json.Marshal(kafkaACLSpec)
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) RestoreCassandra(restoreData v1alpha1.CassandraRestoreFrom) (string, error) {
	url := fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	restoreRequest := struct {
		ClusterNameOverride string                        `json:"clusterNameOverride,omitempty"`
		CDCInfos            []v1alpha1.CassandraRestoreDC `json:"cdcInfos,omitempty"`
		PointInTime         int64                         `json:"pointInTime,omitempty"`
		KeyspaceTables      string                        `json:"keyspaceTables,omitempty"`
		ClusterNetwork      string                        `json:"clusterNetwork,omitempty"`
	}{
		CDCInfos:            restoreData.CDCInfos,
		ClusterNameOverride: restoreData.ClusterNameOverride,
		PointInTime:         restoreData.PointInTime,
		KeyspaceTables:      restoreData.KeyspaceTables,
		ClusterNetwork:      restoreData.ClusterNetwork,
	}

	jsonData, err := json.Marshal(restoreRequest)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		RestoredCluster string `json:"restoredCluster"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return response.RestoredCluster, nil
}

func (c *Client) GetPostgreSQL(id string) ([]byte, error) {
	url := c.serverHostname + PGSQLEndpoint + id
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

	return body, nil
}

func (c *Client) UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error {
	updateRequest := struct {
		DataCentres []*models.PGDataCentre `json:"dataCentres"`
	}{
		DataCentres: dataCentres,
	}

	reqData, err := json.Marshal(updateRequest)
	if err != nil {
		return err
	}
	url := c.serverHostname + PGSQLEndpoint + id
	resp, err := c.DoRequest(url, http.MethodPut, reqData)
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

func (c *Client) GetPostgreSQLConfigs(id string) ([]*models.PGConfigs, error) {
	url := fmt.Sprintf(PGSQLConfigEndpoint, c.serverHostname, id)
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

	response := []*models.PGConfigs{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) UpdatePostgreSQLConfiguration(id, name, value string) error {
	request := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}

	reqData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(PGSQLConfigManagementEndpoint+"%s", c.serverHostname, id+"|"+name)
	resp, err := c.DoRequest(url, http.MethodPut, reqData)
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

func (c *Client) CreatePostgreSQLConfiguration(id, name, value string) error {
	request := struct {
		Name      string `json:"name"`
		ClusterID string `json:"clusterId"`
		Value     string `json:"value"`
	}{
		Name:      name,
		ClusterID: id,
		Value:     value,
	}

	reqData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(PGSQLConfigManagementEndpoint, c.serverHostname)
	resp, err := c.DoRequest(url, http.MethodPost, reqData)
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

func (c *Client) ResetPostgreSQLConfiguration(id, name string) error {
	url := fmt.Sprintf(PGSQLConfigManagementEndpoint+"%s", c.serverHostname, id+"|"+name)
	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetCadence(id string) ([]byte, error) {
	url := c.serverHostname + CadenceEndpoint + id
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

	return body, nil
}

func (c *Client) UpdatePostgreSQLDefaultUserPassword(id, password string) error {
	request := struct {
		Password string `json:"password"`
	}{
		Password: password,
	}

	reqData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(PGSQLUpdateDefaultUserPasswordEndpoint, c.serverHostname, id)
	resp, err := c.DoRequest(url, http.MethodPut, reqData)
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

func (c *Client) ListClusters() ([]*models.ActiveClusters, error) {
	url := c.serverHostname + ClustersEndpoint
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

	response := []*models.ActiveClusters{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) CreateEncryptionKey(
	encryptionKeySpec any,
) (*clusterresourcesv1alpha1.AWSEncryptionKeyStatus, error) {
	jsonEncryptionKey, err := json.Marshal(encryptionKeySpec)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + AWSEncryptionKeyEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonEncryptionKey)
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

	var creationResponse *clusterresourcesv1alpha1.AWSEncryptionKeyStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) GetEncryptionKeyStatus(
	encryptionKeyID string,
	encryptionKeyEndpoint string,
) (*clusterresourcesv1alpha1.AWSEncryptionKeyStatus, error) {
	url := c.serverHostname + encryptionKeyEndpoint + encryptionKeyID

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

	var encryptionKeyStatus *clusterresourcesv1alpha1.AWSEncryptionKeyStatus
	err = json.Unmarshal(body, &encryptionKeyStatus)
	if err != nil {
		return nil, err
	}

	return encryptionKeyStatus, nil
}

func (c *Client) DeleteEncryptionKey(
	encryptionKeyID string,
) error {
	url := c.serverHostname + AWSEncryptionKeyEndpoint + encryptionKeyID

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
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

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}
