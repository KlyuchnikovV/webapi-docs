package types

import (
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/utils"
)

type ImportedType struct {
	*typeBase

	Package string
	ts      *ast.SelectorExpr
	imp     *ast.ImportSpec
}

func NewImported(file *ast.File, selector *ast.SelectorExpr, tag *ast.BasicLit) *ImportedType {
	var name = getBaseTypeAlias(selector, 0)

	var path, imp = utils.FindImport(*file, name)

	return &ImportedType{
		typeBase: newTypeBase(file, selector.Sel.Name, tag),

		Package: path,
		imp:     imp,

		ts: selector,
	}
}

func NewSimpleImported(name, path string) ImportedType {
	return ImportedType{
		typeBase: &typeBase{
			name: name,
		},
		Package: path,
	}
}

func (i ImportedType) AddMethod(FuncType) {}

func (i ImportedType) Field(name string) Type {
	return i.Field(name)
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

func (i ImportedType) Schema() Schema {
	return nil
}

func (i ImportedType) EqualTo(t Type) bool {
	it, ok := t.(ImportedType)
	if !ok {
		return false
	}

	if i.Package != it.Package {
		return false
	}

	return i.typeBase.EqualTo(it)
}
