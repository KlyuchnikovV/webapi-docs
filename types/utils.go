package types

import (
	"go/ast"
	"strings"
)

func ConvertFieldType(t string) SchemaType {
	var result SchemaType

	switch t {
	case "byte", "rune",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"int", "int8", "int16", "int32", "int64":
		result = IntegerSchemaType
	case "float32", "float64":
		result = NumberSchemaType
	case "bool":
		result = BooleanSchemaType
	default:
		result = StringSchemaType
	}

	return result
}

func GetFieldTypeFormat(t string) string {
	switch t {
	case "int32", "int64":
		return t
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		return ""
	}
}

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
