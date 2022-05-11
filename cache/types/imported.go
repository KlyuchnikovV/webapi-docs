package types

import (
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/utils"
)

type ImportedType struct {
	typeBase

	Package string
	ts      *ast.SelectorExpr
	imp     *ast.ImportSpec
}

func NewImported(file *ast.File, selector *ast.SelectorExpr) *ImportedType {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil
	}

	var path, imp = utils.FindImport(*file, ident.Name)

	return &ImportedType{
		typeBase: newTypeBase(file, selector.Sel.Name),

		Package: path,
		imp:     imp,

		ts: selector,
	}
}

func NewSimpleImported(name, path string) ImportedType {
	return ImportedType{
		typeBase: typeBase{
			name: name,
		},
	}
}

func (i ImportedType) AddMethod(FuncType) {}


func (i ImportedType) Field(name string) Type {
	// if i.fields != nil {
		return i.Field(name)
	// }

	// ParsePackage

}

func (i ImportedType) Method(name string) *FuncType {
	var field = i.fields[name]

	if fun, ok := field.(FuncType); ok {
		return &fun
	}

	return nil
}

func (i ImportedType) Fields() map[string]Type {
	// TODO:
	return nil
}

