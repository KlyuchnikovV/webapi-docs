package types

import (
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
		typeBase: newTypeBase(file, name, tag),

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

	// for _, stmt := range body {
	// 	switch typed := stmt.(type) {
	// 	case *ast.ReturnStmt:
	// 		r := NewReturn(file, *typed)
	// 		bytes, _ := json.Marshal(r)

	// 		fmt.Printf("stmt: %s\n", string(bytes))
	// 	}
	// }

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

	return f.typeBase.EqualTo(fun)
}
