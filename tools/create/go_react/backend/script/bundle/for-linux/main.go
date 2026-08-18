// usage: go run ./script/bundle/for-linux
//
// Cross-compiles __PROJECT_NAME__ for linux/amd64.
// The resulting artifact is written to ./__PROJECT_NAME__-linux-amd64 in the module root.
//
// Proposed behavior (sketch):
//   1. Parse -o/--output (default __PROJECT_NAME__-linux-amd64).
//   2. Cross-compile with GOOS=linux GOARCH=amd64 CGO_ENABLED=0.
//   3. Write the binary under the module root and print its path.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/less-gen/flags"
)

const defaultOutput = "__PROJECT_NAME__-linux-amd64"

var help = `
Usage: go run ./script/bundle/for-linux [options]

Cross-compiles the binary for linux/amd64.

Options:
  -o, --output PATH   Output binary path (default: __PROJECT_NAME__-linux-amd64)
  -h, --help          Show this help message
`

func main() {
	if err := Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var output string
	_, err := flags.
		String("-o,--output", &output).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if output == "" {
		output = defaultOutput
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	fmt.Printf("module root: %s\n", root)

	env := append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	out := filepath.Join(root, output)
	fmt.Printf("\nBundle ready: %s\n", out)
	return nil
}
