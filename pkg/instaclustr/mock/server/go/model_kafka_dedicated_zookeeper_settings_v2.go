/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaDedicatedZookeeperSettingsV2 struct {

	// Size of the nodes provisioned as dedicated Zookeeper nodes.
	ZookeeperNodeSize string `json:"zookeeperNodeSize"`

	// Number of dedicated Zookeeper node count, it must be 3 or 5.
	ZookeeperNodeCount int32 `json:"zookeeperNodeCount"`
}

// AssertKafkaDedicatedZookeeperSettingsV2Required checks if the required fields are not zero-ed
func AssertKafkaDedicatedZookeeperSettingsV2Required(obj KafkaDedicatedZookeeperSettingsV2) error {
	elements := map[string]interface{}{
		"zookeeperNodeSize":  obj.ZookeeperNodeSize,
		"zookeeperNodeCount": obj.ZookeeperNodeCount,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKafkaDedicatedZookeeperSettingsV2Constraints checks if the values respects the defined constraints
func AssertKafkaDedicatedZookeeperSettingsV2Constraints(obj KafkaDedicatedZookeeperSettingsV2) error {
	return nil
}
