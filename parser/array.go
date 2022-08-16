package parser

import (
	"go/ast"
)

type Array struct {
	Name string `json:"name,omitempty"`
	Type Type   `json:"type,omitempty"`

	arrayType *ast.ArrayType `json:"-"`
	file      *ast.File      `json:"-"`
}

func NewArray(name string, spec *ast.ArrayType, file *ast.File) (*Array, error) {
	// TODO: check empty type's name
	t, err := TypeFromExpr("", spec.Elt, file)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = "array"
	}

	return &Array{
		Name: name,
		Type: t,

		arrayType: spec,
		file:      file,
	}, nil
}

func (a Array) GetName() string {
	return a.Name
}

func (a Array) EqualTo(t Type) bool {
	typed, ok := t.(*Array)
	if !ok {
		return false
	}

	return a.Name == typed.Name && a.Type.EqualTo(typed.Type)
}

// func (a Array) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(struct{
// 		Name string

// 	}{

// 	})
// }
