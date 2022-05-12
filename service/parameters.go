package service

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/cache"
	"github.com/KlyuchnikovV/webapi-docs/types"
	"github.com/KlyuchnikovV/webapi-docs/utils"
)

func (srv *Service) ParseParameter(route *types.Route, argument ast.CallExpr) error {
	if len(argument.Args) == 0 {
		return fmt.Errorf("no arguments found")
	}

	selector, ok := argument.Fun.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("not a selector")
	}

	switch selector.Sel.Name {
	case "Body", "CustomBody":
		name, schema := srv.NewInBody(route, argument.Args[0])
		if schema == nil {
			panic("nil")
		}

		srv.Components.Schemas[name] = schema
		srv.Components.RequestBodies[name] = types.NewRequestBody(*types.NewReference(name, "schemas"))

		route.RequestBody = types.NewReference(name, "requestBodies")
	default:
		var (
			prefix string
			name   string
		)

		if strings.HasPrefix(selector.Sel.Name, "Query") {
			prefix = "query"
			name = selector.Sel.Name[strings.Index(selector.Sel.Name, "Query")+len("Query"):]
		} else if strings.HasPrefix(selector.Sel.Name, "InPath") {
			prefix = "path"
			name = selector.Sel.Name[strings.Index(selector.Sel.Name, "InPath")+len("InPath"):]
		}

		srv.AddParameter(route, NewParameter(prefix, name, argument.Args))
	}

	return nil
}

func (srv *Service) NewInBody(route *types.Route, arg ast.Expr) (string, types.Schema) {
	var selector ast.SelectorExpr

	ast.Inspect(arg, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			selector = *sel
			return false
		}
		return true
	})

	_, imp := utils.FindImport(*srv.receiver.File(), selector.X.(*ast.Ident).Name)
	selector.X.(*ast.Ident).Name = strings.Trim(imp.Path.Value, "\"")

	model := cache.FindModel(selector)
	if model == nil {
		return "", nil
	}

	return model.Name(), model.Schema()
}

func (srv *Service) AddParameter(route *types.Route, param types.IParameter) {
	var (
		name  string
		ok    = true
		saved types.IParameter
	)

	for i := 0; ok; i++ {
		name = fmt.Sprintf("%s-%s-%d", param.NameParam(), param.Type(), i)
		saved, ok = srv.Components.Parameters[name]

		if ok && saved.EqualTo(param) {
			break
		}
	}

	route.Parameters = append(route.Parameters, types.NewReference(name, "parameters"))

	if saved == nil {
		srv.Components.Parameters[name] = param
	}
}
