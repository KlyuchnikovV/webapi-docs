package service

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/cache/types"
)

func GetTypeSpecification(arg ast.Expr) (string, types.Type) {
	var expr ast.Expr

	switch typed := arg.(type) {
	case *ast.UnaryExpr:
		expr = typed.X
	case *ast.CompositeLit:
		expr = typed.Type
	case *ast.SelectorExpr:
		spec := cache.FindModel(*typed)
		// if spec == nil {
		// 	_, spec = GetTypeSpecification(typed.X)
		// }

		if spec == nil {
			_, baseType := GetTypeSpecification(typed.X)
			if baseType == nil {
				return typed.Sel.Name, nil
			}

			funcDecl := baseType.Method(typed.Sel.Name)
			return funcDecl.Results[0].Name(), funcDecl.Results[0]
		}

		return typed.Sel.Name, spec
	case *ast.Ident:
		if typed.Obj == nil {
			return "", nil
		}

		switch decl := typed.Obj.Decl.(type) {
		case *ast.TypeSpec:
			model, err := cache.FindModelByName(typed.Name)
			if err != nil {
				panic(err)
			}
			return typed.Name, model
		case *ast.Field:
			return GetTypeSpecification(decl.Type)
		case *ast.AssignStmt:
			for i, variable := range decl.Lhs {
				ident, ok := variable.(*ast.Ident)
				if !ok {
					continue
				}

				if ident.Name == typed.Name {
					return GetTypeSpecification(decl.Rhs[i])
				}
			}

			return "", nil
		default:
			panic("not ok")
		}
	case *ast.ArrayType:
		expr = typed.Elt
	case *ast.StarExpr:
		return GetTypeSpecification(typed.X)
	case *ast.CallExpr:
		return GetTypeSpecification(typed.Fun)
	default:
		panic(fmt.Sprintf("unknown type %#v", typed))
	}

	return GetTypeSpecification(expr)
}

func extractMethod(arg ast.Expr) string {
	if sel, ok := arg.(*ast.SelectorExpr); ok {
		return strings.ToLower(sel.Sel.Name)
	}

	return ""
}

func extractPath(arg ast.Expr) string {
	if sel, ok := arg.(*ast.BasicLit); ok {
		return fmt.Sprintf("/%s", strings.Trim(sel.Value, "\"/"))
	}

	return ""
}
