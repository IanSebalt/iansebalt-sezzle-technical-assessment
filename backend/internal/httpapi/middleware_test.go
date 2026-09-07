package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

func newTestAPI() *api {
	return &api{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestRecoverPanic_Returns500(t *testing.T) {
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("handler exploded")
	})

	rec := httptest.NewRecorder()
	newTestAPI().recoverPanic(panicking).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v (body: %s)", err, rec.Body.String())
	}
	if body.Error.Code != apierror.CodeInternal {
		t.Errorf("code = %q, want %q", body.Error.Code, apierror.CodeInternal)
	}
	if body.Error.Message == "" {
		t.Error("error message is empty")
	}
}

func TestRecoverPanic_LeavesHealthyResponsesAlone(t *testing.T) {
	healthy := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("fine"))
	})

	rec := httptest.NewRecorder()
	newTestAPI().recoverPanic(healthy).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "fine" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "fine")
	}
}
