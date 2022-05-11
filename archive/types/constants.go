package types

import "net/http"

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

func GetResultCode(t string) int {
	switch t {
	case "OK":
		return http.StatusOK
	case "Created":
		return http.StatusCreated
	case "NoContent":
		return http.StatusNoContent
	case "BadRequest":
		return http.StatusBadRequest
	case "Forbidden":
		return http.StatusForbidden
	case "NotFound":
		return http.StatusNotFound
	case "MethodNotAllowed":
		return http.StatusMethodNotAllowed
	case "InternalServerError":
		return http.StatusInternalServerError
	default:
		return -1
	}
}

var CodeSelector = map[string]int{
	"StatusContinue":           http.StatusContinue,
	"StatusSwitchingProtocols": http.StatusSwitchingProtocols,
	"StatusProcessing":         http.StatusProcessing,
	"StatusEarlyHints":         http.StatusEarlyHints,

	"StatusOK":                   http.StatusOK,
	"StatusCreated":              http.StatusCreated,
	"StatusAccepted":             http.StatusAccepted,
	"StatusNonAuthoritativeInfo": http.StatusNonAuthoritativeInfo,
	"StatusNoContent":            http.StatusNoContent,
	"StatusResetContent":         http.StatusResetContent,
	"StatusPartialContent":       http.StatusPartialContent,
	"StatusMultiStatus":          http.StatusMultiStatus,
	"StatusAlreadyReported":      http.StatusAlreadyReported,
	"StatusIMUsed":               http.StatusIMUsed,

	"StatusMultipleChoices":   http.StatusMultipleChoices,
	"StatusMovedPermanently":  http.StatusMovedPermanently,
	"StatusFound":             http.StatusFound,
	"StatusSeeOther":          http.StatusSeeOther,
	"StatusNotModified":       http.StatusNotModified,
	"StatusUseProxy":          http.StatusUseProxy,
	"StatusTemporaryRedirect": http.StatusTemporaryRedirect,
	"StatusPermanentRedirect": http.StatusPermanentRedirect,

	"StatusBadRequest":                   http.StatusBadRequest,
	"StatusUnauthorized":                 http.StatusUnauthorized,
	"StatusPaymentRequired":              http.StatusPaymentRequired,
	"StatusForbidden":                    http.StatusForbidden,
	"StatusNotFound":                     http.StatusNotFound,
	"StatusMethodNotAllowed":             http.StatusMethodNotAllowed,
	"StatusNotAcceptable":                http.StatusNotAcceptable,
	"StatusProxyAuthRequired":            http.StatusProxyAuthRequired,
	"StatusRequestTimeout":               http.StatusRequestTimeout,
	"StatusConflict":                     http.StatusConflict,
	"StatusGone":                         http.StatusGone,
	"StatusLengthRequired":               http.StatusLengthRequired,
	"StatusPreconditionFailed":           http.StatusPreconditionFailed,
	"StatusRequestEntityTooLarge":        http.StatusRequestEntityTooLarge,
	"StatusRequestURITooLong":            http.StatusRequestURITooLong,
	"StatusUnsupportedMediaType":         http.StatusUnsupportedMediaType,
	"StatusRequestedRangeNotSatisfiable": http.StatusRequestedRangeNotSatisfiable,
	"StatusExpectationFailed":            http.StatusExpectationFailed,
	"StatusTeapot":                       http.StatusTeapot,
	"StatusMisdirectedRequest":           http.StatusMisdirectedRequest,
	"StatusUnprocessableEntity":          http.StatusUnprocessableEntity,
	"StatusLocked":                       http.StatusLocked,
	"StatusFailedDependency":             http.StatusFailedDependency,
	"StatusTooEarly":                     http.StatusTooEarly,
	"StatusUpgradeRequired":              http.StatusUpgradeRequired,
	"StatusPreconditionRequired":         http.StatusPreconditionRequired,
	"StatusTooManyRequests":              http.StatusTooManyRequests,
	"StatusRequestHeaderFieldsTooLarge":  http.StatusRequestHeaderFieldsTooLarge,
	"StatusUnavailableForLegalReasons":   http.StatusUnavailableForLegalReasons,

	"StatusInternalServerError":           http.StatusInternalServerError,
	"StatusNotImplemented":                http.StatusNotImplemented,
	"StatusBadGateway":                    http.StatusBadGateway,
	"StatusServiceUnavailable":            http.StatusServiceUnavailable,
	"StatusGatewayTimeout":                http.StatusGatewayTimeout,
	"StatusHTTPVersionNotSupported":       http.StatusHTTPVersionNotSupported,
	"StatusVariantAlsoNegotiates":         http.StatusVariantAlsoNegotiates,
	"StatusInsufficientStorage":           http.StatusInsufficientStorage,
	"StatusLoopDetected":                  http.StatusLoopDetected,
	"StatusNotExtended":                   http.StatusNotExtended,
	"StatusNetworkAuthenticationRequired": http.StatusNetworkAuthenticationRequired,
}
