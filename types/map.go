package types

import "go/ast"

type MapType struct {
	*typeBase

	Key   Type
	Value Type

	ts *ast.MapType
}

func NewMap(file *ast.File, name string, str *ast.MapType, tag *ast.BasicLit) MapType {
	return MapType{
		typeBase: newTypeBase(file, name, nil, EmptySchemaType),

		Key:   NewType(file, "", &str.Key, nil),
		Value: NewType(file, "", &str.Value, nil),

		ts: str,
	}
}

func NewSimpleMap(name string, key Type, value Type) MapType {
	return MapType{
		typeBase: &typeBase{
			name: name,
		},

		Key:   key,
		Value: value,
	}
}

func (m MapType) EqualTo(t Type) bool {
	mapType, ok := t.(MapType)
	if !ok {
		return false
	}

	if !m.Key.EqualTo(mapType.Key) {
		return false
	}

	if !m.Value.EqualTo(mapType.Value) {
		return false
	}

	return m.typeBase.EqualTo(mapType.typeBase)
}
