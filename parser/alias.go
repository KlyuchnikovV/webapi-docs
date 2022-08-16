package parser

import (
	"go/ast"
)

type Alias struct {
	Name string `json:"name,omitempty"`
	Type Type   `json:"type,omitempty"`

	identType *ast.Ident `json:"-"`
	File      *ast.File  `json:"-"`
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
		File:      file,
	}, nil
}

func (a Alias) GetName() string {
	return a.Name
}

func (a Alias) EqualTo(t Type) bool {
	typed, ok := t.(*Alias)
	if !ok {
		return false
	}

	return a.Name == typed.Name && a.Type.EqualTo(typed.Type)
}
