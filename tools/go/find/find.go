package find

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/xhd2015/kool/pkgs/reflectfield"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/less-gen/go/astinfo"
	"github.com/xhd2015/less-gen/go/load"
)

const help = `
kool go find helps to find names across go project

Usage: kool go find <name> [OPTIONS]

<name> can be a field name, a method name, a type name, a package name, etc.

Options:
  --dir <dir>                      project directory
  --load-args <args>               load go packages, default is ./...
  --set                            find assignments
  --get                            find access
  -v,--verbose                     show verbose info  

Examples:
  kool go find InterestingField           find set and get
  kool go find InterestingField --set     find set
  kool go find SomeStruct.InterestingField --set   find field set of SomeStruct
`

// kool go find --dir tools/go/find/testfind TestData.TestField
func Handle(args []string) error {
	var dir string
	var set bool
	var get bool
	var verbose bool
	var loadArgs []string
	args, err := flags.String("--dir", &dir).
		StringSlice("--load-args,--load-arg", &loadArgs).
		Bool("--set", &set).
		Bool("--get", &get).
		Bool("-v,--verbose", &verbose).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	name, err := flags.OnlyArg(args)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	loadDir := dir
	if loadDir == "" {
		loadDir = "."
	}

	if len(loadArgs) == 0 {
		loadArgs = []string{"./..."}
	}

	pkgs, err := load.Load(loadDir, loadArgs...)
	if err != nil {
		return err
	}

	var typeName string
	fieldName := name

	dotIdx := strings.Index(name, ".")
	if dotIdx >= 0 {
		typeName = name[:dotIdx]
		fieldName = name[dotIdx+1:]
	}

	fset := pkgs.Fset
	// find all set to field EnableMTAdminFeeDiscount of struct MemoryRuleDiscountingInfo
	for _, pkg := range pkgs.Packages {
		for _, file := range pkg.Syntax {
			// find assignment, Literal
			ast.Inspect(file, func(n ast.Node) bool {

				// pkg.TypesInfo.Defs
				switch n := n.(type) {
				case *ast.AssignStmt:
					for _, lhs := range n.Lhs {
						switch h := lhs.(type) {
						case *ast.SelectorExpr:
							if h.Sel.Name == fieldName {
								// get type of h.X
								if matchTypeName(pkg.TypesInfo, typeName, mode_var, h.X) {
									fmt.Printf("found %s at %s\n", h.Sel.Name, astinfo.FileLine(fset, h.Sel.Pos()))
								}
							}
						case *ast.Ident:
							if h.Name == fieldName {
								if typeName == "" {
									fmt.Printf("found %s at %s\n", h.Name, astinfo.FileLine(fset, h.NamePos))
								}
							}
						}
					}
				case *ast.CompositeLit:
					if matchTypeName(pkg.TypesInfo, typeName, mode_type, n.Type) {
						for _, elt := range n.Elts {
							switch elt := elt.(type) {
							case *ast.KeyValueExpr:
								// fmt.Printf("elt: %s\n", astinfo.FileLine(fset, elt.Pos()))
								keyIdent, ok := elt.Key.(*ast.Ident)
								if ok && keyIdent.Name == fieldName {
									fmt.Printf("found %s at %s\n", keyIdent.Name, astinfo.FileLine(fset, keyIdent.NamePos))
								}
							}
						}
					}
				}
				return true
			})
		}
	}

	return nil
}

type mode int

const (
	mode_type mode = iota
	mode_var
)

func matchTypeName(typesInfo *types.Info, typeName string, mode mode, expr ast.Expr) bool {
	if typeName == "" {
		return true
	}
	exprType := typesInfo.Types[expr]
	switch mode {
	case mode_type:
		if !exprType.IsType() {
			return false
		}
	case mode_var:
		if !exprType.IsValue() {
			return false
		}
	}

	// resolve pointer
	expType := exprType.Type
	if ptr, ok := expType.(*types.Pointer); ok {
		expType = ptr.Elem()
	}

	if chainCheck(typeName, expType) {
		return true
	}
	return false
}

func getNamedFromRHS(namedType *types.Named) types.Type {
	fromRHS := reflect.ValueOf(namedType).Elem().FieldByName("fromRHS")
	return reflectfield.GetUnexportedValue(fromRHS).(types.Type)
}
func getAliasFromRHS(aliasType *types.Alias) types.Type {
	fromRHS := reflect.ValueOf(aliasType).Elem().FieldByName("fromRHS")
	return reflectfield.GetUnexportedValue(fromRHS).(types.Type)
}

func chainCheck(typeName string, typ types.Type) bool {
	if namedType, ok := typ.(*types.Named); ok {
		if namedType.Obj().Name() == typeName {
			return true
		}
		fromRHS := getNamedFromRHS(namedType)
		return chainCheck(typeName, fromRHS)
	}

	if alias, ok := typ.(*types.Alias); ok {
		if alias.Obj().Name() == typeName {
			return true
		}
		fromRHS := getAliasFromRHS(alias)
		return chainCheck(typeName, fromRHS)
	}
	return false
}
