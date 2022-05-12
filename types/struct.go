package types

import "go/ast"

type StructType struct {
	*typeBase

	StructType *ast.StructType
}

func NewStruct(file *ast.File, name string, str *ast.StructType, tag *ast.BasicLit) StructType {
	var result = StructType{
		typeBase: newTypeBase(file, name, tag),

		StructType: str,
	}

	for _, field := range str.Fields.List {
		name, t := NewTypeFromField(file, field)

		result.fields[name] = t
	}

	return result
}

func (s StructType) Schema() Schema {
	var props = make(map[string]Schema)

	for _, item := range s.Fields() {
		props[item.Tag()] = item.Schema()
	}

	return &ObjectSchema{
		Type:       "object",
		Properties: props,
	}
}

type ObjectSchema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
}

func (o ObjectSchema) EqualTo(s Schema) bool {
	os, ok := s.(ObjectSchema)
	if !ok {
		return false
	}

	if o.Type != os.Type {
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

func (o ObjectSchema) SchemaType() string {
	return o.Type
}
