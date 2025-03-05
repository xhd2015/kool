package main

import (
	"fmt"

	"github.com/xhd2015/kool/tools/go_update"
)

func handleGo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: replace,update")
	}
	switch args[0] {
	case "replace":
		return handleGoReplace(args[1:])
	case "update":
		return handleGoUpdate(args[1:])
	}
	return fmt.Errorf("unknown command: %s", args[0])
}

func handleGoReplace(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 1 {
		return fmt.Errorf("too many argments: %v", args)
	}
	_, _, err := go_update.Replace(args[0])
	return err
}

func handleGoUpdate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 1 {
		return fmt.Errorf("too many argments: %v", args)
	}
	return go_update.Update(args[0])
}
