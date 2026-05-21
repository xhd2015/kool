package main

import (
	"fmt"
	"os"

	"MODULE_NAME/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
