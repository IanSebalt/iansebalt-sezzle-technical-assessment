package httpapi

import (
	"net/http"
	"runtime/debug"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

// recoverPanic keeps a single failing request from taking the process down, and makes sure the
// client still receives the documented error envelope instead of an empty connection.
func (a *api) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				a.logger.Error("recovered from panic",
					"panic", recovered,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()))
				a.writeError(w, apierror.Internal())
			}
		}()

		next.ServeHTTP(w, r)
	})
}
