package types

import "go/ast"

type InterfaceType struct {
	*typeBase

	ts *ast.InterfaceType
}

func NewInterface(file *ast.File, name string, interf *ast.InterfaceType, tag *ast.BasicLit) InterfaceType {
	var result = InterfaceType{
		typeBase: newTypeBase(file, name, tag, EmptySchemaType),
		ts:       interf,
	}

	if interf.Methods != nil {
		for _, method := range interf.Methods.List {
			name, t := NewMethodFromField(file, method)

			result.fields[name] = *t
		}
	}

	return result
}

func NewInterfaceFields(fields map[string]Type) InterfaceType {
	return InterfaceType{
		typeBase: &typeBase{
			fields: fields,
		},
	}
}

func RoutersInterface() InterfaceType {
	return NewInterfaceFields(map[string]Type{
		"Prefix": NewFuncDeclaration("Prefix", nil,
			[]Type{
				NewSimpleBasicType("string"),
			},
		),
		"Routers": RoutersFunc(),
	})
}

func RoutersFuncInterface() InterfaceType {
	return NewInterfaceFields(map[string]Type{
		"Routers": RoutersFunc(),
	})
}

func RoutersFunc() Type {
	return NewFuncDeclaration("Routers", nil,
		[]Type{
			NewSimpleMap("",
				NewSimpleBasicType("string"),
				NewSimpleImported("RouterByPath", "github.com/KlyuchnikovV/webapi"),
			),
		},
	)
}
