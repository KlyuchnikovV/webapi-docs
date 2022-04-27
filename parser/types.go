package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type (
	Server struct {
		URL         string `json:"url"`
		Description string `json:"description"`
	}

	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	}

	Response struct {
		Description string      `json:"description"`
		Schema      interface{} `json:"schema,omitempty"`
	}

	Components struct {
		Schemas       map[string]Schema      `json:"schemas,omitempty"`
		Parameters    map[string]IParameter  `json:"parameters,omitempty"`
		RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`

		Responses map[string]interface{} `json:"responses,omitempty"`
	}

	Route struct {
		Summary     string           `json:"summary,omitempty"`
		Tags        []string         `json:"tags,omitempty"`
		Parameters  []IParameter     `json:"parameters,omitempty"`
		RequestBody *Reference       `json:"requestBody,omitempty"`
		Responses   map[int]Response `json:"responses"`
	}

	SwaggerSpec struct {
		Openapi    string                      `json:"openapi"`
		Info       Info                        `json:"info"`
		Servers    []Server                    `json:"servers"`
		Components Components                  `json:"components"`
		Paths      map[string]map[string]Route `json:"paths"`
	}
)

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

func NewRoute() *Route {
	return &Route{
		Parameters: make([]IParameter, 0),
		Responses: map[int]Response{
			// TODO: Mocked for now
			200: {
				Description: "success",
			},
		},
		Tags: make([]string, 0),
	}
}

type (
	IParameter interface {
		NameParam() string
		Type() string
		EqualTo(interface{}) bool
	}

	Parameter struct {
		In          string     `json:"in"`
		Name        string     `json:"name"`
		Required    bool       `json:"required"`
		Minimum     int        `json:"minimum,omitempty"`
		Description string     `json:"description,omitempty"`
		RequestBody *Reference `json:"requestBody,omitempty"`
		Schema      Schema     `json:"schema,omitempty"`
	}
)

func NewParameter(paramType string, t string, args []ast.Expr) Parameter {
	var parameter = Parameter{
		In:       paramType,
		Required: true,
		Schema: Object{
			Type: types.ConvertFieldType(types.TypeParamsMap[t]),
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

func (i Parameter) NameParam() string {
	return i.Name
}

func (i Parameter) Type() string {
	return i.Schema.SchemaType()
}

func (i Parameter) EqualTo(p interface{}) bool {
	typed, ok := p.(Parameter)
	if !ok {
		return false
	}

	if i.Schema != nil && !i.Schema.EqualTo(typed.Schema) {
		return false
	}

	if i.RequestBody != nil && !i.RequestBody.EqualTo(typed.RequestBody) {
		return false
	}

	return typed.In == i.In &&
		typed.Name == i.Name &&
		typed.Required == i.Required &&
		typed.Description == i.Description
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
