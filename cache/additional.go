package cache

import "github.com/KlyuchnikovV/webapi-docs/cache/types"

// func GetBySelector(selector ast.SelectorExpr) {
// 	switch typed := selector.X.(type) {
// 	case *ast.Ident:

// 	}
// }

func UnwrapImportedType(s types.ImportedType) (types.Type, error) {
	if err := ParsePackage(s.Package); err != nil {
		return nil, err
	}

	return FindModelByName(s.Name())
}
