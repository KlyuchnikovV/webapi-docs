package service

import (
	"fmt"
	"go/ast"
	"net/http"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/constants"
	"github.com/KlyuchnikovV/webapi-docs/types"
)

func (srv *Service) getResponses(service, method string, returns []ast.ReturnStmt, route *types.Route) error {
	var err error

	for _, returnStmt := range returns {
		for _, result := range returnStmt.Results {
			callExpr, ok := result.(*ast.CallExpr)
			if !ok {
				return fmt.Errorf("not a call expr")
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return fmt.Errorf("not a selector")
			}

			var code int
			// TODO: templating
			if selExpr.Sel.Name == "JSON" {
				var t = types.NewType(nil, "", callExpr.Args[0], nil)
				code, _ = constants.GetResultCode(strings.Trim(t.Name(), "Status"))
			} else {
				code, _ = constants.GetResultCode(selExpr.Sel.Name)
			}

			if code == -1 {
				continue
			}

			route.Responses[code], err = srv.defineResponse(service, method, selExpr.Sel.Name, callExpr.Args)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (srv *Service) defineResponse(service, method, fun string, args []ast.Expr) (*types.Reference, error) {
	var code, _ = constants.GetResultCode(fun)

	switch fun {
	case "Created", "NoContent":
		return srv.noContentResponse(code), nil
	case "JSON":
		var t = types.NewType(nil, "", args[0], nil)
		code, _ = constants.GetResultCode(strings.Trim(t.Name(), "Status"))

		return srv.objectResponse(service, method, code, args)
	case "OK":
		return srv.objectResponse(service, method, code, args)
	case "InternalServerError", "BadRequest", "Forbidden", "MethodNotAllowed", "NotFound":
		return srv.errorResponse(service, method, code, args)
	}

	return nil, fmt.Errorf("method unknown")
}

func (srv *Service) noContentResponse(code int) *types.Reference {
	var id = srv.newResponseBodyID(nil, "NoContent")

	if _, ok := srv.Components.Responses[id]; !ok {
		srv.Components.Responses[id] = *types.NewResponse(http.StatusText(code))
	}

	return types.NewReference(id, "responses")
}

func (srv *Service) objectResponse(service, method string, code int, args []ast.Expr) (*types.Reference, error) {
	t, err := srv.getReturnType(args[0], 0)
	if err != nil {
		return nil, err
	}

	switch typed := t.(type) {
	case types.ImportedType:
		t, err = srv.parser.UnwrapImportedType(typed)
		if err != nil {
			return nil, err
		}
	case *types.StringType, types.StringType:
	case *types.StructType, types.StructType:
	case nil:
		return nil, nil
	default:
		// fmt.Printf("%T\n", typed)
		panic(fmt.Sprintf("%T\n", typed))
	}

	var id = srv.newResponseBodyID(t, service, method)

	srv.Components.Responses[id] = *types.NewResponse(
		http.StatusText(code), *types.NewReference(id, "schemas"),
	)

	srv.Components.Schemas[id] = t

	return types.NewReference(id, "responses"), nil
}

func (srv *Service) errorResponse(service, method string, code int, args []ast.Expr) (*types.Reference, error) {
	t, err := srv.getReturnType(args[0], 0)
	if err != nil {
		return nil, err
	}

	var desc string

	switch typed := t.(type) {
	case types.ImportedType:
		t, err = srv.parser.UnwrapImportedType(typed)
		if err != nil {
			return nil, err
		}
	case types.StringType:
		desc = typed.Data
	case nil:
		return nil, nil
	}

	var id = srv.newResponseBodyID(t, service, method)

	if r, ok := srv.Components.Responses[id]; ok {
		r.Description = strings.Join([]string{r.Description, desc}, ", ")
	} else {
		srv.Components.Responses[id] = *types.NewErrorResponse(
			fmt.Sprintf("%s: %s", http.StatusText(code), desc), *types.NewReference(id, "schemas"),
		)

		srv.Components.Schemas[id] = t
	}

	return types.NewReference(id, "responses"), nil
}

func (srv *Service) newResponseBodyID(schema types.Type, pieces ...string) string {
	var (
		responseID string
		i          int
		prefix     = fmt.Sprintf("%s-response", strings.Join(pieces, "."))
	)

	for id, response := range srv.Components.Schemas {
		if !strings.HasPrefix(id, prefix) {
			continue
		}

		i++

		if response.EqualTo(schema) {
			responseID = id
			break
		}
	}

	if responseID != "" {
		return responseID
	}

	return fmt.Sprintf("%s-%d", prefix, i)
}
