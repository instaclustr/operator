package dcomparison_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/instaclustr/operator/pkg/utils/dcomparison"
)

func TestCompareMaps(t *testing.T) {
	t.Parallel()

	type _t struct {
		map1     map[string]any
		map2     map[string]any
		expected dcomparison.ObjectDiffs
	}

	cases := map[string]_t{
		"both maps are nil": {},
		"map2 is nil": {
			map1: map[string]any{
				"field1": "value1",
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.field1", Value1: "value1"},
			},
		},
		"map1 is nil": {
			map2: map[string]any{
				"field1": "value1",
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.field1", Value2: "value1"},
			},
		},
		"maps do not contain nested maps and they are equal": {
			map1: map[string]any{
				"field1": "value1",
			},
			map2: map[string]any{
				"field1": "value1",
			},
		},
		"maps do not contain nested maps and they are not equal": {
			map1: map[string]any{
				"field1": "value1",
			},
			map2: map[string]any{
				"field1": "value2",
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.field1", Value1: "value1", Value2: "value2"},
			},
		},
		"maps contain nested maps and they are equal": {
			map1: map[string]any{
				"nestedMap": map[string]any{
					"field1": "value1",
				},
			},
			map2: map[string]any{
				"nestedMap": map[string]any{
					"field1": "value1",
				},
			},
		},
		"maps contain nested maps and they are not equal": {
			map1: map[string]any{
				"nestedMap": map[string]any{
					"field1": "value1",
				},
			},
			map2: map[string]any{
				"nestedMap": map[string]any{
					"field1": "value2",
				},
			},
			expected: []dcomparison.ObjectDiff{
				{Field: "spec.nestedMap.field1", Value1: "value1", Value2: "value2"},
			},
		},
		"maps contain slices and they are equal": {
			map1: map[string]any{
				"slice": []int{1, 2, 3},
			},
			map2: map[string]any{
				"slice": []int{1, 2, 3},
			},
		},
		"maps contain slices and they are not equal": {
			map1: map[string]any{
				"slice": []int{1, 2, 3},
			},
			map2: map[string]any{
				"slice": []int{1, 2},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[2]", Value1: 3},
			},
		},
		"maps contain slices of nested maps and they are equal": {
			map1: map[string]any{
				"slice": []map[string]any{
					{
						"field1": "value1",
						"field2": "value2",
					},
				},
			},
			map2: map[string]any{
				"slice": []map[string]any{
					{
						"field1": "value1",
						"field2": "value2",
					},
				},
			},
		},
		"maps contain slices of nested maps and they are not equal": {
			map1: map[string]any{
				"slice": []map[string]any{
					{
						"field1": "value1",
						"field2": "value2",
					},
				},
			},
			map2: map[string]any{
				"slice": []map[string]any{
					{
						"field1": "value1",
						"field2": "value1",
					},
				},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[0].field2", Value1: "value2", Value2: "value1"},
			},
		},
		"maps contain fields with different types": {
			map1: map[string]any{
				"field1": "1",
			},
			map2: map[string]any{
				"field1": 1,
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.field1", Value1: "1", Value2: 1},
			},
		},
		"the first map contains slice with length bigger than the second contains": {
			map1: map[string]any{
				"slice": []int{1, 2, 3, 4},
			},
			map2: map[string]any{
				"slice": []int{1, 2, 3},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[3]", Value1: 4, Value2: nil},
			},
		},
		"the second map contains slice with length bigger than the first contains": {
			map1: map[string]any{
				"slice": []int{1, 2, 3},
			},
			map2: map[string]any{
				"slice": []int{1, 2, 3, 4},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[3]", Value1: nil, Value2: 4},
			},
		},
		"the first map contains slice with length bigger than the second contains by two elements": {
			map1: map[string]any{
				"slice": []int{1, 2, 3, 4, 5},
			},
			map2: map[string]any{
				"slice": []int{1, 2, 3},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[3]", Value1: 4, Value2: nil},
				{Field: "spec.slice[4]", Value1: 5, Value2: nil},
			},
		},
		"the second map contains slice with length bigger than the first contains by two elements": {
			map1: map[string]any{
				"slice": []int{1, 2, 3},
			},
			map2: map[string]any{
				"slice": []int{1, 2, 3, 4, 5},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice[3]", Value1: nil, Value2: 4},
				{Field: "spec.slice[4]", Value1: nil, Value2: 5},
			},
		},
		"the first map contains nested map with length bigger than the second contains": {
			map1: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
				},
			},
			map2: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
				},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice.field3", Value1: "value3", Value2: nil},
			},
		},
		"the second map contains nested map with length bigger than the first contains": {
			map1: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
				},
			},
			map2: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
				},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice.field3", Value1: nil, Value2: "value3"},
			},
		},
		"the first map contains nested map with length bigger than the second contains by two elements": {
			map1: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
					"field4": "value4",
				},
			},
			map2: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
				},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice.field3", Value1: "value3", Value2: nil},
				{Field: "spec.slice.field4", Value1: "value4", Value2: nil},
			},
		},
		"the second map contains nested map with length bigger than the first contains by two elements": {
			map1: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
				},
			},
			map2: map[string]any{
				"slice": map[string]any{
					"field1": "value1",
					"field2": "value2",
					"field3": "value3",
					"field4": "value4",
				},
			},
			expected: dcomparison.ObjectDiffs{
				{Field: "spec.slice.field3", Value1: nil, Value2: "value3"},
				{Field: "spec.slice.field4", Value1: nil, Value2: "value4"},
			},
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			_ = name
			got := dcomparison.MapsDiff("spec", c.map1, c.map2)

			if !reflect.DeepEqual(got, c.expected) {
				gotJSON, _ := json.MarshalIndent(got, " ", " ")
				expectedJSON, _ := json.MarshalIndent(got, " ", " ")
				t.Errorf("expected: %s, got: %s", gotJSON, expectedJSON)
			}
		})
	}
}
