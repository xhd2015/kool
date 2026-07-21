// Command for-linux cross-compiles kool into a static Linux amd64 binary.
//
// Usage (from the kool module root):
//
//	go run ./script/bundle/for-linux
//
// The resulting artifact is written to ./kool-linux-amd64 in the module root.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const outputName = "kool-linux-amd64"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	fmt.Printf("module root: %s\n", root)

	out := filepath.Join(root, outputName)
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", out, ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS=linux",
		"GOARCH=amd64",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("building: GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o %s .\n", outputName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	fmt.Printf("\nBundle ready: %s\n", out)
	return nil
}
