package objects

import "go/ast"

type Object struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
}

// TODO: check embedded types and anonymous

func (c *Components) NewObject(t ast.StructType) (*Object, error) {
	var props = make(map[string]Schema)

	for _, item := range t.Fields.List {
		var (
			prop Schema
			err  error
		)

		switch typed := item.Type.(type) {
		case *ast.Ident:
			prop = c.NewField(*typed)
		case *ast.StructType:
			prop, err = c.NewObject(*typed)
		case *ast.ArrayType:
			prop, err = c.NewArray(*typed)
		}

		if err != nil {
			return nil, err
		}

		props[getFieldName(item)] = prop
	}

	return &Object{
		Type:       "object",
		Properties: props,
	}, nil
}

func (o Object) SchemaType() string {
	return o.Type
}

func (o Object) EqualTo(s interface{}) bool {
	typed, ok := s.(Object)
	if !ok {
		return false
	}

	if o.Type != typed.Type {
		return false
	}

	if len(o.Properties) != len(typed.Properties) {
		return false
	}

	for name, prop := range o.Properties {
		if typedProp, ok := typed.Properties[name]; !ok || !prop.EqualTo(typedProp) {
			return false
		}
	}

	return true
}
