package typed

import (
	"go/types"
	"testing"
)

func TestMakeDefault(t *testing.T) {
	tests := []struct {
		name     string
		typ      types.Type
		opts     MakeDefaultOptions
		validate func(t *testing.T, val interface{})
	}{
		{
			name: "basic int",
			typ:  types.Typ[types.Int],
			validate: func(t *testing.T, val interface{}) {
				if val != 0 {
					t.Errorf("expected 0, got %v", val)
				}
			},
		},
		{
			name: "basic string",
			typ:  types.Typ[types.String],
			validate: func(t *testing.T, val interface{}) {
				if val != "" {
					t.Errorf("expected empty string, got %v", val)
				}
			},
		},
		{
			name: "basic bool",
			typ:  types.Typ[types.Bool],
			validate: func(t *testing.T, val interface{}) {
				if val != false {
					t.Errorf("expected false, got %v", val)
				}
			},
		},
		{
			name: "slice of int",
			typ:  types.NewSlice(types.Typ[types.Int]),
			validate: func(t *testing.T, val interface{}) {
				slice, ok := val.([]interface{})
				if !ok {
					t.Errorf("expected []interface{}, got %T", val)
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
			typ:  types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			validate: func(t *testing.T, val interface{}) {
				m, ok := val.(map[string]interface{})
				if !ok {
					t.Errorf("expected map[string]interface{}, got %T", val)
				}
				if len(m) != 1 {
					t.Errorf("expected map with one element, got %v", m)
				}
				if v, ok := m[""]; !ok || v != 0 {
					t.Errorf("expected m[\"\"] to be 0, got %v", v)
				}
			},
		},
		{
			name: "struct",
			typ: types.NewStruct([]*types.Var{
				types.NewVar(0, nil, "IntField", types.Typ[types.Int]),
				types.NewVar(0, nil, "StringField", types.Typ[types.String]),
				types.NewVar(0, nil, "SliceField", types.NewSlice(types.Typ[types.Int])),
				types.NewVar(0, nil, "MapField", types.NewMap(types.Typ[types.String], types.Typ[types.Int])),
			}, nil),
			validate: func(t *testing.T, val interface{}) {
				m, ok := val.(map[string]interface{})
				if !ok {
					t.Errorf("expected map[string]interface{}, got %T", val)
				}
				if m["IntField"] != 0 {
					t.Errorf("expected IntField to be 0, got %v", m["IntField"])
				}
				if m["StringField"] != "" {
					t.Errorf("expected StringField to be empty, got %v", m["StringField"])
				}
				if len(m["SliceField"].([]interface{})) != 1 {
					t.Errorf("expected SliceField to have one element, got %v", m["SliceField"])
				}
				mapField, ok := m["MapField"].(map[string]interface{})
				if !ok {
					t.Errorf("expected MapField to be map[string]interface{}, got %T", m["MapField"])
				}
				if len(mapField) != 1 {
					t.Errorf("expected MapField to have one element, got %v", mapField)
				}
			},
		},
		{
			name: "with DefaultValueProvider",
			typ:  types.Typ[types.Int],
			opts: MakeDefaultOptions{
				DefaultValueProvider: func(t types.Type) (interface{}, bool) {
					if t == types.Typ[types.Int] {
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
		{
			name: "array",
			typ:  types.NewArray(types.Typ[types.Int], 3),
			validate: func(t *testing.T, val interface{}) {
				arr, ok := val.([]interface{})
				if !ok {
					t.Errorf("expected []interface{}, got %T", val)
				}
				if len(arr) != 3 {
					t.Errorf("expected array with 3 elements, got %v", arr)
				}
				for i, v := range arr {
					if v != 0 {
						t.Errorf("expected arr[%d] to be 0, got %v", i, v)
					}
				}
			},
		},
		{
			name: "pointer",
			typ:  types.NewPointer(types.Typ[types.Int]),
			validate: func(t *testing.T, val interface{}) {
				ptr, ok := val.(*interface{})
				if !ok {
					t.Errorf("expected *interface{}, got %T", val)
				}
				if ptr == nil {
					t.Error("expected non-nil pointer")
				}
				if *ptr != 0 {
					t.Errorf("expected pointed value to be 0, got %v", *ptr)
				}
			},
		},
		{
			name: "interface",
			typ:  types.NewInterfaceType(nil, nil),
			validate: func(t *testing.T, val interface{}) {
				if val != nil {
					t.Errorf("expected nil, got %v", val)
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
