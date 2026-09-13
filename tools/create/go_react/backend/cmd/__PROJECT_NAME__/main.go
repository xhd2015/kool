//go:build ignore

package main

import (
	"fmt"
	"os"

	__PACKAGE_NAME__ "__MODULE_NAME__"
	"__MODULE_NAME__/run"
	"__MODULE_NAME__/server"
)

func main() {
	server.Init(__PACKAGE_NAME__.Dist, __PACKAGE_NAME__.Template)
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
