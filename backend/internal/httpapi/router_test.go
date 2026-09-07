package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

func TestRouter_MethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			rec := request(t, method, calculatePath, "", "application/json")

			assertAPIError(t, rec, http.StatusMethodNotAllowed, apierror.CodeMethodNotAllowed)
			if got := rec.Header().Get("Allow"); got != http.MethodPost {
				t.Errorf("Allow = %q, want %q", got, http.MethodPost)
			}
		})
	}
}

func TestRouter_NotFound(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "root", method: http.MethodGet, path: "/"},
		{name: "unknown endpoint", method: http.MethodPost, path: "/api/v1/unknown"},
		{name: "unknown version", method: http.MethodPost, path: "/api/v2/calculate"},
		{name: "sub path of the endpoint", method: http.MethodPost, path: calculatePath + "/extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, tt.method, tt.path, "", "application/json")
			assertAPIError(t, rec, http.StatusNotFound, apierror.CodeNotFound)
		})
	}
}

func TestRouter_ErrorEnvelopeCarriesNothingElse(t *testing.T) {
	rec := post(t, `{"operation":"divide","a":1,"b":0}`)

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if len(body) != 1 {
		t.Errorf("error response has %d top-level keys, want only \"error\": %s", len(body), rec.Body.String())
	}

	var detail map[string]string
	if err := json.Unmarshal(body["error"], &detail); err != nil {
		t.Fatalf("error member is not an object of strings: %v", err)
	}
	if len(detail) != 2 || detail["code"] == "" || detail["message"] == "" {
		t.Errorf("error member = %v, want exactly a code and a message", detail)
	}
}
