package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/script/dev/util"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
Usage:
  build [flags]

Flags:
  -o, --output string   Output binary path (default: server)
  --debug               Build with debug symbols (no optimization)
  -v, --verbose         Verbose output
  -h, --help            Show help

Example:
  go run ./script/build
  go run ./script/build -o /tmp/server
  go run ./script/build --debug
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var output string
	var debug bool
	var verbose bool

	args, err := flags.
		Help("-h,--help", help).
		String("-o,--output", &output).
		Bool("--debug", &debug).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	if output == "" {
		output = "server"
	}

	projectRoot, err := util.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Build command arguments
	cmdArgs := []string{"build", "-o", output}

	if debug {
		// Disable optimizations and inlining for debugging
		cmdArgs = append(cmdArgs, "-gcflags", "all=-N -l")
	}

	cmdArgs = append(cmdArgs, ".")

	if verbose {
		fmt.Printf("Running: go %v\n", cmdArgs)
		fmt.Printf("Working directory: %s\n", projectRoot)
	}

	c := cmd.Dir(projectRoot).Debug()
	return c.Run("go", cmdArgs...)
}
