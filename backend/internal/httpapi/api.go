// Package httpapi is the HTTP transport for the calculator. It owns decoding, validation of the
// request envelope, and the translation of domain failures into the wire error contract.
package httpapi

import "log/slog"

type api struct {
	logger *slog.Logger
}
