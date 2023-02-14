/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"net/http"
	"strings"
)

// SwaggerV2ApiController binds http requests to an api service and writes the service results to the http response
type SwaggerV2ApiController struct {
	service      SwaggerV2ApiServicer
	errorHandler ErrorHandler
}

// SwaggerV2ApiOption for how the controller is set up.
type SwaggerV2ApiOption func(*SwaggerV2ApiController)

// WithSwaggerV2ApiErrorHandler inject ErrorHandler into controller
func WithSwaggerV2ApiErrorHandler(h ErrorHandler) SwaggerV2ApiOption {
	return func(c *SwaggerV2ApiController) {
		c.errorHandler = h
	}
}

// NewSwaggerV2ApiController creates a default api controller
func NewSwaggerV2ApiController(s SwaggerV2ApiServicer, opts ...SwaggerV2ApiOption) Router {
	controller := &SwaggerV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the SwaggerV2ApiController
func (c *SwaggerV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2SwaggerForTerraformYamlGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/swagger-for-terraform.yaml",
			c.ClusterManagementV2SwaggerForTerraformYamlGet,
		},
	}
}

// ClusterManagementV2SwaggerForTerraformYamlGet - Retrieve Swagger YAML
func (c *SwaggerV2ApiController) ClusterManagementV2SwaggerForTerraformYamlGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.ClusterManagementV2SwaggerForTerraformYamlGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
