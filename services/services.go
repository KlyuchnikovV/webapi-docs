package services

// import (
// 	"fmt"
// 	"go/ast"
// 	"net/http"
// 	"path/filepath"

// 	"github.com/KlyuchnikovV/webapi-docs/cache"
// 	cacheTypes "github.com/KlyuchnikovV/webapi-docs/cache/types"
// 	"github.com/KlyuchnikovV/webapi-docs/constants"
// 	"github.com/KlyuchnikovV/webapi-docs/objects"
// 	"github.com/KlyuchnikovV/webapi-docs/types"
// 	"github.com/KlyuchnikovV/webapi-docs/utils"
// )

// type Service struct {
// 	// parser *Parser

// 	pkgName  string
// 	pkg      cacheTypes.Package
// 	selector ast.FuncDecl

// 	apiPrefix     string
// 	servicePrefix string
// 	webapiAlias   string

// 	Components objects.Components
// 	Paths      map[string]map[string]objects.Route

// 	currentFile ast.File

// 	// typeSpec ast.TypeSpec
// }

// func NewService(pkg cacheTypes.Package, apiPrefix string) *Service {
// 	return &Service{
// 		// selector:   funcDecl,
// 		pkg:        pkg,
// 		apiPrefix:  apiPrefix,
// 		Paths:      make(map[string]map[string]objects.Route),
// 		Components: objects.NewComponents(),
// 	}
// }

// func (srv *Service) ParseService(file ast.File, funcDecl ast.FuncDecl) error {
// 	srv.currentFile = file
// 	srv.webapiAlias = cache.FindAliasOfWebAPIInFile(srv.currentFile)

// 	typeSpec, err := getReturnType(funcDecl)
// 	if err != nil {
// 		return err
// 	}

// 	// if !srv.pkg.Implements(*typeSpec, types.ServiceInterface(srv.webapiAlias)) {
// 	// 	return nil
// 	// }

// 	srv.servicePrefix = GetPrefix(funcDecl.Body.List)

// 	routers, err := srv.parseRoutersMethod(
// 		*typeSpec,
// 		*cache.FindMethod(*typeSpec, *types.RoutersFuncDecl(srv.webapiAlias)),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	for _, route := range routers {
// 		err := srv.parseRoutes(route)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (srv *Service) parseRoutersMethod(
// 	receiver ast.TypeSpec, funcDecl ast.FuncDecl,
// ) ([][]ast.Expr, error) {
// 	var (
// 		returns = GetReturnStatements(receiver, funcDecl)
// 		result  = make([][]ast.Expr, 0)
// 	)

// 	for _, returnStmt := range returns {
// 		if len(returnStmt.Results) == 0 {
// 			continue
// 		}

// 		compositeLit, ok := returnStmt.Results[0].(*ast.CompositeLit)
// 		if !ok {
// 			continue
// 		}

// 		routersType := types.RoutersFuncType(srv.webapiAlias)

// 		if !utils.SameNodes(compositeLit.Type, routersType.Results.List[0].Type) {
// 			// if err := CheckRoutersResultType(compositeLit.Type); err != nil {
// 			return nil, fmt.Errorf("'Routers' is of wrong return type")
// 		}

// 		if len(funcDecl.Recv.List) == 0 {
// 			continue
// 		}

// 		// starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr)
// 		// if !ok {
// 		// 	continue
// 		// }

// 		// ident, ok := starExpr.X.(*ast.Ident)
// 		// if !ok {
// 		// 	continue
// 		// }

// 		// srv.typeSpec = *ident.Obj.Decl.(*ast.TypeSpec)
// 		result = append(result, compositeLit.Elts)

// 		// return srv.parseRoutes(file, compositeLit.Elts)
// 	}

// 	return result, nil
// }

// func (srv *Service) parseRoutes(expressions []ast.Expr) error {
// 	// var apiPrefix = "/" + filepath.Join(srv.apiPrefix, srv.servicePrefix)

// 	for _, expression := range expressions {
// 		keyValue, ok := expression.(*ast.KeyValueExpr)
// 		if !ok {
// 			return fmt.Errorf("not a key-value")
// 		}

// 		if err := srv.parseRoute(*keyValue); err != nil {
// 			return err
// 		}

// 		// route.Tags = append(route.Tags, srv.servicePrefix)

// 		// var resultPath = filepath.Join(apiPrefix, path)

// 		// if _, ok := srv.paths[resultPath]; !ok {
// 		// 	srv.paths[resultPath] = make(map[string]objects.Route)
// 		// }

// 		// srv.paths[resultPath][method] = *route
// 	}

// 	return nil
// }

// func (srv *Service) parseRoute(keyValue ast.KeyValueExpr) error {
// 	callExpr, ok := keyValue.Value.(*ast.CallExpr)
// 	if !ok {
// 		return fmt.Errorf("value not a function")
// 	}

// 	var route = objects.NewRoute(srv.servicePrefix)

// 	for _, arg := range callExpr.Args {
// 		switch argument := arg.(type) {
// 		case *ast.CallExpr:
// 			if err := srv.ParseParameter(route, *argument); err != nil {
// 				return err
// 			}
// 		case *ast.SelectorExpr:
// 			// funcDecl := GetFunctionDecl(srv.currentFile, argument)

// 			// returns := GetReturnStatements(
// 			// 	*funcDecl.Receiver,
// 			// 	*types.RouteHandlerFuncDecl(argument.Sel.Name, srv.webapiAlias),
// 			// )

// 			// if err := srv.getResponses(returns, route); err != nil {
// 			// 	return err
// 			// }
// 		}
// 	}

// 	var path = filepath.Join("/", srv.apiPrefix, srv.servicePrefix, extractPath(keyValue.Key))

// 	if _, ok := srv.Paths[path]; !ok {
// 		srv.Paths[path] = make(map[string]objects.Route)
// 	}

// 	srv.Paths[path][extractMethod(callExpr.Fun)] = *route

// 	return nil
// }

// func (srv *Service) getResponses(returns []ast.ReturnStmt, route *objects.Route) error {
// 	for _, returnStmt := range returns {
// 		for _, result := range returnStmt.Results {
// 			callExpr, ok := result.(*ast.CallExpr)
// 			if !ok {
// 				return fmt.Errorf("not a call expr")
// 			}

// 			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
// 			if !ok {
// 				return fmt.Errorf("not a selector")
// 			}

// 			code, codeType := constants.GetResultCode(selExpr.Sel.Name)
// 			if code == -1 {
// 				fmt.Printf("unknown code")
// 				continue
// 			}

// 			if err := srv.extractResponse(code, codeType, *callExpr, route); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func (srv *Service) extractResponse(code int, codeType constants.CodeType, callExpr ast.CallExpr, route *objects.Route) error {
// 	// TODO:
// 	//    define return code
// 	//    if (no error && with body) -> define result value
// 	//    else if (no error)         -> exit
// 	//    else                       -> exit?

// 	switch codeType {
// 	case constants.Success:
// 		return srv.extractSuccessResponse(code, callExpr, route)
// 	case constants.ServerError:
// 		var schema = srv.Components.NewField(*ast.NewIdent("string"))

// 		srv.Components.Schemas["error"] = schema
// 		srv.Components.Responses["error"] = *objects.NewResponse(
// 			http.StatusText(code),
// 			*objects.NewReference("error", "schemas"),
// 		)

// 		route.Responses[code] = objects.NewReference("error", "responses")

// 		// case constants.Info, constants.Redirection:
// 		// case constants.ClientError:
// 		// default:

// 	}

// 	return nil
// }

// func (srv *Service) extractSuccessResponse(code int, callExpr ast.CallExpr, route *objects.Route) error {
// 	if code == http.StatusNoContent || code == http.StatusCreated {
// 		var id = fmt.Sprintf("nocontent%d", code)

// 		if _, ok := srv.Components.Responses[id]; !ok {
// 			srv.Components.Responses[id] = *objects.NewResponse(http.StatusText(code))
// 		}

// 		route.Responses[code] = objects.NewReference(id, "responses")
// 		return nil
// 	}

// 	if len(callExpr.Args) != 1 {
// 		return fmt.Errorf("wrong number of arguments")
// 	}

// 	name, schema, err := srv.getVariablesSchema(callExpr.Args[0])
// 	if err != nil {
// 		return err
// 	}

// 	var id = fmt.Sprintf("success%d%s", code, name)

// 	srv.Components.Schemas[id] = schema
// 	srv.Components.Responses[id] = *objects.NewResponse(
// 		http.StatusText(code),
// 		*objects.NewReference("success", "responses"),
// 	)

// 	route.Responses[code] = objects.NewReference(id, "responses")
// 	return nil
// }

// func (srv *Service) getVariablesSchema(expr ast.Expr) (string, objects.Schema, error) {
// 	switch argument := expr.(type) {
// 	case *ast.Ident:
// 		return srv.parseIdent(argument.Name, *argument)
// 	case *ast.BasicLit:
// 		fmt.Printf("lit %#v\n", argument)
// 	case *ast.SelectorExpr:
// 		fmt.Printf("selector %#v\n", argument)
// 	case *ast.CallExpr:
// 		fmt.Printf("call %#v\n", argument)
// 	default:
// 		panic("not handled")
// 	}

// 	return "", nil, nil
// }

// func (srv *Service) parseIdent(name string, ident ast.Ident) (string, objects.Schema, error) {
// 	var (
// 		err      error
// 		typeSpec *ast.TypeSpec
// 		typeName string
// 	)

// 	switch typed := ident.Obj.Decl.(type) {
// 	case *ast.AssignStmt:
// 		for i, variable := range typed.Lhs {
// 			ident, ok := variable.(*ast.Ident)
// 			if !ok {
// 				continue
// 			}

// 			if ident.Name == name {
// 				typeName, typeSpec = GetTypeSpecification(srv.currentFile, typed.Rhs[i])
// 				break
// 			}
// 		}

// 	default:
// 		panic("not handled")
// 	}

// 	if typeSpec == nil {
// 		return typeName, nil, fmt.Errorf("no definition for '%s' found", typeName)
// 	}

// 	schema, err := srv.Components.NewSchema(*typeSpec)

// 	return typeName, schema, err
// }

// // func (srv *Service) ExtractReturnData(callExpr ast.CallExpr, fun ast.SelectorExpr, route *Route) error {
// // 	var (
// // 		code = types.GetResultCode(fun.Sel.Name)
// // 		args = make([]string, 0, len(callExpr.Args))
// // 		ref  *Reference
// // 		err  error
// // 	)

// // 	// TODO: args depends on code
// // 	for _, arg := range callExpr.Args {
// // 		switch typed := arg.(type) {
// // 		case *ast.BasicLit:
// // 			args = append(args, typed.Value)
// // 		case *ast.Ident:
// // 			name, schema, err := srv.defineResultSchema(*typed)
// // 			if err != nil {
// // 				return err
// // 			}

// // 			if schema == nil {
// // 				continue
// // 			}

// // 			srv.parser.Spec.Components.Schemas[name] = schema

// // 			srv.parser.Spec.Components.Responses[name] = NewResponse(
// // 				http.StatusText(code), *srv.parser.NewReference(name, "schemas"),
// // 			)

// // 			ref = srv.parser.NewReference(name, "responses")
// // 		case *ast.SelectorExpr:
// // 			code = types.CodeSelector[typed.Sel.Name]
// // 		case *ast.CallExpr:
// // 			name, typeSpec := srv.findDefinition(typed)

// // 			schema, err := srv.parser.NewSchema(typeSpec)
// // 			if err != nil {
// // 				return err
// // 			}

// // 			if schema == nil {
// // 				continue
// // 			}

// // 			srv.parser.Spec.Components.Schemas[name] = schema

// // 			srv.parser.Spec.Components.Responses[name] = NewResponse(
// // 				http.StatusText(code), *srv.parser.NewReference(name, "schemas"),
// // 			)

// // 			ref = srv.parser.NewReference(name, "responses")
// // 		}
// // 	}

// // 	if code == -1 {
// // 		code, err = strconv.Atoi(strings.Trim(args[0], "\""))
// // 		if err != nil {
// // 			return err
// // 		}
// // 	}

// // 	switch code {
// // 	case http.StatusBadRequest, http.StatusForbidden,
// // 		http.StatusNotFound, http.StatusMethodNotAllowed,
// // 		http.StatusInternalServerError:
// // 		schema, err := srv.parser.NewField(*ast.NewIdent("string"))
// // 		if err != nil {
// // 			return err
// // 		}

// // 		srv.parser.Spec.Components.Schemas["error"] = schema
// // 		srv.parser.Spec.Components.Responses["error"] = NewResponse(
// // 			fmt.Sprintf("%s: %s", http.StatusText(code), strings.Join(args, ",")),
// // 			*srv.parser.NewReference("error", "schemas"),
// // 		)

// // 		route.Responses[code] = srv.parser.NewReference("error", "responses")
// // 	default:
// // 		route.Responses[code] = ref
// // 	}

// // 	return nil
// // }

// // func (srv *Service) defineResultSchema(sel ast.Ident) (string, Schema, error) {
// // 	var (
// // 		err      error
// // 		typeSpec *ast.TypeSpec
// // 		name     string
// // 	)

// // 	switch typed := sel.Obj.Decl.(type) {
// // 	case *ast.StructType:
// // 		schema, err := srv.parser.NewObject(*typed)
// // 		return sel.Name, schema, err
// // 	case *ast.ArrayType:
// // 		schema, err := srv.parser.NewArray(*typed)
// // 		return sel.Name, schema, err
// // 	case *ast.AssignStmt:
// // 		for i, variable := range typed.Lhs {
// // 			ident, ok := variable.(*ast.Ident)
// // 			if !ok {
// // 				continue
// // 			}

// // 			if ident.Name == sel.Name {
// // 				name, typeSpec = srv.findDefinition(typed.Rhs[i])
// // 			}
// // 		}
// // 	}

// // 	if err != nil {
// // 		return "", nil, err
// // 	}

// // 	schema, err := srv.parser.NewSchema(typeSpec)

// // 	return name, schema, err
// // }

// // func (srv *Service) findDefinition(expr ast.Expr) (string, *ast.TypeSpec) {
// // 	var (
// // 		name     string
// // 		typeSpec *ast.TypeSpec
// // 		selector *ast.SelectorExpr
// // 	)

// // 	switch typed := expr.(type) {
// // 	case *ast.Ident:
// // 		return srv.identDef(*typed)
// // 	case *ast.CallExpr:
// // 		return srv.findDefinition(typed.Fun)
// // 	case *ast.StarExpr:
// // 		return srv.findDefinition(typed.X)
// // 	case *ast.SelectorExpr:
// // 		selector = typed
// // 		name, typeSpec = srv.findDefinition(typed.X)

// // 		if typeSpec == nil {
// // 			return srv.findDefinition(typed.Sel)
// // 		}
// // 	default:
// // 		return "", nil
// // 	}

// // 	typed, ok := typeSpec.Type.(*ast.StructType)
// // 	if !ok {
// // 		return name, typeSpec
// // 	}

// // 	for _, field := range typed.Fields.List {
// // 		for _, name := range field.Names {
// // 			if name.Name != selector.Sel.Name {
// // 				continue
// // 			}

// // 			return srv.findDefinition(field.Type)
// // 		}
// // 	}

// // 	fd := srv.parser.FindMethodByName(*typeSpec, selector.Sel.Name)

// // 	if fd == nil || len(fd.Type.Results.List) == 0 {
// // 		return "", nil
// // 	}

// // 	return srv.findDefinition(fd.Type.Results.List[0].Type)
// // }

// // func (srv *Service) identDef(ident ast.Ident) (string, *ast.TypeSpec) {
// // 	var (
// // 		obj      = ident.Obj
// // 		typeSpec *ast.TypeSpec
// // 	)

// // 	if obj == nil {
// // 		model, err := srv.parser.FindModel(ident)
// // 		if err != nil {
// // 			return "", nil
// // 		}

// // 		obj = model
// // 	}

// // 	switch typed := obj.Decl.(type) {
// // 	case *ast.Field:
// // 		return srv.findDefinition(typed.Type)
// // 	case *ast.TypeSpec:
// // 		typeSpec = typed
// // 	default:
// // 		return "", nil
// // 	}

// // 	return typeSpec.Name.Name, typeSpec
// // }
