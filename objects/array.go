package objects

import "go/ast"

type Array struct {
	Type  string `json:"type"`
	Items Schema `json:"items"`
}

func (c *Components) NewArray(t ast.ArrayType) (*Array, error) {
	var (
		items Schema
		err   error
	)

	switch typed := t.Elt.(type) {
	case *ast.Ident:
		items = c.NewField(*typed)
	case *ast.StructType:
		items, err = c.NewObject(*typed)
	case *ast.ArrayType:
		items, err = c.NewArray(*typed)
	}

	return &Array{
		Type:  "array",
		Items: items,
	}, err
}

func (a Array) SchemaType() string {
	return a.Type
}

func (a Array) EqualTo(s interface{}) bool {
	typed, ok := s.(Array)
	if !ok {
		return false
	}

	if a.Type != typed.Type {
		return false
	}

	return a.Items.EqualTo(typed.Items)
}
