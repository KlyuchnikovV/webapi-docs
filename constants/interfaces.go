package constants

import (
	"github.com/KlyuchnikovV/webapi-docs/types"
)

func RoutersInterface() types.InterfaceType {
	return types.NewInterfaceFields(map[string]types.Type{
		"Routers": types.NewFuncDeclaration(
			"Routers",
			nil,
			[]types.Type{
				types.NewSimpleMap("",
					types.NewSimpleBasicType("string"),
					types.NewSimpleImported("RouterByPath", "github.com/KlyuchnikovV/webapi"),
				),
			},
		),
	})
}
