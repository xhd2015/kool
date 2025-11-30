package run

import (
	"fmt"
	"strings"

	"MODULE_NAME/server"

	"github.com/xhd2015/kool/pkgs/web"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: PROJECT_NAME <subcommand>

Subcommands:
  create    Create a new presentation
`

func Run(args []string) error {
	var devFlag bool
	var component string
	args, err := flags.
		Bool("--dev", &devFlag).
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

	// next port
	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
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
			OpenBrowserUrl: func(port int, url string) string {
				if devFlag {
					return fmt.Sprintf("%s/?component=%s", url, component)
				}
				return url
			},
		})
	}

	return server.Serve(port, devFlag)
}
