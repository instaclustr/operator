package v1alpha1

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func (kacl *KafkaACL) validateCreate() error {
	if len(kacl.Spec.ACLs) == 0 {
		return fmt.Errorf("acls field should not be empty")
	}

	for _, acl := range kacl.Spec.ACLs {
		principalMatched, err := regexp.Match(models.ACLPrincipalRegExp, []byte(acl.Principal))
		if !principalMatched || err != nil {
			return fmt.Errorf("acl principal should fit pattern: %s, err: %s",
				models.ACLPrincipalRegExp,
				err.Error())
		}

		err = validate(acl)
		if err != nil {
			return err
		}
	}

	if strings.HasPrefix(kacl.Spec.UserQuery, models.ACLUserPrefix) {
		return fmt.Errorf("user query should not have %s prefix", models.ACLUserPrefix)
	}

	return nil
}

func (kacl *KafkaACL) validateUpdate(oldKafkaACL *KafkaACL) error {
	if len(kacl.Spec.ACLs) == 0 {
		return fmt.Errorf("acls field should not be empty")
	}
	if kacl.Spec.ClusterID != oldKafkaACL.Spec.ClusterID {
		return fmt.Errorf("clusterId field is immutable")
	}
	if kacl.Spec.UserQuery != oldKafkaACL.Spec.UserQuery {
		return fmt.Errorf("userQuery field is immutable")
	}
	if strings.HasPrefix(kacl.Spec.UserQuery, models.ACLUserPrefix) {
		return fmt.Errorf("user query should not have %s prefix", models.ACLUserPrefix)
	}

	for _, acl := range kacl.Spec.ACLs {
		if strings.TrimPrefix(acl.Principal, models.ACLUserPrefix) != kacl.Spec.UserQuery {
			return fmt.Errorf("principal value should be the same as userQuery " +
				"and must start with \"User:\" including the wildcard")
		}

		err := validate(acl)
		if err != nil {
			return err
		}
	}

	return nil
}

func validate(acl ACL) error {
	if !validation.Contains(acl.PermissionType, models.ACLPermissionType) {
		return fmt.Errorf("acl permission type %s is unavailable, available values: %v",
			acl.PermissionType, models.ACLPermissionType)
	}
	if !validation.Contains(acl.PatternType, models.ACLPatternType) {
		return fmt.Errorf("acl pattern type %s is unavailable, available values: %v",
			acl.PatternType, models.ACLPatternType)
	}
	if !validation.Contains(acl.Operation, models.ACLOperation) {
		return fmt.Errorf("acl operation %s is unavailable, available values: %v",
			acl.Operation, models.ACLOperation)
	}
	if !validation.Contains(acl.ResourceType, models.ACLResourceType) {
		return fmt.Errorf("acl resource type %s is unavailable, available values: %v",
			acl.ResourceType, models.ACLResourceType)
	}

	return nil
}
