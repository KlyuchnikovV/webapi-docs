package constants

import "net/http"

var TypeParamsMap = map[string]string{
	"Bool":    "bool",
	"Time":    "time",
	"Float":   "float64",
	"String":  "string",
	"Integer": "int64",
}

func ConvertFieldType(t string) (string, bool) {
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
	case "string", "time", "error":
		result = "string"
	}

	return result, len(result) != 0
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

func GetResultCode(t string) (int, CodeType) {
	code, ok := ServerErrorsCodes[t]
	if ok {
		return code, ServerError
	}

	code, ok = SuccessCodes[t]
	if ok {
		return code, Success
	}

	code, ok = ClientErrorCodes[t]
	if ok {
		return code, ClientError
	}

	code, ok = RedirectionCodes[t]
	if ok {
		return code, Redirection
	}

	code, ok = InfoCodes[t]
	if ok {
		return code, Info
	}

	return -1, Undefined
}

type CodeType string

const (
	Undefined   CodeType = ""
	Info        CodeType = "info"
	Success     CodeType = "success"
	Redirection CodeType = "redirection"
	ClientError CodeType = "client error"
	ServerError CodeType = "server error"
)

var InfoCodes = map[string]int{
	"Continue":           http.StatusContinue,
	"SwitchingProtocols": http.StatusSwitchingProtocols,
	"Processing":         http.StatusProcessing,
	"EarlyHints":         http.StatusEarlyHints,
}

var SuccessCodes = map[string]int{
	"OK":                   http.StatusOK,
	"Created":              http.StatusCreated,
	"Accepted":             http.StatusAccepted,
	"NonAuthoritativeInfo": http.StatusNonAuthoritativeInfo,
	"NoContent":            http.StatusNoContent,
	"ResetContent":         http.StatusResetContent,
	"PartialContent":       http.StatusPartialContent,
	"MultiStatus":          http.StatusMultiStatus,
	"AlreadyReported":      http.StatusAlreadyReported,
	"IMUsed":               http.StatusIMUsed,
}

var RedirectionCodes = map[string]int{
	"MultipleChoices":   http.StatusMultipleChoices,
	"MovedPermanently":  http.StatusMovedPermanently,
	"Found":             http.StatusFound,
	"SeeOther":          http.StatusSeeOther,
	"NotModified":       http.StatusNotModified,
	"UseProxy":          http.StatusUseProxy,
	"TemporaryRedirect": http.StatusTemporaryRedirect,
	"PermanentRedirect": http.StatusPermanentRedirect,
}

var ClientErrorCodes = map[string]int{
	"BadRequest":                   http.StatusBadRequest,
	"Unauthorized":                 http.StatusUnauthorized,
	"PaymentRequired":              http.StatusPaymentRequired,
	"Forbidden":                    http.StatusForbidden,
	"NotFound":                     http.StatusNotFound,
	"MethodNotAllowed":             http.StatusMethodNotAllowed,
	"NotAcceptable":                http.StatusNotAcceptable,
	"ProxyAuthRequired":            http.StatusProxyAuthRequired,
	"RequestTimeout":               http.StatusRequestTimeout,
	"Conflict":                     http.StatusConflict,
	"Gone":                         http.StatusGone,
	"LengthRequired":               http.StatusLengthRequired,
	"PreconditionFailed":           http.StatusPreconditionFailed,
	"RequestEntityTooLarge":        http.StatusRequestEntityTooLarge,
	"RequestURITooLong":            http.StatusRequestURITooLong,
	"UnsupportedMediaType":         http.StatusUnsupportedMediaType,
	"RequestedRangeNotSatisfiable": http.StatusRequestedRangeNotSatisfiable,
	"ExpectationFailed":            http.StatusExpectationFailed,
	"Teapot":                       http.StatusTeapot,
	"MisdirectedRequest":           http.StatusMisdirectedRequest,
	"UnprocessableEntity":          http.StatusUnprocessableEntity,
	"Locked":                       http.StatusLocked,
	"FailedDependency":             http.StatusFailedDependency,
	"TooEarly":                     http.StatusTooEarly,
	"UpgradeRequired":              http.StatusUpgradeRequired,
	"PreconditionRequired":         http.StatusPreconditionRequired,
	"TooManyRequests":              http.StatusTooManyRequests,
	"RequestHeaderFieldsTooLarge":  http.StatusRequestHeaderFieldsTooLarge,
	"UnavailableForLegalReasons":   http.StatusUnavailableForLegalReasons,
}

var ServerErrorsCodes = map[string]int{
	"InternalServerError":           http.StatusInternalServerError,
	"NotImplemented":                http.StatusNotImplemented,
	"BadGateway":                    http.StatusBadGateway,
	"ServiceUnavailable":            http.StatusServiceUnavailable,
	"GatewayTimeout":                http.StatusGatewayTimeout,
	"HTTPVersionNotSupported":       http.StatusHTTPVersionNotSupported,
	"VariantAlsoNegotiates":         http.StatusVariantAlsoNegotiates,
	"InsufficientStorage":           http.StatusInsufficientStorage,
	"LoopDetected":                  http.StatusLoopDetected,
	"NotExtended":                   http.StatusNotExtended,
	"NetworkAuthenticationRequired": http.StatusNetworkAuthenticationRequired,
}
