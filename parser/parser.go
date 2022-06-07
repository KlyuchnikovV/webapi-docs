package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types/pkg"
)

func Parse(path string) (pkg.Packages, *token.FileSet, error) {
	var (
		pkgs = make(pkg.Packages)
		fset = token.NewFileSet()
	)

	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	if !info.IsDir() {
		err = parseFile(path, pkgs, fset)
	} else {
		err = parseDir(path, pkgs, fset)
	}

	return pkgs, fset, err
}

func parseDir(path string, pkgs pkg.Packages, fset *token.FileSet) error {
	astPackages, err := parser.ParseDir(fset, path, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	for _, astPkg := range astPackages {
		tPkg, err := pkg.NewPackage(*astPkg)
		if err != nil {
			return err
		}

		pkgs[path] = *tPkg
	}

	list, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}

		if err := parseDir(filepath.Join(path, entry.Name()), pkgs, fset); err != nil {
			return err
		}
	}

	return nil
}

func parseFile(path string, pkgs pkg.Packages, fset *token.FileSet) error {
	astFile, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	var (
		slashIndex = strings.LastIndexByte(path, '/')
		name       = path[slashIndex+1:]
	)

	pkg, err := pkg.NewPackage(ast.Package{
		Name:  name,
		Files: map[string]*ast.File{astFile.Name.Name: astFile},
	})
	if err != nil {
		return err
	}

	pkgs[path] = *pkg

	return nil
}
