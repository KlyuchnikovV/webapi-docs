package types

import "go/ast"

type StructType struct {
	typeBase

	StructType *ast.StructType
}

func NewStruct(file *ast.File, name string, str *ast.StructType) StructType {
	var result = StructType{
		typeBase: newTypeBase(file, name),

		StructType: str,
	}

	for _, field := range str.Fields.List {
		name, t := NewTypeFromField(file, field)

		result.fields[name] = t
	}

	return result
}
