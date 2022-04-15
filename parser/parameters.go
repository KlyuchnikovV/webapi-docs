package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type File struct {
	Openapi    string                            `json:"openapi"`
	Info       types.Info                        `json:"info"`
	Components types.Components                  `json:"components"`
	Paths      map[string]map[string]types.Route `json:"paths"`
}

func NewFile() *File {
	return &File{
		Openapi: "3.0.3",
		Info: types.Info{
			Version: "3.0.3",
		},
		Paths: make(map[string]map[string]types.Route),
		Components: types.Components{
			Schemas:       make(map[string]types.Schema),
			Parameters:    make(map[string]types.Parameter),
			RequestBodies: make(map[string]types.RequestBody),
		},
	}
}

func (f *File) ParseRoute(keyValue ast.KeyValueExpr) (string, string, *types.Route, error) {
	var route = types.NewRoute()

	callExpression, ok := keyValue.Value.(*ast.CallExpr)
	if !ok {
		return "", "", nil, fmt.Errorf("value not a function")
	}

	for i, arg := range callExpression.Args {
		switch argument := arg.(type) {
		case *ast.CallExpr:
			if err := f.ParseParameter(route, *argument); err != nil {
				return "", "", nil, err
			}
		case *ast.SelectorExpr:
			// TODO: name of functions to parse
			fmt.Printf("\t1arg %d is %#v\n", i, argument.Sel.Name)
		default:
			fmt.Printf("\t2arg %d is %#v\n", i, argument)
		}
	}

	return extractMethod(callExpression.Fun),
		extractPath(keyValue.Key),
		route, nil
}

func (f *File) ParseParameter(route *types.Route, argument ast.CallExpr) error {
	if len(argument.Args) == 0 {
		return fmt.Errorf("no arguments found")
	}

	selector, ok := argument.Fun.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("not a selector")
	}

	switch selector.Sel.Name {
	case "WithBody", "WithCustomBody":
		name, schema := types.NewInBody(argument.Args[0])

		f.Components.Schemas[name] = schema
		f.Components.RequestBodies[name] = types.NewRequestBody(*types.NewReference(name, "schemas"))

		route.RequestBody = types.NewReference(name, "requestBodies")
	default:
		var param = types.NewInQuery(selector.Sel.Name, argument.Args)

		f.Components.Parameters[param.Name] = param

		route.Parameters = append(route.Parameters, types.NewReference(param.Name, "parameters"))
	}

	return nil
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
