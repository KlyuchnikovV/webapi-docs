package parser

import (
	"go/ast"
	"strings"
)

type Struct struct {
	Name    string              `json:"name,omitempty"`
	Fields  map[string]*Field    `json:"fields,omitempty"`
	Methods map[string]*FuncDecl `json:"methods,omitempty"`

	structType *ast.StructType `json:"-"`
	file       *ast.File       `json:"-"`
}

func NewStruct(name string, spec *ast.StructType, file *ast.File) (*Struct, error) {
	var result = Struct{
		Name:    name,
		Fields:  make(map[string]*Field),
		Methods: make(map[string]*FuncDecl),

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

		result.Fields[field.Name] = field
	}

	return &result, nil
}

func (s Struct) GetName() string {
	return s.Name
}

func (s Struct) EqualTo(t Type) bool {
	typed, ok := t.(*Struct)
	if !ok {
		return false
	}

	if len(s.Fields) != len(typed.Fields) {
		return false
	}

	for name, field1 := range s.Fields {
		field2, ok := typed.Fields[name]
		if !ok {
			return false
		}

		if !field1.EqualTo(field2) {
			return false
		}
	}

	return s.Name == typed.Name
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

func (f Field) GetName() string {
	return f.Name
}

func (f Field) EqualTo(t Type) bool {
	field, ok := t.(*Field)
	if !ok {
		return false
	}

	if len(f.Tags) != len(field.Tags) {
		return false
	}

	for name, tags1 := range f.Tags {
		tags2, ok := field.Tags[name]
		if !ok {
			return false
		}

		if len(tags1) != len(tags2) {
			return false
		}

		for i, tag1 := range tags1 {
			if tag1 != tags2[i] {
				return false
			}
		}
	}

	return f.Name == field.Name && f.Type.EqualTo(field.Type)
}
