package types

import (
	"go/ast"
)

type FuncType struct {
	*typeBase

	Receiver   Type
	Parameters []Type
	Results    []ReturnStatement
	Statements []interface{}

	// Base AST variables
	Body []ast.Stmt
	decl *ast.FuncType
}

func NewFunc(file *ast.File, decl ast.FuncDecl, name string, tag *ast.BasicLit) FuncType {
	var result = FuncType{
		typeBase: newTypeBase(file, name, tag),

		Parameters: make([]Type, 0),
		Results:    make([]ReturnStatement, 0),

		decl: decl.Type,
		Body: decl.Body.List,
	}

	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			_, t := NewTypeFromField(file, param)
			result.Parameters = append(result.Parameters, t)
		}
	}

	if decl.Type.Results != nil {
		for _, param := range decl.Type.Results.List {
			_, t := NewTypeFromField(file, param)

			result.Results = append(result.Results, ReturnStatement{
				Type: t,
			})
		}
	}

	for _, stmt := range decl.Body.List {
		switch typed := stmt.(type) {
		case *ast.ReturnStmt:
			for i, r := range typed.Results {
				result.Results[i] = NewValue(r, result.Results[i].Type)
			}
		}
	}

	return result
}

func NewFuncFromType(file *ast.File, ft *ast.FuncType, name string) FuncType {
	return NewFunc(file, ast.FuncDecl{
		Type: ft,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}, name, nil)
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

	var t = NewFuncFromType(file, ft, name)
	if name == "" {
		name = t.Name()
	}

	return name, &t
}

func NewFuncDeclaration(name string, params []Type, results []Type) FuncType {
	var returns = make([]ReturnStatement, len(results))

	for i := range results {
		returns[i] = ReturnStatement{
			Type: results[i],
		}
	}

	return FuncType{
		typeBase: &typeBase{
			name: name,
		},
		Parameters: params,
		Results:    returns,
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

	return f.typeBase.EqualTo(fun)
}

type ReturnStatement struct {
	Type

	Value interface{}
}
