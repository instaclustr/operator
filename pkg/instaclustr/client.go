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
	apiv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	apiv2convertors "github.com/instaclustr/operator/pkg/instaclustr/api/v2/convertors"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
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

func (c *Client) GetClusterSpec(id, clusterEndpoint string) (*models.ClusterSpec, error) {
	url := c.serverHostname + clusterEndpoint + id + TerraformDescription

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

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	clusterSpec := &models.ClusterSpec{}

	err = json.Unmarshal(body, clusterSpec)
	if err != nil {
		return nil, err
	}

	return clusterSpec, nil
}

func (c *Client) GetRedisSpec(id, clusterEndpoint string) (*models.RedisCluster, error) {
	url := c.serverHostname + clusterEndpoint + id

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

	redisSpec := &models.RedisCluster{}

	err = json.Unmarshal(body, redisSpec)
	if err != nil {
		return nil, err
	}

	return redisSpec, nil
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
	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
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

	t.Status, err = apiv2convertors.TopicStatusFromInstAPI(body)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateKafkaTopic(url string, t *kafkamanagementv1alpha1.Topic) error {
	data, err := json.Marshal(apiv2convertors.TopicConfigsUpdateToInstAPI(t.Spec.TopicConfigs))
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

	t.Status, err = apiv2convertors.TopicStatusFromInstAPI(body)
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
	kafkaUser *modelsv2.KafkaUserAPIv2,
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
	kafkaUserID,
	kafkaUserEndpoint string,
	kafkaUserSpec *modelsv2.KafkaUserAPIv2,
) error {
	url := c.serverHostname + kafkaUserEndpoint + kafkaUserID

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
	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error {
	data, err := json.Marshal(m.Spec)
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

	err = json.Unmarshal(body, &m.Status)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetMirrorStatus(id, mirrorEndpoint string) (*kafkamanagementv1alpha1.MirrorStatus, error) {
	url := c.serverHostname + mirrorEndpoint + id

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

	clusterStatus := &kafkamanagementv1alpha1.MirrorStatus{}

	err = json.Unmarshal(body, &clusterStatus)
	if err != nil {
		return nil, err
	}

	return clusterStatus, nil
}

func (c *Client) DeleteKafkaMirror(url, id string) error {
	url = c.serverHostname + url + id

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateKafkaMirror(url string, m *kafkamanagementv1alpha1.Mirror) error {
	updateTargetLatency := &struct {
		TargetLatency int32 `json:"targetLatency"`
	}{TargetLatency: m.Spec.TargetLatency}

	data, err := json.Marshal(updateTargetLatency)
	if err != nil {
		return err
	}

	url = c.serverHostname + url + m.Status.ID

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

	err = json.Unmarshal(body, &m.Status)
	if err != nil {
		return err
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

func (c *Client) CreateExclusionWindow(url string,
	me *clusterresourcesv1alpha1.MaintenanceEventsSpec,
) (*clusterresourcesv1alpha1.MaintenanceEventsStatus,
	error,
) {
	url = c.serverHostname + url

	type exclusionWindowRequest struct {
		ClusterID       string `json:"clusterId"`
		DayOfWeek       string `json:"dayOfWeek"`
		StartHour       int32  `json:"startHour"`
		DurationInHours int32  `json:"durationInHours"`
	}

	exclusionWindow := &exclusionWindowRequest{
		ClusterID:       me.ClusterID,
		DayOfWeek:       me.DayOfWeek,
		StartHour:       me.StartHour,
		DurationInHours: me.DurationInHours,
	}

	data, err := json.Marshal(exclusionWindow)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRequest(url, http.MethodPost, data)
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

	var creationResponse *clusterresourcesv1alpha1.MaintenanceEventsStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) GetMaintenanceEventStatus(
	clusterID string,
	endpoint string,
) (*clusterresourcesv1alpha1.MaintenanceEventsStatus, error) {
	url := c.serverHostname + endpoint + "?clusterId=" + clusterID

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

	var meStatus *clusterresourcesv1alpha1.MaintenanceEventsStatus
	err = json.Unmarshal(body, &meStatus)
	if err != nil {
		return nil, err
	}

	return meStatus, nil
}

func (c *Client) GetExclusionWindowStatus(
	clusterId string,
	endpoint string,
) (*clusterresourcesv1alpha1.MaintenanceEventsStatus, error) {
	url := c.serverHostname + endpoint + "?clusterId=" + clusterId

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

	var meStatus *clusterresourcesv1alpha1.MaintenanceEventsStatus
	err = json.Unmarshal(body, &meStatus)
	if err != nil {
		return nil, err
	}

	return meStatus, nil
}

func (c *Client) DeleteExclusionWindow(
	meStatus *clusterresourcesv1alpha1.MaintenanceEventsStatus,
	endpoint string,
) error {
	url := c.serverHostname + endpoint + "/" + meStatus.ID

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

func (c *Client) UpdateMaintenanceEvent(me *clusterresourcesv1alpha1.MaintenanceEventsSpec,
	endpoint string,
) error {
	url := c.serverHostname + endpoint + me.EventID
	type request struct {
		ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	}
	requestBody := &request{
		ScheduledStartTime: me.ScheduledStartTime,
	}

	data, err := json.Marshal(requestBody)
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("me code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateNodeReload(bundle,
	nodeID string,
	nr *modelsv1.NodeReload,
) error {
	data, err := json.Marshal(nr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(c.serverHostname+NodeReloadEndpoint, bundle, nodeID)
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

func (c *Client) GetNodeReloadStatus(
	bundle,
	nodeID string,
) (*modelsv1.NodeReloadStatusAPIv1, error) {
	url := fmt.Sprintf(c.serverHostname+NodeReloadEndpoint, bundle, nodeID)

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

	var nodeReload *modelsv1.NodeReloadStatusAPIv1
	err = json.Unmarshal(body, &nodeReload)
	if err != nil {
		return nil, err
	}

	return nodeReload, nil
}

func (c *Client) RestorePgCluster(restoreData *v1alpha1.PgRestoreFrom) (string, error) {
	url := fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	var restoreRequest struct {
		ClusterNameOverride string                       `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.PgRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                        `json:"pointInTime"`
	}
	restoreRequest.CDCInfos = restoreData.CDCInfos
	restoreRequest.ClusterNameOverride = restoreData.ClusterNameOverride
	restoreRequest.PointInTime = restoreData.PointInTime

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

	var restoreRequest struct {
		ClusterNameOverride string                          `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.RedisRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                           `json:"pointInTime,omitempty"`
		IndexNames          string                          `json:"indexNames,omitempty"`
	}
	restoreRequest.CDCInfos = restoreData.CDCInfos
	restoreRequest.ClusterNameOverride = restoreData.ClusterNameOverride
	restoreRequest.PointInTime = restoreData.PointInTime
	restoreRequest.IndexNames = restoreData.IndexNames

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

	var restoreRequest struct {
		ClusterNameOverride string                               `json:"clusterNameOverride,omitempty"`
		CDCInfos            []*v1alpha1.OpenSearchRestoreCDCInfo `json:"cdcInfos,omitempty"`
		PointInTime         int64                                `json:"pointInTime,omitempty"`
		IndexNames          string                               `json:"indexNames,omitempty"`
	}
	restoreRequest.CDCInfos = restoreData.CDCInfos
	restoreRequest.ClusterNameOverride = restoreData.ClusterNameOverride
	restoreRequest.PointInTime = restoreData.PointInTime
	restoreRequest.IndexNames = restoreData.IndexNames

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

func (c *Client) GetCassandra(id, clusterEndpoint string) (*modelsv2.CassandraCluster, error) {
	url := c.serverHostname + clusterEndpoint + id

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

	cassandraSpec := &modelsv2.CassandraCluster{}
	err = json.Unmarshal(body, cassandraSpec)
	if err != nil {
		return nil, err
	}

	return cassandraSpec, nil
}

func (c *Client) RestoreCassandra(url string, restoreData v1alpha1.CassandraRestoreFrom) (string, error) {
	url = fmt.Sprintf(APIv1RestoreEndpoint, c.serverHostname, restoreData.ClusterID)

	var restoreRequest struct {
		ClusterNameOverride string                        `json:"clusterNameOverride,omitempty"`
		CDCInfos            []v1alpha1.CassandraRestoreDC `json:"cdcInfos,omitempty"`
		PointInTime         int64                         `json:"pointInTime,omitempty"`
		KeyspaceTables      string                        `json:"keyspaceTables,omitempty"`
	}
	restoreRequest.CDCInfos = restoreData.CDCInfos
	restoreRequest.ClusterNameOverride = restoreData.ClusterNameOverride
	restoreRequest.PointInTime = restoreData.PointInTime
	restoreRequest.KeyspaceTables = restoreData.KeyspaceTables

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

func (c *Client) GetPostgreSQL(id string) (*models.PGStatus, error) {
	url := c.serverHostname + PostgreSQLEndpoint + id
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

	pgCluster := &models.PGStatus{}
	err = json.Unmarshal(body, pgCluster)
	if err != nil {
		return nil, err
	}

	return pgCluster, nil
}

func (c *Client) UpdatePostgreSQLDataCentres(id string, dataCentres []*models.PGDataCentre) error {
	updateRequest := struct {
		DataCentres []*models.PGDataCentre `json:"dataCentres"`
	}{
		DataCentres: dataCentres,
	}

	reqData, err := json.Marshal(updateRequest)
	url := c.serverHostname + PostgreSQLEndpoint + id
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
	url := fmt.Sprintf(PostgreSQLConfigEndpoint, c.serverHostname, id)
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
	url := fmt.Sprintf(PostgreSQLConfigManagementEndpoint+"%s", c.serverHostname, id+"|"+name)
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
	url := fmt.Sprintf(PostgreSQLConfigManagementEndpoint, c.serverHostname)
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
	url := fmt.Sprintf(PostgreSQLConfigManagementEndpoint+"%s", c.serverHostname, id+"|"+name)
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
