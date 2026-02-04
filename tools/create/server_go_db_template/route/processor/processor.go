package processor

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/routehelp"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/routewrap"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/service/session"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

var processor = &routewrap.Processor{
	SessionFactory: session.SessionFactoryImpl{},
	IDType:         reflect.TypeOf(types.ID(0)),
	ParseRequest:   routehelp.ParseRequest,
	ProcessResult: func(result interface{}) (interface{}, error) {
		result = routehelp.FillJsonNull(result)
		return result, nil
	},
}

func Gin(f interface{}) func(ctx *gin.Context) {
	return processor.Gin(f)
}
