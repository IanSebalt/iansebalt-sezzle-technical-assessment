package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

func TestRequireJSONContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantErr     bool
	}{
		{name: "plain json", contentType: "application/json"},
		{name: "json with charset", contentType: "application/json; charset=utf-8"},
		{name: "json with uppercase", contentType: "Application/JSON"},
		{name: "missing", contentType: "", wantErr: true},
		{name: "plain text", contentType: "text/plain", wantErr: true},
		{name: "unparseable", contentType: "application/json;;", wantErr: true},
		{name: "json subtype suffix", contentType: "application/vnd.api+json", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			apiErr := requireJSONContentType(req)
			if tt.wantErr && apiErr == nil {
				t.Errorf("requireJSONContentType(%q) = nil, want an error", tt.contentType)
			}
			if !tt.wantErr && apiErr != nil {
				t.Errorf("requireJSONContentType(%q) = %v, want nil", tt.contentType, apiErr)
			}
		})
	}
}

func TestDecodeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{name: "body too large", err: &http.MaxBytesError{Limit: maxBodyBytes}, wantCode: apierror.CodePayloadTooLarge},
		{name: "wrong type for an operand", err: &json.UnmarshalTypeError{Field: "a"}, wantCode: apierror.CodeInvalidOperand},
		{name: "wrong type for the operation", err: &json.UnmarshalTypeError{Field: "operation"}, wantCode: apierror.CodeInvalidOperation},
		{name: "wrong type for the whole body", err: &json.UnmarshalTypeError{}, wantCode: apierror.CodeMalformedJSON},
		{name: "syntax error", err: &json.SyntaxError{}, wantCode: apierror.CodeMalformedJSON},
		{name: "empty body", err: io.EOF, wantCode: apierror.CodeMalformedJSON},
		{name: "truncated body", err: io.ErrUnexpectedEOF, wantCode: apierror.CodeMalformedJSON},
		{name: "unknown field", err: errors.New(`json: unknown field "c"`), wantCode: apierror.CodeUnknownField},
		{name: "anything else", err: errors.New("boom"), wantCode: apierror.CodeMalformedJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeError(tt.err).Code; got != tt.wantCode {
				t.Errorf("decodeError(%v) code = %q, want %q", tt.err, got, tt.wantCode)
			}
		})
	}
}

func TestDecodeCalculateRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"operation":"add","a":1.5,"b":-2}`))
	req.Header.Set("Content-Type", "application/json")

	decoded, apiErr := decodeCalculateRequest(httptest.NewRecorder(), req)
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	if decoded.Operation == nil || *decoded.Operation != "add" {
		t.Errorf("operation = %v, want add", decoded.Operation)
	}
	if decoded.A == nil || *decoded.A != 1.5 {
		t.Errorf("a = %v, want 1.5", decoded.A)
	}
	if decoded.B == nil || *decoded.B != -2 {
		t.Errorf("b = %v, want -2", decoded.B)
	}
}

func TestDecodeCalculateRequestDistinguishesAbsentFromZero(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"operation":"add","a":0}`))
	req.Header.Set("Content-Type", "application/json")

	decoded, apiErr := decodeCalculateRequest(httptest.NewRecorder(), req)
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	if decoded.A == nil || *decoded.A != 0 {
		t.Errorf("a = %v, want a pointer to 0", decoded.A)
	}
	if decoded.B != nil {
		t.Errorf("b = %v, want nil for an absent field", *decoded.B)
	}
}
