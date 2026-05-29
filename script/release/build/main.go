package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/script/lib"
)

func main() {
	_, err := lib.BuildRelease(lib.DefaultSpecs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
