package parser

import (
	"go/ast"
	"go/token"

	"github.com/KlyuchnikovV/webapi-docs/parser/fun"
)

type FuncType struct {
	Name    string `json:"name,omitempty"`
	Params  []Type `json:"params,omitempty"`
	Results []Type `json:"results,omitempty"`

	funcType *ast.FuncType `json:"-"`
	file     *ast.File     `json:"-"`
}

func NewFuncType(name string, spec *ast.FuncType, file *ast.File) (*FuncType, error) {
	var result = FuncType{
		Name:    name,
		Params:  make([]Type, 0),
		Results: make([]Type, 0),

		funcType: spec,
		file:     file,
	}

	if spec.Params != nil {
		for _, arg := range spec.Params.List {
			var name string
			if len(arg.Names) > 0 {
				name = arg.Names[0].Name
			}

			field, err := TypeFromExpr(name, arg.Type, file)
			if err != nil {
				return nil, err
			}

			result.Params = append(result.Params, field)
		}
	}

	if spec.Results != nil {
		for _, res := range spec.Results.List {
			var name string
			if len(res.Names) > 0 {
				name = res.Names[0].Name
			}

			field, err := TypeFromExpr(name, res.Type, file)
			if err != nil {
				return nil, err
			}

			result.Results = append(result.Results, field)
		}
	}

	return &result, nil
}

func (fg FuncType) GetName() string {
	return fg.Name
}

func (fg FuncType) EqualTo(t Type) bool {
	typed, ok := t.(*FuncType)
	if !ok {
		return false
	}

	if len(fg.Params) != len(typed.Params) || len(fg.Results) != len(typed.Results) {
		return false
	}

	for i, param := range fg.Params {
		if !param.EqualTo(typed.Params[i]) {
			return false
		}
	}

	for i, result := range fg.Results {
		if !result.EqualTo(typed.Results[i]) {
			return false
		}
	}

	return fg.Name == typed.Name
}

// func NewParam(arg *ast.Field, file *ast.File) (*Field, error) {
// 	var name string
// 	if len(arg.Names) > 0 {
// 		name = arg.Names[0].Name
// 	}

// 	return NewField(name, arg.Type, "", file)
// }

type FuncDecl struct {
	FuncType

	// Body             *ast.BlockStmt   `json:"-"`

	fun.Body
	// ReturnStatements []ast.ReturnStmt `json:"-"`

	funcDecl *ast.FuncDecl `json:"-"`
	File     *ast.File     `json:"-"`
}

func NewFuncDecl(name string, spec *ast.FuncDecl, file *ast.File) (*FuncDecl, error) {
	funcType, err := NewFuncType(name, spec.Type, file)
	if err != nil {
		return nil, err
	}

	var decl = &FuncDecl{
		FuncType: *funcType,
		funcDecl: spec,
		Body:     *fun.NewBody(spec.Body),
		File:     file,
	}

	if spec.Recv != nil && len(spec.Recv.List) == 1 {
		var receiver = spec.Recv.List[0]

		tt, err := TypeFromExpr("", receiver.Type, file)
		if err != nil {
			return nil, err
		}

		t, ok := Pkgs[file.Name.Name].Types[tt.GetName()]
		if !ok {
			panic("not ok")
		}

		switch typed := t.(type) {
		case *Struct:
			typed.Methods[name] = decl
		default:
			panic("not handled")
		}

		Pkgs[file.Name.Name].Types[tt.GetName()] = t
	}

	return decl, nil
}

func FilterByType(t Type) func(as ast.AssignStmt) bool {
	return func(as ast.AssignStmt) bool {
		for _, decl := range as.Rhs {
			typed, err := TypeFromExpr("", decl, nil)
			if err != nil || !t.EqualTo(typed) {
				continue
			}
			return true
		}

		return false
	}
}

func (f *FuncDecl) Variables(pkgName string, filters ...func(ast.AssignStmt) bool) map[string]*Variable {
	var result = make(map[string]*Variable)

	ast.Inspect(f.funcDecl.Body, func(n ast.Node) bool {
		decl, ok := n.(*ast.AssignStmt)
		if !ok || decl.Tok != token.DEFINE {
			return true
		}

		for _, filter := range filters {
			if !filter(*decl) {
				return true
			}
		}

		variables, err := NewVariables(decl, f.File, pkgName)
		if err != nil {
			panic(err)
		}

		var pkg = Pkgs[pkgName]
		pkg.Variables = variables
		Pkgs[pkgName] = pkg

		for name, variable := range variables {
			result[name] = variable
		}

		return true
	})

	return result
}
