package services

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/utils"
)

func getReturnType(funcDecl ast.FuncDecl) (*ast.TypeSpec, error) {
	if len(funcDecl.Type.Results.List) == 0 {
		return nil, fmt.Errorf("no results")
	}

	var typeObj *ast.Object

	switch typed := funcDecl.Type.Results.List[0].Type.(type) {
	case *ast.StarExpr:
		ident, ok := typed.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("pointer not ident")
		}

		typeObj = ident.Obj
	case *ast.SelectorExpr:
		typeObj = typed.Sel.Obj
	default:
		return nil, fmt.Errorf("can't parse %#v", typed)
	}

	if typeObj == nil {
		return nil, fmt.Errorf("no type")
	}

	typeSpec, ok := typeObj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("not a type")
	}

	return typeSpec, nil
}

// TODO: merge with previous
func GetPrefix(statements []ast.Stmt) string {
	var result string

	for _, stmt := range statements {
		ast.Inspect(stmt, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if !IsMethod(*callExpr, NewSelector("webapi", "NewService")) {
				return true
			}

			if len(callExpr.Args) < 2 {
				return true
			}

			lit, ok := callExpr.Args[1].(*ast.BasicLit)
			if !ok {
				return true
			}

			result = strings.Trim(lit.Value, "\"")

			return false
		})
	}

	return result
}

func IsMethod(call ast.CallExpr, selector ast.SelectorExpr) bool {
	callSel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	for ok {
		var typed *ast.SelectorExpr
		typed, ok = callSel.X.(*ast.SelectorExpr)

		if ok {
			callSel = typed
		}
	}

	return utils.SameNodes(callSel, &selector)
}

func NewSelector(prefix, fun string) ast.SelectorExpr {
	return ast.SelectorExpr{
		X:   &ast.Ident{Name: prefix},
		Sel: &ast.Ident{Name: fun},
	}
}

func GetReturnStatements(typeSpec ast.TypeSpec, method ast.FuncDecl) []ast.ReturnStmt {
	funcDecl := cache.FindMethod(
		typeSpec,
		method,
	)
	if funcDecl == nil {
		return nil
	}

	var returns = make([]ast.ReturnStmt, 0)

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		returnStmt, ok := n.(*ast.ReturnStmt)
		if ok {
			returns = append(returns, *returnStmt)
		}
		return true
	})

	return returns
}

// Retrieves only global specs
func GetTypeSpecification(file ast.File, arg ast.Expr) (string, *ast.TypeSpec) {
	var expr ast.Expr

	switch typed := arg.(type) {
	case *ast.UnaryExpr:
		expr = typed.X
	case *ast.CompositeLit:
		expr = typed.Type
	case *ast.SelectorExpr:
		// spec := cache.FindModel(file, *typed)
		// if spec == nil {
		// 	return GetTypeSpecification(file, typed.X)
		// }

		// if spec == nil {
		// 	_, baseType := GetTypeSpecification(file, typed.X)
		// 	if baseType == nil {
		// 		return typed.Sel.Name, nil
		// 	}

		// 	funcDecl := cache.FindMethodByName(*baseType, typed.Sel.Name)
		// 	if funcDecl == nil {
		// 		return typed.Sel.Name, nil
		// 	}

		// 	return GetTypeSpecification(file, funcDecl.Type.Results.List[0].Type)
		// }

		// return typed.Sel.Name, spec
	case *ast.Ident:
		if typed.Obj == nil {
			return "", nil
		}

		switch decl := typed.Obj.Decl.(type) {
		case *ast.TypeSpec:
			return typed.Name, decl
		case *ast.Field:
			return GetTypeSpecification(file, decl.Type)
		default:
			panic("not ok")
		}
	case *ast.ArrayType:
		expr = typed.Elt
	case *ast.StarExpr:
		return GetTypeSpecification(file, typed.X)
	case *ast.CallExpr:
		return GetTypeSpecification(file, typed.Fun)
	default:
		panic(fmt.Sprintf("unknown type %#v", typed))
	}

	return GetTypeSpecification(file, expr)
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

func ModelFromSelector(selector ast.SelectorExpr) {

}

func GetTypeSpecification2(arg ast.Expr) (string, *ast.TypeSpec) {
	var expr ast.Expr

	switch typed := arg.(type) {
	case *ast.UnaryExpr:
		expr = typed.X
	case *ast.CompositeLit:
		expr = typed.Type
	case *ast.SelectorExpr:
		// spec, err := cache.FindModelByName(typed.Sel.Name)
		// if err != nil {
		// 	_, baseType := GetTypeSpecification2(typed.X)
		// 	if baseType == nil {
		// 		return typed.Sel.Name, nil
		// 	}

		// 	funcDecl := cache.FindMethodByName(*baseType, typed.Sel.Name)
		// 	if funcDecl == nil {
		// 		return typed.Sel.Name, nil
		// 	}

		// 	return GetTypeSpecification2(funcDecl.Type.Results.List[0].Type)
		// }

		// typeSpec, ok := spec.Decl.(*ast.TypeSpec)
		// if !ok {
		return "", nil
		// }

		// return typed.Sel.Name, typeSpec
	case *ast.Ident:
		if typed.Obj == nil {
			return "", nil
		}

		switch decl := typed.Obj.Decl.(type) {
		case *ast.TypeSpec:
			return typed.Name, decl
		case *ast.Field:
			return GetTypeSpecification2(decl.Type)
		default:
			panic("not ok")
		}
	case *ast.ArrayType:
		expr = typed.Elt
	case *ast.StarExpr:
		return GetTypeSpecification2(typed.X)
	case *ast.CallExpr:
		return GetTypeSpecification2(typed.Fun)
	default:
		panic(fmt.Sprintf("unknown type %#v", typed))
	}

	return GetTypeSpecification2(expr)
}
