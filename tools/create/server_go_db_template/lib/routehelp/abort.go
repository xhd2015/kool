package routehelp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/log"
)

func AbortWithErr(ctx *gin.Context, err error) {
	AbortWithErrCode(ctx, http.StatusInternalServerError, err)
}

func AbortWithErrCode(ctx *gin.Context, code int, err error) {
	if true {
		log.Error(ctx, err.Error())
		ctx.Abort()
		ctx.Writer.WriteString(fmt.Sprintf(`{"code":%d, "msg":%q}`, code, err.Error()))
		return
	}
	ctx.AbortWithStatus(code)
	ctx.Writer.WriteString(err.Error())
	// ctx.AbortWithError(code, &gin.Error{Err: err, Type: gin.ErrorTypePublic})
}
