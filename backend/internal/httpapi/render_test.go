package httpapi

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

// brokenWriter stands in for a client that disappears mid-response.
type brokenWriter struct {
	header http.Header
}

func (b *brokenWriter) Header() http.Header {
	if b.header == nil {
		b.header = http.Header{}
	}
	return b.header
}

func (b *brokenWriter) Write([]byte) (int, error) { return 0, errors.New("connection reset") }

func (b *brokenWriter) WriteHeader(int) {}

func TestWriteJSON_LogsWhenTheBodyCannotBeWritten(t *testing.T) {
	var logged bytes.Buffer
	a := &api{logger: slog.New(slog.NewTextHandler(&logged, nil))}

	a.writeJSON(&brokenWriter{}, http.StatusOK, map[string]int{"result": 1})

	if !strings.Contains(logged.String(), "failed to encode response body") {
		t.Errorf("a failed write was not logged: %q", logged.String())
	}
}

func TestWriteError_LogsServerFaultsOnly(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     *apierror.Error
		wantLogged bool
	}{
		{name: "client fault", apiErr: apierror.DivisionByZero()},
		{name: "server fault", apiErr: apierror.Internal(), wantLogged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logged bytes.Buffer
			a := &api{logger: slog.New(slog.NewTextHandler(&logged, nil))}

			a.writeError(httptest.NewRecorder(), tt.apiErr)

			if gotLogged := logged.Len() > 0; gotLogged != tt.wantLogged {
				t.Errorf("logged = %v, want %v (log: %q)", gotLogged, tt.wantLogged, logged.String())
			}
			// A server fault should be diagnosable from the log line alone.
			if tt.wantLogged && !strings.Contains(logged.String(), tt.apiErr.Error()) {
				t.Errorf("log %q does not carry the failure %q", logged.String(), tt.apiErr.Error())
			}
		})
	}
}
