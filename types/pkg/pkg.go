package pkg

import (
	"fmt"
	"go/ast"

	"github.com/KlyuchnikovV/webapi-docs/types"
)

type (
	// FuncType interface{}

	Package struct {
		Name  string
		Types map[string]types.Type
		// Functions map[string]FuncType

		Pkg ast.Package `json:"-"`
	}

	Packages map[string]Package
)

func NewPackage(pkg ast.Package) (*Package, error) {
	var result = Package{
		Name:  pkg.Name,
		Types: make(map[string]types.Type),
		// Functions: make(map[string]FuncType),

		Pkg: pkg,
	}

	for name, file := range pkg.Files {
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, spec := range typed.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						// Should be 'var' specification
						break
					}

					t, err := types.NewType(*typeSpec, file)
					if err != nil {
						return nil, err
					}

					if t == nil {
						continue
					}

					result.Types[t.GetName()] = t
				}
			case *ast.FuncDecl:
			// 	fun, receiver, err := result.newFromFuncDecl(file, typed)
			// 	if err != nil {
			// 		panic(err)
			// 	}

			// 	if t, ok := result.isConstructorOf(*fun); ok {
			// 		t.AddConstructor(*fun)
			// 		continue
			// 	}

			// 	if receiver != nil {
			// 		result.Types[receiver.Name()].AddMethod(*fun)
			// 	} else {
			// 		result.Functions[fun.Name()] = *fun
			// 	}
			default:
				return nil, fmt.Errorf("%s - %#v", name, typed)
			}
		}
	}

	return &result, nil
}
