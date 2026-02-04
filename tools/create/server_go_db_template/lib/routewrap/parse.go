package routewrap

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Func contains metadata about a handler function
type Func struct {
	// Original function
	Func interface{}

	// Function type
	Type reflect.Type
	// Function value
	Value reflect.Value
	// Parameters information

	FirstArgIsCtx bool
	CtxIsGinCtx   bool

	BindingSessionKeys []*SessionKey

	// can be ID, which resolves to `id` param
	// in body or query
	LastInputParam     reflect.Type
	LastInputParamIsID bool

	// Whether the function returns an error as the last return value
	LastReturnIsError bool

	// Whether the function returns a result value before the error
	FirstReturnIsResult bool
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
var ginCtxType = reflect.TypeOf((*gin.Context)(nil)).Elem()

func validateBindings(bindKeys []*SessionKey) (map[reflect.Type]*SessionKey, map[string]*SessionKey, error) {
	seenTypes := make(map[reflect.Type]*SessionKey, len(bindKeys))
	seenNames := make(map[string]*SessionKey, len(bindKeys))
	for _, key := range bindKeys {
		if key.Key == "" {
			return nil, nil, fmt.Errorf("key cannot be empty")
		}
		if _, ok := seenNames[key.Key]; ok {
			return nil, nil, fmt.Errorf("duplicate key: %s", key.Key)
		}
		// must not be basic type
		if isBasicType(key.Type) {
			return nil, nil, fmt.Errorf("binding key must be named type, got basic type: %s", key.Type)
		}
		// must be unique
		if _, ok := seenTypes[key.Type]; ok {
			return nil, nil, fmt.Errorf("duplicate type: %s", key.Type)
		}
		seenTypes[key.Type] = key
		seenNames[key.Key] = key
	}
	return seenTypes, seenNames, nil
}

var basicTypes = map[string]bool{
	"int64":   true,
	"int32":   true,
	"int16":   true,
	"int8":    true,
	"int":     true,
	"uint64":  true,
	"uint32":  true,
	"uint16":  true,
	"uint8":   true,
	"uint":    true,
	"string":  true,
	"byte":    true,
	"rune":    true,
	"bool":    true,
	"float64": true,
}

func isBasicType(t reflect.Type) bool {
	return basicTypes[t.String()]
}

// parseFuncInfo analyzes a function and returns its metadata
// allow parse string to int64
func parseFuncInfo(f interface{}, bindingByType map[reflect.Type]*SessionKey, idType reflect.Type) (*Func, error) {
	if f == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}

	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("requires func, actual: %T", f)
	}

	t := v.Type()
	numIn := t.NumIn()
	numOut := t.NumOut()

	var lastReturnIsError bool
	var firstReturnIsResult bool
	numRes := numOut
	if numOut > 0 {
		lastRetType := t.Out(numOut - 1)
		if lastRetType.Implements(errType) {
			lastReturnIsError = true
			numRes--
		}
	}
	if numRes > 1 {
		return nil, fmt.Errorf("function must return at most one result value and optionally an error, got %d return values: %T", numOut, f)
	}
	if numRes > 0 {
		firstReturnIsResult = true
	}

	var firstArgIsCtx bool
	var ctxIsGinCtx bool
	numParam := numIn
	var paramIndex int
	if numIn > 0 {
		firstParamType := t.In(0)
		if firstParamType == ginCtxType {
			firstArgIsCtx = true
			ctxIsGinCtx = true
			numParam--
			paramIndex++
		} else if firstParamType.Implements(ctxType) {
			firstArgIsCtx = true
			numParam--
			paramIndex++
		}
	}

	// Parse parameters
	// only the last arg can be arbitrary type
	var bindingSessionKeys []*SessionKey
	var lastInputParam reflect.Type
	var lastInputParamIsID bool
	for i := paramIndex; i < numIn; i++ {
		paramType := t.In(i)
		bindingKey := bindingByType[paramType]
		if bindingKey != nil {
			bindingSessionKeys = append(bindingSessionKeys, bindingKey)
			continue
		}
		if lastInputParam != nil {
			return nil, fmt.Errorf("only the last arg can be arbitrary type, got multiple non-binding args at: params[%d]=%s of %T", i, paramType, f)
		}
		lastInputParam = paramType
		if idType != nil && paramType == idType {
			lastInputParamIsID = true
		}
	}

	return &Func{
		Func:                f,
		Type:                t,
		Value:               v,
		FirstArgIsCtx:       firstArgIsCtx,
		CtxIsGinCtx:         ctxIsGinCtx,
		BindingSessionKeys:  bindingSessionKeys,
		LastInputParam:      lastInputParam,
		LastInputParamIsID:  lastInputParamIsID,
		LastReturnIsError:   lastReturnIsError,
		FirstReturnIsResult: firstReturnIsResult,
	}, nil
}

var errParseRequestFail = fmt.Errorf("parse request")

func bindArgs(ctx *gin.Context, session ISession, parseRequest func(ctx *gin.Context, req interface{}) bool, f *Func) ([]reflect.Value, error) {
	nIn := f.Type.NumIn()
	args := make([]reflect.Value, nIn)

	idx := 0
	if f.FirstArgIsCtx {
		idx++
		if f.CtxIsGinCtx {
			args[0] = reflect.ValueOf(ctx)
		} else {
			var plainContext context.Context = ctx
			args[0] = reflect.ValueOf(plainContext)
		}
	}

	if len(f.BindingSessionKeys) > 0 {
		if session == nil {
			return nil, fmt.Errorf("need to bind %d session keys, but session is nil", len(f.BindingSessionKeys))
		}
		for _, key := range f.BindingSessionKeys {
			sessionKey, ok, err := session.Get(key.Key)
			if err != nil {
				return nil, fmt.Errorf("binding session %s: %w", key.Key, err)
			}
			if !ok {
				return nil, fmt.Errorf("binding session %s: not found", key.Key)
			}
			args[idx] = reflect.ValueOf(sessionKey)
			idx++
		}
	}

	if f.LastInputParam != nil {
		paramPtr := reflect.New(f.LastInputParam)
		if f.LastInputParamIsID {
			id, err := getID(ctx)
			if err != nil {
				return nil, err
			}
			paramPtr.Elem().SetInt(id)
		} else {
			if !parseRequest(ctx, paramPtr.Interface()) {
				return nil, errParseRequestFail
			}
		}
		args[idx] = paramPtr.Elem()
	}

	return args, nil
}

func getID(ctx *gin.Context) (int64, error) {
	var idStr string
	paramID, ok := ctx.Params.Get("id")
	if ok {
		if paramID == "" {
			return 0, fmt.Errorf("requires id")
		}
		idStr = paramID
	}
	if idStr == "" {
		queryID, ok := ctx.GetQuery("id")
		if ok {
			if queryID == "" {
				return 0, fmt.Errorf("requires id")
			}
			idStr = queryID
		}
	}
	if idStr == "" {
		var idBody struct {
			ID *json.Number `json:"id"`
		}
		body, err := ctx.GetRawData()
		if err != nil {
			return 0, fmt.Errorf("read body: %w", err)
		}
		if err := json.Unmarshal(body, &idBody); err != nil {
			return 0, fmt.Errorf("unmarshal body: %w", err)
		}
		if idBody.ID != nil {
			idStr = idBody.ID.String()
		}
	}

	if idStr == "" {
		return 0, fmt.Errorf("requires id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id: %w", err)
	}
	return id, nil
}
