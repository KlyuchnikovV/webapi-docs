package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Server struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

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

func (parser *Parser) AddParameter(route *Route, param Parameter) {
	var (
		name  string
		ok    = true
		saved Parameter
	)

	for i := 0; ok; i++ {
		name = fmt.Sprintf("%s-%s-%d", param.NameParam(), param.Type(), i)
		saved, ok = parser.file.Components.Parameters[name]

		if ok && saved.EqualTo(param) {
			break
		}
	}

	route.Parameters = append(route.Parameters, NewReference(name, "parameters"))

	if saved == nil {
		parser.file.Components.Parameters[name] = param
	}
}

type Parameter interface {
	NameParam() string
	Type() string
	EqualTo(interface{}) bool
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

func (i InQueryParameter) NameParam() string {
	return i.Name
}

func (i InQueryParameter) Type() string {
	return i.Schema.SchemaType()
}

func (i InQueryParameter) EqualTo(p interface{}) bool {
	typed, ok := p.(InQueryParameter)
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

func NewInQuery(t string, args []ast.Expr) InQueryParameter {
	var parameter = InQueryParameter{
		In:       "query",
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

type InPathParameter struct {
	In          string     `json:"in"`
	Name        string     `json:"name"`
	Required    bool       `json:"required"`
	Minimum     int        `json:"minimum,omitempty"`
	Description string     `json:"description,omitempty"`
	RequestBody *Reference `json:"requestBody,omitempty"`
	Schema      Schema     `json:"schema,omitempty"`
}

func (i InPathParameter) NameParam() string {
	return i.Name
}

func (i InPathParameter) Type() string {
	return i.Schema.SchemaType()
}

func (i InPathParameter) EqualTo(p interface{}) bool {
	typed, ok := p.(InPathParameter)
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

func NewInPath(desc string) InPathParameter {
	var (
		param = InPathParameter{
			In:       "path",
			Required: true,
		}
		nameEnd = strings.IndexRune(desc, '}')
	)
	if nameEnd == -1 {
		panic("something wrong")
	}
	param.Name = desc[1:nameEnd]

	if !strings.ContainsRune(desc, '[') {
		param.Schema = Object{Type: "string"}
		return param
	}

	var (
		typeStart = strings.IndexRune(desc, '[')
		typeEnd   = strings.IndexRune(desc, ']')
	)

	if typeStart == -1 || typeEnd == -1 {
		panic("something wrong")
	}

	param.Schema = Object{
		Type: types.ConvertFieldType(desc[typeStart+1 : typeEnd]),
	}

	return param
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
