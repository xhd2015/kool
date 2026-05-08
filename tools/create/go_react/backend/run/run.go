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
  --dev                    run in dev mode (proxies to the vite dev server)
  --port PORT              listen on PORT (default: auto-select starting at 8080)
  --route-prefix PREFIX    mount the whole app under PREFIX, e.g. my-app
  --component NAME         render a single named component (default: full app)
  -h, --help               show this help
`

func Run(args []string) error {
	var devFlag bool
	var component string
	var port int
	var routePrefix string
	args, err := flags.
		Bool("--dev", &devFlag).
		Int("--port", &port).
		String("--route-prefix", &routePrefix).
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
			Dev:         devFlag,
			RoutePrefix: routePrefix,
			Static: server.StaticOptions{
				IndexHtml: html,
			},
		})
	}

	return server.Serve(port, devFlag, routePrefix)
}
