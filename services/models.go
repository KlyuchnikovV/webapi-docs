package services

// import (
// 	"fmt"
// 	"go/ast"

// 	"github.com/KlyuchnikovV/webapi-docs/cache"
// )

// type ModelDeclaration struct {
// 	Name    string
// 	Fields  []ast.Ident
// 	Methods []ast.FuncDecl
// }

// func NewModelDeclaration(file ast.File, ts ast.TypeSpec) *ModelDeclaration {
// 	return nil
// }

// func GetModelDeclaration(file ast.File, arg ast.Expr) *ModelDeclaration {
// 	var expr ast.Expr

// 	switch typed := arg.(type) {
// 	// case *ast.UnaryExpr:
// 	// 	expr = typed.X
// 	// case *ast.CompositeLit:
// 	// 	expr = typed.Type
// 	case *ast.SelectorExpr:
// 		// GetModelDeclaration(file, typed.X)

// 		spec := cache.FindModel(file, *typed)
// 		if spec != nil {
// 			return NewModelDeclaration(file, *spec)
// 		}

// 		panic("not found")

// 		// baseType := cache.FindModel(file, typed.X)
// 		// if baseType == nil {
// 		// 	return nil
// 		// }

// 		// funcDecl := cache.FindMethodByName(*baseType, typed.Sel.Name)
// 		// if funcDecl == nil {
// 		// 	return nil
// 		// }

// 		// return GetFunctionDecl(file, funcDecl.Type.Results.List[0].Type)

// 		// return spec
// 	case *ast.Ident:
// 		if typed.Obj == nil {
// 			return nil
// 		}

// 		switch decl := typed.Obj.Decl.(type) {
// 		case *ast.TypeSpec:
// 			return NewModelDeclaration(file, *decl)
// 		case *ast.Field:
// 			return GetModelDeclaration(file, decl.Type)
// 		default:
// 			panic("not ok")
// 		}
// 	// case *ast.ArrayType:
// 	// 	expr = typed.Elt
// 	case *ast.StarExpr:
// 		return GetModelDeclaration(file, typed.X)
// 	case *ast.CallExpr:
// 		return GetModelDeclaration(file, typed.Fun)
// 	default:
// 		panic(fmt.Sprintf("unknown type %#v", typed))
// 	}

// 	return GetModelDeclaration(file, expr)
// }

// type FuncDeclaration struct {
// 	Name     string
// 	Receiver *ast.TypeSpec
// 	Params   map[string]ast.TypeSpec
// 	Results  map[string]ast.TypeSpec
// }

// func GetFunctionDecl(file ast.File, arg ast.Expr) *FuncDeclaration {
// 	var expr ast.Expr

// 	switch typed := arg.(type) {
// 	// case *ast.UnaryExpr:
// 	// 	expr = typed.X
// 	// case *ast.CompositeLit:
// 	// 	expr = typed.Type
// 	case *ast.SelectorExpr:
// 		model := GetModelDeclaration(file, typed.X)

// 		// spec := cache.FindModel(file)
// 		// if spec != nil {
// 			// return NewFuncDeclaration(
// 			// 	file, *cache.FindMethodByName(*spec, typed.Sel.Name),
// 			// )
// 		// }

// 		fmt.Printf("%#v\n", model)

// 		panic("not found")

// 		// baseType := cache.FindModel(file, typed.X)
// 		// if baseType == nil {
// 		// 	return nil
// 		// }

// 		// funcDecl := cache.FindMethodByName(*baseType, typed.Sel.Name)
// 		// if funcDecl == nil {
// 		// 	return nil
// 		// }

// 		// return GetFunctionDecl(file, funcDecl.Type.Results.List[0].Type)

// 		// return spec
// 	case *ast.Ident:
// 		if typed.Obj == nil {
// 			return nil
// 		}

// 		switch decl := typed.Obj.Decl.(type) {
// 		case *ast.TypeSpec:
// 			return GetFunctionDecl(file, decl.Type)
// 		case *ast.Field:
// 			return GetFunctionDecl(file, decl.Type)
// 		default:
// 			panic("not ok")
// 		}
// 	// case *ast.ArrayType:
// 	// 	expr = typed.Elt
// 	// case *ast.StarExpr:
// 	// 	return GetFunctionDecl(file, typed.X)
// 	case *ast.CallExpr:
// 		return GetFunctionDecl(file, typed.Fun)
// 	default:
// 		panic(fmt.Sprintf("unknown type %#v", typed))
// 	}

// 	return GetFunctionDecl(file, expr)
// }

// func NewFuncDeclaration(file ast.File, decl ast.FuncDecl) *FuncDeclaration {
// 	var result = FuncDeclaration{
// 		Name:    decl.Name.Name,
// 		Params:  make(map[string]ast.TypeSpec),
// 		Results: make(map[string]ast.TypeSpec),
// 	}

// 	if decl.Recv != nil {
// 		_, spec := GetTypeSpecification(file, decl.Recv.List[0].Type)
// 		result.Receiver = spec
// 	}

// 	if decl.Type.Params != nil {
// 		for _, param := range decl.Type.Params.List {
// 			name, spec := GetTypeSpecification(file, param.Type)
// 			result.Params[name] = *spec
// 		}
// 	}

// 	if decl.Type.Results != nil {
// 		for _, param := range decl.Type.Results.List {
// 			name, spec := GetTypeSpecification(file, param.Type)
// 			result.Results[name] = *spec
// 		}
// 	}

// 	return &result
// }
