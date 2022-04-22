package interfaces

import "go/ast"

func ServiceInterface() ast.InterfaceType {
	return ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{{
		Names: []*ast.Ident{{
			Name: "Routers",
			Obj:  &ast.Object{Kind: ast.Fun, Name: "Routers"},
		}},
		Type: RoutersFuncType(),
	}}}}
}

func RoutersFuncDecl() *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "Routers",
			Obj:  &ast.Object{Kind: ast.Fun, Name: "Routers"},
		},
		Type: RoutersFuncType(),
	}
}

func RoutersFuncType() *ast.FuncType {
	return &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{List: []*ast.Field{{
			Type: &ast.MapType{
				Key: &ast.Ident{Name: "string"},
				Value: &ast.SelectorExpr{
					// TODO: check import alias
					X:   &ast.Ident{Name: "webapi"},
					Sel: &ast.Ident{Name: "RouterByPath"},
				},
			},
		}}},
	}
}

// func EngineInterface() ast.InterfaceType {
// 	return ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{{
// 		Names: []*ast.Ident{{
// 			Name: "RegisterServices",
// 			Obj:  &ast.Object{Kind: ast.Fun, Name: "RegisterServices"},
// 		}},
// 		Type: &ast.FuncType{
// 			Params: &ast.FieldList{
// 				List: []*ast.Field{{
// 					Type: &ast.ArrayType{
// 						Elt: &ast.SelectorExpr{
// 							X:   &ast.Ident{Name: "webapi"},
// 							Sel: &ast.Ident{Name: "ServiceAPI"},
// 						},
// 					},
// 				}},
// 			},
// 			Results: &ast.FieldList{List: []*ast.Field{{
// 				Type: &ast.Ident{Name: "error"},
// 			}}},
// 		},
// 	}}}}
// }
