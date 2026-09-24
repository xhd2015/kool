//go:build ignore

package run

import (
	"fmt"
	"strings"

	"__MODULE_NAME__/server"

	"github.com/xhd2015/kool/pkgs/web"
	"github.com/xhd2015/less-flags"
)

const help = `
Usage: __PROJECT_NAME__ [options]

Options:
  --dev                    run in dev mode (proxies to the vite dev server)
  --external-vite          in --dev, proxy to an already-running Vite (do not spawn)
  --vite-port PORT         Vite proxy port when using --external-vite (default: 6193)
  --port PORT              listen on PORT (default: auto-select starting at 8080)
  --route-prefix PREFIX    mount the whole app under PREFIX, e.g. my-app
  --keep-root-route        also serve the unprefixed root route (direct domains)
  --component NAME         render a single named component (default: full app)
  -h, --help               show this help
`

// Main is the CLI entry used by cmd/__PROJECT_NAME__.
func Main(args []string) error {
	return Run(args)
}

func Run(args []string) error {
	var devFlag bool
	var externalVite bool
	var component string
	var port int
	var vitePort int
	var routePrefix string
	var keepRoot bool
	args, err := lessflags.
		Bool("--dev", &devFlag).
		Bool("--external-vite", &externalVite).
		Int("--port", &port).
		Int("--vite-port", &vitePort).
		String("--route-prefix", &routePrefix).
		Bool("--keep-root-route", &keepRoot).
		String("--component", &component).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if externalVite && !devFlag {
		return fmt.Errorf("--external-vite requires --dev")
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
			Dev:          devFlag,
			Route:        server.NormalizeRoute(server.RouteOptions{Prefix: routePrefix, KeepRoot: keepRoot}),
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
		Route:        server.NormalizeRoute(server.RouteOptions{Prefix: routePrefix, KeepRoot: keepRoot}),
		VitePort:     vitePort,
		ExternalVite: externalVite,
	})
}
