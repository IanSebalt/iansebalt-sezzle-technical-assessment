package apierror

import "net/http"

// Codes clients may branch on. They are part of the public contract and must stay stable.
const (
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	CodeNotFound             = "NOT_FOUND"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	CodeMalformedJSON        = "MALFORMED_JSON"
	CodeUnknownField         = "UNKNOWN_FIELD"
	CodeMissingOperation     = "MISSING_OPERATION"
	CodeInvalidOperation     = "INVALID_OPERATION"
	CodeInvalidOperand       = "INVALID_OPERAND"
	CodeMissingOperand       = "MISSING_OPERAND"
	CodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	CodeDivisionByZero       = "DIVISION_BY_ZERO"
	CodeNegativeSquareRoot   = "NEGATIVE_SQUARE_ROOT"
	CodeNonFiniteResult      = "NON_FINITE_RESULT"
	CodeInternal             = "INTERNAL_ERROR"
)

func MethodNotAllowed() *Error {
	return newError(http.StatusMethodNotAllowed, CodeMethodNotAllowed,
		"Only POST is supported for this endpoint.")
}

func NotFound() *Error {
	return newError(http.StatusNotFound, CodeNotFound,
		"Endpoint not found.")
}

func UnsupportedMediaType() *Error {
	return newError(http.StatusUnsupportedMediaType, CodeUnsupportedMediaType,
		"Content-Type must be application/json.")
}

func PayloadTooLarge() *Error {
	return newError(http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
		"Request body is too large.")
}

func MalformedJSON() *Error {
	return newError(http.StatusBadRequest, CodeMalformedJSON,
		"Request body must be a valid JSON object.")
}

func UnknownField() *Error {
	return newError(http.StatusBadRequest, CodeUnknownField,
		`Request contains an unrecognised field. Expected: "operation", "a" and "b".`)
}

func MissingOperation() *Error {
	return newError(http.StatusBadRequest, CodeMissingOperation,
		`Field "operation" is required.`)
}

func InvalidOperation() *Error {
	return newError(http.StatusBadRequest, CodeInvalidOperation,
		`Field "operation" must be text.`)
}

// InvalidOperand names the offending field so the client can point at the right input.
func InvalidOperand(field string) *Error {
	return newError(http.StatusBadRequest, CodeInvalidOperand,
		`Field "`+field+`" must be a number the calculator can represent.`)
}

func MissingOperand(field string) *Error {
	return newError(http.StatusBadRequest, CodeMissingOperand,
		`Field "`+field+`" is required for this operation.`)
}

// UnsupportedOperation receives the allowed names rather than importing the domain, which would
// point this package's dependencies the wrong way.
func UnsupportedOperation(allowed string) *Error {
	return newError(http.StatusBadRequest, CodeUnsupportedOperation,
		"Unsupported operation. Use one of: "+allowed+".")
}

func DivisionByZero() *Error {
	return newError(http.StatusUnprocessableEntity, CodeDivisionByZero,
		"Cannot divide by zero.")
}

func NegativeSquareRoot() *Error {
	return newError(http.StatusUnprocessableEntity, CodeNegativeSquareRoot,
		"Cannot take the square root of a negative number.")
}

func NonFiniteResult() *Error {
	return newError(http.StatusUnprocessableEntity, CodeNonFiniteResult,
		"The result is too large or undefined to represent.")
}

func Internal() *Error {
	return newError(http.StatusInternalServerError, CodeInternal,
		"Something went wrong. Please try again.")
}
