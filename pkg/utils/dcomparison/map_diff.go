/*
Copyright 2023.

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

// Package dcomparison provides a solution for deeply comparing two maps,
// including their nested maps and slices. It is designed to identify differences
// between two maps that can contain a variety of data types, such as strings,
// integers, other maps, and slices.
package dcomparison

import (
	"fmt"
	"reflect"
)

// ObjectDiff stores the path to a value and the differing values from two maps.
type ObjectDiff struct {
	Field  string
	Value1 any
	Value2 any
}

// ObjectDiffs is a slice of ObjectDiff
type ObjectDiffs []ObjectDiff

func (o *ObjectDiffs) Append(diff ObjectDiff) {
	*o = append(*o, diff)
}

// compare recursively compares values and populates diffs.
func compare(path string, value1, value2 any, diffs *ObjectDiffs) {
	if reflect.DeepEqual(value1, value2) {
		return
	}

	v1Type := reflect.TypeOf(value1)
	v2Type := reflect.TypeOf(value2)

	if v1Type == v2Type && v1Type != nil {
		switch v1Type.Kind() {
		case reflect.Map:
			map1 := reflect.ValueOf(value1)
			map2 := reflect.ValueOf(value2)

			for _, key := range map1.MapKeys() {
				currentField := fmt.Sprintf("%s.%v", path, key)

				map1Value := map1.MapIndex(key)
				map2Value := map2.MapIndex(key)
				if map2Value.IsValid() {
					compare(currentField, map1Value.Interface(), map2Value.Interface(), diffs)
				} else {
					diffs.Append(ObjectDiff{
						Field:  currentField,
						Value1: map1Value.Interface(),
						Value2: nil,
					})
				}
			}

			for _, key := range map2.MapKeys() {
				if !map1.MapIndex(key).IsValid() {
					subPath := fmt.Sprintf("%s.%v", path, key)
					diffs.Append(ObjectDiff{
						Field:  subPath,
						Value1: nil,
						Value2: map2.MapIndex(key).Interface(),
					})
				}
			}
		case reflect.Slice:
			slice1 := reflect.ValueOf(value1)
			slice2 := reflect.ValueOf(value2)
			maxLength := max(slice1.Len(), slice2.Len())

			for i := 0; i < maxLength; i++ {
				elem1, elem2 := getSliceElement(slice1, i), getSliceElement(slice2, i)
				subPath := fmt.Sprintf("%s[%d]", path, i)
				compare(subPath, elem1, elem2, diffs)
			}
		default:
			diffs.Append(ObjectDiff{
				Field:  path,
				Value1: value1,
				Value2: value2,
			})
		}
	} else {
		diffs.Append(ObjectDiff{
			Field:  path,
			Value1: value1,
			Value2: value2,
		})
	}
}

// getSliceElement safely gets the element at index i from a reflect.Value of a slice.
func getSliceElement(slice reflect.Value, i int) any {
	if i < slice.Len() {
		return slice.Index(i).Interface()
	}
	return nil
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MapsDiff compares two maps and returns a slice of ObjectDiff with the differences.
func MapsDiff(path string, map1, map2 map[string]any) ObjectDiffs {
	var diffs ObjectDiffs
	compare(path, map1, map2, &diffs)
	return diffs
}
