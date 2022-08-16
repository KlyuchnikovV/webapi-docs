package parser

import (
	"go/ast"
	"strings"
)

type Interface struct {
	Name      string           `json:"name,omitempty"`
	Functions map[string]Field `json:"funcs,omitempty"`

	it   *ast.InterfaceType `json:"-"`
	file *ast.File          `json:"-"`
}

func NewInterface(name string, spec *ast.InterfaceType, file *ast.File) (*Interface, error) {
	var result = Interface{
		Name:      name,
		Functions: make(map[string]Field),

		it:   spec,
		file: file,
	}

	if spec.Methods != nil {
		for _, field := range spec.Methods.List {
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

			result.Functions[name] = *field
		}
	}

	return &result, nil
}

func (it Interface) GetName() string {
	return it.Name
}

func (it Interface) EqualTo(t Type) bool {
	typed, ok := t.(*Interface)
	if !ok {
		return false
	}

	if len(it.Functions) != len(typed.Functions) {
		return false
	}

	for name, field1 := range it.Functions {
		field2, ok := typed.Functions[name]
		if !ok {
			return false
		}

		if !field1.EqualTo(field2) {
			return false
		}
	}

	return it.Name == typed.Name
}
