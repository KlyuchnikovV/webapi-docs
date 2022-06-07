package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type Package struct {
	Name      string
	Types     map[string]types.Type
	Functions map[string]types.FuncType

	Pkg ast.Package
}

func NewPackage(pkg ast.Package) Package {
	var result = Package{
		Pkg:       pkg,
		Name:      pkg.Name,
		Types:     make(map[string]types.Type),
		Functions: make(map[string]types.FuncType),
	}

	for name, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, t := range newFromGenDecl(file, typed) {
					result.Types[t.Name()] = t
				}
			case *ast.FuncDecl:
				fun, receiver, err := result.newFromFuncDecl(file, typed)
				if err != nil {
					panic(err)
				}

				if t, ok := result.isConstructorOf(*fun); ok {
					t.AddConstructor(*fun)
					continue
				}

				if receiver != nil {
					result.Types[receiver.Name()].AddMethod(*fun)
				} else {
					result.Functions[fun.Name()] = *fun
				}
			default:
				panic(fmt.Errorf("%s - %#v", name, typed))
			}
		}
	}

	return result
}

func newFromGenDecl(file *ast.File, decl *ast.GenDecl) []types.Type {
	if decl.Tok != token.TYPE {
		return nil
	}

	var result = make([]types.Type, len(decl.Specs))

	for i, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			panic("not ts")
		}

		result[i] = types.NewType(file, ts.Name.Name, &ts.Type, nil)
	}

	return result
}

func (pkg *Package) newFromFuncDecl(file *ast.File, decl *ast.FuncDecl) (*types.FuncType, types.Type, error) {
	var (
		fun      = types.NewFunc(file, *decl, decl.Name.Name, nil)
		receiver types.Type
	)

	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return &fun, receiver, nil
	}

	// TODO: what if type do not exist yet
	var model = pkg.Types[receiverTypeName(decl.Recv.List[0].Type)]

	if model == nil {
		return nil, nil, fmt.Errorf("for now")
	}

	return &fun, model, nil
}

func receiverTypeName(f ast.Expr) string {
	switch typed := f.(type) {
	case *ast.StarExpr:
		return receiverTypeName(typed.X)
	case *ast.Ident:
		return typed.Name
	default:
		panic("for now")
	}
}

func (pkg *Package) isConstructorOf(fun types.FuncType) (types.Type, bool) {
	if !strings.HasPrefix(fun.Name(), "New") {
		return nil, false
	}

	t, ok := pkg.Types[strings.TrimPrefix(fun.Name(), "New")]
	if !ok {
		return nil, false
	}

	for _, result := range fun.Results {
		if result.Name() == t.Name() {
			return t, true
		}
	}

	return nil, false
}
