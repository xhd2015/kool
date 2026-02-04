package routehelp

import "reflect"

// swift cannot gracefully handle nil
func FillJsonNull(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Struct {
		newStruct := reflect.New(rv.Type())
		// copy
		newStruct.Elem().Set(rv)
		nv := fillJsonNull(newStruct, 8)
		return nv.Interface()
	}
	nv := fillJsonNull(rv, 8)
	return nv.Interface()
}

func fillJsonNull(rv reflect.Value, depth int) reflect.Value {
	if depth <= 0 {
		return rv
	}
	depth--
	switch rv.Kind() {
	case reflect.Interface:
		if !rv.IsNil() {
			rv.Elem().Set(fillJsonNull(rv.Elem(), depth))
			return rv
		}
		return reflect.New(rv.Type().Elem())
	case reflect.Ptr:
		if rv.IsNil() {
			rv = reflect.New(rv.Type().Elem())
		}
		rv.Elem().Set(fillJsonNull(rv.Elem(), depth))
	case reflect.Slice:
		if !rv.IsNil() {
			return rv
		}
		return reflect.MakeSlice(rv.Type(), 0, 0)
	case reflect.Map:
		if !rv.IsNil() {
			return rv
		}
		return reflect.MakeMap(rv.Type())
	case reflect.Struct:
		n := rv.NumField()
		for i := 0; i < n; i++ {
			field := rv.Field(i)
			if !field.CanSet() {
				continue
			}
			field.Set(fillJsonNull(field, depth))
		}
		return rv
	default:
	}
	return rv
}
