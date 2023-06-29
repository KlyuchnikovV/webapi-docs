package types

import (
	"fmt"
	"go/ast"
)

type FuncType struct {
	*typeBase

	Parameters []Type
	Results    []Type
	Statements []interface{}

	Body []ast.Stmt
	decl *ast.FuncType
}

func NewFunc(file *ast.File, decl *ast.FuncType, name string, body []ast.Stmt, tag *ast.BasicLit) FuncType {
	var result = FuncType{
		typeBase: newTypeBase(file, name, tag, EmptySchemaType),

		Parameters: make([]Type, 0),
		Results:    make([]Type, 0),
		decl:       decl,
		Body:       body,
	}

	if decl.Params != nil {
		for _, param := range decl.Params.List {
			_, t := NewTypeFromField(file, param)
			result.Parameters = append(result.Parameters, t)
		}
	}

	if decl.Results != nil {
		for _, param := range decl.Results.List {
			_, t := NewTypeFromField(file, param)
			result.Results = append(result.Results, t)
		}
	}

	result.ReturnStatements2()

	return result
}

func NewMethodFromField(file *ast.File, field *ast.Field) (string, *FuncType) {
	var name string
	if len(field.Names) != 0 {
		name = field.Names[0].Name
	}

	ft, ok := field.Type.(*ast.FuncType)
	if !ok {
		return name, nil
	}

	var t = NewFunc(file, ft, name, nil, nil)
	if name == "" {
		name = t.Name()
	}

	return name, &t
}

func NewFuncDeclaration(name string, params []Type, results []Type) FuncType {
	return FuncType{
		typeBase: &typeBase{
			name: name,
		},
		Parameters: params,
		Results:    results,
	}
}

func (f FuncType) ReturnStatements() []ast.ReturnStmt {
	var returns = make([]ast.ReturnStmt, 0)

	for _, stmt := range f.Body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if ret, ok := n.(*ast.ReturnStmt); ok {
				returns = append(returns, *ret)
				return false
			}

			return true
		})
	}

	return returns
}

func (FuncType) AddMethod(FuncType) {}

func (f FuncType) EqualTo(t Type) bool {
	fun, ok := t.(FuncType)
	if !ok {
		return false
	}

	for i, param := range f.Parameters {
		if !param.EqualTo(fun.Parameters[i]) {
			return false
		}
	}

	for i, result := range f.Results {
		if !result.EqualTo(fun.Results[i]) {
			return false
		}
	}

	return f.typeBase.EqualTo(fun.typeBase)
}

func (f FuncType) Implements(it InterfaceType) bool {
	if len(it.fields) > 1 {
		return false
	}

	method, ok := it.fields[f.name].(FuncType)
	if !ok {
		return false
	}

	return f.typeBase.EqualTo(method.typeBase)
}

func (f FuncType) ReturnStatements2() []ReturnStatement {
	var returns = make([]ReturnStatement, 0)

	for _, stmt := range f.Body {
		ast.Inspect(stmt, func(n ast.Node) bool {
			ret, ok := n.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			result := NewReturn(f.file, *ret)
			if result == nil {
				return true
			}

			returns = append(returns, *result)
			return false
		})
	}

	return returns
}

type ReturnStatement struct {
	t Type
}

func NewReturn(file *ast.File, statements ast.ReturnStmt) *ReturnStatement {
	if len(statements.Results) != 1 {
		return nil // Not webapi response
	}

	var result = ReturnStatement{
		t: NewType(file, "", &statements.Results[0], nil),
	}

	fmt.Printf("%#v - %t\n", result.t, result.t)

	// switch typed := statements.Results[0].(type) {
	// case *ast.BasicLit:
	// 	result.t = NewBasicFromBasicLit(file, "", typed, nil)
	// 	fmt.Printf("result is %#v - %T\n", result.t, result.t)
	// case *ast.CompositeLit:
	// 	result.t = NewType(file, "", &statements.Results[0], nil)
	// 	fmt.Printf("result is %#v - %T\n", result.t, result.t)
	// case *ast.CallExpr:
	// 	selector, ok := typed.Fun.(*ast.SelectorExpr)
	// 	if !ok {
	// 		panic("wrong")
	// 	}

	// 	var imported = NewImported(file, selector, nil)

	// 	if !imported.IsWebAPI() {
	// 		break
	// 	}

	// 	fmt.Printf("func is %s, args are:\n", imported.name)
	// 	for i, arg := range typed.Args {
	// 		fmt.Printf("	%d - %#v\n", i, arg)
	// 	}

	// 	result.t = imported
	// default:
	// 	panic(fmt.Sprintf("%v - %T\n", typed, typed))
	// }

	return nil //statements.Results
}
