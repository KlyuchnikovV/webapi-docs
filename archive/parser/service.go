package parser

// import (
// 	"fmt"
// 	"go/ast"
// 	"net/http"
// 	"path/filepath"
// 	"strconv"
// 	"strings"

// 	"github.com/KlyuchnikovV/webapi-docs/objects"

// 	"github.com/KlyuchnikovV/webapi-docs/services"
// 	"github.com/KlyuchnikovV/webapi-docs/types"
// )

// type Service struct {
// 	parser *Parser

// 	pkg           Package
// 	selector      ast.FuncDecl
// 	servicePrefix string

// 	typeSpec ast.TypeSpec
// }

// func NewService(parser *Parser, pkg Package, funcDecl ast.FuncDecl) *Service {
// 	return &Service{
// 		parser:   parser,
// 		pkg:      pkg,
// 		selector: funcDecl,
// 	}
// }

// func (srv *Service) ParseService(file ast.File, funcDecl ast.FuncDecl) error {
// 	typeSpec, err := GetReturnType(funcDecl)
// 	if err != nil {
// 		return err
// 	}

// 	var alias = findWebapiImport(file)

// 	if !srv.pkg.Implements(*typeSpec, types.ServiceInterface(alias)) {
// 		return nil
// 	}

// 	srv.servicePrefix = GetPrefix(funcDecl.Body.List)

// 	return srv.ParseRouters(
// 		file,
// 		*srv.pkg.FindMethod(*typeSpec, *types.RoutersFuncDecl(alias)),
// 	)
// }

// func (srv *Service) ParseRouters(file ast.File, funcDecl ast.FuncDecl) error {
// 	if err := CheckFuncDeclaration(funcDecl, "Routers", nil, CheckRoutersResultType); err != nil {
// 		return err
// 	}

// 	for _, statement := range funcDecl.Body.List {
// 		returnStmt, ok := statement.(*ast.ReturnStmt)
// 		if !ok {
// 			continue
// 		}

// 		if len(returnStmt.Results) == 0 {
// 			continue
// 		}

// 		compositeLit, ok := returnStmt.Results[0].(*ast.CompositeLit)
// 		if !ok {
// 			continue
// 		}

// 		if err := CheckRoutersResultType(compositeLit.Type); err != nil {
// 			return fmt.Errorf("'Routers' is of wrong return type: %w", err)
// 		}

// 		if len(funcDecl.Recv.List) == 0 {
// 			continue
// 		}

// 		starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr)
// 		if !ok {
// 			continue
// 		}

// 		ident, ok := starExpr.X.(*ast.Ident)
// 		if !ok {
// 			continue
// 		}

// 		srv.typeSpec = *ident.Obj.Decl.(*ast.TypeSpec)

// 		return srv.parseRoutes(file, compositeLit.Elts)
// 	}

// 	return nil
// }

// func (srv *Service) parseRoutes(file ast.File, expressions []ast.Expr) error {
// 	var apiPrefix = "/" + filepath.Join(srv.parser.apiPrefix, srv.servicePrefix)

// 	for _, expression := range expressions {
// 		keyValue, ok := expression.(*ast.KeyValueExpr)
// 		if !ok {
// 			return fmt.Errorf("not a key-value")
// 		}

// 		method, path, route, err := srv.ParseRoute(file, *keyValue)
// 		if err != nil {
// 			return err
// 		}

// 		route.Tags = append(route.Tags, srv.servicePrefix)

// 		var resultPath = filepath.Join(apiPrefix, path)

// 		if _, ok := srv.parser.Spec.Paths[resultPath]; !ok {
// 			srv.parser.Spec.Paths[resultPath] = make(map[string]objects.Route)
// 		}

// 		srv.parser.Spec.Paths[resultPath][method] = *route
// 	}

// 	return nil
// }

// func (srv *Service) ParseRoute(file ast.File, keyValue ast.KeyValueExpr) (string, string, *objects.Route, error) {
// 	var route = objects.NewRoute()

// 	callExpression, ok := keyValue.Value.(*ast.CallExpr)
// 	if !ok {
// 		return "", "", nil, fmt.Errorf("value not a function")
// 	}

// 	for _, arg := range callExpression.Args {
// 		switch argument := arg.(type) {
// 		case *ast.CallExpr:
// 			if err := srv.parser.ParseParameter(file, route, *argument); err != nil {
// 				return "", "", nil, err
// 			}
// 		case *ast.SelectorExpr:
// 			// returns, err := srv.getReturnStatements(findWebapiImport(file), *argument)
// 			// if err != nil {
// 			// 	return "", "", nil, err
// 			// }

// 			// if err := srv.getResponses(returns, route); err != nil {
// 			// 	return "", "", nil, err
// 			// }

// 			_, typeSpec := services.GetTypeSpecification(file, argument)

// 			returns := services.GetReturnStatements(
// 				*typeSpec,
// 				*types.RouteHandlerFuncDecl(argument.Sel.Name, findWebapiImport(file)),
// 			)

// 			if err := srv.getResponses(returns, route); err != nil {
// 				return "", "", nil, err
// 			}

// 			// fmt.Printf("%#v", returns)
// 		}
// 	}

// 	return extractMethod(callExpression.Fun),
// 		extractPath(keyValue.Key),
// 		route, nil
// }

// func (srv *Service) getReturnStatements(alias string, selector ast.SelectorExpr) ([]*ast.ReturnStmt, error) {
// 	funcDecl := srv.pkg.FindMethod(
// 		srv.typeSpec,
// 		*types.RouteHandlerFuncDecl(
// 			selector.Sel.Name,
// 			alias,
// 		),
// 	)
// 	if funcDecl == nil {
// 		return nil, fmt.Errorf("no declaration found")
// 	}

// 	var returns = make([]*ast.ReturnStmt, 0)

// 	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
// 		returnStmt, ok := n.(*ast.ReturnStmt)
// 		if ok {
// 			returns = append(returns, returnStmt)
// 		}
// 		return true
// 	})

// 	return returns, nil
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

// 			if err := srv.ExtractReturnData(*callExpr, *selExpr, route); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func (srv *Service) ExtractReturnData(callExpr ast.CallExpr, fun ast.SelectorExpr, route *objects.Route) error {
// 	var (
// 		code = types.GetResultCode(fun.Sel.Name)
// 		args = make([]string, 0, len(callExpr.Args))
// 		ref  *objects.Reference
// 		err  error
// 	)

// 	// TODO: args depends on code
// 	for _, arg := range callExpr.Args {
// 		switch typed := arg.(type) {
// 		case *ast.BasicLit:
// 			args = append(args, typed.Value)
// 		case *ast.Ident:
// 			name, schema, err := srv.defineResultSchema(*typed)
// 			if err != nil {
// 				return err
// 			}

// 			if schema == nil {
// 				continue
// 			}

// 			srv.parser.Spec.Components.Schemas[name] = schema

// 			srv.parser.Spec.Components.Responses[name] = *objects.NewResponse(
// 				http.StatusText(code), *objects.NewReference(name, "schemas"),
// 			)

// 			ref = objects.NewReference(name, "responses")
// 		case *ast.SelectorExpr:
// 			code = types.CodeSelector[typed.Sel.Name]
// 		case *ast.CallExpr:
// 			name, typeSpec := srv.findDefinition(typed)

// 			schema, err := srv.parser.NewSchema(typeSpec)
// 			if err != nil {
// 				return err
// 			}

// 			if schema == nil {
// 				continue
// 			}

// 			srv.parser.Spec.Components.Schemas[name] = schema

// 			srv.parser.Spec.Components.Responses[name] = *objects.NewResponse(
// 				http.StatusText(code), *objects.NewReference(name, "schemas"),
// 			)

// 			ref = objects.NewReference(name, "responses")
// 		}
// 	}

// 	if code == -1 {
// 		code, err = strconv.Atoi(strings.Trim(args[0], "\""))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	switch code {
// 	case http.StatusBadRequest, http.StatusForbidden,
// 		http.StatusNotFound, http.StatusMethodNotAllowed,
// 		http.StatusInternalServerError:
// 		schema, err := srv.parser.NewField(*ast.NewIdent("string"))
// 		if err != nil {
// 			return err
// 		}

// 		srv.parser.Spec.Components.Schemas["error"] = schema
// 		srv.parser.Spec.Components.Responses["error"] = *objects.NewResponse(
// 			fmt.Sprintf("%s: %s", http.StatusText(code), strings.Join(args, ",")),
// 			*objects.NewReference("error", "schemas"),
// 		)

// 		route.Responses[code] = objects.NewReference("error", "responses")
// 	default:
// 		route.Responses[code] = ref
// 	}

// 	return nil
// }

// func (srv *Service) defineResultSchema(sel ast.Ident) (string, Schema, error) {
// 	var (
// 		err      error
// 		typeSpec *ast.TypeSpec
// 		name     string
// 	)

// 	switch typed := sel.Obj.Decl.(type) {
// 	case *ast.StructType:
// 		schema, err := srv.parser.NewObject(*typed)
// 		return sel.Name, schema, err
// 	case *ast.ArrayType:
// 		schema, err := srv.parser.NewArray(*typed)
// 		return sel.Name, schema, err
// 	case *ast.AssignStmt:
// 		for i, variable := range typed.Lhs {
// 			ident, ok := variable.(*ast.Ident)
// 			if !ok {
// 				continue
// 			}

// 			if ident.Name == sel.Name {
// 				name, typeSpec = srv.findDefinition(typed.Rhs[i])
// 			}
// 		}
// 	}

// 	if err != nil {
// 		return "", nil, err
// 	}

// 	schema, err := srv.parser.NewSchema(typeSpec)

// 	return name, schema, err
// }

// func (srv *Service) findDefinition(expr ast.Expr) (string, *ast.TypeSpec) {
// 	var (
// 		name     string
// 		typeSpec *ast.TypeSpec
// 		selector *ast.SelectorExpr
// 	)

// 	switch typed := expr.(type) {
// 	case *ast.Ident:
// 		return srv.identDef(*typed)
// 	case *ast.CallExpr:
// 		return srv.findDefinition(typed.Fun)
// 	case *ast.StarExpr:
// 		return srv.findDefinition(typed.X)
// 	case *ast.SelectorExpr:
// 		selector = typed
// 		name, typeSpec = srv.findDefinition(typed.X)

// 		if typeSpec == nil {
// 			return srv.findDefinition(typed.Sel)
// 		}
// 	default:
// 		return "", nil
// 	}

// 	typed, ok := typeSpec.Type.(*ast.StructType)
// 	if !ok {
// 		return name, typeSpec
// 	}

// 	for _, field := range typed.Fields.List {
// 		for _, name := range field.Names {
// 			if name.Name != selector.Sel.Name {
// 				continue
// 			}

// 			return srv.findDefinition(field.Type)
// 		}
// 	}

// 	fd := srv.parser.FindMethodByName(*typeSpec, selector.Sel.Name)

// 	if fd == nil || len(fd.Type.Results.List) == 0 {
// 		return "", nil
// 	}

// 	return srv.findDefinition(fd.Type.Results.List[0].Type)
// }

// func (srv *Service) identDef(ident ast.Ident) (string, *ast.TypeSpec) {
// 	var (
// 		obj      = ident.Obj
// 		typeSpec *ast.TypeSpec
// 	)

// 	if obj == nil {
// 		model, err := srv.parser.FindModel(ident)
// 		if err != nil {
// 			return "", nil
// 		}

// 		obj = model
// 	}

// 	switch typed := obj.Decl.(type) {
// 	case *ast.Field:
// 		return srv.findDefinition(typed.Type)
// 	case *ast.TypeSpec:
// 		typeSpec = typed
// 	default:
// 		return "", nil
// 	}

// 	return typeSpec.Name.Name, typeSpec
// }
