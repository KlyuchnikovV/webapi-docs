package service

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Parser interface {
	FindModel(ast.SelectorExpr) types.Type
	UnwrapImportedType(types.ImportedType) (types.Type, error)
}

type Service struct {
	parser Parser

	receiver      types.Type
	servicePrefix string

	Components types.Components
	Paths      map[string]map[string]types.Route
}

func New(parser Parser, receiver types.Type, prefix string) *Service {
	return &Service{
		parser:        parser,
		receiver:      receiver,
		Components:    types.NewComponents(),
		Paths:         make(map[string]map[string]types.Route),
		servicePrefix: prefix,
	}
}

func (srv *Service) Parse() error {
	var routers = srv.receiver.Method("Routers")
	if routers == nil {
		return fmt.Errorf("routers are nil")
	}

	if !routers.Implements(types.RoutersFuncInterface()) {
		return fmt.Errorf("routers are not implemented")
	}

	var returns = routers.ReturnStatements()
	if len(returns) != 1 && len(returns[0].Results) != 1 {
		return fmt.Errorf("routers not implemented")
	}

	composite, ok := returns[0].Results[0].(*ast.CompositeLit)
	if !ok {
		return fmt.Errorf("not composite")
	}

	for _, route := range composite.Elts {
		mapEntry, ok := route.(*ast.KeyValueExpr)
		if !ok {
			return fmt.Errorf("not key-value")
		}

		callExpr, ok := mapEntry.Value.(*ast.CallExpr)
		if !ok {
			return fmt.Errorf("value not a function")
		}

		var path = filepath.Join("/", srv.servicePrefix, extractPath(mapEntry.Key))
		if err := srv.parseRoute(path, callExpr); err != nil {
			return err
		}
	}

	return nil
}

func (srv *Service) parseRoute(path string, value *ast.CallExpr) error {
	var route = types.NewRoute(srv.servicePrefix)

	for _, arg := range value.Args {
		switch typed := arg.(type) {
		case *ast.CallExpr:
			if err := srv.ParseParameter(route, *typed); err != nil {
				return err
			}
		case *ast.SelectorExpr:
			var returns = srv.receiver.Method(typed.Sel.Name).ReturnStatements()

			if err := srv.getResponses(srv.servicePrefix, typed.Sel.Name, returns, route); err != nil {
				return err
			}
		}
	}

	if _, ok := srv.Paths[path]; !ok {
		srv.Paths[path] = make(map[string]types.Route)
	}

	srv.Paths[path][extractMethod(value.Fun)] = *route

	return nil
}
