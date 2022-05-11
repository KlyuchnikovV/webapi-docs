package objects

import (
	"go/ast"
	"strings"
)

func getFieldName(field *ast.Field) string {
	if field.Tag == nil {
		if len(field.Names) != 0 {
			return field.Names[0].Name
		} else {
			return field.Type.(*ast.SelectorExpr).Sel.Name
		}
	}

	var (
		value = field.Tag.Value
		start = strings.Index(value, "json:\"") + len("json:\"")
		end   = strings.Index(value[start:], "\"")
	)

	return value[start : start+end]
}
