package types

import (
	"go/ast"
)

type Array struct {
	Name string `json:"name,omitempty"`
	Type `json:"type,omitempty"`

	arrayType *ast.ArrayType
	file      *ast.File
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

// func (a Array) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(struct{
// 		Name string

// 	}{

// 	})
// }
