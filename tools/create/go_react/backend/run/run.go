package run

import (
	"fmt"
	"strings"

	"MODULE_NAME/server"

	"github.com/xhd2015/kool/pkgs/web"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: PROJECT_NAME [options]

Options:
  --dev              run in dev mode (proxies to the vite dev server on :5173)
  --port PORT        listen on PORT (default: auto-select starting at 8080)
  --component NAME   render a single named component (default: full app)
  -h, --help         show this help
`

func Run(args []string) error {
	var devFlag bool
	var component string
	var port int
	args, err := flags.
		Bool("--dev", &devFlag).
		Int("--port", &port).
		String("--component", &component).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if component == "list" {
		fmt.Println("Available components: App")
		return nil
	}

	if port == 0 {
		port, err = web.FindAvailablePort(8080, 100)
		if err != nil {
			return err
		}
	}

	if component != "" {
		var html string
		if !devFlag {
			html, err = server.FormatTemplateHtml(server.FormatOptions{
				Component: component,
			})
			if err != nil {
				return err
			}
		}
		return server.ServeComponent(port, server.ServeOptions{
			Dev: devFlag,
			Static: server.StaticOptions{
				IndexHtml: html,
			},
		})
	}

	return server.Serve(port, devFlag)
}
