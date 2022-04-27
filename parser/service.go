package parser

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

func (parser *Parser) ParseServices() error {
	for _, selector := range parser.services {
		if selector.Sel == nil {
			continue
		}

		for name, pkg := range parser.packages {
			for _, file := range pkg.Files {
				obj := file.Scope.Lookup(selector.Sel.Name)
				if obj == nil {
					continue
				}

				funcDecl, ok := obj.Decl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if err := parser.ParseService(name, *file, *funcDecl); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (parser *Parser) ParseService(pkg string, file ast.File, funcDecl ast.FuncDecl) error {
	typeSpec, err := parser.GetReturnType(funcDecl)
	if err != nil {
		return err
	}

	var alias = parser.findWebapiImport(file)

	if !parser.Implements(pkg, *typeSpec, types.ServiceInterface(alias)) {
		return nil
	}

	return parser.ParseRouters(
		file,
		parser.GetPrefix(funcDecl.Body.List),
		*parser.FindMethod(pkg, *typeSpec, *types.RoutersFuncDecl(alias)),
	)
}

func (parser *Parser) ParseRouters(file ast.File, apiPrefix string, funcDecl ast.FuncDecl) error {
	if err := CheckFuncDeclaration(funcDecl, "Routers", nil, CheckRoutersResultType); err != nil {
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

		return parser.parseRoutes(file, apiPrefix, compositeLit.Elts)
	}

	return nil
}

func (parser *Parser) GetReturnType(funcDecl ast.FuncDecl) (*ast.TypeSpec, error) {
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

func (parser *Parser) GetPrefix(stmts []ast.Stmt) string {
	var result string

	for _, stmt := range stmts {
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

		if _, ok := parser.Spec.Paths[resultPath]; !ok {
			parser.Spec.Paths[resultPath] = make(map[string]Route)
		}

		parser.Spec.Paths[resultPath][method] = *route
	}

	return nil
}

func (parser *Parser) ParseRoute(file ast.File, keyValue ast.KeyValueExpr) (string, string, *Route, error) {
	var route = NewRoute()

	callExpression, ok := keyValue.Value.(*ast.CallExpr)
	if !ok {
		return "", "", nil, fmt.Errorf("value not a function")
	}

	for i, arg := range callExpression.Args {
		switch argument := arg.(type) {
		case *ast.CallExpr:
			if err := parser.ParseParameter(file, route, *argument); err != nil {
				return "", "", nil, err
			}
		case *ast.SelectorExpr:
			// TODO: name of functions to parse
			fmt.Printf("\t1arg %d is %#v\n", i, argument.Sel.Name)
		}
	}

	return extractMethod(callExpression.Fun),
		extractPath(keyValue.Key),
		route, nil
}
