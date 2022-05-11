package services

// import (
// 	"fmt"
// 	"go/ast"
// 	"strings"

// 	"github.com/KlyuchnikovV/webapi-docs/objects"
// )

// func (srv *Service) ParseParameter(route *objects.Route, argument ast.CallExpr) error {
// 	if len(argument.Args) == 0 {
// 		return fmt.Errorf("no arguments found")
// 	}

// 	selector, ok := argument.Fun.(*ast.SelectorExpr)
// 	if !ok {
// 		return fmt.Errorf("not a selector")
// 	}

// 	switch selector.Sel.Name {
// 	case "Body", "CustomBody":
// 		name, schema, err := srv.NewInBody(srv.currentFile, route, argument.Args[0])
// 		if err != nil {
// 			return err
// 		}

// 		srv.Components.Schemas[name] = schema
// 		srv.Components.RequestBodies[name] = objects.NewRequestBody(*objects.NewReference(name, "schemas"))

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

// 		srv.AddParameter(route, NewParameter(prefix, name, argument.Args))
// 	}

// 	return nil
// }

// func (srv *Service) NewInBody(file ast.File, route *objects.Route, arg ast.Expr) (string, objects.Schema, error) {
// 	identifier, typeSpec := GetTypeSpecification(file, arg)
// 	if typeSpec == nil {
// 		return identifier, nil, nil
// 	}

// 	schema, err := srv.Components.NewSchema(*typeSpec)

// 	return identifier, schema, err
// }

// func (srv *Service) AddParameter(route *objects.Route, param objects.IParameter) {
// 	var (
// 		name  string
// 		ok    = true
// 		saved objects.IParameter
// 	)

// 	for i := 0; ok; i++ {
// 		name = fmt.Sprintf("%s-%s-%d", param.NameParam(), param.Type(), i)
// 		saved, ok = srv.Components.Parameters[name]

// 		if ok && saved.EqualTo(param) {
// 			break
// 		}
// 	}

// 	route.Parameters = append(route.Parameters, objects.NewReference(name, "parameters"))

// 	if saved == nil {
// 		srv.Components.Parameters[name] = param
// 	}
// }
