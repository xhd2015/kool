package routehelp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ParseRequest(ctx *gin.Context, req interface{}) bool {
	if req == nil {
		panic("req is nil")
	}
	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("expect pointer, given: %s", req))
	}

	// body
	if ctx.Request.Body != nil {
		dec := json.NewDecoder(ctx.Request.Body)
		dec.UseNumber()
		err := dec.Decode(req)
		if err != nil && err != io.EOF {
			AbortWithErrCode(ctx, http.StatusBadRequest, err)
			return false
		}
	}

	// query and param
	q := ctx.Request.URL.Query()
	elem := rv.Elem()
	t := elem.Type()
	if t.Kind() == reflect.Pointer {
		if elem.IsNil() {
			elem.Set(reflect.New(t.Elem()))
		}
		elem = elem.Elem()
		t = t.Elem()
	}

	// cannot parse non-struct
	if t.Kind() != reflect.Struct {
		return true
	}

	n := t.NumField()
	for i := 0; i < n; i++ {
		field := t.Field(i)
		jsonField := getJsonFieldName(field.Name, field.Tag.Get("json"))
		if jsonField == "" || jsonField == "-" {
			continue
		}
		val, ok := ctx.Params.Get(jsonField)
		if !ok {
			queryVal, ok := q[jsonField]
			if !ok || len(queryVal) == 0 {
				continue
			}
			val = queryVal[0]
		}

		if err := setField(elem.Field(i), val); err != nil {
			AbortWithErrCode(ctx, http.StatusBadRequest, fmt.Errorf("parse query: %s, %s", field.Name, err))
			return false
		}
	}
	return true
}

func getJsonFieldName(fieldName string, tag string) string {
	idx := strings.Index(tag, ",")
	if idx < 0 {
		idx = len(tag)
	}
	jsonField := strings.TrimSpace(tag[:idx])
	if jsonField != "" {
		return jsonField
	}
	return fieldName
}

// setField allow set int64 from string
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
		return nil
	case reflect.String:
		field.SetString(value)
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
		return nil
	}
	return fmt.Errorf("unsupported field type: %s", field.Kind())
}
