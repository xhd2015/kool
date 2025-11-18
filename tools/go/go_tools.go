package go_tools

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/kool/tools/go/example"
	"github.com/xhd2015/kool/tools/go/find"
	"github.com/xhd2015/kool/tools/go/run"
	"github.com/xhd2015/kool/tools/go/vendortool"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"

	goconfig "github.com/xhd2015/kool/tools/go/config"
	"github.com/xhd2015/kool/tools/go/inspect/typed"
	"github.com/xhd2015/kool/tools/go/replace"
	"github.com/xhd2015/kool/tools/go/resolve"
	go_update "github.com/xhd2015/kool/tools/go/update"
	"golang.org/x/tools/go/packages"
)

const help = `
kool go run helps to debug go command easily

Commands:
  replace
  update
  resolve
  inspect
  refactor
  vendor
  find
  example
  run

Run kool <cmd> --help for more information.

Example:
  kool go run --debug --debug-cwd=<dir> ./ ...
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("commands: replace,update,resolve,inspect,refactor,vendor,example,run")
	}
	gocmd := args[0]
	args = args[1:]
	if gocmd == "help" || gocmd == "--help" {
		fmt.Println(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch gocmd {
	case "replace":
		return HandleReplace(args)
	case "update":
		return HandleUpdate(args)
	case "resolve":
		return HandleResolve(args)
	case "inspect":
		return HandleInspect(args)
	case "refactor":
		return HandleRefactor(args)
	case "vendor":
		return vendortool.Handle(args)
	case "find":
		return find.Handle(args)
	case "example":
		return example.Handle(args)
	case "run":
		return run.Handle(args)
	case "version":
		return cmd.Debug().Run("go", args...)
	case "env":
		return cmd.Debug().Run("go", args...)
	}
	return fmt.Errorf("unknown command: %s", args[0])
}

func HandleReplace(args []string) error {
	var all bool
	var show bool
	args, err := flags.
		Bool("--all", &all).
		Bool("--show", &show).
		Parse(args)
	if err != nil {
		return err
	}

	if all {
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
		}
		if show {
			return goconfig.ShowLocalModulesConfig()
		}
		return replace.ReplaceAll()
	}

	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(args) != 1 {
		return fmt.Errorf("too many argments: %v", args)
	}
	_, _, err = replace.Replace(args[0])
	return err
}

func HandleUpdate(args []string) error {
	var replaced bool
	var all bool
	var show bool
	args, err := flags.
		Bool("--all", &all).
		Bool("--replaced", &replaced).
		Bool("--show", &show).
		Parse(args)
	if err != nil {
		return err
	}

	if all && replaced {
		return fmt.Errorf("cannot use --all and --replaced together")
	}

	if replaced {
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
		}
		return go_update.UpdateReplaced()
	}

	if all {
		if len(args) > 0 {
			return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
		}
		if show {
			return goconfig.ShowLocalModulesConfig()
		}
		return go_update.UpdateAll()
	}

	if len(args) == 0 {
		return fmt.Errorf("requires dir")
	}
	dir := args[0]
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	return go_update.Update(dir)
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
