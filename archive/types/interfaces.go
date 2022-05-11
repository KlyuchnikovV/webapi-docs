package types

import "go/ast"

func ServiceInterface(alias string) ast.InterfaceType {
	return ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{{
		Names: []*ast.Ident{{
			Name: "Routers",
			Obj:  &ast.Object{Kind: ast.Fun, Name: "Routers"},
		}},
		Type: RoutersFuncType(alias),
	}}}}
}

func RoutersFuncDecl(alias string) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "Routers",
			Obj:  &ast.Object{Kind: ast.Fun, Name: "Routers"},
		},
		Type: RoutersFuncType(alias),
	}
}

func RoutersFuncType(alias string) *ast.FuncType {
	return &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{List: []*ast.Field{{
			Type: &ast.MapType{
				Key: &ast.Ident{Name: "string"},
				Value: &ast.SelectorExpr{
					X:   &ast.Ident{Name: alias},
					Sel: &ast.Ident{Name: "RouterByPath"},
				},
			},
		}}},
	}
}

// TODO: move to service space with webapi alias auto-detection
func RouteHandlerFuncDecl(name, alias string) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: &ast.Ident{
			Name: name,
			Obj:  &ast.Object{Kind: ast.Fun, Name: name},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: alias,
								},
								Sel: &ast.Ident{
									Name: "Context",
								},
							},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
	}
}

// GetByID(ctx *webapi.Context) error

// func RoutersFuncDecl(alias string) *ast.FuncDecl {
// 	return &ast.FuncDecl{
// 		Name: &ast.Ident{
// 			Name: "Routers",
// 			Obj:  &ast.Object{Kind: ast.Fun, Name: "Routers"},
// 		},
// 		Type: RoutersFuncType(alias),
// 	}
// }
