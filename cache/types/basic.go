package types

import "go/ast"

type BasicType struct {
	typeBase

	BasicType string
	Ident     *ast.Ident
}

func NewBasic(file *ast.File, name string, ident *ast.Ident) BasicType {
	if name == "" {
		name = ident.Name
	}

	return BasicType{
		typeBase:  newTypeBase(file, name),
		BasicType: ident.Name,

		Ident: ident,
	}
}

func NewSimpleBasicType(name string) BasicType {
	return BasicType{
		typeBase: typeBase{
			name: name,
		},
	}
}

func (b BasicType) EqualTo(t Type) bool {
	basic, ok := t.(BasicType)
	if !ok {
		return false
	}

	if b.BasicType != basic.BasicType {
		return false
	}

	return b.typeBase.EqualTo(basic)
}
