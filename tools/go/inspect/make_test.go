package inspect

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	IntField    int
	StringField string
	SliceField  []int
	MapField    map[string]int
	PtrField    *int
	ArrayField  [2]int
	ByteField   []byte
}

type TestStructWithUnexported struct {
	Exported   int
	unexported int
}

func TestMakeDefault(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		opts     MakeDefaultOptions
		validate func(t *testing.T, val interface{})
	}{
		{
			name: "basic int",
			typ:  reflect.TypeOf(0),
			validate: func(t *testing.T, val interface{}) {
				if val != 0 {
					t.Errorf("expected 0, got %v", val)
				}
			},
		},
		{
			name: "basic string",
			typ:  reflect.TypeOf(""),
			validate: func(t *testing.T, val interface{}) {
				if val != "" {
					t.Errorf("expected empty string, got %v", val)
				}
			},
		},
		{
			name: "basic bool",
			typ:  reflect.TypeOf(false),
			validate: func(t *testing.T, val interface{}) {
				if val != false {
					t.Errorf("expected false, got %v", val)
				}
			},
		},
		{
			name: "slice of int",
			typ:  reflect.TypeOf([]int{}),
			validate: func(t *testing.T, val interface{}) {
				slice, ok := val.([]int)
				if !ok {
					t.Errorf("expected []int, got %T", val)
				}
				if len(slice) != 1 {
					t.Errorf("expected slice with one element, got %v", slice)
				}
				if slice[0] != 0 {
					t.Errorf("expected slice element to be 0, got %v", slice[0])
				}
			},
		},
		{
			name: "map",
			typ:  reflect.TypeOf(map[string]int{}),
			validate: func(t *testing.T, val interface{}) {
				m, ok := val.(map[string]int)
				if !ok {
					t.Errorf("expected map[string]int, got %T", val)
				}
				if len(m) != 1 {
					t.Errorf("expected map with one element, got %v", m)
				}
			},
		},
		{
			name: "struct",
			typ:  reflect.TypeOf(TestStruct{}),
			validate: func(t *testing.T, val interface{}) {
				s, ok := val.(TestStruct)
				if !ok {
					t.Errorf("expected TestStruct, got %T", val)
				}
				if s.IntField != 0 {
					t.Errorf("expected IntField to be 0, got %v", s.IntField)
				}
				if s.StringField != "" {
					t.Errorf("expected StringField to be empty, got %v", s.StringField)
				}
				if len(s.SliceField) != 1 {
					t.Errorf("expected SliceField to have one element, got %v", s.SliceField)
				}
				if len(s.MapField) != 1 {
					t.Errorf("expected MapField to have one element, got %v", s.MapField)
				}
				if s.PtrField == nil {
					t.Errorf("expected PtrField to be non-nil")
				}
				if s.ArrayField != [2]int{0, 0} {
					t.Errorf("expected ArrayField to be [0,0], got %v", s.ArrayField)
				}
				if s.ByteField != nil {
					t.Errorf("expected ByteField to be nil, got %v", s.ByteField)
				}
			},
		},
		{
			name: "struct with unexported",
			typ:  reflect.TypeOf(TestStructWithUnexported{}),
			validate: func(t *testing.T, val interface{}) {
				s, ok := val.(TestStructWithUnexported)
				if !ok {
					t.Errorf("expected TestStructWithUnexported, got %T", val)
				}
				if s.Exported != 0 {
					t.Errorf("expected Exported to be 0, got %v", s.Exported)
				}
			},
		},
		{
			name: "with DefaultValueProvider",
			typ:  reflect.TypeOf(0),
			opts: MakeDefaultOptions{
				DefaultValueProvider: func(t reflect.Type) (interface{}, bool) {
					if t.Kind() == reflect.Int {
						return 42, true
					}
					return nil, false
				},
			},
			validate: func(t *testing.T, val interface{}) {
				if val != 42 {
					t.Errorf("expected 42, got %v", val)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := MakeDefault(tt.typ, tt.opts)
			tt.validate(t, val)
		})
	}
}

func TestMakeDefaultTypeReuse(t *testing.T) {
	type Node struct {
		Value int
		Left  *Node
		Right *Node
	}

	val := MakeDefault(reflect.TypeOf(Node{}), MakeDefaultOptions{})
	node, ok := val.(Node)
	if !ok {
		t.Fatalf("expected Node, got %T", val)
	}

	// The implementation should handle type reuse without panicking
	if node.Left == nil {
		t.Error("expected Left to be non-nil")
	}
	if node.Right == nil {
		t.Error("expected Right to be non-nil")
	}
}
