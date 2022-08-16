package parser

import (
	"go/ast"
)

type Type interface {
	GetName() string
	EqualTo(Type) bool
}

func NewType(spec ast.TypeSpec, file *ast.File) (Type, error) {
	var (
		result Type
		err    error
		name   = spec.Name.Name
	)

	if pkg, ok := Pkgs[file.Name.Name]; ok {
		for typeName, tt := range pkg.Types {
			if typeName == name {
				return tt, nil
			}
		}
	}

	switch typed := spec.Type.(type) {
	case *ast.ArrayType:
		result, err = NewArray(name, typed, file)
	case *ast.StructType:
		result, err = NewStruct(name, typed, file)
	case *ast.InterfaceType:
		result, err = NewInterface(name, typed, file)
	case *ast.Ident:
		result = NewBuiltIn(typed.Name)
	case *ast.SelectorExpr:
		result, err = NewImported(typed, file)
	case *ast.StarExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.MapType:
		result, err = NewMap(name, typed, file)
	case *ast.FuncType:
		result, err = NewFuncType(name, typed, file)
	case *ast.Ellipsis:
		result, err = TypeFromExpr(name, typed.Elt, file)
	case *ast.ChanType:
		result, err = TypeFromExpr(name, typed.Value, file)
	case *ast.UnaryExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.CompositeLit:
		result, err = TypeFromExpr(name, typed.Type, file)
	case *ast.CallExpr:
		result, err = TypeFromExpr(name, typed.Fun, file)
	case *ast.IndexExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.FuncLit:
		result, err = NewFuncType(name, typed.Type, file)
	case *ast.BinaryExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.BasicLit:
		result = NewBuiltIn(typed.Value)
	default:
		panic(typed)
	}

	return result, err
}

func TypeFromExpr(name string, expr ast.Expr, file *ast.File) (Type, error) {
	var (
		result Type
		err    error
	)

	if pkg, ok := Pkgs[file.Name.Name]; ok {
		for typeName, tt := range pkg.Types {
			if typeName == name {
				return tt, nil
			}
		}
	}

	switch typed := expr.(type) {
	case *ast.ArrayType:
		result, err = NewArray(name, typed, file)
	case *ast.StructType:
		result, err = NewStruct(name, typed, file)
	case *ast.InterfaceType:
		result, err = NewInterface(name, typed, file)
	case *ast.Ident:
		if typed.Obj == nil {
			result = NewBuiltIn(typed.Name)
		} else {
			result, err = NewType(*typed.Obj.Decl.(*ast.TypeSpec), file)
		}
	case *ast.SelectorExpr:
		result, err = NewImported(typed, file)
	case *ast.StarExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.MapType:
		result, err = NewMap(name, typed, file)
	case *ast.FuncType:
		result, err = NewFuncType(name, typed, file)
	case *ast.Ellipsis:
		result, err = TypeFromExpr(name, typed.Elt, file)
	case *ast.ChanType:
		result, err = TypeFromExpr(name, typed.Value, file)
	case *ast.UnaryExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.CompositeLit:
		result, err = TypeFromExpr(name, typed.Type, file)
	case *ast.CallExpr:
		result, err = TypeFromExpr(name, typed.Fun, file)
	case *ast.IndexExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.FuncLit:
		result, err = NewFuncType(name, typed.Type, file)
	case *ast.BinaryExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	case *ast.BasicLit:
		result = NewBuiltIn(typed.Value)
	case *ast.TypeAssertExpr:
		result, err = TypeFromExpr(name, typed.X, file)
	default:
		panic(typed)
	}

	return result, err
}

func TypeFromVariable(name string, expr ast.Expr, file *ast.File, pkgName string) (Type, error) {
	// if v, ok := vars[name]; ok {
	// 	return v.Type, nil
	// }

	var (
		result Type
		err    error
	)

	switch typed := expr.(type) {
	// case *ast.ArrayType:
	// 	result, err = NewArray(name, typed, file)
	// case *ast.StructType:
	// 	result, err = NewStruct(name, typed, file)
	// case *ast.InterfaceType:
	// 	result, err = NewInterface(name, typed, file)
	// case *ast.Ident:
	// 	if typed.Obj == nil {
	// 		result = NewBuiltIn(typed.Name)
	// 	} else {
	// 		result, err = NewType(*typed.Obj.Decl.(*ast.TypeSpec), file)
	// 	}
	case *ast.SelectorExpr:
		v, ok := Pkgs[pkgName].Variables[typed.X.(*ast.Ident).Name]
		if !ok {
			return NewImported(typed, file)
		}

		st, ok := v.Type.(*Struct)
		if !ok {
			panic("not ok")
		}

		return st.Methods[typed.Sel.Name].Results[0], nil
	// case *ast.StarExpr:
	// 	result, err = TypeFromExpr(name, typed.X, file)
	// case *ast.MapType:
	// 	result, err = NewMap(name, typed, file)
	// case *ast.FuncType:
	// 	result, err = NewFuncType(name, typed, file)
	// case *ast.Ellipsis:
	// 	result, err = TypeFromExpr(name, typed.Elt, file)
	// case *ast.ChanType:
	// 	result, err = TypeFromExpr(name, typed.Value, file)
	// case *ast.UnaryExpr:
	// 	result, err = TypeFromExpr(name, typed.X, file)
	// case *ast.CompositeLit:
	// 	result, err = TypeFromExpr(name, typed.Type, file)
	case *ast.CallExpr:
		result, err = TypeFromVariable(name, typed.Fun, file, pkgName)
	// case *ast.IndexExpr:
	// 	result, err = TypeFromExpr(name, typed.X, file)
	// case *ast.FuncLit:
	// 	result, err = NewFuncType(name, typed.Type, file)
	// case *ast.BinaryExpr:
	// 	result, err = TypeFromExpr(name, typed.X, file)
	// case *ast.BasicLit:
	// 	result = NewBuiltIn(typed.Value)
	// case *ast.TypeAssertExpr:
	// 	result, err = TypeFromExpr(name, typed.X, file)
	default:
		panic(typed)
	}

	return result, err
}
