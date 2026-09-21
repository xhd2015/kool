//go:build ignore

package run

import (
	"fmt"
)

const rootHelp = `Usage: __PROJECT_NAME__ <command> [options]

__PROJECT_NAME__ — agent-driven web app: a Go API server plus a React UI.
Start the server, then talk to its API with the HTTP verb commands.

Commands:
  server    Start the web UI and API server (default port 8080)
  get       GET <URI> and print the response body
  put       PUT <URI> with an optional JSON body (literal, @file, or - for stdin)
  post      POST <URI> with an optional JSON body
  delete    DELETE <URI>
  skill     Show or install the __PROJECT_NAME__ agent skill

Run '__PROJECT_NAME__ <command> --help' for command-specific options.
`

// Main is the CLI entry used by cmd/__PROJECT_NAME__.
func Main(args []string) error {
	return Run(args)
}

// Run dispatches a sub-command. With no arguments (or -h/--help/help) it
// prints the root help.
func Run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Print(rootHelp)
		return nil
	}

	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "server":
		return handleServer(rest)
	case "get":
		return handleGet(rest)
	case "put":
		return handlePut(rest)
	case "post":
		return handlePost(rest)
	case "delete":
		return handleDelete(rest)
	case "skill":
		return handleSkill(rest)
	default:
		return fmt.Errorf("unknown command: %s\nRun '__PROJECT_NAME__ --help' for usage", cmd)
	}
}
