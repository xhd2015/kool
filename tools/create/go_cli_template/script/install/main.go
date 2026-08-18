// usage: go run ./script/install
// build via go run ./script/build; then go install
//
// Proposed behavior (sketch):
//   1. Build the project via go run ./script/build.
//   2. Install the module with go install .
//   3. Exit non-zero if either step fails.
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := handle(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle() error {
	fmt.Println("==> Building")
	if err := cmd.Debug().Run("go", "run", "./script/build"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println("==> Installing")
	if err := cmd.Debug().Run("go", "install", "."); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}
	fmt.Println("install complete")
	return nil
}
