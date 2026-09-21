//go:build ignore

package run

import (
	"errors"
	"fmt"
	"strings"

	"__MODULE_NAME__/server"

	"github.com/xhd2015/kool/pkgs/web"
	lessflags "github.com/xhd2015/less-flags"
)

const serverHelp = `Usage: __PROJECT_NAME__ server [options]

Start the __PROJECT_NAME__ web UI and API server.

Options:
  --dev                    run in dev mode (proxies to the vite dev server)
  --external-vite          in --dev, proxy to an already-running Vite (do not spawn)
  --vite-port PORT         Vite proxy port when using --external-vite (default: 6193)
  --port PORT              listen on PORT (default: auto-select starting at 8080)
  --route-prefix PREFIX    mount the whole app under PREFIX, e.g. my-app
  --component NAME         render a single named component (default: full app)
  -h, --help               show this help

Examples:
  __PROJECT_NAME__ server
  __PROJECT_NAME__ server --port 9000
  __PROJECT_NAME__ server --dev --route-prefix app
`

func handleServer(args []string) error {
	var devFlag bool
	var externalVite bool
	var component string
	var port int
	var vitePort int
	var routePrefix string
	remain, err := lessflags.
		Bool("--dev", &devFlag).
		Bool("--external-vite", &externalVite).
		Int("--port", &port).
		Int("--vite-port", &vitePort).
		String("--route-prefix", &routePrefix).
		String("--component", &component).
		Help("-h,--help", serverHelp).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	if len(remain) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(remain, " "))
	}

	if externalVite && !devFlag {
		return fmt.Errorf("--external-vite requires --dev")
	}

	if component == "list" {
		fmt.Println("Available components: App")
		return nil
	}

	if port == 0 {
		port, err = web.FindAvailablePort(defaultPort, 100)
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
			Dev:          devFlag,
			RoutePrefix:  routePrefix,
			VitePort:     vitePort,
			ExternalVite: externalVite,
			Static: server.StaticOptions{
				IndexHtml: html,
			},
		})
	}

	return server.ServeWithConfig(server.ServeConfig{
		Port:         port,
		Dev:          devFlag,
		RoutePrefix:  routePrefix,
		VitePort:     vitePort,
		ExternalVite: externalVite,
	})
}
