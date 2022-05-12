package parser

import (
	"go/ast"
)

func SameNodes(t1, t2 ast.Node) bool {
	if t1 == nil || t2 == nil {
		return t1 == t2
	}

	switch typed1 := t1.(type) {
	case *ast.Ident:
		typed2, ok := t2.(*ast.Ident)
		if !ok {
			return false
		}

		return typed1.Name == typed2.Name
	case *ast.FieldList:
		typed2, ok := t2.(*ast.FieldList)
		if !ok {
			return false
		}

		if typed1 == nil || typed2 == nil {
			return typed1 == typed2
		}

		if len(typed1.List) != len(typed2.List) {
			return false
		}

		for i := range typed1.List {
			if !SameNodes(typed1.List[i].Type, typed2.List[i].Type) {
				return false
			}
		}

		return true
	case *ast.ParenExpr:
		typed2, ok := t2.(*ast.ParenExpr)
		if !ok {
			return false
		}

		return SameNodes(typed1.X, typed2.X)
	case *ast.StarExpr:
		typed2, ok := t2.(*ast.StarExpr)
		if !ok {
			return false
		}

		return SameNodes(typed1.X, typed2.X)
	case *ast.SelectorExpr:
		typed2, ok := t2.(*ast.SelectorExpr)
		if !ok {
			return false
		}

		if !SameNodes(typed1.X, typed2.X) {
			return false
		}

		return SameNodes(typed1.Sel, typed2.Sel)
	case *ast.ArrayType:
		typed2, ok := t2.(*ast.ArrayType)
		if !ok {
			return false
		}

		return SameNodes(typed1.Elt, typed2.Elt)
	case *ast.ChanType:
		typed2, ok := t2.(*ast.ChanType)
		if !ok {
			return false
		}

		return SameNodes(typed1.Value, typed2.Value)
	case *ast.FuncType:
		typed2, ok := t2.(*ast.FuncType)
		if !ok {
			return false
		}

		if !SameNodes(typed1.Params, typed2.Params) {
			return false
		}

		return SameNodes(typed1.Results, typed2.Results)
	case *ast.MapType:
		typed2, ok := t2.(*ast.MapType)
		if !ok {
			return false
		}

		if !SameNodes(typed1.Key, typed2.Key) {
			return false
		}

		return SameNodes(typed1.Value, typed2.Value)
	case *ast.StructType:
		typed2, ok := t2.(*ast.StructType)
		if !ok {
			return false
		}

		return SameNodes(typed1.Fields, typed2.Fields)
	case *ast.InterfaceType:
		typed2, ok := t2.(*ast.InterfaceType)
		if !ok {
			return false
		}

		return SameNodes(typed1.Methods, typed2.Methods)
	case *ast.FuncDecl:
		typed2, ok := t2.(*ast.FuncDecl)
		if !ok {
			return false
		}

		if !SameNodes(typed1.Name, typed2.Name) {
			return false
		}

		if !SameNodes(typed1.Type, typed2.Type) {
			return false
		}

		return true
	default:
		return false
	}
}

func IsMethod(call ast.CallExpr, selector ast.SelectorExpr) bool {
	callSel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	for ok {
		var typed *ast.SelectorExpr
		typed, ok = callSel.X.(*ast.SelectorExpr)

		if ok {
			callSel = typed
		}
	}

	// TODO: remove method
	return SameNodes(callSel, &selector)
}

func NewSelector(prefix, fun string) ast.SelectorExpr {
	return ast.SelectorExpr{
		X:   &ast.Ident{Name: prefix},
		Sel: &ast.Ident{Name: fun},
	}
}

