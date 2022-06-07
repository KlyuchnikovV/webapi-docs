package types

import (
	"go/ast"
)

type Alias struct {
	Name string `json:"name,omitempty"`
	Type Type   `json:"type,omitempty"`

	identType *ast.Ident `json:"-"`
	file      *ast.File  `json:"-"`
}

func NewAlias(spec *ast.Ident, file *ast.File) (Type, error) {
	var (
		t   Type
		err error
	)

	// TODO: alias of baisc type
	t, err = NewType(*spec.Obj.Decl.(*ast.TypeSpec), file)
	if err != nil {
		return nil, err
	}

	return &Alias{
		Name: spec.Name,
		Type: t,

		identType: spec,
		file:      file,
	}, nil
}

func (s Alias) GetName() string {
	return s.Name
}
