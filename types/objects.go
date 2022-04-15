package types

import (
	"fmt"
	"go/ast"
	"strings"
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
}

type Schema interface {
	SchemaType() string
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

func NewObject(t ast.StructType) Object {
	var schema = Object{
		Type:       "object",
		Properties: make(map[string]Property),
	}

	for _, item := range t.Fields.List {
		var parsedField Schema

		switch typed := item.Type.(type) {
		case *ast.Ident:
			parsedField = NewField(*typed)
		case *ast.StructType:
			parsedField = NewObject(*typed)
		case *ast.ArrayType:
			parsedField = NewArray(*typed)
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

func NewArray(t ast.ArrayType) Array {
	var schema = Array{Type: "array"}

	switch typed := t.Elt.(type) {
	case *ast.Ident:
		schema.Items = NewField(*typed)
	case *ast.StructType:
		schema.Items = NewObject(*typed)
	case *ast.ArrayType:
		schema.Items = NewArray(*typed)
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

func NewField(ident ast.Ident) Schema {
	var identType = ConvertFieldType(ident.Name)

	if len(identType) == 0 {
		return NewReference(ident.Name, "schemas")
	}

	return Field{
		Type:   identType,
		Format: GetFieldTypeFormat(ident.Name),
	}
}

type Reference struct {
	Ref string `json:"$ref"`
}

func (r Reference) SchemaType() string {
	return ""
}

func NewReference(name, where string) *Reference {
	return &Reference{
		Ref: fmt.Sprintf("#/components/%s/%s", where, name),
	}
}
