package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"
)

func (parser *Parser) ParseRouters(file ast.File, prefix string, funcDecl ast.FuncDecl) error {
	err := CheckFuncDeclaration(
		funcDecl,
		"Routers",
		nil,
		[]func(ast.Expr) error{CheckRoutersResultType},
	)
	if err != nil {
		return err
	}

	for _, statement := range funcDecl.Body.List {
		returnStmt, ok := statement.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		if len(returnStmt.Results) == 0 {
			continue
		}

		compositeLit, ok := returnStmt.Results[0].(*ast.CompositeLit)
		if !ok {
			continue
		}

		if err := CheckRoutersResultType(compositeLit.Type); err != nil {
			return fmt.Errorf("'Routers' is of wrong return type: %w", err)
		}

		return parser.parseRoutes(file, prefix, compositeLit.Elts)
	}

	return nil
}

func CheckFuncDeclaration(
	funcDecl ast.FuncDecl,
	name string,
	params []func(ast.Expr) error,
	results []func(ast.Expr) error,
) error {
	if funcDecl.Name.Name != name {
		return fmt.Errorf("not a '%s' func", name)
	}

	if len(params) != len(funcDecl.Type.Params.List) {
		return fmt.Errorf("'%s' should have %d parameters", name, len(params))
	}

	for i, param := range params {
		if err := param(funcDecl.Type.Params.List[i].Type); err != nil {
			return err
		}
	}

	if len(results) != len(funcDecl.Type.Results.List) {
		return fmt.Errorf("'%s' should have %d results", name, len(params))
	}

	for i, param := range params {
		if err := param(funcDecl.Type.Results.List[i].Type); err != nil {
			return err
		}
	}

	return nil
}

func CheckRoutersResultType(resultType ast.Expr) error {
	mapType, ok := resultType.(*ast.MapType)
	if !ok {
		return fmt.Errorf("not a map")
	}

	ident, ok := mapType.Key.(*ast.Ident)
	if !ok || ident.Name != "string" {
		return fmt.Errorf("map's key is not a 'string'")
	}

	selector, ok := mapType.Value.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("map's value is of wrong type")
	}

	valuePackage, ok := selector.X.(*ast.Ident)
	if !ok {
		return fmt.Errorf("map's value package couldn't be defined")
	}

	if valuePackage.Name != "webapi" {
		return fmt.Errorf("wrong map's value package")
	}

	if selector.Sel.Name != "RouterByPath" {
		return fmt.Errorf("map's value is not a 'RouterByPath'")
	}

	return nil
}

func (parser *Parser) parseRoutes(file ast.File, servicePrefix string, expressions []ast.Expr) error {
	var apiPrefix = "/" + filepath.Join(parser.apiPrefix, servicePrefix)

	for _, expression := range expressions {
		keyValue, ok := expression.(*ast.KeyValueExpr)
		if !ok {
			return fmt.Errorf("not a key-value")
		}

		method, path, route, err := parser.ParseRoute(file, *keyValue)
		if err != nil {
			return err
		}

		route.Tags = append(route.Tags, servicePrefix)

		var resultPath = filepath.Join(apiPrefix, path)

		if _, ok := parser.file.Paths[resultPath]; !ok {
			parser.file.Paths[resultPath] = make(map[string]Route)
		}

		parser.file.Paths[resultPath][method] = *route
	}

	return nil
}
