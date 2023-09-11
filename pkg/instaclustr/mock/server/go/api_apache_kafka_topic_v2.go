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

// ApacheKafkaTopicV2APIController binds http requests to an api service and writes the service results to the http response
type ApacheKafkaTopicV2APIController struct {
	service      ApacheKafkaTopicV2APIServicer
	errorHandler ErrorHandler
}

// ApacheKafkaTopicV2APIOption for how the controller is set up.
type ApacheKafkaTopicV2APIOption func(*ApacheKafkaTopicV2APIController)

// WithApacheKafkaTopicV2APIErrorHandler inject ErrorHandler into controller
func WithApacheKafkaTopicV2APIErrorHandler(h ErrorHandler) ApacheKafkaTopicV2APIOption {
	return func(c *ApacheKafkaTopicV2APIController) {
		c.errorHandler = h
	}
}

// NewApacheKafkaTopicV2APIController creates a default api controller
func NewApacheKafkaTopicV2APIController(s ApacheKafkaTopicV2APIServicer, opts ...ApacheKafkaTopicV2APIOption) Router {
	controller := &ApacheKafkaTopicV2APIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ApacheKafkaTopicV2APIController
func (c *ApacheKafkaTopicV2APIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2DataSourcesKafkaClusterClusterIdTopicsV2Get": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/kafka_cluster/{clusterId}/topics/v2",
			c.ClusterManagementV2DataSourcesKafkaClusterClusterIdTopicsV2Get,
		},
		"ClusterManagementV2OperationsApplicationsKafkaTopicsV2KafkaTopicIdModifyConfigsV2Put": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/operations/applications/kafka/topics/v2/{kafkaTopicId}/modify-configs/v2",
			c.ClusterManagementV2OperationsApplicationsKafkaTopicsV2KafkaTopicIdModifyConfigsV2Put,
		},
		"ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdDelete": Route{
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/kafka/topics/v2/{kafkaTopicId}/",
			c.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdDelete,
		},
		"ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdGet": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/kafka/topics/v2/{kafkaTopicId}/",
			c.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdGet,
		},
		"ClusterManagementV2ResourcesApplicationsKafkaTopicsV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/kafka/topics/v2/",
			c.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2Post,
		},
	}
}

// ClusterManagementV2DataSourcesKafkaClusterClusterIdTopicsV2Get - Get a list of Kafka topics
func (c *ApacheKafkaTopicV2APIController) ClusterManagementV2DataSourcesKafkaClusterClusterIdTopicsV2Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2DataSourcesKafkaClusterClusterIdTopicsV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2OperationsApplicationsKafkaTopicsV2KafkaTopicIdModifyConfigsV2Put - Update Kafka topic configs
func (c *ApacheKafkaTopicV2APIController) ClusterManagementV2OperationsApplicationsKafkaTopicsV2KafkaTopicIdModifyConfigsV2Put(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	kafkaTopicIdParam := params["kafkaTopicId"]
	kafkaTopicConfigsV2Param := KafkaTopicConfigsV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&kafkaTopicConfigsV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKafkaTopicConfigsV2Required(kafkaTopicConfigsV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertKafkaTopicConfigsV2Constraints(kafkaTopicConfigsV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2OperationsApplicationsKafkaTopicsV2KafkaTopicIdModifyConfigsV2Put(r.Context(), kafkaTopicIdParam, kafkaTopicConfigsV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdDelete - Delete the kafka topic
func (c *ApacheKafkaTopicV2APIController) ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	kafkaTopicIdParam := params["kafkaTopicId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdDelete(r.Context(), kafkaTopicIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdGet - Get Kafka Topic details
func (c *ApacheKafkaTopicV2APIController) ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	kafkaTopicIdParam := params["kafkaTopicId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2KafkaTopicIdGet(r.Context(), kafkaTopicIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsKafkaTopicsV2Post - Create a Kafka Topic
func (c *ApacheKafkaTopicV2APIController) ClusterManagementV2ResourcesApplicationsKafkaTopicsV2Post(w http.ResponseWriter, r *http.Request) {
	kafkaTopicV2Param := KafkaTopicV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&kafkaTopicV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKafkaTopicV2Required(kafkaTopicV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertKafkaTopicV2Constraints(kafkaTopicV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaTopicsV2Post(r.Context(), kafkaTopicV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
