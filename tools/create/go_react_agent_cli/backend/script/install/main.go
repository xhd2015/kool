// usage: go run ./script/install
package main

import (
	"fmt"
	"os"
	"strings"

	localinstall "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/localbin/install"
	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
Usage: go run ./script/install [OPTIONS]

Build the React frontend (unless --skip-frontend) and install __PROJECT_NAME__
to the existing PATH copy, or ~/.local/bin/__PROJECT_NAME__ if none.

Options:
  --skip-frontend   skip React frontend build
  -h, --help        show help
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	var skipFrontend bool
	remain, err := lessflags.
		Bool("--skip-frontend", &skipFrontend).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(remain, " "))
	}
	if !skipFrontend {
		fmt.Println("==> frontend")
		if err := cmd.Debug().Run("go", "run", "./script/build-frontend"); err != nil {
			return err
		}
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = localinstall.Install(localinstall.Options{
		Dir:     root,
		Package: "./cmd/__PROJECT_NAME__",
		BinName: "__PROJECT_NAME__",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
	return err
}
