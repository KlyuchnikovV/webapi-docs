package service

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

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

func getReturnType(expr ast.Expr, i int) (types.Type, error) {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		return returnTypeFromSelector(typed, i)
	case *ast.Ident:
		return getReturnTypeFromIdent(*typed)
	case *ast.UnaryExpr:
		return getReturnType(typed.X, i)
	case *ast.CompositeLit:
		return getReturnType(typed.Type, i)
	case *ast.ArrayType:
		return getReturnType(typed.Elt, i)
	case *ast.StarExpr:
		return getReturnType(typed.X, i)
	case *ast.CallExpr:
		return getReturnType(typed.Fun, i)
	case *ast.BasicLit:
		return types.NewString(typed), nil
	default:
		return nil, nil
	}
}

func getReturnTypeFromIdent(ident ast.Ident) (types.Type, error) {
	if ident.Obj == nil {
		return types.NewString(&ast.BasicLit{Value: "string"}), nil
	}

	switch decl := ident.Obj.Decl.(type) {
	case *ast.TypeSpec:
		return cache.FindModelByName(ident.Name)
	case *ast.Field:
		return getReturnType(decl.Type, 0)
	case *ast.AssignStmt:
		for i, variable := range decl.Lhs {
			v, ok := variable.(*ast.Ident)
			if !ok {
				continue
			}

			if ident.Name != v.Name {
				continue
			}

			if len(decl.Rhs) <= i {
				return getReturnType(decl.Rhs[len(decl.Rhs)-1], i)
			}

			return getReturnType(decl.Rhs[i], i)
		}
	}

	return nil, nil
}

func returnTypeFromSelector(selector *ast.SelectorExpr, i int) (types.Type, error) {
	var objectName = selector.Sel.Name

	t, err := getReturnType(selector.X, i)
	if err != nil {
		return nil, err
	}

	switch typed := t.(type) {
	case *types.ImportedType:
		t, err = cache.UnwrapImportedType(*typed)
		if err != nil {
			return nil, err
		}
	case types.BasicType:
		return typed, nil
	case nil:
		return nil, nil
	}

	if field := t.Field(objectName); field != nil {
		return field, nil
	}

	var method = t.Method(objectName)
	if method == nil {
		return t, nil
	}

	if len(method.Results) == 0 {
		return nil, nil
	}

	if len(method.Results) <= i {
		return method.Results[len(method.Results)-1], nil
	}

	return method.Results[i], nil
}
