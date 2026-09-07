// Package apierror defines the error contract the API exposes: a stable machine-readable code, the
// HTTP status it maps to, and the sentence a user is shown. It is the only place that copy lives.
package apierror

// Error is a failure the API is willing to describe to a client.
type Error struct {
	Status  int
	Code    string
	Message string
}

// Error makes *Error satisfy the error interface, so it can be logged, wrapped and matched with
// errors.Is/As like any other failure rather than needing special handling at every call site.
func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func newError(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}
