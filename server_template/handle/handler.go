package handle

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"git.garena.com/shopee/loan-service/credit_backend/credit-pricing-center/test/server/handle/model"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

func Wrap(handler interface{}) func(w http.ResponseWriter, r *http.Request) {
	v := reflect.ValueOf(handler)
	if v.Kind() != reflect.Func {
		panic(fmt.Errorf("requires func, actual: %T", handler))
	}
	t := v.Type().In(1)
	if t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("requires arg[1] to be ptr, actual: %v", t))
	}
	if t.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("requires arg[1] to be struct, actual: %v", t))
	}

	numOut := v.Type().NumOut()
	var noResp bool
	if numOut == 1 {
		if v.Type().Out(0).AssignableTo(errType) {
			noResp = true
		}
	} else if numOut == 2 {
		if !v.Type().Out(1).AssignableTo(errType) {
			panic(fmt.Errorf("requires the second result to be error, actual: %v", v.Type().Out(1)))
		}
	} else if numOut > 2 {
		panic(fmt.Errorf("requires response no more than 2"))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var resp interface{}
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("%v", e)
			}
			if err != nil {
				AbortWithErr(w, err)
				return
			}
			if resp == nil {
				ResponseJSON(w, model.NewSuccessResp())
			} else {
				ResponseJSON(w, model.NewResp(resp))
			}
		}()
		req := reflect.New(t.Elem())
		if err = ParseRequest(w, r, req.Interface()); err != nil {
			return
		}
		ctx := context.Background()
		res := v.Call([]reflect.Value{reflect.ValueOf(ctx), req})
		if numOut == 0 {
			return
		}
		if numOut == 1 {
			if noResp {
				if !res[0].IsNil() {
					err = res[0].Interface().(error)
				}
			} else {
				resp = res[0].Interface()
			}
		} else {
			resp = res[0].Interface()
			if !res[1].IsNil() {
				err = res[1].Interface().(error)
			}
		}
		if len(res) == 0 {
			return
		}
		if len(res) == 1 {
			res[0].Interface()
		}
	}
}
