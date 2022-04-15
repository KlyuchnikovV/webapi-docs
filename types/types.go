package types

import (
	"fmt"
	"go/ast"
	"strings"
)

type Components struct {
	Schemas       map[string]Schema      `json:"schemas,omitempty"`
	Parameters    map[string]Parameter   `json:"parameters,omitempty"`
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`

	Responses map[string]interface{} `json:"responses,omitempty"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Route struct {
	Summary     string           `json:"summary,omitempty"`
	Parameters  []Parameter      `json:"parameters,omitempty"`
	RequestBody *Reference       `json:"requestBody,omitempty"`
	Responses   map[int]Response `json:"responses"`
}

func NewRoute() *Route {
	return &Route{
		Parameters: make([]Parameter, 0),
		Responses: map[int]Response{
			// TODO: Mocked for now
			200: {
				Description: "success",
			},
		},
	}
}

type Parameter interface {
}

type InQueryParameter struct {
	In          string     `json:"in"`
	Name        string     `json:"name"`
	Required    bool       `json:"required"`
	Minimum     int        `json:"minimum,omitempty"`
	Description string     `json:"description,omitempty"`
	RequestBody *Reference `json:"requestBody,omitempty"`
	Schema      Schema     `json:"schema,omitempty"`
}

func NewInQuery(t string, args []ast.Expr) InQueryParameter {
	var parameter = InQueryParameter{
		In:       "query",
		Required: true,
		Schema: Object{
			Type: ConvertFieldType(TypeParamsMap[t]),
		},
	}

	for _, arg := range args {
		switch argument := arg.(type) {
		case *ast.BasicLit:
			if parameter.Name == "" {
				parameter.Name = strings.Trim(argument.Value, "\"")
			} else if parameter.Schema.SchemaType() == "date-time" {
				parameter.Description += fmt.Sprintf("Layout is '%s'", strings.Trim(argument.Value, "\""))
			}
		case *ast.CallExpr:
			name, args := parseParameterOption(argument)
			switch name {
			case "Description":
				parameter.Description = strings.Trim(fmt.Sprintf("%s %s",
					strings.Trim(argument.Args[0].(*ast.BasicLit).Value, "\""),
					parameter.Description,
				), " ")
			case "AND":
				parameter.Description = strings.Trim(fmt.Sprintf("%s %s",
					parameter.Description,
					fmt.Sprintf("Must be: %s", strings.Join(args, " and ")),
				), " ")
			case "OR":
				parameter.Description = strings.Trim(fmt.Sprintf("%s %s",
					parameter.Description,
					fmt.Sprintf("Must be: %s", strings.Join(args, " or ")),
				), " ")
			}
		case *ast.SelectorExpr:
			switch argument.Sel.Name {
			case "NotEmpty":
				parameter.Description = strings.Trim(fmt.Sprintf("%s %s",
					parameter.Description,
					"Shouldn't be empty.",
				), " ")
			}
		}
	}

	return parameter
}

type Response struct {
	Description string      `json:"description"`
	Schema      interface{} `json:"schema,omitempty"`
}

func parseParameterOption(expr *ast.CallExpr) (string, []string) {
	var (
		name      = expr.Fun.(*ast.SelectorExpr).Sel.Name
		arguments = make([]string, len(expr.Args))
	)

	for i, arg := range expr.Args {
		switch argument := arg.(type) {
		case *ast.SelectorExpr:
			switch argument.Sel.Name {
			case "NotEmpty":
				arguments[i] = "not empty"
			}
		case *ast.CallExpr:
			switch argument.Fun.(*ast.SelectorExpr).Sel.Name {
			case "Description":
				arguments[i] = strings.Trim(argument.Args[0].(*ast.BasicLit).Value, "\"")
			case "Greater":
				arguments[i] = fmt.Sprintf("greater than '%s'", argument.Args[0].(*ast.BasicLit).Value)
			case "Less":
				arguments[i] = fmt.Sprintf("less than '%s'", argument.Args[0].(*ast.BasicLit).Value)
			case "NotEmpty":
				arguments[i] = "shouldn't be empty"
			}
		}
	}

	return name, arguments
}

func NewInBody(arg ast.Expr) (string, Schema) {
	var (
		identifier *ast.Ident
	)

	switch typed := arg.(type) {
	case *ast.UnaryExpr:
		argument, ok := typed.X.(*ast.CompositeLit)
		if !ok {
			return "", nil
		}

		identifier, ok = argument.Type.(*ast.Ident)
		if !ok {
			return "", nil
		}
	case *ast.CompositeLit:
		argument, ok := typed.Type.(*ast.ArrayType)
		if !ok {
			return "", nil
		}

		identifier, ok = argument.Elt.(*ast.Ident)
		if !ok {
			return "", nil
		}
	}

	typeSpec, ok := identifier.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return "", nil
	}

	switch typed := typeSpec.Type.(type) {
	case *ast.StructType:
		return identifier.Name, NewObject(*typed)
	case *ast.ArrayType:
		return identifier.Name, NewArray(*typed)
	}

	return "", nil
}
