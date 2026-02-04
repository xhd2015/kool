package route

import (
	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/handle/example"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/route/processor"
)

func Init(r *gin.Engine) {
	// Example routes - demonstrating how to use processor.Gin
	//
	// processor.Gin automatically handles:
	// - Request parsing (query params, JSON body, path params)
	// - Response formatting (wraps result in {"code":0,"data":...})
	// - Error handling (returns {"code":-1,"msg":"error message"})
	// - Session injection (if handler accepts session types)
	//
	// Usage examples:
	//   GET  /api/example?id=1
	//   POST /api/example/create with JSON body {"name":"My Item"}
	r.GET("/api/example", processor.Gin(example.Get))
	r.POST("/api/example/create", processor.Gin(example.Create))

	// Add your routes here...
}
