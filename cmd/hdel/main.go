package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/bash/history"
)

func main() {
	err := history.HandleDel(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
