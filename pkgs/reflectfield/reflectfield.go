package reflectfield

import (
	"reflect"
	"unsafe"
)

func GetUnexportedValue(value reflect.Value) interface{} {
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().Interface()
}

func SetUnexportedValue(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
