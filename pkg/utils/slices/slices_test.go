package slices

import "testing"

func TestEqualsUnordered(t *testing.T) {
	type args[T comparable] struct {
		s1 []T
		s2 []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}

	tests := []testCase[int]{
		{
			name: "empty",
			args: args[int]{},
			want: true,
		},
		{
			name: "first is empty",
			args: args[int]{
				s2: []int{1},
			},
			want: false,
		},
		{
			name: "second is empty",
			args: args[int]{
				s1: []int{1},
			},
			want: false,
		},
		{
			name: "length differs",
			args: args[int]{
				s1: []int{1},
				s2: []int{1, 1},
			},
			want: false,
		},
		{
			name: "different values",
			args: args[int]{
				s1: []int{1},
				s2: []int{2},
			},
			want: false,
		},
		{
			name: "equals with the same order",
			args: args[int]{
				s1: []int{1, 2, 3},
				s2: []int{1, 2, 3},
			},
			want: true,
		},
		{
			name: "equals but unordered",
			args: args[int]{
				s1: []int{1, 2, 3},
				s2: []int{2, 3, 1},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualsUnordered(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("EqualsUnordered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	type args[T comparable] struct {
		s1 []T
		s2 []T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}

	tests := []testCase[int]{
		{
			name: "empty",
			args: args[int]{},
			want: true,
		},
		{
			name: "first is empty",
			args: args[int]{
				s2: []int{1},
			},
			want: false,
		},
		{
			name: "second is empty",
			args: args[int]{
				s1: []int{1},
			},
			want: false,
		},
		{
			name: "length differs",
			args: args[int]{
				s1: []int{1},
				s2: []int{1, 1},
			},
			want: false,
		},
		{
			name: "different values",
			args: args[int]{
				s1: []int{1},
				s2: []int{2},
			},
			want: false,
		},
		{
			name: "equals with the same order",
			args: args[int]{
				s1: []int{1, 2, 3},
				s2: []int{1, 2, 3},
			},
			want: true,
		},
		{
			name: "unordered",
			args: args[int]{
				s1: []int{1, 2, 3},
				s2: []int{2, 3, 1},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equals(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptrOf[T any](values ...T) []*T {
	var ptrs []*T
	for i := 0; i < len(values); i++ {
		ptrs = append(ptrs, &values[i])
	}

	return ptrs
}

func TestEqualsPtr(t *testing.T) {
	type args[T comparable] struct {
		s1 []*T
		s2 []*T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}

	tests := []testCase[int]{
		{
			name: "empty",
			args: args[int]{},
			want: true,
		},
		{
			name: "same values",
			args: args[int]{
				s1: ptrOf(1),
				s2: ptrOf(1),
			},
			want: true,
		},
		{
			name: "different values",
			args: args[int]{
				s1: ptrOf(1),
				s2: ptrOf(2),
			},
			want: false,
		},
		{
			name: "different length",
			args: args[int]{
				s1: ptrOf(1),
				s2: ptrOf(1, 1),
			},
			want: false,
		},
		{
			name: "equals",
			args: args[int]{
				s1: ptrOf(1, 2, 3),
				s2: ptrOf(1, 2, 3),
			},
			want: true,
		},
		{
			name: "unordered",
			args: args[int]{
				s1: ptrOf(1, 2, 3),
				s2: ptrOf(3, 2, 1),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualsPtr(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("EqualsPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqualsUnorderedPtr(t *testing.T) {
	type args[T comparable] struct {
		s1 []*T
		s2 []*T
	}
	type testCase[T comparable] struct {
		name string
		args args[T]
		want bool
	}
	tests := []testCase[int]{
		{
			name: "empty",
			args: args[int]{},
			want: true,
		},
		{
			name: "first is empty",
			args: args[int]{
				s2: ptrOf(1),
			},
			want: false,
		},
		{
			name: "second is empty",
			args: args[int]{
				s1: ptrOf(1),
			},
			want: false,
		},
		{
			name: "length differs",
			args: args[int]{
				s1: ptrOf(1),
				s2: ptrOf(1, 1),
			},
			want: false,
		},
		{
			name: "different values",
			args: args[int]{
				s1: ptrOf(1),
				s2: ptrOf(2),
			},
			want: false,
		},
		{
			name: "equals with the same order",
			args: args[int]{
				s1: ptrOf(1, 2, 3),
				s2: ptrOf(1, 2, 3),
			},
			want: true,
		},
		{
			name: "equals but unordered",
			args: args[int]{
				s1: ptrOf(1, 2, 3),
				s2: ptrOf(2, 3, 1),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualsUnorderedPtr(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("EqualsUnorderedPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}
