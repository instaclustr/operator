package requiredfieldsvalidator

import "testing"

type TestStruct struct {
	StrField string `json:"str_field"`
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		args    any
		wantErr bool
	}{
		{
			name: "empty required string",
			args: struct {
				Some string `json:"some"`
			}{
				Some: "",
			},
			wantErr: true,
		},
		{
			name: "filled required string",
			args: struct {
				Some string `json:"some"`
			}{
				Some: "some",
			},
			wantErr: false,
		},
		{
			name: "empty optional string",
			args: struct {
				Some string `json:"some,omitempty"`
			}{
				Some: "",
			},
			wantErr: false,
		},
		{
			name: "filled optional string",
			args: struct {
				Some string `json:"some,omitempty"`
			}{
				Some: "some",
			},
			wantErr: false,
		},
		{
			name: "empty required int",
			args: struct {
				Some int `json:"some"`
			}{
				Some: 0,
			},
			wantErr: true,
		},
		{
			name: "filled required int",
			args: struct {
				Some int `json:"some"`
			}{
				Some: 1,
			},
			wantErr: false,
		},
		{
			name: "empty optional int",
			args: struct {
				Some int `json:"some,omitempty"`
			}{
				Some: 0,
			},
			wantErr: false,
		},
		{
			name: "filled optional int",
			args: struct {
				Some int `json:"some,omitempty"`
			}{
				Some: 1,
			},
			wantErr: false,
		},
		{
			name: "empty required slice",
			args: struct {
				Some []int `json:"some"`
			}{
				Some: []int{},
			},
			wantErr: true,
		},
		{
			name: "filled required slice",
			args: struct {
				Some []int `json:"some"`
			}{
				Some: []int{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "empty optional slice",
			args: struct {
				Some []int `json:"some,omitempty"`
			}{
				Some: []int{},
			},
			wantErr: false,
		},
		{
			name: "filled optional slice",
			args: struct {
				Some []int `json:"some,omitempty"`
			}{
				Some: []int{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "empty required embedded struct",
			args: struct {
				TestStruct `json:",inline"`
			}{
				TestStruct: TestStruct{},
			},
			wantErr: true,
		},
		{
			name: "filled required embedded struct",
			args: struct {
				TestStruct `json:",inline"`
			}{
				TestStruct: TestStruct{
					StrField: "some",
				},
			},
			wantErr: false,
		},
		{
			name: "empty optional embedded struct",
			args: struct {
				TestStruct `json:",inline,omitempty"`
			}{
				TestStruct: TestStruct{},
			},
			wantErr: false,
		},
		{
			name: "filled optional embedded struct",
			args: struct {
				TestStruct `json:",inline"`
			}{
				TestStruct: TestStruct{
					StrField: "some",
				},
			},
			wantErr: false,
		},
		{
			name: "empty required field in filled optional embedded struct",
			args: struct {
				TestStruct `json:",inline"`
			}{
				TestStruct: TestStruct{
					StrField: "",
				},
			},
			wantErr: true,
		},
		{
			name: "empty required pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct"`
			}{},
			wantErr: true,
		},
		{
			name: "filled required pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct"`
			}{
				Struct: &TestStruct{
					StrField: "some",
				},
			},
			wantErr: false,
		},
		{
			name: "empty optional pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct,omitempty"`
			}{},
			wantErr: false,
		},
		{
			name: "filled optional pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct,omitempty"`
			}{
				Struct: &TestStruct{
					StrField: "some",
				},
			},
			wantErr: false,
		},
		{
			name: "empty required field in filled optional pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct,omitempty"`
			}{
				Struct: &TestStruct{
					StrField: "",
				},
			},
			wantErr: true,
		},
		{
			name: "empty required field in filled required pointer to struct",
			args: struct {
				Struct *TestStruct `json:"struct"`
			}{
				Struct: &TestStruct{
					StrField: "",
				},
			},
			wantErr: true,
		},
		{
			name: "empty required slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some"`
			}{
				Some: []*TestStruct{},
			},
			wantErr: true,
		},
		{
			name: "filled required slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some"`
			}{
				Some: []*TestStruct{{
					StrField: "some",
				}},
			},
			wantErr: false,
		},
		{
			name: "empty required field in filled required slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some"`
			}{
				Some: []*TestStruct{{
					StrField: "",
				}},
			},
			wantErr: true,
		},
		{
			name: "empty required field in filled optional slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some,omitempty"`
			}{
				Some: []*TestStruct{{
					StrField: "",
				}},
			},
			wantErr: true,
		},
		{
			name: "filled required field in filled optional slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some,omitempty"`
			}{
				Some: []*TestStruct{{
					StrField: "some",
				}},
			},
			wantErr: false,
		},
		{
			name: "filled optional slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some,omitempty"`
			}{
				Some: []*TestStruct{{
					StrField: "some",
				}},
			},
			wantErr: false,
		},
		{
			name: "empty optional slice of pointers to struct",
			args: struct {
				Some []*TestStruct `json:"some,omitempty"`
			}{
				Some: []*TestStruct{},
			},
			wantErr: false,
		},
		{
			name: "empty required map of string to pointers on struct",
			args: struct {
				Some map[string]*TestStruct `json:"some"`
			}{
				Some: map[string]*TestStruct{},
			},
			wantErr: true,
		},
		{
			name: "filled required map of string to pointers on struct",
			args: struct {
				Some map[string]*TestStruct `json:"some"`
			}{
				Some: map[string]*TestStruct{"some": {
					StrField: "some",
				}},
			},
			wantErr: false,
		},
		{
			name: "empty required field in filled required map of string to pointers on struct",
			args: struct {
				Some map[string]*TestStruct `json:"some"`
			}{
				Some: map[string]*TestStruct{"some": {
					StrField: "",
				}},
			},
			wantErr: true,
		},
		{
			name: "filled required field in filled optional map of string to pointers on struct",
			args: struct {
				Some map[string]*TestStruct `json:"some,omitempty"`
			}{
				Some: map[string]*TestStruct{"some": {
					StrField: "some",
				}},
			},
			wantErr: false,
		},
		{
			name: "empty required field in filled optional map of string to pointers on struct",
			args: struct {
				Some map[string]*TestStruct `json:"some,omitempty"`
			}{
				Some: map[string]*TestStruct{"some": {
					StrField: "",
				}},
			},
			wantErr: true,
		},
		{
			name: "empty required map",
			args: struct {
				Some map[string]string `json:"some"`
			}{
				Some: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "filled required map",
			args: struct {
				Some map[string]string `json:"some"`
			}{
				Some: map[string]string{"some": "some"},
			},
			wantErr: false,
		},
		{
			name: "empty optional map",
			args: struct {
				Some map[string]string `json:"some,omitempty"`
			}{
				Some: nil,
			},
			wantErr: false,
		},
		{
			name: "filed optional map",
			args: struct {
				Some map[string]string `json:"some,omitempty"`
			}{
				Some: map[string]string{"some": "some"},
			},
			wantErr: false,
		},
		{
			name: "filed optional map",
			args: struct {
				Some string `json:"omitempty"`
			}{
				Some: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRequiredFields(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
