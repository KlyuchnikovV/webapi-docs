package types

import (
	"fmt"
	"go/ast"
)

type ImportedType struct {
	Name    string
	Package string

	selectorType *ast.SelectorExpr
	// imp          *ast.ImportSpec
	file *ast.File
}

func NewImported(spec *ast.SelectorExpr, file *ast.File) (*ImportedType, error) {
	ident, ok := spec.X.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("not ident")
	}

	return &ImportedType{
		Name:    spec.Sel.Name,
		Package: ident.Name,

		selectorType: spec,
		file:         file,
	}, nil
}

func (i ImportedType) GetName() string {
	return i.Name
}
