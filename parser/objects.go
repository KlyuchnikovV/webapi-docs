package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Content struct {
	Schema Reference `json:"schema"`
}

type RequestBody struct {
	Description string             `json:"description,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Content     map[string]Content `json:"content"`
}

func NewRequestBody(ref Reference) RequestBody {
	return RequestBody{
		Content: map[string]Content{
			// TODO: multiple schemas
			"application/json": {
				Schema: ref,
			},
		},
	}
}

type Schema interface {
	SchemaType() string
	EqualTo(interface{}) bool
}

const jsonTag = "json:\""

func GetFieldName(field *ast.Field) string {
	if field.Tag == nil {
		return field.Names[0].Name
	}

	var (
		value = field.Tag.Value
		start = strings.Index(value, jsonTag) + len(jsonTag)
		end   = strings.Index(value[start:], "\"")
	)

	return value[start : start+end]
}

type ObjectProperty interface {
	EqualTo(interface{}) bool
}

type Object struct {
	Type       string                    `json:"type"`
	Properties map[string]ObjectProperty `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

func (parser *Parser) NewObject(t ast.StructType) (*Object, error) {
	var schema = Object{
		Type:       "object",
		Properties: make(map[string]ObjectProperty),
	}

	for _, item := range t.Fields.List {
		var (
			parsedField Schema
			err         error
		)

		switch typed := item.Type.(type) {
		case *ast.Ident:
			parsedField, err = parser.NewField(*typed)
		case *ast.StructType:
			parsedField, err = parser.NewObject(*typed)
		case *ast.ArrayType:
			parsedField, err = parser.NewArray(*typed)
		}

		if err != nil {
			return nil, err
		}

		schema.Properties[GetFieldName(item)] = parsedField
	}

	return &schema, nil
}

func (o Object) SchemaType() string {
	return o.Type
}

func (o Object) EqualTo(s interface{}) bool {
	typed, ok := s.(Object)
	if !ok {
		return false
	}

	if o.Type != typed.Type {
		return false
	}

	if len(o.Properties) != len(typed.Properties) {
		return false
	}

	for name, prop := range o.Properties {
		if typedProp, ok := typed.Properties[name]; !ok || !prop.EqualTo(typedProp) {
			return false
		}
	}

	if len(o.Required) == len(typed.Required) {
		return true
	}

	for _, value := range o.Required {
		var found bool

		for _, secondValue := range typed.Required {
			if value == secondValue {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

type Array struct {
	Type  string `json:"type"`
	Items Schema `json:"items"`
}

func (parser *Parser) NewArray(t ast.ArrayType) (*Array, error) {
	var (
		schema = Array{Type: "array"}
		err    error
	)
	switch typed := t.Elt.(type) {
	case *ast.Ident:
		schema.Items, err = parser.NewField(*typed)
	case *ast.StructType:
		schema.Items, err = parser.NewObject(*typed)
	case *ast.ArrayType:
		schema.Items, err = parser.NewArray(*typed)
	}

	return &schema, err
}

func (a Array) SchemaType() string {
	return a.Type
}

func (a Array) EqualTo(s interface{}) bool {
	typed, ok := s.(Array)
	if !ok {
		return false
	}

	if a.Type != typed.Type {
		return false
	}

	return a.Items.EqualTo(typed.Items)
}

type Field struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

func (parser *Parser) NewField(ident ast.Ident) (Schema, error) {
	if identType := types.ConvertFieldType(ident.Name); len(identType) != 0 {
		return Field{
			Type:   identType,
			Format: types.GetFieldTypeFormat(ident.Name),
		}, nil
	}

	var obj = ident.Obj
	if obj == nil {
		model, err := parser.FindModel(ident)
		if err != nil {
			return nil, err
		}
		obj = model
	}

	typeSpec, ok := obj.Decl.(ast.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("field could be parsed only to type specification (got: '%T')", obj.Decl)
	}

	switch t := typeSpec.Type.(type) {
	case *ast.Ident:
		return parser.NewField(*t)
	case *ast.StructType:

		if _, ok := parser.loopController[ident.Name]; !ok {
			var err error

			parser.loopController[ident.Name] = struct{}{}
			parser.Spec.Components.Schemas[ident.Name], err = parser.NewObject(*t)
			delete(parser.loopController, ident.Name)

			if err != nil {
				return nil, err
			}
		}

		return parser.NewReference(ident.Name, "schemas"), nil

	default:
		return nil, fmt.Errorf("fields inner type not handled")
	}
}

func (f Field) SchemaType() string {
	return f.Type
}

func (f Field) EqualTo(s interface{}) bool {
	typed, ok := s.(Field)
	if !ok {
		return false
	}

	return f.Type == typed.Type && f.Format == typed.Format
}

type Reference struct {
	Ref string `json:"$ref"`

	name string
}

func (parser *Parser) NewReference(name, where string) *Reference {
	return &Reference{
		Ref:  fmt.Sprintf("#/components/%s/%s", where, name),
		name: name,
	}
}

func (r Reference) SchemaType() string {
	return ""
}

func (r Reference) NameParam() string {
	return r.name
}

func (r Reference) Type() string {
	return "ref"
}

func (r Reference) EqualTo(p interface{}) bool {
	typed, ok := p.(Reference)
	if !ok {
		return false
	}

	return typed.Ref == r.Ref
}
