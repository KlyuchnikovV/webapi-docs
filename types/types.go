package types

import (
	"go/ast"
)

type Type interface {
	GetName() string
}

func NewType(spec ast.TypeSpec, file *ast.File) (Type, error) {
	return TypeFromExpr(spec.Name.Name, spec.Type, file)
}

func TypeFromExpr(name string, expr ast.Expr, file *ast.File) (Type, error) {
	switch spec := expr.(type) {
	case *ast.ArrayType:
		return NewArray(name, spec, file)
	case *ast.Ident:
		if spec.Obj == nil {
			return &Alias{
				Name: name,
				Type: NewBuiltIn(spec.Name),
				file: file,
			}, nil
		}

		return NewAlias(spec, file)
	case *ast.SelectorExpr:
		return NewImported(spec, file)
	case *ast.StructType:
		return NewStruct(name, spec, file)
	default:
		return nil, nil
		// panic(fmt.Errorf("not handled: %T", typed))
	}
}
