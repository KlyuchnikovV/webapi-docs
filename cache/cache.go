package cache

import (
	"fmt"
	"go/ast"
	goParser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/parser"
)

var packages Packages

type (
	Package struct {
		Name      string
		Types     map[string]parser.Type
		Functions map[string]*parser.FuncDecl
		Variables map[string]*parser.Variable

		Pkg ast.Package `json:"-"`
	}

	Packages map[string]*Package
)

func NewPackage(pkg ast.Package) (*Package, error) {
	var result = Package{
		Name: pkg.Name,
		// Types:     make(map[string]Type),
		// Functions: make(map[string]*FuncDecl),
		// Variables: make(map[string]*Variable),

		Pkg: pkg,
	}

	for name, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, spec := range typed.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						// Should be 'var' or 'import' specification
						break
					}

					t, err := parser.NewType(*typeSpec, file)
					if err != nil {
						return nil, err
					}

					if t == nil {
						continue
					}

					result.Types[t.GetName()] = t
				}
			case *ast.FuncDecl:
				decl, err := parser.NewFuncDecl(typed.Name.Name, typed, file, result)
				if err != nil {
					return nil, err
				}

				result.Functions[typed.Name.Name] = decl
			default:
				return nil, fmt.Errorf("%s - %#v", name, typed)
			}
		}
	}

	if len(packages) == 0 {
		packages = make(Packages)
	}

	packages[result.Name] = &result

	return &result, nil
}

func Parse(path string) (Packages, *token.FileSet, error) {
	var (
		pkgs = make(Packages)
		fset = token.NewFileSet()
	)

	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	srcBasedPath, err := getSourcesRelativePath(path)
	if err != nil {
		return nil, nil, err
	}

	if !info.IsDir() {
		err = parseFile(path, srcBasedPath, pkgs, fset)
	} else {
		err = parseDir(path, srcBasedPath, pkgs, fset)
	}

	return pkgs, fset, err
}

func parseDir(path, srcBasedPath string, pkgs Packages, fset *token.FileSet) error {
	astPackages, err := goParser.ParseDir(fset, path, nil, goParser.AllErrors)
	if err != nil {
		return err
	}

	for _, astPkg := range astPackages {
		tPkg, err := NewPackage(*astPkg)
		if err != nil {
			return err
		}

		pkgs[srcBasedPath] = tPkg
	}

	list, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}

		if err := parseDir(
			filepath.Join(path, entry.Name()),
			filepath.Join(srcBasedPath, entry.Name()),
			pkgs,
			fset,
		); err != nil {
			return err
		}
	}

	return nil
}

func parseFile(path, srcBasedPath string, pkgs Packages, fset *token.FileSet) error {
	astFile, err := goParser.ParseFile(fset, path, nil, goParser.AllErrors)
	if err != nil {
		return err
	}

	var (
		slashIndex = strings.LastIndexByte(path, '/')
		name       = path[slashIndex+1:]
	)

	pkg, err := NewPackage(ast.Package{
		Name:  name,
		Files: map[string]*ast.File{astFile.Name.Name: astFile},
	})
	if err != nil {
		return err
	}

	pkgs[srcBasedPath] = pkg

	return nil
}

func (pkgs *Packages) FindType(pkgPath, name string) (parser.Type, error) {
	pkg, ok := (*pkgs)[pkgPath]
	if ok {
		return pkg.FindType(name), nil
	}

	packages, _, err := Parse(filepath.Join(os.Getenv("GOPATH"), "src", pkgPath))
	if err != nil {
		return nil, err
	}

	for name, pkg := range packages {
		(*pkgs)[name] = pkg
	}

	return pkgs.FindType(pkgPath, name)
}

func (pkg *Package) FindType(name string) parser.Type {
	t, ok := pkg.Types[name]
	if ok {
		return t
	}

	return pkg.Functions[name]
}

func getSourcesRelativePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var (
		gopath         = os.Getenv("GOPATH") // TODO: move to configs
		srcPath        = filepath.Join(gopath, "src")
		srcIndexInPath = strings.LastIndex(absolutePath, srcPath) + len(srcPath)
	)

	return absolutePath[srcIndexInPath+1:], nil
}
