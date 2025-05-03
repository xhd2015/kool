package move

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
	"github.com/xhd2015/xgo/support/goinfo"
	"golang.org/x/tools/go/packages"
)

// move a package from one directory to another
// it's part of refactor
func Handle(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: kool go move <src> <dst>")
	}
	if len(args) > 2 {
		return fmt.Errorf("unrecognized extra argments: %v", args[2:])
	}
	src := args[0]
	dst := args[1]

	cleanDir := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	if cleanDir == cleanDst {
		fmt.Println("source and destination are the same")
		return nil
	}
	// dst must not exists
	if _, err := os.Stat(cleanDst); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", cleanDst)
	}

	srcSubPaths, srcMainModule, err := goinfo.ResolveMainModule(src)
	if err != nil {
		return fmt.Errorf("resolve src main module: %w", err)
	}

	srcRoot := src
	for range srcSubPaths {
		srcRoot = filepath.Dir(srcRoot)
	}

	// check dst is under srcRoot
	srcRootAbs, err := filepath.Abs(srcRoot)
	if err != nil {
		return fmt.Errorf("get abs path of srcRoot: %w", err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("get abs path of dst: %w", err)
	}
	if !strings.HasPrefix(dstAbs, srcRootAbs) {
		return fmt.Errorf("dst(%s) is not under srcRoot(%s)", dstAbs, srcRootAbs)
	}

	var dstSubPaths []string
	testDst := dstAbs
	for {
		if testDst == srcRootAbs {
			break
		}
		base := filepath.Base(testDst)
		dstSubPaths = append([]string{base}, dstSubPaths...)
		testDst = filepath.Dir(testDst)
	}
	dstPkgPath := srcMainModule
	if len(dstSubPaths) > 0 {
		dstPkgPath = dstPkgPath + "/" + strings.Join(dstSubPaths, "/")
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
		Dir:  srcRoot,
	}, "./...")
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	srcPkgPath := srcMainModule
	if len(srcSubPaths) > 0 {
		srcPkgPath = srcPkgPath + "/" + strings.Join(srcSubPaths, "/")
	}

	var srcPkg *packages.Package
	var callerPkgs []*packages.Package
	packages.Visit(pkgs, func(p *packages.Package) bool {
		if p.PkgPath == srcPkgPath {
			srcPkg = p
		} else {
			if _, ok := p.Imports[srcPkgPath]; ok {
				callerPkgs = append(callerPkgs, p)
			}
		}
		return true
	}, nil)
	if srcPkg == nil {
		return fmt.Errorf("package %s not found", srcPkgPath)
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("create dst directory: %w", err)
	}

	origName := srcPkg.Name
	dstName := filepath.Base(dst)

	var oldUseName string
	// allow package rename, unless it's main
	if origName != "main" && dstName != "main" && origName != dstName {
		oldUseName = origName

		for _, goFile := range srcPkg.GoFiles {
			err := renameFilePackageName(goFile, dstName)
			if err != nil {
				return fmt.Errorf("rename package: %w", err)
			}
		}
	}

	err = os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("rename src to dst: %w", err)
	}

	// TODO: handle rewrite
	for _, callerPkg := range callerPkgs {
		err := rewritePackageImport(callerPkg, srcPkgPath, dstPkgPath, oldUseName)
		if err != nil {
			return fmt.Errorf("rewrite package import: %s %w", callerPkg.PkgPath, err)
		}
	}

	return nil
}

func rewritePackageImport(caller *packages.Package, srcPkgPath string, dstPkgPath string, oldUseName string) error {
	for _, goFile := range caller.GoFiles {
		err := rewriteFileImport(goFile, srcPkgPath, dstPkgPath, oldUseName)
		if err != nil {
			return fmt.Errorf("rewrite file import: %s %w", goFile, err)
		}
	}
	return nil
}

func rewriteFileImport(goFile string, srcPkgPath string, dstPkgPath string, oldUseName string) error {
	content, err := os.ReadFile(goFile)
	if err != nil {
		return fmt.Errorf("read go file: %w", err)
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, goFile, content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse go file: %w", err)
	}
	edit := goedit.NewWithBytes(fset, content)
	for _, importSpec := range astFile.Imports {
		importPathQuoted := importSpec.Path.Value
		importPath, _ := strconv.Unquote(importPathQuoted)
		if importPath == "" {
			continue
		}
		if importPath == srcPkgPath {
			pos := importSpec.Path.Pos()
			end := importSpec.Path.End()
			importStmt := strconv.Quote(dstPkgPath)
			if oldUseName != "" && importSpec.Name == nil {
				importStmt = oldUseName + " " + importStmt
			}
			edit.Replace(pos, end, importStmt)
		}
	}
	if !edit.HasEdit() {
		return nil
	}

	err = os.WriteFile(goFile, edit.Buffer().Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("write go file: %w", err)
	}
	return nil
}

func renameFilePackageName(goFile string, newName string) error {
	content, err := os.ReadFile(goFile)
	if err != nil {
		return fmt.Errorf("read go file: %w", err)
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, goFile, content, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse go file: %w", err)
	}

	edit := goedit.NewWithBytes(fset, content)

	edit.Replace(astFile.Name.Pos(), astFile.Name.End(), newName)
	err = os.WriteFile(goFile, edit.Buffer().Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("write go file: %w", err)
	}
	return nil
}
