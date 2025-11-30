package web

import (
	"embed"

	"github.com/xhd2015/kool/tools/web/server"
)

//go:embed all:react/dist
var distFS embed.FS

//go:embed react/template.html
var templateHTML string

func init() {
	server.Init(distFS, templateHTML)
}
