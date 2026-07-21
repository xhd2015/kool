// Command bundle compiles kool for the host OS/arch into a single binary.
//
// Usage (from the kool module root):
//
//	go run ./script/bundle
//
// The resulting artifact is written to ./kool-<goos>-<goarch> in the module
// root so host and cross-compiled bundles can coexist. For Linux releases use
// `go run ./script/bundle/for-linux` instead.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

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

	outputName := fmt.Sprintf("kool-%s-%s", runtime.GOOS, runtime.GOARCH)
	out := filepath.Join(root, outputName)

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", out, ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("building: CGO_ENABLED=0 go build -o %s .\n", outputName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	fmt.Printf("\nBundle ready: %s\n", out)
	return nil
}
