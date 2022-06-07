package types

import (
	"go/ast"
	"strings"
)

type Struct struct {
	Name   string           `json:"name,omitempty"`
	Fields map[string]Field `json:"fields,omitempty"`

	structType *ast.StructType
	file       *ast.File
}

func NewStruct(name string, spec *ast.StructType, file *ast.File) (*Struct, error) {
	var result = Struct{
		Name:   name,
		Fields: make(map[string]Field),

		structType: spec,
		file:       file,
	}

	for _, field := range spec.Fields.List {
		var (
			name string
			tag  string
		)

		if len(field.Names) > 0 {
			name = field.Names[0].Name
		}

		if field.Tag != nil {
			tag = strings.Trim(field.Tag.Value, "`")
		}

		field, err := NewField(name, field.Type, tag, file)
		if err != nil {
			return nil, err
		}

		result.Fields[field.Name] = *field
	}

	return &result, nil
}

func (s Struct) GetName() string {
	return s.Name
}

type Field struct {
	Name string              `json:"name,omitempty"`
	Type Type                `json:"type,omitempty"`
	Tags map[string][]string `json:"tags,omitempty"`
}

func NewField(name string, spec ast.Expr, tag string, file *ast.File) (*Field, error) {
	var tags = make(map[string][]string)

	if len(tag) > 0 {
		for _, value := range strings.Split(tag, " ") {
			var values = strings.Split(value, ":")

			if len(values) < 2 {
				continue
			}

			tags[values[0]] = strings.Split(
				strings.Trim(values[1], `"`),
				",",
			)
		}
	}

	t, err := TypeFromExpr("", spec, file)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = t.GetName()
	}

	return &Field{
		Name: name,

		Type: t,
		Tags: tags,
	}, nil
}
