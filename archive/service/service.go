package service

import (
	"fmt"
	"go/ast"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/pkg"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Service struct {
	pkg pkg.Package

	receiver types.Type

	servicePrefix string

	Components types.Components
	Paths      map[string]map[string]types.Route
}

func New(pkg pkg.Package, receiver types.Type, prefix string) *Service {
	return &Service{
		pkg:           pkg,
		receiver:      receiver,
		Components:    types.NewComponents(),
		Paths:         make(map[string]map[string]types.Route),
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

	var route = types.NewRoute(srv.servicePrefix)

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
		srv.Paths[path] = make(map[string]types.Route)
	}

	srv.Paths[path][extractMethod(callExpr.Fun)] = *route

	return nil
}

func (srv *Service) getResponses(returns []ast.ReturnStmt, route *types.Route) error {
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

			// TODO: templating
			code, _ := constants.GetResultCode(selExpr.Sel.Name)
			if code == -1 {
				continue
			}

			srv.defineResponse(selExpr.Sel.Name, callExpr.Args, route)
		}
	}

	return nil
}

func (srv *Service) objectResponse(code int, args []ast.Expr, route *types.Route) error {
	t, err := getReturnType(args[0], 0)
	if err != nil {
		return err
	}

	switch typed := t.(type) {
	case *types.ImportedType:
		t, err = cache.UnwrapImportedType(*typed)
		if err != nil {
			return err
		}
	case nil:
		return nil
	}

	var id = t.Name()

	srv.Components.Responses[id] = *types.NewResponse(
		http.StatusText(code), *types.NewReference(id, "schemas"),
	)

	srv.Components.Schemas[id] = t.Schema()
	route.Responses[code] = types.NewReference(id, "responses")

	return nil
}

func (srv *Service) defineResponse(fun string, args []ast.Expr, route *types.Route) {
	var code, _ = constants.GetResultCode(fun)

	switch fun {
	case "Created", "NoContent":
		var id = fmt.Sprintf("nocontent%d", code)

		if _, ok := srv.Components.Responses[id]; !ok {
			srv.Components.Responses[id] = *types.NewResponse(http.StatusText(code))
		}

		route.Responses[code] = types.NewReference(id, "responses")
	case "OK":
		srv.objectResponse(code, args, route)
	case "InternalServerError", "BadRequest",
		"Forbidden", "MethodNotAllowed", "NotFound":
		srv.errorResponse(code, args, route)
	}
}

func (srv *Service) errorResponse(code int, args []ast.Expr, route *types.Route) error {
	t, err := getReturnType(args[0], 0)
	if err != nil {
		return err
	}

	var desc string

	switch typed := t.(type) {
	case *types.ImportedType:
		t, err = cache.UnwrapImportedType(*typed)
		if err != nil {
			return err
		}
	case types.StringType:
		desc = typed.Data
	case nil:
		return nil
	}

	var id = fmt.Sprintf("%s-%s", t.Name(), desc)

	if r, ok := srv.Components.Responses[id]; ok {
		r.Description = strings.Join([]string{r.Description, desc}, ", ")
	} else {
		srv.Components.Responses[id] = *types.NewErrorResponse(
			fmt.Sprintf("%s: %s", http.StatusText(code), desc), *types.NewReference(id, "schemas"),
		)

		srv.Components.Schemas[id] = t.Schema()
	}

	route.Responses[code] = types.NewReference(id, "responses")

	return nil
}
