/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// AWSVPCPeerV2APIController binds http requests to an api service and writes the service results to the http response
type AWSVPCPeerV2APIController struct {
	service      AWSVPCPeerV2APIServicer
	errorHandler ErrorHandler
}

// AWSVPCPeerV2APIOption for how the controller is set up.
type AWSVPCPeerV2APIOption func(*AWSVPCPeerV2APIController)

// WithAWSVPCPeerV2APIErrorHandler inject ErrorHandler into controller
func WithAWSVPCPeerV2APIErrorHandler(h ErrorHandler) AWSVPCPeerV2APIOption {
	return func(c *AWSVPCPeerV2APIController) {
		c.errorHandler = h
	}
}

// NewAWSVPCPeerV2APIController creates a default api controller
func NewAWSVPCPeerV2APIController(s AWSVPCPeerV2APIServicer, opts ...AWSVPCPeerV2APIOption) Router {
	controller := &AWSVPCPeerV2APIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the AWSVPCPeerV2APIController
func (c *AWSVPCPeerV2APIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/providers/aws/vpc-peers/v2",
			c.ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get,
		},
		"ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/providers/aws/vpc-peers/v2",
			c.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post,
		},
		"ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete": Route{
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/providers/aws/vpc-peers/v2/{vpcPeerId}",
			c.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete,
		},
		"ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/providers/aws/vpc-peers/v2/{vpcPeerId}",
			c.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet,
		},
		"ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/providers/aws/vpc-peers/v2/{vpcPeerId}",
			c.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut,
		},
	}
}

// ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get - List all AWS VPC Peering requests
func (c *AWSVPCPeerV2APIController) ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post - Create AWS VPC Peering Request
func (c *AWSVPCPeerV2APIController) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post(w http.ResponseWriter, r *http.Request) {
	awsVpcPeerV2Param := AwsVpcPeerV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&awsVpcPeerV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertAwsVpcPeerV2Required(awsVpcPeerV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertAwsVpcPeerV2Constraints(awsVpcPeerV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post(r.Context(), awsVpcPeerV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete - Delete AWS VPC Peering Connection
func (c *AWSVPCPeerV2APIController) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	vpcPeerIdParam := params["vpcPeerId"]
	result, err := c.service.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete(r.Context(), vpcPeerIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet - Get AWS VPC Peering Connection info
func (c *AWSVPCPeerV2APIController) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	vpcPeerIdParam := params["vpcPeerId"]
	result, err := c.service.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet(r.Context(), vpcPeerIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut - Update AWS VPC Peering Connection info
func (c *AWSVPCPeerV2APIController) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	vpcPeerIdParam := params["vpcPeerId"]
	awsVpcPeerUpdateV2Param := AwsVpcPeerUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&awsVpcPeerUpdateV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertAwsVpcPeerUpdateV2Required(awsVpcPeerUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertAwsVpcPeerUpdateV2Constraints(awsVpcPeerUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut(r.Context(), vpcPeerIdParam, awsVpcPeerUpdateV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
