package service

import (
	"fmt"
	"go/ast"
	"strings"

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

func (srv *Service) getReturnType(expr ast.Expr, i int) (types.Type, error) {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		return srv.returnTypeFromSelector(typed, i)
	case *ast.Ident:
		return srv.getReturnTypeFromIdent(*typed)
	case *ast.UnaryExpr:
		return srv.getReturnType(typed.X, i)
	case *ast.CompositeLit:
		return srv.getReturnType(typed.Type, i)
	case *ast.ArrayType:
		return srv.getReturnType(typed.Elt, i)
	case *ast.StarExpr:
		return srv.getReturnType(typed.X, i)
	case *ast.CallExpr:
		return srv.getReturnType(typed.Fun, i)
	case *ast.BasicLit:
		return types.NewString(typed), nil
	case *ast.StructType:
		return types.NewStruct(nil,
			fmt.Sprintf("ResponseBody-%d", len(srv.Components.Responses)),
			typed,
			nil,
		), nil
	default:
		return nil, nil
	}
}

func (srv *Service) getReturnTypeFromIdent(ident ast.Ident) (types.Type, error) {
	if ident.Obj == nil {
		return types.NewString(&ast.BasicLit{Value: "string"}), nil
	}

	switch decl := ident.Obj.Decl.(type) {
	case *ast.TypeSpec:
		panic("not handled")
		// return srv.parser.FindModelByName(ident.Name)
	case *ast.Field:
		return srv.getReturnType(decl.Type, 0)
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
				return srv.getReturnType(decl.Rhs[len(decl.Rhs)-1], i)
			}

			return srv.getReturnType(decl.Rhs[i], i)
		}
	}

	return nil, nil
}

func (srv *Service) returnTypeFromSelector(selector *ast.SelectorExpr, i int) (types.Type, error) {
	var objectName = selector.Sel.Name

	t, err := srv.getReturnType(selector.X, i)
	if err != nil {
		return nil, err
	}

	switch typed := t.(type) {
	case types.ImportedType:
		t, err = srv.parser.UnwrapImportedType(typed)
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
