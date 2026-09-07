package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

const calculatePath = "/api/v1/calculate"

// NewRouter builds the API handler. The catch-all registrations exist so that a wrong method or an
// unknown path still answers with the documented JSON envelope rather than net/http's plain text.
func NewRouter(logger *slog.Logger) http.Handler {
	a := &api{logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("POST "+calculatePath, a.calculate)
	mux.HandleFunc(calculatePath, a.methodNotAllowed)
	mux.HandleFunc("/", a.notFound)

	return a.recoverPanic(mux)
}

func (a *api) methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", http.MethodPost)
	a.writeError(w, apierror.MethodNotAllowed())
}

func (a *api) notFound(w http.ResponseWriter, _ *http.Request) {
	a.writeError(w, apierror.NotFound())
}
