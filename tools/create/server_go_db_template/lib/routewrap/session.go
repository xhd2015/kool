package routewrap

import (
	"fmt"
	"reflect"
)

func ExtractStructSessionKeys(sessionStruct interface{}) []*SessionKey {
	rv := reflect.ValueOf(sessionStruct)
	if rv.Kind() != reflect.Struct {
		panic("sessionStruct must be a struct")
	}
	rt := rv.Type()

	n := rt.NumField()
	sessionKeys := make([]*SessionKey, 0, n)
	for i := 0; i < n; i++ {
		field := rt.Field(i)
		if field.Anonymous {
			panic(fmt.Sprintf("sessionStruct cannot have anonymous fields, found: %s", field.Name))
		}
		sessionKeys = append(sessionKeys, &SessionKey{
			Key:  field.Name,
			Type: field.Type,
		})
	}
	return sessionKeys
}
