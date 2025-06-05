package find

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/less-gen/go/astinfo"
	"github.com/xhd2015/less-gen/go/load"
)

const help = `
kool go find helps to find names across go project

Usage: kool go find <name> [OPTIONS]

Options:
  --dir <dir>                      project directory
  --load <args>                    load go packages, default is ./...
  --set                            find assignments
  --get                            find access
  -v,--verbose                     show verbose info  

Examples:
  kool go find InterestingField
`

// kool go find MemoryRuleDiscountingInfo.EnableMTAdminFeeDiscount --set
func Handle(args []string) error {
	n := len(args)
	var remainArgs []string
	var verbose bool
	var dir string
	var loadArgs []string
	for i := 0; i < n; i++ {
		flag, value := flags.ParseIndex(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		switch flag {
		case "--dir":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a value", flag)
			}
			dir = value
		case "--load-args", "--load-arg":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a value", flag)
			}
			loadArgs = append(loadArgs, value)
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(help, "\n"))
			return nil
		default:
			return fmt.Errorf("unrecognized flag: %s", flag)
		}
	}
	_ = verbose
	if len(remainArgs) == 0 {
		return fmt.Errorf("requires name")
	}
	if len(remainArgs) > 1 {
		return fmt.Errorf("requires only one name,given: %s", strings.Join(remainArgs[1:], ","))
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

	name := remainArgs[0]

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
							if h.Sel.Name == name {
								fmt.Printf("found %s at %s\n", h.Sel.Name, astinfo.FileLine(fset, h.Sel.Pos()))
							}
						case *ast.Ident:
							if h.Name == name {
								fmt.Printf("found %s at %s\n", h.Name, astinfo.FileLine(fset, h.NamePos))
							}
						}
					}
				case *ast.CompositeLit:
					for _, elt := range n.Elts {
						switch elt := elt.(type) {
						case *ast.KeyValueExpr:
							// fmt.Printf("elt: %s\n", astinfo.FileLine(fset, elt.Pos()))
							keyIdent, ok := elt.Key.(*ast.Ident)
							if ok && keyIdent.Name == name {
								fmt.Printf("found %s at %s\n", keyIdent.Name, astinfo.FileLine(fset, keyIdent.NamePos))
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
