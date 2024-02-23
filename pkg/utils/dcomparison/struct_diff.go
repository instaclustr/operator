/*
Copyright 2024.

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

package dcomparison

import (
	"fmt"
	"reflect"
	"strings"
)

func StructsDiff[T any](path string, obj1, obj2 T) (ObjectDiffs, error) {
	val1 := reflect.ValueOf(obj1)
	val2 := reflect.ValueOf(obj2)

	cmp := structsComparer{}

	if !cmp.isStruct(val1.Type()) {
		return nil, fmt.Errorf("expected struct, got: %s", val1.Kind())
	}

	if val1.Kind() == reflect.Ptr {
		val1 = val1.Elem()
		val2 = val2.Elem()
	}

	cmp.compare(val1, val2, path)
	return cmp.diffs, nil
}

const (
	skipTag   = "dcomparisonSkip"
	skipValue = "true"

	jsonTag              = "json"
	jsonInlineConstraint = ",inline"
)

type structsComparer struct {
	diffs ObjectDiffs
}

func (s *structsComparer) compare(obj1, obj2 reflect.Value, path string) {
	switch obj1.Kind() {
	case reflect.Ptr:
		s.comparePtrs(obj1, obj2, path)
	case reflect.Struct:
		s.compareStructs(obj1, obj2, path)
	case reflect.Slice, reflect.Array:
		s.compareSlicesOrArrays(obj1, obj2, path)
	case reflect.Map:
		s.compareMaps(obj1, obj2, path)
	default:
		val1 := s.getInterfaceValueIfValid(obj1)
		val2 := s.getInterfaceValueIfValid(obj2)

		if val1 != val2 {
			s.diffs.Append(ObjectDiff{
				Field:  path,
				Value1: val1,
				Value2: val2,
			})
		}
	}
}

func (s *structsComparer) isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func (s *structsComparer) comparePtrs(obj1, obj2 reflect.Value, subpath string) {
	switch {
	case obj1.IsValid() && obj2.IsValid():
		s.compare(obj1.Elem(), obj2.Elem(), subpath)

	case obj1.IsZero() && obj2.IsZero():
		return

	default:
		s.diffs.Append(ObjectDiff{
			Field:  subpath,
			Value1: s.getElemIfPtrOrInterface(obj1),
			Value2: s.getElemIfPtrOrInterface(obj2),
		})
	}
}

func (s *structsComparer) compareStructs(obj1, obj2 reflect.Value, path string) {
	n := obj1.NumField()
	for i := 0; i < n; i++ {
		field1 := obj1.Type().Field(i)

		if s.shouldSkip(field1) {
			continue
		}

		// If there is an embedded struct with the json `,inline` constraint:
		// type s struct {
		// 	EmbeddedStruct `json:",inline"
		// }
		// We should skip adding the name of the EmbeddedStruct to sub path
		subPath := path
		if !s.hasJSONInlineConstraint(field1) {
			subPath += "." + s.getFieldName(field1)
		}

		s.compare(obj1.Field(i), obj2.Field(i), subPath)
	}
}

// shouldSkip indicates should the field be skipped during comparing.
// It is skipped only when:
// 1. field is not exported
// 2. field has a skipTag
// 3. value of the skipTag == skipValue
func (s *structsComparer) shouldSkip(field reflect.StructField) bool {
	if !field.IsExported() {
		return true
	}

	val, has := field.Tag.Lookup(skipTag)
	return has && val == skipValue
}

func (s *structsComparer) getJsonFieldName(tag reflect.StructTag) string {
	val, has := tag.Lookup(jsonTag)
	if !has {
		return ""
	}

	return strings.Split(val, ",")[0]
}

func (s *structsComparer) getFieldName(field reflect.StructField) string {
	fieldName := s.getJsonFieldName(field.Tag)
	if fieldName == "" || fieldName == "-" {
		// If there is no json tag use the name of field directly
		fieldName = field.Name
	}

	return fieldName
}

func (s *structsComparer) compareSlicesOrArrays(slice1, slice2 reflect.Value, path string) {
	maxLen := max(slice1.Len(), slice2.Len())
	for i := 0; i < maxLen; i++ {
		val1, val2 := s.getSliceElement(slice1, i), s.getSliceElement(slice2, i)
		subPath := fmt.Sprintf("%s[%d]", path, i)
		s.compare(val1, val2, subPath)
	}
}

func (s *structsComparer) getSliceElement(slice reflect.Value, i int) reflect.Value {
	if i < slice.Len() {
		return slice.Index(i)
	}

	return reflect.ValueOf(nil)
}

func (s *structsComparer) compareMaps(map1, map2 reflect.Value, path string) {
	for _, key := range map1.MapKeys() {
		val1 := map1.MapIndex(key)
		val2 := map2.MapIndex(key)

		subPath := fmt.Sprintf("%s[%v]", path, key.Interface())

		if val2.IsValid() {
			s.compare(val1, val2, subPath)
		} else {
			s.diffs.Append(ObjectDiff{
				Field:  subPath,
				Value1: val1.Interface(),
				Value2: nil,
			})
		}
	}

	for _, key := range map2.MapKeys() {
		subPath := fmt.Sprintf("%s[%v]", path, key.Interface())
		if !map1.MapIndex(key).IsValid() {
			s.diffs.Append(ObjectDiff{
				Field:  subPath,
				Value1: nil,
				Value2: map2.MapIndex(key).Interface(),
			})
		}
	}
}

func (s *structsComparer) getInterfaceValueIfValid(value reflect.Value) any {
	if value.IsValid() {
		return value.Interface()
	}

	return nil
}

func (s *structsComparer) getElemIfPtrOrInterface(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}

	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		return value.Elem().Interface()
	default:
		return value.Interface()
	}
}

func (s *structsComparer) hasJSONInlineConstraint(field reflect.StructField) bool {
	val, has := field.Tag.Lookup(jsonTag)

	return has && strings.Contains(val, jsonInlineConstraint)
}
