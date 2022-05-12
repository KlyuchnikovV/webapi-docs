package utils

import (
	"go/ast"
	"strings"
)

func FindImport(file ast.File, alias string) (string, *ast.ImportSpec) {
	var (
		result = alias
		spec   *ast.ImportSpec
	)

	for i, imp := range file.Imports {
		var curAlias string

		if imp.Name != nil {
			curAlias = imp.Name.Name
		} else {
			curAlias = strings.Trim(
				imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1:],
				"\"",
			)
		}

		if curAlias == alias {
			result = strings.Trim(imp.Path.Value, "\"")
			spec = file.Imports[i]

			break
		}
	}

	return result, spec
}

func FindImportWithPath(file ast.File, path string) (string, *ast.ImportSpec) {
	var (
		result string
		spec   *ast.ImportSpec
	)

	for i, imp := range file.Imports {
		if strings.Trim(imp.Path.Value, "\"") != path {
			continue
		}

		var curAlias string

		if imp.Name != nil {
			curAlias = imp.Name.Name
		} else {
			curAlias = strings.Trim(
				imp.Path.Value[strings.LastIndex(imp.Path.Value, "/")+1:],
				"\"",
			)
		}

		result = curAlias
		spec = file.Imports[i]

		break
	}

	return result, spec
}

func ConvertFieldType(t string) string {
	var result string

	switch t {
	case "byte", "rune",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"int", "int8", "int16", "int32", "int64":
		result = "integer"
	case "float32", "float64":
		result = "number"
	case "bool":
		result = "boolean"
	default:
		result = "string"
	}

	return result
}

func GetFieldTypeFormat(t string) string {
	switch t {
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		return ""
	}
}
