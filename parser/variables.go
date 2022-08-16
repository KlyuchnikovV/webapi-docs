package parser

import (
	"fmt"
	"go/ast"
)

type Variable struct {
	Name string
	Type Type

	File *ast.File `json:"-"`
}

func NewVariables(decl *ast.AssignStmt, file *ast.File, pkgName string) (map[string]*Variable, error) {
	var vars map[string]*Variable
	if pkg, ok := Pkgs[pkgName]; ok {
		vars = pkg.Variables
	}

	for i, stmt := range decl.Lhs {
		ident, ok := stmt.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("not ident")
		}

		if ident.Name == "_" {
			continue
		}

		var expr ast.Expr
		if len(decl.Rhs) <= i {
			expr = decl.Rhs[len(decl.Rhs)-1]
		} else {
			expr = decl.Rhs[i]
		}

		v, err := NewVariable(ident.Name, expr, i, file, pkgName)
		if err != nil {
			continue
		}

		vars[v.Name] = v
	}

	return vars, nil
}

func NewVariable(name string, value ast.Expr, i int, file *ast.File, pkgName string) (*Variable, error) {
	// if v, ok := vars[name]; ok {
	// 	return v, nil
	// }

	t, err := TypeFromVariable(name, value, file, pkgName)
	if err != nil {
		return nil, err
	}

	if t, err = unpackTypeOfVar(t, i); err != nil {
		return nil, err
	}

	return &Variable{
		Name: name,
		Type: t,
		File: file,
	}, nil
}

func unpackTypeOfVar(t Type, i int) (Type, error) {
	switch typed := t.(type) {
	case *ImportedType:
		t, err := typed.Unpack()
		if err != nil {
			return nil, err
		}

		return unpackTypeOfVar(t, i)
	case *FuncDecl:
		return typed.Results[i], nil
	default:
		return t, nil
		// panic("not handled")
	}
}
