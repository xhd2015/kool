package routewrap

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/routehelp"
)

type Processor struct {
	SessionFactory SessionFactory
	IDType         reflect.Type
	ParseRequest   func(ctx *gin.Context, req interface{}) bool

	ProcessResult func(result interface{}) (interface{}, error)
}

type ISession interface {
	Get(key string) (interface{}, bool, error)
}

type SessionFactory interface {
	SessionKeys() []*SessionKey
	GetSession(ctx *gin.Context) ISession
}

// bind by type
// so types must be unique
type SessionKey struct {
	Key  string
	Type reflect.Type
}

// Gin converts a function to a gin.HandlerFunc. The function can have parameters
// that will be automatically populated from:
// - Path parameters
// - Query parameters
// - Request body (for struct types)
// - Session values
//
// The function should return (interface{}, error) and the result will be
// automatically formatted as a JSON response.
func (c *Processor) Gin(f interface{}) func(ctx *gin.Context) {
	// Parse function info
	var sessionKeys []*SessionKey
	if c.SessionFactory != nil {
		sessionKeys = c.SessionFactory.SessionKeys()
	}
	bindingByType, _, err := validateBindings(sessionKeys)
	if err != nil {
		panic(err)
	}

	funcInfo, err := parseFuncInfo(f, bindingByType, c.IDType)
	if err != nil {
		panic(err)
	}

	parseReq := c.ParseRequest
	if c.ParseRequest == nil {
		parseReq = routehelp.ParseRequest
	}

	// Create the gin handler
	return func(ctx *gin.Context) {
		var session ISession
		if c.SessionFactory != nil {
			session = c.SessionFactory.GetSession(ctx)
		}

		args, err := bindArgs(ctx, session, parseReq, funcInfo)
		if err != nil {
			if err == errParseRequestFail {
				// if parse request fail, the parser will consume
				// the error itself
				return
			}
			routehelp.AbortWithErrCode(ctx, http.StatusBadRequest, err)
			return
		}

		// Call the function
		returnValues := funcInfo.Value.Call(args)

		// Handle the return values
		if funcInfo.LastReturnIsError {
			// Check if there was an error
			errVal := returnValues[len(returnValues)-1].Interface()
			if errVal != nil {
				errTyped, ok := errVal.(error)
				if ok {
					routehelp.AbortWithErrCode(ctx, http.StatusInternalServerError, errTyped)
					return
				}
			}
		}

		// Handle success with a result value
		if funcInfo.FirstReturnIsResult {
			result := returnValues[0].Interface()
			if c.ProcessResult != nil {
				var err error
				result, err = c.ProcessResult(result)
				if err != nil {
					routehelp.AbortWithErrCode(ctx, http.StatusInternalServerError, err)
					return
				}
			}
			routehelp.Success(ctx, result)
			return
		} else {
			routehelp.OK(ctx)
			return
		}
	}
}
