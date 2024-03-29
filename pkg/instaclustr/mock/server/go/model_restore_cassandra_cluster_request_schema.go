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
	"errors"
)

type RestoreCassandraClusterRequestSchema struct {

	// The display name of the restored cluster.<br><br>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored.
	ClusterNameOverride string `json:"clusterNameOverride,omitempty"`

	// An optional list of Cluster Data Centres for which custom VPC settings will be used.<br><br>This property must be populated for all Cluster Data Centres or none at all.
	CdcInfos []RestoreCdcInfoSchema `json:"cdcInfos,omitempty"`

	// Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.<br><br>Defaults to the current date and time.
	PointInTime int64 `json:"pointInTime,omitempty"`

	// A comma separated list of keyspace/table names which follows the format <code>&lt;keyspace&gt;.&lt;table1&gt;,&lt;keyspace&gt;.&lt;table2&gt;</code><br><br>Only data for the specified tables will be restored, for the point in time.
	KeyspaceTables string `json:"keyspaceTables,omitempty"`

	// The cluster network for this cluster to be restored to.
	ClusterNetwork string `json:"clusterNetwork,omitempty"`
}

// AssertRestoreCassandraClusterRequestSchemaRequired checks if the required fields are not zero-ed
func AssertRestoreCassandraClusterRequestSchemaRequired(obj RestoreCassandraClusterRequestSchema) error {
	for _, el := range obj.CdcInfos {
		if err := AssertRestoreCdcInfoSchemaRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRestoreCassandraClusterRequestSchemaConstraints checks if the values respects the defined constraints
func AssertRestoreCassandraClusterRequestSchemaConstraints(obj RestoreCassandraClusterRequestSchema) error {
	if obj.PointInTime < 1420070400000 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	return nil
}
