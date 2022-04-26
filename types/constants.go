package types

var TypeParamsMap = map[string]string{
	"Bool":    "bool",
	"Time":    "time",
	"Float":   "float64",
	"String":  "string",
	"Integer": "int64",
}

func ConvertFieldType(t string) string {
	switch t {
	case "byte", "rune",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"int", "int8", "int16", "int32", "int64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string", "time":
		return "string"
	default:
		return ""
	}
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
