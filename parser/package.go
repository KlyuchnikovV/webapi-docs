package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Package struct {
	Name string
	Path string

	Types     map[string]types.Type
	Functions map[string]types.FuncType

	pkg *ast.Package
}

func NewPackage(pkg *ast.Package, path string) (*Package, error) {
	var result = Package{
		Name:      pkg.Name,
		Types:     make(map[string]types.Type),
		Functions: make(map[string]types.FuncType),

		pkg: pkg,
	}

	for name, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, t := range newFromGenDecl(file, typed) {
					result.Types[t.Name()] = t
				}
			case *ast.FuncDecl:
				fun, receiver, err := newFromFuncDecl(result, file, typed)
				if err != nil {
					return nil, err
				}

				if receiver != nil {
					result.Types[receiver.Name()].AddMethod(*fun)
				} else {
					result.Functions[fun.Name()] = *fun
				}
			default:
				return nil, fmt.Errorf("unknown type: %s - %#v", name, typed)
			}
		}
	}

	return &result, nil
}

func (pkg *Package) FindModelByName(
	name string,
	unwrapper func(types.ImportedType) (types.Type, error),
) (types.Type, error) {
	var obj types.Type

	for _, t := range pkg.Types {
		if t.Name() == name {
			obj = t
			break
		}
	}

	if obj == nil {
		return nil, fmt.Errorf("no model found for name: '%s'", name)
	}

	if imp, ok := obj.(types.ImportedType); ok {
		return unwrapper(imp)
	}

	return obj, nil
}

func (pkg *Package) GetServices() map[string]types.Type {
	var result = make(map[string]types.Type)

	for i, t := range pkg.Types {
		if !t.Implements(types.RoutersInterface()) {
			continue
		}

		var prefix = t.Method("Prefix")
		if prefix == nil {
			continue
		}

		var stmts = prefix.ReturnStatements()
		if len(stmts) != 1 && len(stmts[0].Results) != 1 {
			continue
		}

		lit, ok := stmts[0].Results[0].(*ast.BasicLit)
		if !ok {
			continue
		}

		result[strings.Trim(lit.Value, "\"")] = pkg.Types[i]
	}

	return result
}
