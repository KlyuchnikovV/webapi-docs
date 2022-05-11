package objects

import (
	"fmt"
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/constants"
)

type Field struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`

	// WasFound bool      `json:"-" xml:"-"`
	// Ident    ast.Ident `json:"-" xml:"-"`
}

func (c *Components) NewField(ident ast.Ident) Schema {
	identType, ok := constants.ConvertFieldType(ident.Name)
	if ok {
		return Field{
			Type:   identType,
			Format: constants.GetFieldTypeFormat(ident.Name),
		}
	}

	var obj = ident.Obj
	if obj == nil {
		return nil
		// model, err := cache.FindModelByName(ident.Name)
		// if err != nil {
		// 	panic(err)
		// }

		// obj = model
	}

	typeSpec, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		panic(fmt.Errorf("field could be parsed only to type specification (got: '%T')", obj.Decl))
	}

	switch t := typeSpec.Type.(type) {
	case *ast.Ident:
		return c.NewField(*t)
	case *ast.StructType:
		if _, ok := c.loopController[ident.Name]; !ok {
			var err error

			c.loopController[ident.Name] = struct{}{}
			c.Schemas[ident.Name], err = c.NewObject(*t)
			delete(c.loopController, ident.Name)

			if err != nil {
				panic(err)
			}
		}

		return NewReference(ident.Name, "schemas")

	default:
		panic(fmt.Errorf("fields inner type not handled"))
	}
}

func (f Field) SchemaType() string {
	return f.Type
}

func (f Field) EqualTo(s interface{}) bool {
	typed, ok := s.(Field)
	if !ok {
		return false
	}

	return f.Type == typed.Type && f.Format == typed.Format
}
