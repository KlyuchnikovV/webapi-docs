package parser

import (
	"fmt"
	"go/ast"
	"strings"
)

type SwaggerSpec struct {
	Openapi    string                      `json:"openapi"`
	Info       Info                        `json:"info"`
	Servers    []Server                    `json:"servers"`
	Components Components                  `json:"components"`
	Paths      map[string]map[string]Route `json:"paths"`
}

func NewSwaggerSpec(servers ...Server) *SwaggerSpec {
	return &SwaggerSpec{
		Openapi: "3.0.3",
		Info: Info{
			Version: "3.0.3",
		},
		Servers: servers,
		Paths:   make(map[string]map[string]Route),
		Components: Components{
			Schemas:       make(map[string]Schema),
			Parameters:    make(map[string]IParameter),
			RequestBodies: make(map[string]RequestBody),
		},
	}
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
		default:
			fmt.Printf("\t2arg %d is %#v\n", i, argument)
		}
	}

	return extractMethod(callExpression.Fun),
		extractPath(keyValue.Key),
		route, nil
}

func (parser *Parser) ParseParameter(file ast.File, route *Route, argument ast.CallExpr) error {
	if len(argument.Args) == 0 {
		return fmt.Errorf("no arguments found")
	}

	selector, ok := argument.Fun.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("not a selector")
	}

	switch selector.Sel.Name {
	case "Body", "CustomBody":
		name, schema := parser.NewInBody(file, argument.Args[0])

		parser.Spec.Components.Schemas[name] = schema
		parser.Spec.Components.RequestBodies[name] = NewRequestBody(*NewReference(name, "schemas"))

		route.RequestBody = NewReference(name, "requestBodies")
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

		parser.AddParameter(route, NewParameter(prefix, name, argument.Args))
	}

	return nil
}

func (parser *Parser) NewInBody(file ast.File, arg ast.Expr) (string, Schema) {
	identifier, typeSpec := parser.getTypeSpecification(file, arg)
	if typeSpec == nil {
		return identifier, nil
	}

	switch typed := typeSpec.Type.(type) {
	case *ast.StructType:
		return identifier, parser.NewObject(*typed)
	case *ast.ArrayType:
		return identifier, parser.NewArray(*typed)
	}

	return "", nil
}

func (parser *Parser) getTypeSpecification(file ast.File, arg ast.Expr) (string, *ast.TypeSpec) {
	var expr ast.Expr

	switch typed := arg.(type) {
	case *ast.UnaryExpr:
		expr = typed.X
	case *ast.CompositeLit:
		expr = typed.Type
	case *ast.SelectorExpr:
		return typed.Sel.Name, parser.findModel(file, *typed)
	case *ast.Ident:
		typeSpec, ok := typed.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			panic("not ok")
		}

		return typed.Name, typeSpec
	case *ast.ArrayType:
		expr = typed.Elt
	default:
		panic(fmt.Sprintf("unknown type %#v", typed))
	}

	return parser.getTypeSpecification(file, expr)
}

func (parser *Parser) findModel(file ast.File, selector ast.SelectorExpr) *ast.TypeSpec {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil
	}

	var pkg = parser.getPackage(file, ident.Name)
	if pkg == nil {
		return nil
	}

	for _, file := range pkg.Files {
		obj := file.Scope.Lookup(selector.Sel.Name)
		if obj != nil {
			return obj.Decl.(*ast.TypeSpec)
		}
	}

	return nil
}

func (parser *Parser) getPackage(file ast.File, alias string) *ast.Package {
	var path string

	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == alias {
				path = strings.Trim(imp.Path.Value, "\"")
				break
			}

			continue
		}

		if strings.HasSuffix(strings.Trim(imp.Path.Value, "\""), alias) {
			path = strings.Trim(imp.Path.Value, "\"")
			break
		}
	}

	pkg, ok := parser.packages[path]
	if ok {
		return &pkg
	}

	if err := parser.ImportPackage(path); err != nil {
		panic(err)
	}

	pkg, ok = parser.packages[path]
	if ok {
		return &pkg
	}

	parser.notFoundImports = append(parser.notFoundImports, path)

	return nil
}
