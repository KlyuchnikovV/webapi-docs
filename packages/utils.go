package packages

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
)

// func (parser *Parser) FindMethodByName(t ast.TypeSpec, selector string) *ast.FuncDecl {
// 	var result *ast.FuncDecl

// 	for _, pkg := range parser.packages {

// 		for _, file := range pkg.Files {
// 			ast.Inspect(file, func(n ast.Node) bool {
// 				funcDecl, ok := n.(*ast.FuncDecl)
// 				if !ok {
// 					return true
// 				}

// 				if !IsReceiverOfType(*funcDecl, t) {
// 					return true
// 				}

// 				if funcDecl.Name.Name == selector {
// 					result = funcDecl
// 					return false
// 				}

// 				return true
// 			})
// 		}
// 	}

// 	return result
// }

// func (parser *Parser) ImportPackage(path string) error {
// 	absolutePath := filepath.Join(parser.gopath, "src", path)

// 	pkgs, err := goparser.ParseDir(&token.FileSet{}, absolutePath, nil, goparser.AllErrors|goparser.ParseComments)
// 	if err != nil {
// 		return err
// 	}

// 	if len(path) > 0 {
// 		for name, pkg := range pkgs {
// 			var (
// 				index   = strings.Index(name, "/")
// 				newPath = path
// 			)

// 			if index != -1 {
// 				newPath = fmt.Sprintf("%s/%s", path, name[index+1:])
// 			}

// 			parser.packages[newPath] = NewPackage(*pkg)
// 		}
// 	}

// 	return nil
// }

// func (parser *Parser) FindModel(ident ast.Ident) (*ast.Object, error) {
// 	var obj *ast.Object

// 	for _, pkg := range parser.packages {
// 		for _, file := range pkg.Files {
// 			obj = file.Scope.Lookup(ident.Name)
// 			if obj != nil {
// 				break
// 			}
// 		}

// 		if obj != nil {
// 			break
// 		}
// 	}

// 	if obj == nil {
// 		return nil, fmt.Errorf("no model found for name: '%s'", ident.Name)
// 	}

// 	return obj, nil
// }

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

	return SameNodes(&t, typeSpec)
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
	case *ast.TypeSpec:
		typed2, ok := t2.(*ast.TypeSpec)
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

func CheckRoutersResultType(resultType ast.Expr) error {
	mapType, ok := resultType.(*ast.MapType)
	if !ok {
		return fmt.Errorf("not a map")
	}

	ident, ok := mapType.Key.(*ast.Ident)
	if !ok || ident.Name != "string" {
		return fmt.Errorf("map's key is not a 'string'")
	}

	selector, ok := mapType.Value.(*ast.SelectorExpr)
	if !ok {
		return fmt.Errorf("map's value is of wrong type")
	}

	valuePackage, ok := selector.X.(*ast.Ident)
	if !ok {
		return fmt.Errorf("map's value package couldn't be defined")
	}

	if valuePackage.Name != "webapi" {
		return fmt.Errorf("wrong map's value package")
	}

	if selector.Sel.Name != "RouterByPath" {
		return fmt.Errorf("map's value is not a 'RouterByPath'")
	}

	return nil
}

func CheckFuncDeclaration(
	funcDecl ast.FuncDecl,
	name string,
	params []func(ast.Expr) error,
	results ...func(ast.Expr) error,
) error {
	if funcDecl.Name.Name != name {
		return fmt.Errorf("not a '%s' func", name)
	}

	if len(params) != len(funcDecl.Type.Params.List) {
		return fmt.Errorf("'%s' should have %d parameters", name, len(params))
	}

	for i, param := range params {
		if err := param(funcDecl.Type.Params.List[i].Type); err != nil {
			return err
		}
	}

	if len(results) != len(funcDecl.Type.Results.List) {
		return fmt.Errorf("'%s' should have %d results", name, len(params))
	}

	for i, param := range params {
		if err := param(funcDecl.Type.Results.List[i].Type); err != nil {
			return err
		}
	}

	return nil
}

func getDirectoriesPaths(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, nil
	}

	paths, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	var directoriesPaths = []string{path}

	for _, innerPath := range paths {
		subPaths, err := getDirectoriesPaths(filepath.Join(path, innerPath))
		if err != nil {
			return nil, err
		}

		directoriesPaths = append(directoriesPaths, subPaths...)
	}

	return directoriesPaths, nil
}

func GetReturnType(funcDecl ast.FuncDecl) (*ast.TypeSpec, error) {
	if len(funcDecl.Type.Results.List) == 0 {
		return nil, fmt.Errorf("no results")
	}

	var typeObj *ast.Object

	switch typed := funcDecl.Type.Results.List[0].Type.(type) {
	case *ast.StarExpr:
		ident, ok := typed.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("pointer not ident")
		}

		typeObj = ident.Obj
	case *ast.SelectorExpr:
		typeObj = typed.Sel.Obj
	default:
		return nil, fmt.Errorf("can't parse %#v", typed)
	}

	if typeObj == nil {
		return nil, fmt.Errorf("no type")
	}

	typeSpec, ok := typeObj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("not a type")
	}

	return typeSpec, nil
}

func GetPrefix(statements []ast.Stmt) string {
	var result string

	for _, stmt := range statements {
		ast.Inspect(stmt, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if !IsMethod(*callExpr, NewSelector("webapi", "NewService")) {
				return true
			}

			if len(callExpr.Args) < 2 {
				return true
			}

			lit, ok := callExpr.Args[1].(*ast.BasicLit)
			if !ok {
				return true
			}

			result = strings.Trim(lit.Value, "\"")

			return false
		})
	}

	return result
}
