package routehelp

import (
	"testing"

	"github.com/xhd2015/xgo/support/assert"
)

func TestFillJsonNull(t *testing.T) {
	type EmptyStruct struct{}

	type StructWithSlice struct {
		List []int
	}
	var emptyString string
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Integer input",
			input:    42,
			expected: 42,
		},
		{
			name:     "String input",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "Empty string input",
			input:    "",
			expected: "", // Adjust based on actual fillJsonNull behavior
		},
		{
			name: "Struct input",
			input: struct {
				Name  string
				Age   int
				Email string
			}{Name: "Alice", Age: 30, Email: ""},
			expected: struct {
				Name  string
				Age   int
				Email string
			}{Name: "Alice", Age: 30, Email: ""},
		},
		{
			name:     "Slice input",
			input:    []string{"a", "", "b"},
			expected: []string{"a", "", "b"},
		},
		{
			name:     "Map input",
			input:    map[string]interface{}{"key1": "value1", "key2": ""},
			expected: map[string]interface{}{"key1": "value1", "key2": ""}, // Adjust if fillJsonNull sets empty to nil
		},
		{
			name:     "Nil pointer input",
			input:    (*string)(nil),
			expected: (*string)(&emptyString),
		},
		{
			name:     "Nil empty struct input",
			input:    (*EmptyStruct)(nil),
			expected: &EmptyStruct{},
		},
		{
			name:  "Nil slice struct input",
			input: (*StructWithSlice)(nil),
			expected: &StructWithSlice{
				List: []int{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call FillJsonNull
			result := FillJsonNull(tt.input)

			if diff := assert.Diff(tt.expected, result); diff != "" {
				t.Error(diff)
			}
		})
	}
}
