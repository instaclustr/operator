package dcomparison

import (
	"reflect"
	"testing"
)

type unexportedField struct {
	unexported string
}

type intFieldWithSkipTrue struct {
	Int int `dcomparisonSkip:"true"`
}

type intFieldWithSkipFalse struct {
	Int int `dcomparisonSkip:"false"`
}

type stringFieldWithJsonTag struct {
	String string `json:"string"`
}

type intSlice struct {
	Slice []int `json:"slice"`
}

type stringPointer struct {
	StringPointer *string `json:"stringPointer"`
}

type structPointer struct {
	Ptr *stringFieldWithJsonTag `json:"ptr"`
}

type structField struct {
	Field stringFieldWithJsonTag `json:"field"`
}

type sliceOfPointers struct {
	Slice []*stringFieldWithJsonTag `json:"slice"`
}

type mapStringToInt struct {
	Map map[string]int `json:"map"`
}

type mapStringToStructPtr struct {
	Map map[string]*stringFieldWithJsonTag `json:"map"`
}

type ExportedStruct struct {
	Field string `json:"field"`
}

type structWithEmbeddedStructWithJSONInlineConstraint struct {
	ExportedStruct `json:",inline"`
}

func ptrOf[T any](v T) *T {
	return &v
}

func TestStructDiff(t *testing.T) {
	type args struct {
		obj1 any
		obj2 any
	}
	type testCase struct {
		name    string
		args    args
		want    ObjectDiffs
		wantErr bool
	}

	tests := []testCase{
		{
			name:    "non-struct type",
			args:    args{obj1: 1, obj2: 1},
			wantErr: true,
		},
		{
			name: "structs with unexported field, should be skipped",
			args: args{
				obj1: unexportedField{unexported: "unexported"},
				obj2: nil,
			},
		},
		{
			name: "structs with exported int field, same values",
			args: args{
				obj1: struct {
					Exported int
				}{1},
				obj2: struct {
					Exported int
				}{1},
			},
			want: nil,
		},
		{
			name: "structs with exported int field, different values",
			args: args{
				obj1: struct {
					Int int
				}{1},
				obj2: struct {
					Int int
				}{2},
			},
			want: ObjectDiffs{
				{Field: "spec.Int", Value1: 1, Value2: 2},
			},
		},
		{
			name: "structs with exported int field, same values, skip",
			args: args{
				obj1: intFieldWithSkipTrue{1},
				obj2: intFieldWithSkipTrue{1},
			},
		},
		{
			name: "structs with exported int field, different values, skip",
			args: args{
				obj1: intFieldWithSkipTrue{1},
				obj2: intFieldWithSkipTrue{2},
			},
		},
		{
			name: "structs with exported int field, same values, has skip tag but do not skip",
			args: args{
				obj1: intFieldWithSkipFalse{1},
				obj2: intFieldWithSkipFalse{1},
			},
		},
		{
			name: "structs with exported int field, different values, has skip tag but do not skip",
			args: args{
				obj1: intFieldWithSkipFalse{2},
				obj2: intFieldWithSkipFalse{1},
			},
			want: ObjectDiffs{
				{Field: "spec.Int", Value1: 2, Value2: 1},
			},
		},
		{
			name: "exported string field with json tag, same values, do not skip",
			args: args{
				obj1: stringFieldWithJsonTag{"test"},
				obj2: stringFieldWithJsonTag{"test"},
			},
		},
		{
			name: "exported string field with json tag, diff values, do not skip",
			args: args{
				obj1: stringFieldWithJsonTag{"test1"},
				obj2: stringFieldWithJsonTag{"test2"},
			},
			want: ObjectDiffs{
				{Field: "spec.string", Value1: "test1", Value2: "test2"},
			},
		},
		{
			name: "nil slices",
			args: args{
				obj1: intSlice{Slice: nil},
				obj2: intSlice{Slice: nil},
			},
		},
		{
			name: "empty slices",
			args: args{
				obj1: intSlice{Slice: []int{}},
				obj2: intSlice{Slice: []int{}},
			},
		},
		{
			name: "slices with different length",
			args: args{
				obj1: intSlice{Slice: []int{1, 1, 1}},
				obj2: intSlice{Slice: []int{1, 1}},
			},
			want: ObjectDiffs{
				{Field: "spec.slice[2]", Value1: 1, Value2: nil},
			},
		},
		{
			name: "slices same length, different values",
			args: args{
				obj1: intSlice{Slice: []int{1, 2, 3}},
				obj2: intSlice{Slice: []int{3, 2, 1}},
			},
			want: ObjectDiffs{
				{Field: "spec.slice[0]", Value1: 1, Value2: 3},
				{Field: "spec.slice[2]", Value1: 3, Value2: 1},
			},
		},
		{
			name: "structs with pointer to string, nil pointer, equal",
			args: args{
				obj1: stringPointer{StringPointer: nil},
				obj2: stringPointer{StringPointer: nil},
			},
		},
		{
			name: "structs with pointer to string, one of them is nil",
			args: args{
				obj1: stringPointer{StringPointer: ptrOf("test")},
				obj2: stringPointer{StringPointer: nil},
			},
			want: ObjectDiffs{{
				Field:  "spec.stringPointer",
				Value1: "test",
				Value2: nil,
			}},
		},
		{
			name: "structs with pointer to string, both are not nil, equal",
			args: args{
				obj1: stringPointer{StringPointer: ptrOf("test")},
				obj2: stringPointer{StringPointer: ptrOf("test")},
			},
		},
		{
			name: "structs with pointer to string, both are not nil, not equal",
			args: args{
				obj1: stringPointer{StringPointer: ptrOf("test1")},
				obj2: stringPointer{StringPointer: ptrOf("test2")},
			},
			want: ObjectDiffs{{
				Field:  "spec.stringPointer",
				Value1: "test1",
				Value2: "test2",
			}},
		},
		{
			name: "structs with pointer to other struct, nil, equal",
			args: args{
				obj1: structPointer{Ptr: nil},
				obj2: structPointer{Ptr: nil},
			},
		},
		{
			name: "structs with pointer to other struct, not nil, equal",
			args: args{
				obj1: structPointer{Ptr: &stringFieldWithJsonTag{String: "string"}},
				obj2: structPointer{Ptr: &stringFieldWithJsonTag{String: "string"}},
			},
		},
		{
			name: "structs with pointer to other struct, one of them is nil, not equal",
			args: args{
				obj1: structPointer{Ptr: nil},
				obj2: structPointer{Ptr: &stringFieldWithJsonTag{String: "string"}},
			},
			want: ObjectDiffs{{
				Field:  "spec.ptr",
				Value1: nil,
				Value2: stringFieldWithJsonTag{String: "string"},
			}},
		},
		{
			name: "structs with pointer to other struct, not nil, not equal",
			args: args{
				obj1: structPointer{Ptr: &stringFieldWithJsonTag{String: "string1"}},
				obj2: structPointer{Ptr: &stringFieldWithJsonTag{String: "string2"}},
			},
			want: ObjectDiffs{{
				Field:  "spec.ptr.string",
				Value1: "string1",
				Value2: "string2",
			}},
		},
		{
			name: "structs with struct field, empty, equal",
			args: args{
				obj1: structField{},
				obj2: structField{},
			},
		},
		{
			name: "structs with struct field, one of them is empty, not equal",
			args: args{
				obj1: structField{Field: stringFieldWithJsonTag{String: "test"}},
				obj2: structField{},
			},
			want: ObjectDiffs{{
				Field:  "spec.field.string",
				Value1: "test",
				Value2: "",
			}},
		},
		{
			name: "structs with struct field, not empty, equal",
			args: args{
				obj1: structField{Field: stringFieldWithJsonTag{String: "test"}},
				obj2: structField{Field: stringFieldWithJsonTag{String: "test"}},
			},
		},
		{
			name: "structs with struct field, not empty, not equal",
			args: args{
				obj1: structField{Field: stringFieldWithJsonTag{String: "test1"}},
				obj2: structField{Field: stringFieldWithJsonTag{String: "test2"}},
			},
			want: ObjectDiffs{{
				Field:  "spec.field.string",
				Value1: "test1",
				Value2: "test2",
			}},
		},
		{
			name: "structs with slice of pointers to struct, empty, equal",
			args: args{
				obj1: sliceOfPointers{},
				obj2: sliceOfPointers{},
			},
		},
		{
			name: "structs with slice of pointers to struct, one of them is empty, not equal",
			args: args{
				obj1: sliceOfPointers{Slice: []*stringFieldWithJsonTag{{String: "test"}}},
				obj2: sliceOfPointers{},
			},
			want: ObjectDiffs{{
				Field:  "spec.slice[0]",
				Value1: stringFieldWithJsonTag{String: "test"},
				Value2: nil,
			}},
		},
		{
			name: "structs with map string to int, both nil, equal",
			args: args{
				obj1: mapStringToInt{Map: nil},
				obj2: mapStringToInt{Map: nil},
			},
		},
		{
			name: "structs with map string to int, one of them is nil, not equal",
			args: args{
				obj1: mapStringToInt{Map: map[string]int{"int": 1}},
				obj2: mapStringToInt{Map: nil},
			},
			want: ObjectDiffs{{
				Field:  "spec.map[int]",
				Value1: 1,
				Value2: nil,
			}},
		},
		{
			name: "structs with map string to int, not empty, equal",
			args: args{
				obj1: mapStringToInt{Map: map[string]int{"int": 1}},
				obj2: mapStringToInt{Map: map[string]int{"int": 1}},
			},
		},
		{
			name: "structs with map string to int, not empty, both maps have extra pair",
			args: args{
				obj1: mapStringToInt{Map: map[string]int{"int": 1, "extra1": 1}},
				obj2: mapStringToInt{Map: map[string]int{"int": 1, "extra2": 1}},
			},
			want: ObjectDiffs{
				{
					Field:  "spec.map[extra1]",
					Value1: 1,
					Value2: nil,
				},
				{
					Field:  "spec.map[extra2]",
					Value1: nil,
					Value2: 1,
				},
			},
		},
		{
			name: "structs with map string to int, not empty, different value of the same key",
			args: args{
				obj1: mapStringToInt{Map: map[string]int{"int": 1}},
				obj2: mapStringToInt{Map: map[string]int{"int": 2}},
			},
			want: ObjectDiffs{{
				Field:  "spec.map[int]",
				Value1: 1,
				Value2: 2,
			}},
		},
		{
			name: "map string to struct ptr",
			args: args{
				obj1: mapStringToStructPtr{Map: map[string]*stringFieldWithJsonTag{
					"ptr1": {String: "test"},
				}},
				obj2: mapStringToStructPtr{Map: map[string]*stringFieldWithJsonTag{
					"ptr1": {String: "test"},
				}},
			},
		},
		{
			name: "map string to struct ptr, not equal",
			args: args{
				obj1: mapStringToStructPtr{Map: map[string]*stringFieldWithJsonTag{
					"ptr1": {String: "test1"},
				}},
				obj2: mapStringToStructPtr{Map: map[string]*stringFieldWithJsonTag{
					"ptr1": {String: "test2"},
				}},
			},
			want: ObjectDiffs{{
				Field:  "spec.map[ptr1].string",
				Value1: "test1",
				Value2: "test2",
			}},
		},
		{
			name: "struct with embedded struct with json inline constraint, not equal",
			args: args{
				obj1: structWithEmbeddedStructWithJSONInlineConstraint{
					ExportedStruct: ExportedStruct{Field: "test1"},
				},
				obj2: structWithEmbeddedStructWithJSONInlineConstraint{
					ExportedStruct: ExportedStruct{Field: "test2"},
				},
			},
			want: ObjectDiffs{{
				Field:  "spec.field",
				Value1: "test1",
				Value2: "test2",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructsDiff("spec", tt.args.obj1, tt.args.obj2)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructsDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructsDiff() got = %v, want %v", got, tt.want)
			}
		})
	}
}
