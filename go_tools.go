package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"time"

	"github.com/xhd2015/kool/tools/go_update"
	"github.com/xhd2015/kool/tools/go_view_typed"
	"golang.org/x/tools/go/packages"
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
	case "view":
		return handleGoView(args[1:])
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

func handleGoView(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usaage: go view <pkg> <T>")
	}
	if len(args) > 2 {
		return fmt.Errorf("unrecognized extra argments: %v", args[2:])
	}
	pkg := args[0]
	typeName := args[1]

	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-time.After(1 * time.Second):
			fmt.Fprintf(os.Stderr, "loading type info...\n")
		}
	}()

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
	}, pkg)
	close(done)
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("package not found: %s", pkg)
	}
	if len(pkgs) > 1 {
		return fmt.Errorf("multiple packages found: %v", pkgs)
	}
	actPkg := pkgs[0]

	scope := actPkg.Types.Scope()
	obj := scope.Lookup(typeName)
	if obj == nil {
		return fmt.Errorf("type not found: %s", typeName)
	}

	resolvedType, ok := obj.(*types.TypeName)
	if !ok {
		return fmt.Errorf("%s is not a named type: %s %T", typeName, obj, obj)
	}
	t := resolvedType.Type()

	v := go_view_typed.MakeDefault(t, go_view_typed.MakeDefaultOptions{})

	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}
