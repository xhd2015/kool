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
	case "resolve":
		return handleGoResolve(args[1:])
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

func handleGoResolve(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 2 {
		return fmt.Errorf("requires mod path and version")
	}
	dir, modPath, err := go_update.ResolveModPathFromPossibleDir(args[0])
	if err != nil {
		return err
	}
	version, err := go_update.GoResolve(dir, modPath, args[1])
	if err != nil {
		return err
	}
	fmt.Printf("%s@%s\n", modPath, version)
	return nil
}
