package service

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

const (
	webapiPath    = "github.com/KlyuchnikovV/webapi"
	optionsPath   = "github.com/KlyuchnikovV/webapi/options"
	parameterPath = "github.com/KlyuchnikovV/webapi/parameter"
)

// func OperatorDescriptionHandler(file *ast.File) types.ParameterOptionHandler {
// 	var orDeclaration = types.NewSimpleImported("OR", optionsPath)

// 	return func(parameter types.Type, call types.Call) error {
// 		for _, param := range call.Parameters {
// 			paramCall, ok := param.(types.Call)
// 			if !ok {
// 				continue
// 			}

// 			if !paramCall.Call.EqualTo(orDeclaration) {
// 				continue
// 			}

// 			if len(paramCall.Parameters) < 1 {
// 				return fmt.Errorf("option 'Description' has no argument")
// 			}

// 			description, ok := paramCall.Parameters[0].(types.BasicType)
// 			if !ok {
// 				return fmt.Errorf("option 'Description' isn't string")
// 			}

// 			parameter.SetDescription(description.Value)

// 			parameter.SetDescription(strings.Trim(
// 				fmt.Sprintf("%s %s",
// 					parameter.GetDescription(),
// 					fmt.Sprintf("Must be: %s", strings.Join(args, " or ")),
// 				), " "),
// 			)

// 			break
// 		}

// 		return nil
// 	}
// }

// func ParameterDescriptionHandler(file *ast.File) types.ParameterOptionHandler {
// 	var descDeclaration = types.NewSimpleImported("Description", optionsPath)

// 	return func(parameter types.Type, call types.Call) error {
// 		for _, param := range call.Parameters {
// 			paramCall, ok := param.(types.Call)
// 			if !ok {
// 				continue
// 			}

// 			if !paramCall.Call.EqualTo(descDeclaration) {
// 				continue
// 			}

// 			if len(paramCall.Parameters) < 1 {
// 				return fmt.Errorf("option 'Description' has no argument")
// 			}

// 			description, ok := paramCall.Parameters[0].(types.BasicType)
// 			if !ok {
// 				return fmt.Errorf("option 'Description' isn't string")
// 			}

// 			parameter.SetDescription(description.Value)

// 			break
// 		}

// 		return nil
// 	}
// }

func RouteDescriptionHandler(file *ast.File) types.RouteOptionHanlder {
	var descDeclaration = types.NewSimpleImported("Description", parameterPath)

	return func(route *types.Route, call types.Call) error {
		if !call.Call.EqualTo(descDeclaration) {
			return nil
		}

		if len(call.Parameters) < 1 {
			return fmt.Errorf("option 'Description' has no argument")
		}

		description, ok := call.Parameters[0].(types.BasicType)
		if !ok {
			return fmt.Errorf("option 'Description' isn't string")
		}

		route.Description = strings.Trim(description.Value, `"`)

		return nil
	}
}

func BodyHandler(srv *Service, handlers ...types.ParameterOptionHandler) types.RouteOptionHanlder {
	var (
		bodyDeclaration       = types.NewSimpleImported("Body", parameterPath)
		customBodyDeclaration = types.NewSimpleImported("CustomBody", parameterPath)
	)

	return func(route *types.Route, call types.Call) error {
		if !call.Call.EqualTo(bodyDeclaration) && !call.Call.EqualTo(customBodyDeclaration) {
			return nil
		}

		if len(call.Parameters) < 1 {
			return fmt.Errorf("option 'Body' has no argument")
		}

		name, schema := srv.NewInBody(route, call.Parameters[0])
		if schema == nil {
			return fmt.Errorf("nil schema")
		}

		for _, handler := range handlers {
			if err := handler(schema, call); err != nil {
				return err
			}
		}

		srv.Components.Schemas[name] = schema
		srv.Components.RequestBodies[name] = types.NewRequestBody(*types.NewReference(name, "schemas"))

		route.RequestBody = types.NewReference(name, "requestBodies")

		return nil
	}
}

func ParameterHandler(srv *Service, handlers ...types.ParameterOptionHandler) types.RouteOptionHanlder {
	var params = []types.Type{
		types.NewSimpleImported("Bool", parameterPath),
		types.NewSimpleImported("Time", parameterPath),
		types.NewSimpleImported("Float", parameterPath),
		types.NewSimpleImported("String", parameterPath),
		types.NewSimpleImported("Integer", parameterPath),
	}

	return func(route *types.Route, call types.Call) error {
		var isFound bool

		for _, t := range params {
			if call.Call.EqualTo(t) {
				isFound = true
			}
		}

		if !isFound {
			return nil
		}

		if len(call.Parameters) < 2 {
			return fmt.Errorf("option '%s' has not enough arguments", call.Call.Name())
		}

		var (
			prefix  string
			placing = call.Parameters[1].Name()
		)

		switch placing {
		case "InQuery":
			prefix = "query"
		case "InPath":
			prefix = "path"
			// TODO: add more options
		}

		for _, handler := range handlers {
			if err := handler(call.Call, call); err != nil {
				return err
			}
		}

		// srv.AddParameter(route, NewParameter(prefix, call.Call.Name(), call.Decl().Args))
		srv.AddParameter(route, NewParameter(prefix, call))

		return nil
	}
}

// func (srv *Service) ParseParameter(route *types.Route, argument ast.CallExpr) error {
// 	if len(argument.Args) == 0 {
// 		return fmt.Errorf("no arguments found")
// 	}

// 	selector, ok := argument.Fun.(*ast.SelectorExpr)
// 	if !ok {
// 		return fmt.Errorf("not a selector")
// 	}

// 	switch selector.Sel.Name {
// 	case "Body", "CustomBody":
// 		// name, schema := srv.NewInBody(route, argument.Args[0])
// 		// if schema == nil {
// 		// 	panic("nil")
// 		// }

// 		// srv.Components.Schemas[name] = schema
// 		// srv.Components.RequestBodies[name] = types.NewRequestBody(*types.NewReference(name, "schemas"))

// 		// route.RequestBody = types.NewReference(name, "requestBodies")
// 	case "Description":
// 		// if len(argument.Args) < 1 {
// 		// 	return fmt.Errorf("option 'Description' has no argument")
// 		// }

// 		// description, ok := argument.Args[0].(*ast.BasicLit)
// 		// if !ok {
// 		// 	return fmt.Errorf("option 'Description' isn't string")
// 		// }

// 		// route.Description = strings.Trim(description.Value, `"`)
// 	default:
// 		if len(argument.Args) < 2 {
// 			return fmt.Errorf("option '%s' has not enough arguments", selector.Sel.Name)
// 		}

// 		var (
// 			prefix  string
// 			placing = argument.Args[1].(*ast.SelectorExpr).Sel.Name
// 		)

// 		switch placing {
// 		case "InQuery":
// 			prefix = "query"
// 		case "InPath":
// 			prefix = "path"
// 		}

// 		srv.AddParameter(route, NewParameter(prefix, selector.Sel.Name, argument.Args))
// 	}

// 	return nil
// }

func (srv *Service) NewInBody(route *types.Route, arg types.Type) (string, types.Type) {
	switch typed := arg.(type) {
	case types.Call:
		return srv.NewInBody(route, typed.Parameters[0])
	case types.ArrayType:
		var name = typed.ItemType.Name()

		name, typed.ItemType = srv.NewInBody(route, typed.ItemType)

		return fmt.Sprintf("%sArray", name), typed
	case types.ImportedType:
		model := srv.parser.FindModel(*typed.Selector())
		if model == nil {
			return "", nil
		}

		return model.Name(), model
	default:
		panic("what")
	}

	// var selector ast.SelectorExpr

	// ast.Inspect(arg, func(n ast.Node) bool {
	// 	if sel, ok := n.(*ast.SelectorExpr); ok {
	// 		selector = *sel
	// 		return false
	// 	}
	// 	return true
	// })

	// _, imp := types.FindImport(*srv.receiver.File(), selector.X.(*ast.Ident).Name)
	// selector.X.(*ast.Ident).Name = strings.Trim(imp.Path.Value, "\"")

	// model := srv.parser.FindModel(selector)
	// if model == nil {
	// 	return "", nil
	// }

	// return model.Name(), model
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
