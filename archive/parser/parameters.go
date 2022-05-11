package parser

// import (
// 	"fmt"
// 	"go/ast"
// 	"strings"

// 	"github.com/KlyuchnikovV/webapi-docs/objects"
// )

// func (parser *Parser) ParseParameter(file ast.File, route *objects.Route, argument ast.CallExpr) error {
// 	if len(argument.Args) == 0 {
// 		return fmt.Errorf("no arguments found")
// 	}

// 	selector, ok := argument.Fun.(*ast.SelectorExpr)
// 	if !ok {
// 		return fmt.Errorf("not a selector")
// 	}

// 	switch selector.Sel.Name {
// 	case "Body", "CustomBody":
// 		name, schema, err := parser.NewInBody(file, argument.Args[0])
// 		if err != nil {
// 			return err
// 		}

// 		parser.Spec.Components.Schemas[name] = schema
// 		parser.Spec.Components.RequestBodies[name] = objects.NewRequestBody(*objects.NewReference(name, "schemas"))

// 		route.RequestBody = objects.NewReference(name, "requestBodies")
// 	default:
// 		var (
// 			prefix string
// 			name   string
// 		)

// 		if strings.HasPrefix(selector.Sel.Name, "Query") {
// 			prefix = "query"
// 			name = selector.Sel.Name[strings.Index(selector.Sel.Name, "Query")+len("Query"):]
// 		} else if strings.HasPrefix(selector.Sel.Name, "InPath") {
// 			prefix = "path"
// 			name = selector.Sel.Name[strings.Index(selector.Sel.Name, "InPath")+len("InPath"):]
// 		}

// 		parser.AddParameter(route, NewParameter(prefix, name, argument.Args))
// 	}

// 	return nil
// }

// func (parser *Parser) NewInBody(file ast.File, arg ast.Expr) (string, Schema, error) {
// 	identifier, typeSpec := parser.getTypeSpecification(file, arg)
// 	if typeSpec == nil {
// 		return identifier, nil, nil
// 	}

// 	schema, err := parser.NewSchema(typeSpec)

// 	return identifier, schema, err
// }

// func (parser *Parser) AddParameter(route *objects.Route, param IParameter) {
// 	var (
// 		name  string
// 		ok    = true
// 		saved IParameter
// 	)

// 	for i := 0; ok; i++ {
// 		name = fmt.Sprintf("%s-%s-%d", param.NameParam(), param.Type(), i)
// 		saved, ok = parser.Spec.Components.Parameters[name]

// 		if ok && saved.EqualTo(param) {
// 			break
// 		}
// 	}

// 	route.Parameters = append(route.Parameters, parser.NewReference(name, "parameters"))

// 	if saved == nil {
// 		parser.Spec.Components.Parameters[name] = param
// 	}
// }

// func (parser *Parser) getTypeSpecification(file ast.File, arg ast.Expr) (string, *ast.TypeSpec) {
// 	var expr ast.Expr

// 	switch typed := arg.(type) {
// 	case *ast.UnaryExpr:
// 		expr = typed.X
// 	case *ast.CompositeLit:
// 		expr = typed.Type
// 	case *ast.SelectorExpr:
// 		return typed.Sel.Name, parser.findModel(file, *typed)
// 	case *ast.Ident:
// 		typeSpec, ok := typed.Obj.Decl.(*ast.TypeSpec)
// 		if !ok {
// 			panic("not ok")
// 		}

// 		return typed.Name, typeSpec
// 	case *ast.ArrayType:
// 		expr = typed.Elt
// 	default:
// 		panic(fmt.Sprintf("unknown type %#v", typed))
// 	}

// 	return parser.getTypeSpecification(file, expr)
// }

// func (parser *Parser) findModel(file ast.File, selector ast.SelectorExpr) *ast.TypeSpec {
// 	ident, ok := selector.X.(*ast.Ident)
// 	if !ok {
// 		return nil
// 	}

// 	var pkg = parser.getPackage(file, ident.Name)
// 	if pkg == nil {
// 		return nil
// 	}

// 	for _, file := range pkg.Files {
// 		obj := file.Scope.Lookup(selector.Sel.Name)
// 		if obj != nil {
// 			return obj.Decl.(*ast.TypeSpec)
// 		}
// 	}

// 	return nil
// }

// func (parser *Parser) getPackage(file ast.File, alias string) *Package {
// 	var path string

// 	for _, imp := range file.Imports {
// 		if imp.Name != nil {
// 			if imp.Name.Name == alias {
// 				path = strings.Trim(imp.Path.Value, "\"")
// 				break
// 			}

// 			continue
// 		}

// 		if strings.HasSuffix(strings.Trim(imp.Path.Value, "\""), alias) {
// 			path = strings.Trim(imp.Path.Value, "\"")
// 			break
// 		}
// 	}

// 	pkg, ok := parser.packages[path]
// 	if ok {
// 		return pkg
// 	}

// 	if err := parser.ImportPackage(path); err != nil {
// 		panic(err)
// 	}

// 	pkg, ok = parser.packages[path]
// 	if ok {
// 		return pkg
// 	}

// 	parser.notFoundImports = append(parser.notFoundImports, path)

// 	return nil
// }
