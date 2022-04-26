package parser

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

func (parser *Parser) FindMethod(pkg string, t ast.TypeSpec, selector ast.FuncDecl) *ast.FuncDecl {
	var result *ast.FuncDecl

	for _, file := range parser.packages[pkg].Files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !IsReceiverOfType(*funcDecl, t) {
				return true
			}

			if SameNodes(funcDecl, &selector) {
				result = funcDecl
				return false
			}

			return true
		})
	}

	return result
}

func (parser *Parser) Implements(pkg string, t ast.TypeSpec, i ast.InterfaceType) bool {
	var (
		methods = make(map[string]*ast.FuncType)
		found   = make(map[string]struct{})
	)

	for _, method := range i.Methods.List {
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		methods[method.Names[0].Name] = funcType
	}

	for _, file := range parser.packages[pkg].Files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !IsReceiverOfType(*funcDecl, t) {
				return true
			}

			// TODO: for now - simple name and types checking
			interfaceFunc, ok := methods[funcDecl.Name.Name]
			if !ok {
				return true
			}

			if SameNodes(funcDecl.Type, interfaceFunc) {
				found[funcDecl.Name.Name] = struct{}{}
			}

			return true
		})
	}

	for key := range methods {
		if _, ok := found[key]; !ok {
			return false
		}
	}

	return true
}

func (parser *Parser) ImportPackage(path string) error {
	absolutePath := filepath.Join(parser.gopath, "src", path)

	pkgs, err := goparser.ParseDir(&token.FileSet{}, absolutePath, nil, goparser.AllErrors|goparser.ParseComments)
	if err != nil {
		return err
	}

	if len(path) > 0 {
		for name, pkg := range pkgs {
			var (
				index   = strings.Index(name, "/")
				newPath = path
			)

			if index != -1 {
				newPath = fmt.Sprintf("%s/%s", path, name[index+1:])
			}

			parser.packages[newPath] = *pkg
		}
	}

	return nil
}

func (parser *Parser) FindModel(ident ast.Ident) *ast.Object {
	var obj *ast.Object

	for _, pkg := range parser.packages {
		for _, file := range pkg.Files {
			obj = file.Scope.Lookup(ident.Name)
			if obj != nil {
				break
			}
		}

		if obj != nil {
			break
		}
	}

	return obj
}

func IsReceiverOfType(decl ast.FuncDecl, t ast.TypeSpec) bool {
	if decl.Recv == nil {
		return false
	}

	if len(decl.Recv.List) == 0 {
		return false
	}

	var ident ast.Ident

	switch typed := decl.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		// TODO: Could be problems
		i, ok := typed.X.(*ast.Ident)
		if !ok {
			return false
		}

		ident = *i
	case *ast.Ident:
		ident = *typed
	default:
		return false
	}

	typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return false
	}

	return *typeSpec == t
}

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
		// panic(fmt.Errorf("can't parse %#v", typed1))
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

	return SameNodes(callSel, &selector)
}

func NewSelector(prefix, fun string) ast.SelectorExpr {
	return ast.SelectorExpr{
		X:   &ast.Ident{Name: prefix},
		Sel: &ast.Ident{Name: fun},
	}
}

func extractMethod(arg ast.Expr) string {
	if sel, ok := arg.(*ast.SelectorExpr); ok {
		return strings.ToLower(sel.Sel.Name)
	}

	return ""
}

func extractPath(arg ast.Expr) string {
	if sel, ok := arg.(*ast.BasicLit); ok {
		return fmt.Sprintf("/%s", strings.Trim(sel.Value, "\"/"))
	}

	return ""
}
