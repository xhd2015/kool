//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	lessflags "github.com/xhd2015/less-flags"
	"__MODULE_NAME__/server"
)

const (
	defaultPort = __DEFAULT_PORT__
	defaultHost = "localhost"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "serve":
		if err := server.RunServe(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		cmdStatus(args)
	case "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func serverURL(path string) string {
	return fmt.Sprintf("http://%s:%d%s", defaultHost, defaultPort, path)
}

func printUsage() {
	fmt.Print(`Usage: __DAEMON_NAME__ <command> [flags]

Commands:
  serve    Start the HTTP daemon server
  status   Check if the daemon server is running

Run '__DAEMON_NAME__ <command> --help' for more details.
`)
}

func cmdStatus(args []string) {
	helpText := `Usage: __DAEMON_NAME__ status [flags]

Check if the daemon server is running.

Flags:
  -h, --help  show help
`

	_, err := lessflags.Help("-h,--help", helpText).Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	url := serverURL("/api/health")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Server is not running (port %d): %v\n", defaultPort, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server returned status %d: %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var payload struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || !payload.OK {
		fmt.Printf("Server is running on port %d but health payload is invalid\n", defaultPort)
		os.Exit(1)
	}

	fmt.Printf("Server is running on port %d\n", defaultPort)
}