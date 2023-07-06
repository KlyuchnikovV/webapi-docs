package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

func ParseDirectory(path, gopath string) (map[string]Package, error) {
	entries, err := getPaths(path)
	if err != nil {
		return nil, err
	}

	var (
		packages   = make(map[string]Package)
		srcDirPath = filepath.Join(gopath, "src/")
		basePath   = strings.TrimSuffix(
			path[strings.LastIndex(path, srcDirPath)+len(srcDirPath):],
			strings.TrimLeft(path, "."),
		)
	)

	for _, entry := range entries {
		pkgs, err := parseDirectory(entry)
		if err != nil {
			return nil, err
		}

		for _, p := range pkgs {
			var name = basePath
			if p.Name != "main" {
				name = fmt.Sprintf("%s/%s", strings.Trim(basePath, "/"), p.Name)
			}

			packages[name] = p
		}
	}

	return packages, nil
}

func ParsePackage(gopath, path string) (map[string]Package, error) {
	var (
		packages     = make(map[string]Package)
		absolutePath = filepath.Join(gopath, "src", path)
	)

	pkgs, err := parser.ParseDir(&token.FileSet{}, absolutePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	if len(path) == 0 {
		return nil, nil
	}

	for name, p := range pkgs {
		var (
			index   = strings.Index(name, "/")
			newPath = path
		)

		if index != -1 {
			newPath = fmt.Sprintf("%s/%s", path, name[index+1:])
		}

		pkg, err := NewPackage(p, path)
		if err != nil {
			return nil, err
		}

		packages[newPath] = *pkg
	}

	return packages, nil
}

func getPaths(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	paths, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	var result = []string{path}

	for _, innerPath := range paths {
		subPaths, err := getPaths(filepath.Join(path, innerPath))
		if err != nil {
			return nil, err
		}

		result = append(result, subPaths...)
	}

	return result, nil
}

func parseDirectory(path string) (map[string]Package, error) {
	pkgs, err := parser.ParseDir(&token.FileSet{}, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	if len(pkgs) > 1 {
		panic("several packages in one directory!")
	}

	var result = make(map[string]Package)

	for _, pkg := range pkgs {
		p, err := NewPackage(pkg, path)
		if err != nil {
			return nil, err
		}

		result[path] = *p
	}

	return result, nil
}

func newFromGenDecl(file *ast.File, decl *ast.GenDecl) []types.Type {
	if decl.Tok != token.TYPE {
		return nil
	}

	var result = make([]types.Type, len(decl.Specs))

	for i, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			panic("not ts")
		}

		result[i] = types.NewType(file, ts.Name.Name, ts.Type, nil)
	}

	return result
}

func newFromFuncDecl(pkg Package, file *ast.File, decl *ast.FuncDecl) (*types.FuncType, types.Type, error) {
	var (
		fun      = types.NewFunc(file, decl.Name.Name, decl.Type, decl.Body.List, nil)
		receiver types.Type
	)

	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return &fun, receiver, nil
	}

	var model = pkg.Types[receiverTypeName(decl.Recv.List[0].Type)]

	if model == nil {
		return nil, nil, fmt.Errorf("for now")
	}

	return &fun, model, nil
}

func receiverTypeName(f ast.Expr) string {
	switch typed := f.(type) {
	case *ast.StarExpr:
		return receiverTypeName(typed.X)
	case *ast.Ident:
		return typed.Name
	default:
		panic("for now")
	}
}
