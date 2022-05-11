package types

import (
	"go/ast"
)

type ArrayType struct {
	typeBase

	ItemType  Type
	ArrayType *ast.ArrayType
}

func NewArray(file *ast.File, name string, array *ast.ArrayType, innerType *ast.Expr) ArrayType {
	return ArrayType{
		typeBase:  newTypeBase(file, name),
		ItemType:  NewType(file, "", innerType),
		ArrayType: array,
	}
}

func NewSimpleArrayType(name string, itemType Type) ArrayType {
	return ArrayType{
		typeBase: typeBase{
			name: name,
		},
		ItemType: itemType,
	}
}

func (a ArrayType) EqualTo(t Type) bool {
	array, ok := t.(ArrayType)
	if !ok {
		return false
	}

	if a.ItemType != array.ItemType {
		return false
	}

	return a.typeBase.EqualTo(array)
}
