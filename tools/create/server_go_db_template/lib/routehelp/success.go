package routehelp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, SuccessResp{
		Code: 0,
		Data: data,
	})
}

func OK(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, SuccessResp{
		Code: 0,
	})
}
