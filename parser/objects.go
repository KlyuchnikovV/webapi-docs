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

type Property interface {
	EqualTo(interface{}) bool
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

type Object struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
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

	if len(o.Required) != len(typed.Required) {
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

	return true
}

func (parser *Parser) NewObject(t ast.StructType) Object {
	var schema = Object{
		Type:       "object",
		Properties: make(map[string]Property),
	}

	for _, item := range t.Fields.List {
		var parsedField Schema

		switch typed := item.Type.(type) {
		case *ast.Ident:
			parsedField = parser.NewField(*typed)
		case *ast.StructType:
			parsedField = parser.NewObject(*typed)
		case *ast.ArrayType:
			parsedField = parser.NewArray(*typed)
		}

		schema.Properties[GetFieldName(item)] = parsedField
	}

	return schema
}

type Array struct {
	Type  string `json:"type"`
	Items Schema `json:"items"`
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

func (parser *Parser) NewArray(t ast.ArrayType) Array {
	var schema = Array{Type: "array"}

	switch typed := t.Elt.(type) {
	case *ast.Ident:
		schema.Items = parser.NewField(*typed)
	case *ast.StructType:
		schema.Items = parser.NewObject(*typed)
	case *ast.ArrayType:
		schema.Items = parser.NewArray(*typed)
	}

	return schema
}

type Field struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
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

func (parser *Parser) NewField(ident ast.Ident) Schema {
	var identType = types.ConvertFieldType(ident.Name)

	if len(identType) != 0 {
		return Field{
			Type:   identType,
			Format: types.GetFieldTypeFormat(ident.Name),
		}
	}

	var obj ast.Object
	if ident.Obj == nil {
		obj = *parser.FindModel(ident)
	} else {
		obj = *ident.Obj
	}

	switch typed := obj.Decl.(type) {
	case *ast.TypeSpec:
		switch t := typed.Type.(type) {
		case *ast.Ident:
			return Field{
				Type:   types.ConvertFieldType(t.Name),
				Format: types.GetFieldTypeFormat(ident.Name),
			}
		case *ast.StructType:
			if _, ok := parser.loopController[ident.Name]; !ok {
				parser.loopController[ident.Name] = struct{}{}
				parser.file.Components.Schemas[ident.Name] = parser.NewObject(*t)
				delete(parser.loopController, ident.Name)
			}

			return NewReference(ident.Name, "schemas")
		default:
			panic("not handled inner")
		}
	default:
		// return NewReference(ident.Name, "schemas")
		panic("not handled")
	}
}

type Reference struct {
	Ref string `json:"$ref"`

	name string
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

func NewReference(name, where string) *Reference {
	return &Reference{
		Ref:  fmt.Sprintf("#/components/%s/%s", where, name),
		name: name,
	}
}
