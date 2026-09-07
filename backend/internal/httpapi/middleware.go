package httpapi

import (
	"net/http"
	"runtime/debug"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

// trackingWriter records whether anything has reached the client, so a panic handler can tell a
// clean failure from one that arrives too late to change the response.
type trackingWriter struct {
	http.ResponseWriter
	wrote bool
}

func (w *trackingWriter) WriteHeader(status int) {
	w.wrote = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *trackingWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}

// Unwrap keeps http.ResponseController able to reach the underlying writer.
func (w *trackingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// recoverPanic keeps a single failing request from taking the process down, and makes sure the
// client still receives the documented error envelope instead of an empty connection.
func (a *api) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracked := &trackingWriter{ResponseWriter: w}

		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			a.logger.Error("recovered from panic",
				"panic", recovered,
				"method", r.Method,
				"path", r.URL.Path,
				"stack", string(debug.Stack()))

			// A status line already on the wire cannot be replaced; a second one would corrupt the
			// response the client is mid-way through reading.
			if tracked.wrote {
				return
			}

			a.writeError(tracked, apierror.Internal())
		}()

		next.ServeHTTP(tracked, r)
	})
}
