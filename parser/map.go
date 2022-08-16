package parser

import "go/ast"

type MapType struct {
	Name  string `json:"name,omitempty"`
	Key   Type   `json:"key,omitempty"`
	Value Type   `json:"value,omitempty"`

	mapType *ast.MapType `json:"-"`
	file    *ast.File    `json:"-"`
}

func NewMap(name string, spec *ast.MapType, file *ast.File) (*MapType, error) {
	keyType, err := NewField("", spec.Key, "", file)
	if err != nil {
		return nil, err
	}

	valueType, err := NewField("", spec.Value, "", file)
	if err != nil {
		return nil, err
	}

	return &MapType{
		Name:    name,
		Key:     keyType.Type,
		Value:   valueType.Type,
		mapType: spec,
		file:    file,
	}, nil
}

func (m MapType) GetName() string {
	return m.Name
}

func (m MapType) EqualTo(t Type) bool {
	typed, ok := t.(*MapType)
	if !ok {
		return false
	}

	return m.Name == typed.Name &&
		m.Key.EqualTo(typed.Key) && m.Value.EqualTo(typed.Value)
}
