package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/script/dev/util"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
)

// Default database credentials matching podman-compose.yml
const (
	defaultMySQLUser     = "app"
	defaultMySQLPassword = "apppassword"
	defaultMySQLDatabase = "app_db"
)

const help = `
Usage:
  dev [flags]

Flags:
  --debug           Enable debug mode with xgo
  --port PORT       Server port (default: 8080)
  --static DIR      Static files directory
  -v, --verbose     Verbose output
  -h, --help        Show help

Example:
  go run ./script/dev
  go run ./script/dev --debug
  go run ./script/dev --port 3000
  go run ./script/dev --static ./dist
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var debug bool
	var port int
	var static string
	var verbose bool

	args, err := flags.
		Help("-h,--help", help).
		Bool("--debug", &debug).
		Int("--port", &port).
		String("--static", &static).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	projectRoot, err := util.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Build kool command arguments
	cmdArgs := []string{"go", "run"}
	if debug {
		cmdArgs = append([]string{"debug"}, cmdArgs...)
	}
	cmdArgs = append(cmdArgs, "./")

	// Add server flags - default port is 8080
	if port == 0 {
		port = 8080
	}
	cmdArgs = append(cmdArgs, fmt.Sprintf("--port=%d", port))

	// Add MySQL credentials matching podman-compose.yml
	cmdArgs = append(cmdArgs,
		fmt.Sprintf("--mysql-user=%s", defaultMySQLUser),
		fmt.Sprintf("--mysql-password=%s", defaultMySQLPassword),
		fmt.Sprintf("--mysql-database=%s", defaultMySQLDatabase),
	)

	if static != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--static=%s", static))
	}

	if verbose {
		fmt.Printf("Running: kool %v\n", cmdArgs)
		fmt.Printf("Working directory: %s\n", projectRoot)
	}

	c := cmd.Dir(projectRoot).Debug()
	return c.Run("kool", cmdArgs...)
}
