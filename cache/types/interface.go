package types

import "go/ast"

type InterfaceType struct {
	typeBase

	ts *ast.InterfaceType
}

func NewInterface(file *ast.File, name string, interf *ast.InterfaceType) InterfaceType {
	var result = InterfaceType{
		typeBase: newTypeBase(file, name),
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
		typeBase: typeBase{
			fields: fields,
		},
	}
}
