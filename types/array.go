package types

import (
	"encoding/json"
	"go/ast"
)

type ArrayType struct {
	*typeBase

	ItemType  Type
	ArrayType *ast.ArrayType
}

func NewArray(file *ast.File, name string, array *ast.ArrayType, innerType ast.Expr, tag *ast.BasicLit) ArrayType {
	return ArrayType{
		typeBase:  newTypeBase(file, name, tag, ArraySchemaType),
		ItemType:  NewType(file, "", innerType, nil),
		ArrayType: array,
	}
}

func NewSimpleArrayType(name string, itemType Type) ArrayType {
	return ArrayType{
		typeBase: &typeBase{
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

	return a.typeBase.EqualTo(array.typeBase)
}

func (a ArrayType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type        SchemaType `json:"type"`
		Description string     `json:"description,omitempty"`
		Example     string     `json:"example,omitempty"`
		Required    bool       `json:"required,omitempty"`
		Items       Type       `json:"items,omitempty"`
	}{
		Type:        a.Type,
		Description: a.Description,
		Example:     a.Example,
		Required:    a.Required,
		Items:       a.ItemType,
	})
}
