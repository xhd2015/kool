//go:build ignore

package main

import (
	"os"

	"__MODULE_NAME__/internal/dev"
	devserver "github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"
)

func main() {
	if err := dev.Run(os.Args[1:]); err != nil {
		devserver.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}
