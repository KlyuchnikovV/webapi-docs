package types

import (
	"encoding/json"
	"go/ast"
	"strings"
)

type StructType struct {
	*typeBase

	StructType *ast.StructType

	// *BaseSchema
	Properties map[string]Type `json:"properties,omitempty"`
}

func NewStruct(file *ast.File, name string, str *ast.StructType, tag *ast.BasicLit) StructType {
	var result = StructType{
		typeBase: newTypeBase(file, name, tag, ObjectSchemaType),

		StructType: str,
		Properties: make(map[string]Type),
	}

	for _, field := range str.Fields.List {
		name, t := NewTypeFromField(file, field)
		// TODO: what if we want xml schema
		tagName := strings.TrimSuffix(t.Tag("json"), ",omitempty")

		if tagName == "" {
			tagName = name
		}

		result.fields[name] = t
		result.Properties[tagName] = t
	}

	return result
}

func (o StructType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type        SchemaType      `json:"type"`
		Description string          `json:"description,omitempty"`
		Example     string          `json:"example,omitempty"`
		Required    bool            `json:"required,omitempty"`
		Properties  map[string]Type `json:"properties,omitempty"`
	}{
		Type:        o.Type,
		Description: o.Description,
		Example:     o.Example,
		Required:    o.Required,
		Properties:  o.Properties,
	})
}

func (o StructType) EqualSchema(s Type) bool {
	os, ok := s.(StructType)
	if !ok {
		return false
	}

	for name, prop := range o.Properties {
		sProp, ok := os.Properties[name]
		if !ok {
			return false
		}

		if !prop.EqualTo(sProp) {
			return false
		}
	}

	return true
}
