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
	args, err := flags.Bool("--dev", &devFlag).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	// next port
	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
	}
	return server.Serve(port, devFlag)
}
