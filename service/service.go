package service

import (
	"fmt"
	"go/ast"
	"net/http"
	"path/filepath"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/cache/types"
	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/objects"
)

type Service struct {
	pkg types.Package

	receiver types.Type

	servicePrefix string

	Components objects.Components
	Paths      map[string]map[string]objects.Route
}

func New(pkg types.Package, receiver types.Type, prefix string) *Service {
	return &Service{
		pkg:           pkg,
		receiver:      receiver,
		Components:    objects.NewComponents(),
		Paths:         make(map[string]map[string]objects.Route),
		servicePrefix: prefix,
	}
}

func (srv *Service) Parse() error {
	var (
		returns = srv.receiver.Method("Routers").ReturnStatements()
	)

	for _, ret := range returns {
		for _, result := range ret.Results {
			composite, ok := result.(*ast.CompositeLit)
			if !ok {
				continue
			}

			for _, elt := range composite.Elts {
				keyValue, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				if err := srv.parseRoute(*keyValue); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (srv *Service) parseRoute(keyValue ast.KeyValueExpr) error {
	callExpr, ok := keyValue.Value.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("value not a function")
	}

	var route = objects.NewRoute(srv.servicePrefix)

	for _, arg := range callExpr.Args {
		switch typed := arg.(type) {
		case *ast.CallExpr:
			if err := srv.ParseParameter(route, *typed); err != nil {
				return err
			}
		case *ast.SelectorExpr:
			returns := srv.receiver.Method(typed.Sel.Name).ReturnStatements()

			if err := srv.getResponses(returns, route); err != nil {
				return err
			}
		}
	}

	var path = filepath.Join("/", srv.servicePrefix, extractPath(keyValue.Key))

	if _, ok := srv.Paths[path]; !ok {
		srv.Paths[path] = make(map[string]objects.Route)
	}

	srv.Paths[path][extractMethod(callExpr.Fun)] = *route

	return nil
}

func (srv *Service) getResponses(returns []ast.ReturnStmt, route *objects.Route) error {
	for _, returnStmt := range returns {
		for _, result := range returnStmt.Results {
			callExpr, ok := result.(*ast.CallExpr)
			if !ok {
				return fmt.Errorf("not a call expr")
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return fmt.Errorf("not a selector")
			}

			code, _ := constants.GetResultCode(selExpr.Sel.Name)
			if code == -1 {
				continue
			}

			if err := srv.extractResponse(code, callExpr.Args, route); err != nil {
				return err
			}
		}
	}

	return nil
}

func (srv *Service) extractResponse(code int, args []ast.Expr, route *objects.Route) error {
	if len(args) == 0 {
		args = append(args, nil)
	}

	for _, arg := range args {
		t, err := getReturnType(arg, 0)
		if err != nil {
			return err
		}

		if imp, ok := t.(*types.ImportedType); ok {
			t, err = cache.UnwrapImportedType(*imp)
			if err != nil {
				return err
			}
		}

		schema, err := srv.Components.NewSchema2(t)
		if err != nil {
			return err
		}

		var id string

		if schema == nil {
			id = fmt.Sprintf("nocontent%d", code)

			if _, ok := srv.Components.Responses[id]; !ok {
				srv.Components.Responses[id] = *objects.NewResponse(http.StatusText(code))
			}
		} else {
			id = t.Name()
			srv.Components.Schemas[id] = schema

			srv.Components.Responses[id] = *objects.NewResponse(
				http.StatusText(code), *objects.NewReference(id, "schemas"),
			)
		}

		route.Responses[code] = objects.NewReference(id, "responses")
	}

	return nil
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
	default:
		return nil, nil
	}
}

func getReturnTypeFromIdent(ident ast.Ident) (types.Type, error) {
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
	}

	field, ok := t.Fields()[objectName]
	if ok {
		return field, nil
	}

	var method = t.Method(objectName)
	if method == nil {
		panic("not ok")
	}

	if len(method.Results) == 0 {
		return nil, nil
	}

	if len(method.Results) <= i {
		return method.Results[len(method.Results)-1], nil
	}

	return method.Results[i], nil

}
