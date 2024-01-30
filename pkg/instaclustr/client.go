/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instaclustr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	kafkamanagementv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

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

func (c *Client) CreateClusterRaw(url string, cluster any) ([]byte, error) {
	jsonDataCreate, err := json.Marshal(cluster)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return body, nil
}

func (c *Client) GetOpenSearch(id string) (*models.OpenSearchCluster, error) {
	url := c.serverHostname + OpenSearchEndpoint + id

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

	var model models.OpenSearchCluster
	err = json.Unmarshal(body, &model)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to model, err: %w", err)
	}

	return &model, nil
}

func (c *Client) UpdateOpenSearch(id string, o models.OpenSearchInstAPIUpdateRequest) error {
	url := c.serverHostname + OpenSearchEndpoint + id
	data, err := json.Marshal(o)
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetRedis(id string) ([]byte, error) {
	url := c.serverHostname + RedisEndpoint + id

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

	return body, nil
}

func (c *Client) GetRedisUser(id string) (*models.RedisUser, error) {
	url := c.serverHostname + RedisUserEndpoint + id

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

	userRedis := &models.RedisUser{}
	err = json.Unmarshal(body, userRedis)
	if err != nil {
		return nil, err
	}

	return userRedis, nil
}

func (c *Client) UpdateRedis(id string, r *models.RedisDataCentreUpdate) error {
	url := c.serverHostname + RedisEndpoint + id
	data, err := json.Marshal(r)
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	return body, nil
}

func (c *Client) GetKafkaConnect(id string) ([]byte, error) {
	url := c.serverHostname + KafkaConnectEndpoint + id

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
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

	return body, nil
}

func (c *Client) UpdateDescriptionAndTwoFactorDelete(clusterEndpoint, clusterID, description string, twoFactorDelete *v1beta1.TwoFactorDelete) error {
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetPeeringStatus(peerID,
	peeringEndpoint string,
) (*clusterresourcesv1beta1.PeeringStatus, error) {
	url := c.serverHostname + peeringEndpoint + peerID

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

	var peeringStatus clusterresourcesv1beta1.PeeringStatus
	err = json.Unmarshal(body, &peeringStatus)
	if err != nil {
		return nil, err
	}

	return &peeringStatus, nil
}

func (c *Client) CreateAzureVNetPeering(peeringSpec *clusterresourcesv1beta1.AzureVNetPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error) {
	payload := &struct {
		PeerSubnets            []string `json:"peerSubnets"`
		PeerResourceGroup      string   `json:"peerResourceGroup"`
		PeerSubscriptionID     string   `json:"peerSubscriptionId"`
		PeerADObjectID         string   `json:"peerAdObjectId,omitempty"`
		PeerVirtualNetworkName string   `json:"peerVirtualNetworkName"`
		CDCID                  string   `json:"cdcId"`
	}{
		PeerSubnets:            peeringSpec.PeerSubnets,
		PeerResourceGroup:      peeringSpec.PeerResourceGroup,
		PeerADObjectID:         peeringSpec.PeerADObjectID,
		PeerSubscriptionID:     peeringSpec.PeerSubscriptionID,
		PeerVirtualNetworkName: peeringSpec.PeerVirtualNetworkName,
		CDCID:                  cdcId,
	}

	jsonDataCreate, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + AzurePeeringEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *clusterresourcesv1beta1.PeeringStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) CreateAWSVPCPeering(peeringSpec *clusterresourcesv1beta1.AWSVPCPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error) {
	payload := &struct {
		PeerSubnets      []string `json:"peerSubnets"`
		PeerAWSAccountID string   `json:"peerAwsAccountId"`
		PeerVPCID        string   `json:"peerVpcId"`
		PeerRegion       string   `json:"peerRegion,omitempty"`
		CDCID            string   `json:"cdcId"`
	}{
		PeerSubnets:      peeringSpec.PeerSubnets,
		PeerAWSAccountID: peeringSpec.PeerAWSAccountID,
		PeerVPCID:        peeringSpec.PeerVPCID,
		PeerRegion:       peeringSpec.PeerRegion,
		CDCID:            cdcId,
	}

	jsonDataCreate, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + AWSPeeringEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *clusterresourcesv1beta1.PeeringStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) CreateGCPVPCPeering(peeringSpec *clusterresourcesv1beta1.GCPVPCPeeringSpec, cdcId string) (*clusterresourcesv1beta1.PeeringStatus, error) {
	payload := &struct {
		PeerSubnets        []string `json:"peerSubnets"`
		PeerVPCNetworkName string   `json:"peerVpcNetworkName"`
		PeerProjectID      string   `json:"peerProjectId"`
		CDCID              string   `json:"cdcId"`
	}{
		PeerSubnets:        peeringSpec.PeerSubnets,
		PeerVPCNetworkName: peeringSpec.PeerVPCNetworkName,
		PeerProjectID:      peeringSpec.PeerProjectID,
		CDCID:              cdcId,
	}

	jsonDataCreate, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + GCPPeeringEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonDataCreate)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}
	var creationResponse *clusterresourcesv1beta1.PeeringStatus
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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
) (*clusterresourcesv1beta1.FirewallRuleStatus, error) {
	url := c.serverHostname + firewallRuleEndpoint + firewallRuleID

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

	var firewallRuleStatus *clusterresourcesv1beta1.FirewallRuleStatus
	err = json.Unmarshal(body, &firewallRuleStatus)
	if err != nil {
		return nil, err
	}

	return firewallRuleStatus, nil
}

func (c *Client) CreateClusterNetworkFirewallRule(
	firewallRuleSpec *clusterresourcesv1beta1.ClusterNetworkFirewallRuleSpec,
	clusterID string,
) (*clusterresourcesv1beta1.FirewallRuleStatus, error) {
	payload := &struct {
		ClusterID string `json:"clusterId"`
		Type      string `json:"type"`
		Network   string `json:"network"`
	}{
		ClusterID: clusterID,
		Type:      firewallRuleSpec.Type,
		Network:   firewallRuleSpec.Network,
	}

	jsonFirewallRule, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + ClusterNetworkFirewallRuleEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonFirewallRule)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}
	var creationResponse *clusterresourcesv1beta1.FirewallRuleStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}
	return creationResponse, nil
}

func (c *Client) CreateAWSSecurityGroupFirewallRule(
	firewallRuleSpec *clusterresourcesv1beta1.AWSSecurityGroupFirewallRuleSpec,
	clusterID string,
) (*clusterresourcesv1beta1.FirewallRuleStatus, error) {
	payload := &struct {
		SecurityGroupID string `json:"securityGroupId"`
		ClusterID       string `json:"clusterId,omitempty"`
		Type            string `json:"type"`
	}{
		SecurityGroupID: firewallRuleSpec.SecurityGroupID,
		ClusterID:       clusterID,
		Type:            firewallRuleSpec.Type,
	}

	jsonFirewallRule, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + AWSSecurityGroupFirewallRuleEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonFirewallRule)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *clusterresourcesv1beta1.FirewallRuleStatus
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	return body, nil
}

func (c *Client) CreateKafkaTopic(url string, t *kafkamanagementv1beta1.Topic) error {
	data, err := json.Marshal(t.Spec)
	if err != nil {
		return err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) UpdateKafkaTopic(url string, t *kafkamanagementv1beta1.Topic) error {
	data, err := json.Marshal(t.Spec.TopicConfigsUpdateToInstAPI())
	if err != nil {
		return err
	}

	url = c.serverHostname + url

	resp, err := c.DoRequest(url, http.MethodPut, data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	t.Status, err = t.Status.FromInstAPI(body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateKafkaUser(
	url string,
	kafkaUser *models.KafkaUser,
) (*kafkamanagementv1beta1.KafkaUserStatus, error) {
	jsonKafkaUser, err := json.Marshal(kafkaUser)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonKafkaUser)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *kafkamanagementv1beta1.KafkaUserStatus
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateKafkaUserCertificate(certRequest *models.CertificateRequest) (*models.Certificate, error) {
	data, err := json.Marshal(certRequest)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + KafkauserCertificatesEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	cert := &models.Certificate{}
	err = json.Unmarshal(body, cert)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *Client) DeleteKafkaUserCertificate(certificateID string) error {
	url := c.serverHostname + KafkauserCertificatesEndpoint + certificateID
	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) RenewKafkaUserCertificate(certificateID string) (*models.Certificate, error) {
	payload := &struct {
		CertificateID string `json:"certificateId"`
	}{
		CertificateID: certificateID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + KafkaUserCertificatesRenewEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cert := &models.Certificate{}
	err = json.Unmarshal(body, cert)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *Client) CreateKafkaMirror(m *kafkamanagementv1beta1.MirrorSpec) (*kafkamanagementv1beta1.MirrorStatus, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + KafkaMirrorEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	status := &kafkamanagementv1beta1.MirrorStatus{}

	err = json.Unmarshal(body, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (c *Client) GetMirrorStatus(id string) (*kafkamanagementv1beta1.MirrorStatus, error) {
	url := c.serverHostname + KafkaMirrorEndpoint + id

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

	m := &kafkamanagementv1beta1.MirrorStatus{}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetClusterBackups(clusterID, clusterKind string) (*models.ClusterBackup, error) {
	url := fmt.Sprintf(GetClusterBackupEndpoint, c.serverHostname, clusterKind, clusterID)

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

	clusterBackups := &models.ClusterBackup{}
	err = json.Unmarshal(body, clusterBackups)
	if err != nil {
		return nil, err
	}

	return clusterBackups, nil
}

func (c *Client) TriggerClusterBackup(clusterID, clusterKind string) error {
	url := fmt.Sprintf(TriggerClusterBackupEndpoint, c.serverHostname, clusterKind, clusterID)

	resp, err := c.DoRequest(url, http.MethodPost, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateExclusionWindow(clusterID string, window *clusterresourcesv1beta1.ExclusionWindowSpec) (string, error) {
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

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

func (c *Client) GetExclusionWindowsStatus(windowID string) (string, error) {
	url := c.serverHostname + ExclusionWindowEndpoint + windowID

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

	response := &struct {
		ID string `json:"id"`
	}{}

	err = json.Unmarshal(body, response)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func (c *Client) DeleteExclusionWindow(id string) error {
	url := c.serverHostname + ExclusionWindowEndpoint + id

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetMaintenanceEvents(clusterID, eventType string) ([]*clusterresourcesv1beta1.MaintenanceEventStatus, error) {
	url := fmt.Sprintf(MaintenanceEventStatusEndpoint, c.serverHostname, clusterID, eventType)

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

	holder := []struct {
		MaintenanceEvents []*clusterresourcesv1beta1.MaintenanceEventStatus `json:"maintenanceEvents"`
	}{}

	err = json.Unmarshal(body, &holder)
	if err != nil {
		return nil, err
	}

	return holder[0].MaintenanceEvents, nil
}

func (c *Client) FetchMaintenanceEventStatuses(
	clusterID string,
) ([]*clusterresourcesv1beta1.ClusteredMaintenanceEventStatus,
	error) {

	inProgressMEStatus, err := c.GetMaintenanceEvents(clusterID, models.InProgressME)
	if err != nil {
		return nil, err
	}

	pastMEStatus, err := c.GetMaintenanceEvents(clusterID, models.PastME)
	if err != nil {
		return nil, err
	}

	upcomingMEStatus, err := c.GetMaintenanceEvents(clusterID, models.UpcomingME)
	if err != nil {
		return nil, err
	}

	if len(inProgressMEStatus) == 0 && len(pastMEStatus) == 0 && len(upcomingMEStatus) == 0 {
		return nil, nil
	}

	iMEStatuses := []*clusterresourcesv1beta1.ClusteredMaintenanceEventStatus{
		{
			InProgress: inProgressMEStatus,
			Past:       pastMEStatus,
			Upcoming:   upcomingMEStatus,
		},
	}

	return iMEStatuses, nil
}

func (c *Client) RescheduleMaintenanceEvent(me *clusterresourcesv1beta1.MaintenanceEventReschedule) error {
	url := fmt.Sprintf(MaintenanceEventRescheduleEndpoint, c.serverHostname, me.MaintenanceEventID)

	requestBody := &struct {
		ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	}{
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) CreateNodeReload(nr *clusterresourcesv1beta1.Node) error {
	url := fmt.Sprintf(NodeReloadEndpoint, c.serverHostname, nr.ID)

	data, err := json.Marshal(nr)
	if err != nil {
		return err
	}

	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

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

	var nodeReload *models.NodeReloadStatus
	err = json.Unmarshal(body, &nodeReload)
	if err != nil {
		return nil, err
	}

	return nodeReload, nil
}

func (c *Client) CreateKafkaACL(
	url string,
	kafkaACL *kafkamanagementv1beta1.KafkaACLSpec,
) (*kafkamanagementv1beta1.KafkaACLStatus, error) {
	jsonKafkaACLUser, err := json.Marshal(kafkaACL)
	if err != nil {
		return nil, err
	}

	url = c.serverHostname + url
	resp, err := c.DoRequest(url, http.MethodPost, jsonKafkaACLUser)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *kafkamanagementv1beta1.KafkaACLStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) GetKafkaACLStatus(
	kafkaACLID,
	kafkaACLEndpoint string,
) (*kafkamanagementv1beta1.KafkaACLStatus, error) {
	url := c.serverHostname + kafkaACLEndpoint + kafkaACLID

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

	var kafkaACLStatus kafkamanagementv1beta1.KafkaACLStatus
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) RestoreCluster(restoreData any, clusterKind string) (string, error) {
	url := fmt.Sprintf(RestoreEndpoint, c.serverHostname, clusterKind)

	jsonData, err := json.Marshal(restoreData)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		ClusterID string `json:"restoredClusterId"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return response.ClusterID, nil
}

func (c *Client) GetPostgreSQL(id string) ([]byte, error) {
	url := c.serverHostname + PGSQLEndpoint + id
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

	return body, nil
}

func (c *Client) UpdatePostgreSQL(id string, r *models.PGClusterUpdate) error {
	reqData, err := json.Marshal(r)
	if err != nil {
		return err
	}

	url := c.serverHostname + PGSQLEndpoint + id
	resp, err := c.DoRequest(url, http.MethodPut, reqData)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

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
) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error) {
	jsonEncryptionKey, err := json.Marshal(encryptionKeySpec)
	if err != nil {
		return nil, err
	}

	url := c.serverHostname + AWSEncryptionKeyEndpoint
	resp, err := c.DoRequest(url, http.MethodPost, jsonEncryptionKey)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var creationResponse *clusterresourcesv1beta1.AWSEncryptionKeyStatus
	err = json.Unmarshal(body, &creationResponse)
	if err != nil {
		return nil, err
	}

	return creationResponse, nil
}

func (c *Client) GetEncryptionKeyStatus(
	encryptionKeyID string,
	encryptionKeyEndpoint string,
) (*clusterresourcesv1beta1.AWSEncryptionKeyStatus, error) {
	url := c.serverHostname + encryptionKeyEndpoint + encryptionKeyID

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

	var encryptionKeyStatus *clusterresourcesv1beta1.AWSEncryptionKeyStatus
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

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) ListAppVersions(app string) ([]*models.AppVersions, error) {
	url := fmt.Sprintf(ListAppsVersionsEndpoint, c.serverHostname, app)

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	var appsVersions []*struct {
		AppVersions []*models.AppVersions `json:"applicationVersions"`
	}

	err = json.Unmarshal(body, &appsVersions)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal body, error: %s", err.Error())
	}

	var appVersions []*models.AppVersions
	for _, appVersion := range appsVersions {
		appVersions = append(appVersions, appVersion.AppVersions...)
	}

	return appVersions, nil
}

func (c *Client) CreateUser(userSpec any, clusterID, app string) error {
	var data []byte
	data, err := json.Marshal(userSpec)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(APIv1UserEndpoint, c.serverHostname, clusterID, app)

	resp, err := c.DoRequest(url, http.MethodPost, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) DeleteUser(username, clusterID, app string) error {
	var deletionRequest struct {
		Username string `json:"username"`
	}
	deletionRequest.Username = username

	var data []byte
	data, err := json.Marshal(deletionRequest)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(APIv1UserEndpoint, c.serverHostname, clusterID, app)

	resp, err := c.DoRequest(url, http.MethodDelete, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) FetchUsers(clusterID, app string) ([]string, error) {
	users := make([]string, 0)

	url := fmt.Sprintf(APIv1UserEndpoint, c.serverHostname, clusterID, app)

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) GetDefaultCredentialsV1(clusterID string) (string, string, error) {
	url := c.serverHostname + ClustersEndpointV1 + clusterID

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	creds := &struct {
		Username                string `json:"username"`
		InstaclustrUserPassword string `json:"instaclustrUserPassword"`
	}{}

	err = json.Unmarshal(b, &creds)
	if err != nil {
		return "", "", err
	}

	return creds.Username, creds.InstaclustrUserPassword, nil
}

func (c *Client) UpdateClusterSettings(clusterID string, settings *models.ClusterSettings) error {
	url := fmt.Sprintf(ClusterSettingsEndpoint, c.serverHostname, clusterID)

	data, err := json.Marshal(settings)
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetAWSEndpointServicePrincipal(principalID string) (*models.AWSEndpointServicePrincipal, error) {
	url := c.serverHostname + AWSEndpointServicePrincipalEndpoint + principalID

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	var principal models.AWSEndpointServicePrincipal
	err = json.Unmarshal(b, &principal)
	if err != nil {
		return nil, err
	}

	return &principal, nil
}

func (c *Client) CreateAWSEndpointServicePrincipal(spec clusterresourcesv1beta1.AWSEndpointServicePrincipalSpec, CDCID string) ([]byte, error) {
	payload := &struct {
		ClusterDataCenterID string `json:"clusterDataCenterId"`
		EndPointServiceID   string `json:"endPointServiceId,omitempty"`
		PrincipalARN        string `json:"principalArn"`
	}{
		ClusterDataCenterID: CDCID,
		EndPointServiceID:   spec.EndPointServiceID,
		PrincipalARN:        spec.PrincipalARN,
	}

	url := c.serverHostname + AWSEndpointServicePrincipalEndpoint

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRequest(url, http.MethodPost, b)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	return b, nil
}

func (c *Client) DeleteAWSEndpointServicePrincipal(principalID string) error {
	url := c.serverHostname + AWSEndpointServicePrincipalEndpoint + principalID
	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	return nil
}

func (c *Client) GetResizeOperationsByClusterDataCentreID(cdcID string) ([]*v1beta1.ResizeOperation, error) {
	url := c.serverHostname + fmt.Sprintf(ClusterResizeOperationsEndpoint, cdcID) + "?activeOnly=true"
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	resize := struct {
		Operations []*v1beta1.ResizeOperation `json:"operations"`
	}{}

	err = json.Unmarshal(b, &resize)
	if err != nil {
		return nil, err
	}

	return resize.Operations, nil
}

func (c *Client) GetAWSVPCPeering(peerID string) (*models.AWSVPCPeering, error) {
	url := c.serverHostname + AWSPeeringEndpoint + peerID
	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	var vpcPeering models.AWSVPCPeering
	err = json.Unmarshal(b, &vpcPeering)
	if err != nil {
		return nil, err
	}

	return &vpcPeering, nil
}

func (c *Client) CreateOpenSearchEgressRules(rule *clusterresourcesv1beta1.OpenSearchEgressRules) (string, error) {
	payload := &struct {
		ClusterID           string `json:"clusterId,omitempty"`
		OpenSearchBindingID string `json:"openSearchBindingId"`
		Source              string `json:"source"`
		Type                string `json:"type,omitempty"`
	}{
		ClusterID:           rule.Status.ClusterID,
		OpenSearchBindingID: rule.Spec.OpenSearchBindingID,
		Source:              rule.Spec.Source,
		Type:                rule.Spec.Type,
	}

	url := c.serverHostname + OpenSearchEgressRulesEndpoint
	status := &clusterresourcesv1beta1.OpenSearchEgressRulesStatus{}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.DoRequest(url, http.MethodPost, b)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	err = json.Unmarshal(b, status)
	if err != nil {
		return "", err
	}

	return status.ID, nil
}

func (c *Client) GetOpenSearchEgressRule(id string) (*clusterresourcesv1beta1.OpenSearchEgressRulesStatus, error) {
	rule := clusterresourcesv1beta1.OpenSearchEgressRulesStatus{}
	url := c.serverHostname + OpenSearchEgressRulesEndpoint + "/" + id

	resp, err := c.DoRequest(url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	err = json.Unmarshal(b, &rule)
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (c *Client) DeleteOpenSearchEgressRule(id string) error {
	url := c.serverHostname + OpenSearchEgressRulesEndpoint + "/" + id

	resp, err := c.DoRequest(url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return NotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code: %d, message: %s", resp.StatusCode, b)
	}

	return nil
}
