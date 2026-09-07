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

func TestRecoverPanic_DoesNotRewriteAResponseAlreadySent(t *testing.T) {
	lateFailure := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":1}`))
		panic("failed after the response went out")
	})

	rec := httptest.NewRecorder()
	newTestAPI().recoverPanic(lateFailure).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want the 200 the handler already sent", rec.Code)
	}
	if got := rec.Body.String(); got != `{"result":1}` {
		t.Errorf("body = %q, want the original response with no error envelope appended", got)
	}
}

func TestRecoverPanic_ExposesTheUnderlyingWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	tracked := &trackingWriter{ResponseWriter: recorder}

	unwrapper, ok := http.ResponseWriter(tracked).(interface{ Unwrap() http.ResponseWriter })
	if !ok {
		t.Fatal("trackingWriter does not expose Unwrap, so http.ResponseController cannot reach the writer")
	}
	if unwrapper.Unwrap() != http.ResponseWriter(recorder) {
		t.Error("Unwrap did not return the wrapped writer")
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
