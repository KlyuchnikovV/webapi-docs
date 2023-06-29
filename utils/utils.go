package utils

// func FindImportWithPath(file ast.File, path string) (string, *ast.ImportSpec) {
// 	var (
// 		result string
// 		spec   *ast.ImportSpec
// 	)

// 	for i, imp := range file.Imports {
// 		if strings.Trim(imp.Path.Value, "\"") != path {
// 			continue
// 		}

// 		var curAlias string

// 		if imp.Name != nil {
// 			curAlias = imp.Name.Name
// 		} else {
// 			curAlias = strings.Trim(
// 				imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1:],
// 				"\"",
// 			)
// 		}

// 		result = curAlias
// 		spec = file.Imports[i]

// 		break
// 	}

// 	return result, spec
// }
