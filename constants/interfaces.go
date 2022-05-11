package constants

import (
	"github.com/KlyuchnikovV/webapi-docs/cache/types"
)

// func EngineInterface() types.InterfaceType {
// 	return types.NewInterfaceFields(map[string]types.Type{
// 		"RegisterServices": types.NewFuncDeclaration(
// 			"RegisterServices",
// 			[]types.Type{
// 				types.NewSimpleArrayType("", types.ImportedType{}),
// 				// services ...webapi.ServiceAPI
// 			},
// 			[]types.Type{
// 				types.NewSimpleBasicType("error"),
// 			},
// 		),
// 	})
// }

func RoutersInterface(alias, path string) types.InterfaceType {
	return types.NewInterfaceFields(map[string]types.Type{
		"Routers": types.NewFuncDeclaration(
			"Routers",
			nil,
			[]types.Type{
				types.NewSimpleMap("",
					types.NewSimpleBasicType("string"),
					types.NewSimpleImported(alias, path),
				),
			},
		),
	})
}
