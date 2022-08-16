package parser

import (
	"fmt"
	"go/ast"
	"strings"
)

type ImportedType struct {
	Name        string
	PackageName string
	PackagePath string

	innerType Type

	selectorType *ast.SelectorExpr `json:"-"`
	file         *ast.File         `json:"-"`
	// imp          *ast.ImportSpec
}

func NewImported(spec *ast.SelectorExpr, file *ast.File) (*ImportedType, error) {
	ident, ok := spec.X.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("not ident")
	}

	var path string

	if file != nil {
		for _, imp := range file.Imports {
			var importName string
			if imp.Name == nil {
				importName = strings.Trim(
					imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1:],
					"\"",
				)
			} else {
				importName = imp.Name.Name
			}

			if importName == ident.Name {
				path = strings.Trim(imp.Path.Value, "\"")
				break
			}
		}
	}

	if path == "" {
		panic("path is nil")
	}

	return &ImportedType{
		Name:        spec.Sel.Name,
		PackageName: ident.Name,
		PackagePath: path,

		selectorType: spec,
		file:         file,
	}, nil
}

func (i ImportedType) GetName() string {
	return i.Name
}

func (i ImportedType) EqualTo(t Type) bool {
	typed, ok := t.(*ImportedType)
	if !ok {
		return false
	}

	return i.Name == typed.Name && i.PackageName == typed.PackageName
}

func (i ImportedType) Unpack() (Type, error) {
	if i.innerType == nil {
		var err error

		if i.innerType, err = Pkgs.FindType(i.PackagePath, i.Name); err != nil {
			return nil, err
		}
	}

	return i.innerType, nil
}

var ImportedEngine = ImportedType{
	Name:        "New",
	PackageName: "webapi",
}
