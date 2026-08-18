// usage: go run ./script/build (go build -o bin/__PROJECT_NAME__)
//
// Proposed behavior (sketch):
//   1. Parse optional flags if any (default: native go build).
//   2. Run go build -o bin/__PROJECT_NAME__ for the module root.
//   3. Exit non-zero on build failure.
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	if err := Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	fmt.Println("==> Building")
	return cmd.Debug().Run("go", "build", "-o", "bin/__PROJECT_NAME__", ".")
}
