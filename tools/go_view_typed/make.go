package go_view_typed

import (
	"fmt"
	"go/types"
	"strings"
)

type MakeDefaultOptions struct {
	DefaultValueProvider func(t types.Type) (val interface{}, ok bool)
}

func MakeDefault(t types.Type, opts MakeDefaultOptions) interface{} {
	val := makeDefault(t, nil, opts, make(map[types.Type]interface{}))
	return val
}

// makeDefault creates a default value for a type.
func makeDefault(t types.Type, path []string, opts MakeDefaultOptions, seen map[types.Type]interface{}) interface{} {
	// Handle cyclic references
	if len(path) > 1000 {
		panic(fmt.Errorf("makeDefault possibly cyclic reference:%v... ", strings.Join(path[:10], ".")))
	}

	// Check if we've seen this type before
	if val, ok := seen[t]; ok {
		return val
	}
	seen[t] = nil // Mark as seen to prevent cycles

	// Try custom provider first
	if opts.DefaultValueProvider != nil {
		if val, ok := opts.DefaultValueProvider(t); ok {
			seen[t] = val
			return val
		}
	}

	var val interface{}
	switch t := t.(type) {
	case *types.Pointer:
		elemVal := makeDefault(t.Elem(), append(path, "&"), opts, seen)
		if elemVal != nil {
			val = &elemVal
		}
	case *types.Interface:
		// For interfaces, return nil
		val = nil
	case *types.Array:
		length := int(t.Len())
		arr := make([]interface{}, length)
		for i := 0; i < length; i++ {
			arr[i] = makeDefault(t.Elem(), append(path, fmt.Sprintf("[%d]", i)), opts, seen)
		}
		val = arr
	case *types.Slice:
		// For slices, create a single-element slice
		elemVal := makeDefault(t.Elem(), append(path, "[]"), opts, seen)
		if elemVal != nil {
			val = []interface{}{elemVal}
		} else {
			val = []interface{}{}
		}
	case *types.Map:
		// For maps, create a single-entry map with string key
		m := make(map[string]interface{})
		keyVal := makeDefault(t.Key(), append(path, "$key"), opts, seen)
		elemVal := makeDefault(t.Elem(), append(path, "$value"), opts, seen)
		if keyVal != nil && elemVal != nil {
			// Convert key to string
			keyStr := fmt.Sprint(keyVal)
			m[keyStr] = elemVal
		}
		val = m
	case *types.Struct:
		// For structs, create a map with default values for exported fields
		m := make(map[string]interface{})
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			if !field.Exported() {
				continue
			}
			fieldVal := makeDefault(field.Type(), append(path, field.Name()), opts, seen)
			if fieldVal != nil {
				m[field.Name()] = fieldVal
			}
		}
		val = m
	case *types.Basic:
		// Handle basic types
		switch t.Kind() {
		case types.Bool:
			val = false
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
			val = 0
		case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			val = uint(0)
		case types.Float32, types.Float64:
			val = 0.0
		case types.String:
			val = ""
		case types.Complex64, types.Complex128:
			val = complex(0, 0)
		default:
			val = nil
		}
	case *types.Named:
		// For named types, use the underlying type
		val = makeDefault(t.Underlying(), path, opts, seen)
	default:
		val = nil
	}

	seen[t] = val
	return val
}
