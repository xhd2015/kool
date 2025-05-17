package go_tools

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"time"

	"github.com/xhd2015/kool/tools/go/inspect/typed"
	"github.com/xhd2015/kool/tools/go/replace"
	"github.com/xhd2015/kool/tools/go/resolve"
	go_update "github.com/xhd2015/kool/tools/go/update"
	"golang.org/x/tools/go/packages"
)

func Handle(args []string, flagSnippet string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: replace,update")
	}
	switch args[0] {
	case "replace":
		return HandleReplace(args[1:])
	case "update":
		return HandleUpdate(args[1:])
	case "resolve":
		return HandleResolve(args[1:])
	case "inspect":
		return HandleInspect(args[1:])
	case "refactor":
		return HandleRefactor(args[1:])
	case "example":
		return HandleExample(args[1:], flagSnippet)
	}
	return fmt.Errorf("unknown command: %s", args[0])
}

func HandleReplace(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 1 {
		return fmt.Errorf("too many argments: %v", args)
	}
	_, _, err := replace.Replace(args[0])
	return err
}

func HandleUpdate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 1 {
		return fmt.Errorf("too many argments: %v", args)
	}
	return go_update.Update(args[0])
}

func HandleResolve(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: kool go resolve <dir|mod_path> <version>")
	}
	if len(args) > 2 {
		return fmt.Errorf("unrecognized extra argments: %v", args[2:])
	}
	dir, modPath, err := resolve.ResolveModPathFromPossibleDir(args[0])
	if err != nil {
		return err
	}
	version, err := resolve.GoResolveVersion(dir, modPath, args[1])
	if err != nil {
		return err
	}
	fmt.Printf("%s@%s\n", modPath, version)
	return nil
}

func HandleInspect(args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("unrecognized extra argments: %v", args[2:])
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: go inspect <pkg> <T>")
	}
	pkg := args[0]

	var typeName string
	if len(args) == 2 {
		typeName = args[1]
	}

	actPkg, err := resolveOnlyPkg(pkg)
	if err != nil {
		return err
	}
	var t types.Type

	scope := actPkg.Types.Scope()

	if typeName != "" {
		obj := scope.Lookup(typeName)
		if obj == nil {
			return fmt.Errorf("type not found: %s", typeName)
		}
		resolvedType, ok := obj.(*types.TypeName)
		if !ok {
			return fmt.Errorf("%s is not a named type: %s %T", typeName, obj, obj)
		}
		t = resolvedType.Type()
	} else {
		// all
		var fields []*types.Var
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}
			typeName, ok := obj.(*types.TypeName)
			if !ok {
				continue
			}
			fields = append(fields, types.NewVar(obj.Pos(), obj.Pkg(), obj.Name(), typeName.Type()))
		}
		t = types.NewStruct(fields, nil)
	}

	v := typed.MakeDefault(t, typed.MakeDefaultOptions{})
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func resolveOnlyPkg(pkg string) (*packages.Package, error) {
	var loadDir string
	var loadPkg string
	pkgStat, _ := os.Stat(pkg)
	if pkgStat != nil && pkgStat.IsDir() {
		loadDir = pkg
		loadPkg = "./"
	} else {
		loadPkg = pkg
	}

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
		Dir:  loadDir,
		Mode: packages.LoadAllSyntax,
	}, loadPkg)
	close(done)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package not found: %s", pkg)
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("multiple packages found: %v", pkgs)
	}
	return pkgs[0], nil
}

func HandleExample(args []string, flagSnippet string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kool go example <snippet>\navailable snippets: parse-flag")
	}
	snippet := args[0]
	switch snippet {
	case "parse-flag":
		fmt.Print(flagSnippet)
	}
	return nil
}
