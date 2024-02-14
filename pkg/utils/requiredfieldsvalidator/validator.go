package requiredfieldsvalidator

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	omitemptyConstraint = ",omitempty"
	jsonTag             = "json"
)

func ValidateRequiredFields(v any) error {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type - got the %s", val.Kind())
	}

	err := validateValue(val)
	if err != nil {
		return fmt.Errorf("error have occured: %w", err)
	}

	return nil
}

func validateValue(val reflect.Value) error {
	switch val.Kind() {
	case reflect.Struct:
		return validateStruct(val)

	case reflect.Array, reflect.Slice:
		return validateArrayOrSlice(val)

	case reflect.Map:
		return validateMap(val)

	case reflect.Ptr:
		return validatePtr(val)

	case reflect.Bool:
		return nil

	default:
		if val.IsZero() {
			return fmt.Errorf("required field is empty")
		}
	}
	return nil
}

func structValOrPtr(val reflect.Value) bool {
	return val.Kind() == reflect.Struct || isStructPtr(val)
}

func isStructPtr(fieldValue reflect.Value) bool {
	return fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Struct
}

func tagContainsOmitempty(tag string) bool {
	return strings.Contains(tag, omitemptyConstraint)
}

func validateStruct(val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		if skipValidation(field, fieldValue) {
			continue
		}

		err := validateValue(fieldValue)
		if err != nil {
			return fmt.Errorf("%s, %w", field.Name, err)
		}
	}

	return nil
}

func isCollectionType(val reflect.Value) bool {
	return val.Kind() == reflect.Array || val.Kind() == reflect.Slice || val.Kind() == reflect.Map
}

func skipValidation(field reflect.StructField, val reflect.Value) bool {
	if tag, ok := field.Tag.Lookup(jsonTag); ok && tagContainsOmitempty(tag) {
		if isCollectionType(val) && val.Len() == 0 {
			return true
		}
		if val.Kind() != reflect.Ptr && !isCollectionType(val) || val.IsZero() {
			return true
		}
	}
	return false
}

func validateArrayOrSlice(val reflect.Value) error {
	if val.Len() < 1 {
		return fmt.Errorf("len is 0")
	}
	for i := 0; i < val.Len(); i++ {
		fieldValue := val.Index(i)
		if structValOrPtr(fieldValue) {
			err := validateValue(val.Index(i))
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func validateMap(val reflect.Value) error {
	if val.Len() < 1 {
		return fmt.Errorf("empty map")
	}
	iter := val.MapRange()
	for iter.Next() {
		v := iter.Value()
		if structValOrPtr(v) {
			err := validateValue(v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validatePtr(val reflect.Value) error {
	if val.IsNil() {
		return fmt.Errorf("pointer to nil")
	}
	return validateValue(val.Elem())
}
