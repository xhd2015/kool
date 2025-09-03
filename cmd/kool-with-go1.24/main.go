package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/cmd/kool-with-go1.24/with_go"
)

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	if len(args) > 0 && args[0] == "debug" {
		return with_go.Handle(args, nil)
	}
	withGoArgs := make([]string, len(args)+1)
	withGoArgs[0] = "go1.24"
	copy(withGoArgs[1:], args)
	return with_go.Handle(withGoArgs, nil)
}
