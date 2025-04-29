package go_view

import (
	"fmt"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

type MakeDefaultOptions struct {
	DefaultValueProvider func(t reflect.Type) (val interface{}, ok bool)
}

func MakeDefault(t reflect.Type, opts MakeDefaultOptions) interface{} {
	val := makeDefault(t, nil, opts, make(map[reflect.Type]reflect.Value))
	if !val.IsValid() {
		return nil
	}
	return val.Interface()
}

var ByteSliceType = reflect.TypeOf([]byte(nil))

// makeDefault mock default value for a type.
func makeDefault(t reflect.Type, path []string, opts MakeDefaultOptions, seen map[reflect.Type]reflect.Value) reflect.Value {
	var v reflect.Value
	v, ok := seen[t]
	if ok {
		// p may be invalid
		return v
	}
	seen[t] = reflect.Value{} // set an invalid value
	defer func() {
		seen[t] = v // update the value
	}()

	if len(path) > 1000 {
		panic(fmt.Errorf("makeDefault possibly cyclic reference:%v... ", strings.Join(path[:10], ".")))
	}
	defer func() {
		if len(path) == 0 {
			if e := recover(); e != nil {
				panic(fmt.Errorf("makeDefault err:%v %v", strings.Join(path, "."), e))
			}
		}
	}()

	if opts.DefaultValueProvider != nil {
		defaultValue, ok := opts.DefaultValueProvider(t)
		if ok {
			return reflect.ValueOf(defaultValue)
		}
	}

	kind := t.Kind()
	switch kind {
	case reflect.Ptr:
		p := reflect.New(t.Elem())
		val := makeDefault(t.Elem(), append(path, "&"), opts, seen)
		if val.IsValid() {
			p.Elem().Set(val)
		}
		v = p
		return p
	case reflect.Interface:
		// v := reflect.New(t.Elem()) // interface type has no Elem()
		// v.Elem().Set(mockType(t.Elem(), append(path, "#"))) // not needed
		// return v
		// return mockSpecialInterfaceType(t)
		v = reflect.New(t).Elem()
		return v
	case reflect.Array:
		arr := reflect.New(t).Elem()
		for i := 0; i < arr.Len(); i++ {
			val := makeDefault(t.Elem(), append(path, strconv.FormatInt(int64(i), 10)), opts, seen)
			if val.IsValid() {
				arr.Index(i).Set(val)
			}
		}
		v = arr
		return arr
	case reflect.Slice:
		// []byte. Elem is reflect.Uint8, but we cannot tell if is []byte, or []uint8
		// []byte compitable:
		if t.Elem().Kind() == ByteSliceType.Elem().Kind() {
			// empty slice
			v = reflect.ValueOf([]byte(nil))
			return v
		}
		slice := reflect.New(t).Elem()
		val := makeDefault(t.Elem(), append(path, "[]"), opts, seen)
		if val.IsValid() {
			v = reflect.Append(slice, val)
			return v
		}
		v = slice
		return slice
	case reflect.Map:
		m := reflect.New(t).Elem()
		m.Set(reflect.MakeMapWithSize(t, 1)) // must make map, otherwise panic: assignment to entry in nil map
		valK := makeDefault(t.Key(), append(path, "$key"), opts, seen)
		valV := makeDefault(t.Elem(), append(path, "$value"), opts, seen)
		if valK.IsValid() && valV.IsValid() {
			m.SetMapIndex(valK, valV)
		}
		v = m
		return m
	case reflect.Struct:
		strct := reflect.New(t).Elem()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name := t.Field(i).Name
			if !token.IsExported(name) {
				// must be exported
				continue
			}
			val := makeDefault(field.Type, append(path, name), opts, seen)
			if val.IsValid() {
				strct.Field(i).Set(val)
			}
		}
		v = strct
		return strct
	case reflect.Chan, reflect.Func:
		// ignore
		v = reflect.Value{}
		return v
	default:
		v = reflect.New(t).Elem()
		return v
	}
}
