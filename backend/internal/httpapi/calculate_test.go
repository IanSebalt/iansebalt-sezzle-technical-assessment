package httpapi_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
	"github.com/iansebalt/sezzle-calculator/internal/httpapi"
)

const calculatePath = "/api/v1/calculate"

func newTestRouter() http.Handler {
	return httpapi.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func post(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	return request(t, http.MethodPost, calculatePath, body, "application/json")
}

func request(t *testing.T, method, path, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	newTestRouter().ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not JSON: %v (body: %s)", err, rec.Body.String())
	}
	return body
}

func floatField(t *testing.T, body map[string]any, key string) float64 {
	t.Helper()

	value, ok := body[key]
	if !ok {
		t.Fatalf("response has no %q field: %v", key, body)
	}
	number, ok := value.(float64)
	if !ok {
		t.Fatalf("field %q = %v, want a number", key, value)
	}
	return number
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, wantStatus, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error body is not JSON: %v (body: %s)", err, rec.Body.String())
	}
	if body.Error == nil {
		t.Fatalf("response is missing the error envelope: %s", rec.Body.String())
	}
	if body.Error.Code != wantCode {
		t.Errorf("error code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Error("error message is empty")
	}
}

func TestCalculate_Success(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    float64
		binary  bool
		wantOp  string
		wantOpA float64
	}{
		{name: "add", body: `{"operation":"add","a":2,"b":3}`, want: 5, binary: true, wantOp: "add", wantOpA: 2},
		{name: "subtract", body: `{"operation":"subtract","a":10,"b":4}`, want: 6, binary: true, wantOp: "subtract", wantOpA: 10},
		{name: "multiply", body: `{"operation":"multiply","a":6,"b":7}`, want: 42, binary: true, wantOp: "multiply", wantOpA: 6},
		{name: "divide", body: `{"operation":"divide","a":10,"b":4}`, want: 2.5, binary: true, wantOp: "divide", wantOpA: 10},
		{name: "power", body: `{"operation":"power","a":2,"b":10}`, want: 1024, binary: true, wantOp: "power", wantOpA: 2},
		{name: "percentage", body: `{"operation":"percentage","a":15,"b":200}`, want: 30, binary: true, wantOp: "percentage", wantOpA: 15},
		{name: "sqrt", body: `{"operation":"sqrt","a":81}`, want: 9, binary: false, wantOp: "sqrt", wantOpA: 81},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(t, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			body := decodeBody(t, rec)
			if got := body["operation"]; got != tt.wantOp {
				t.Errorf("operation = %v, want %q", got, tt.wantOp)
			}
			if got := floatField(t, body, "a"); got != tt.wantOpA {
				t.Errorf("a = %v, want %v", got, tt.wantOpA)
			}
			if got := floatField(t, body, "result"); got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}

			_, hasB := body["b"]
			if hasB != tt.binary {
				t.Errorf("response contains b = %v, want %v", hasB, tt.binary)
			}
		})
	}
}

func TestCalculate_SqrtIgnoresASuppliedB(t *testing.T) {
	rec := post(t, `{"operation":"sqrt","a":25,"b":99}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if got := floatField(t, body, "result"); got != 5 {
		t.Errorf("result = %v, want 5", got)
	}
	if _, hasB := body["b"]; hasB {
		t.Error("response echoes b for a unary operation")
	}
}

func TestCalculate_RequestErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name: "malformed json", body: `{"operation":`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMalformedJSON,
		},
		{
			name: "empty body", body: ``,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMalformedJSON,
		},
		{
			name: "json array instead of object", body: `[1,2]`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMalformedJSON,
		},
		{
			name: "trailing data after the object", body: `{"operation":"add","a":1,"b":2}{"extra":true}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMalformedJSON,
		},
		{
			name: "unknown field", body: `{"operation":"add","a":1,"b":2,"c":3}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeUnknownField,
		},
		{
			name: "missing operation", body: `{"a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMissingOperation,
		},
		{
			name: "operation is not text", body: `{"operation":5,"a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeInvalidOperation,
		},
		{
			name: "unsupported operation", body: `{"operation":"modulo","a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeUnsupportedOperation,
		},
		{
			name: "missing operand a", body: `{"operation":"add","b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMissingOperand,
		},
		{
			name: "missing operand b", body: `{"operation":"add","a":1}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMissingOperand,
		},
		{
			name: "missing operand a for a unary operation", body: `{"operation":"sqrt"}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeMissingOperand,
		},
		{
			name: "non numeric operand", body: `{"operation":"add","a":"ten","b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeInvalidOperand,
		},
		{
			name: "boolean operand", body: `{"operation":"add","a":true,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeInvalidOperand,
		},
		{
			name: "operand literal out of range", body: `{"operation":"add","a":1e400,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: apierror.CodeInvalidOperand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertAPIError(t, post(t, tt.body), tt.wantStatus, tt.wantCode)
		})
	}
}

func TestCalculate_ArithmeticErrors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{name: "division by zero", body: `{"operation":"divide","a":10,"b":0}`, wantCode: apierror.CodeDivisionByZero},
		{name: "zero divided by zero", body: `{"operation":"divide","a":0,"b":0}`, wantCode: apierror.CodeDivisionByZero},
		{name: "square root of a negative", body: `{"operation":"sqrt","a":-9}`, wantCode: apierror.CodeNegativeSquareRoot},
		{name: "multiplication overflows", body: `{"operation":"multiply","a":1e308,"b":10}`, wantCode: apierror.CodeNonFiniteResult},
		{name: "power overflows", body: `{"operation":"power","a":10,"b":400}`, wantCode: apierror.CodeNonFiniteResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertAPIError(t, post(t, tt.body), http.StatusUnprocessableEntity, tt.wantCode)
		})
	}
}

func TestCalculate_InvalidOperandNamesTheField(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		field string
	}{
		{name: "a is text", body: `{"operation":"add","a":"ten","b":2}`, field: "a"},
		{name: "b is text", body: `{"operation":"add","a":1,"b":"two"}`, field: "b"},
		{name: "a is out of range", body: `{"operation":"add","a":1e400,"b":2}`, field: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(t, tt.body)
			if !strings.Contains(rec.Body.String(), `\"`+tt.field+`\"`) {
				t.Errorf("message does not name field %q: %s", tt.field, rec.Body.String())
			}
		})
	}
}

func TestCalculate_ContentTypeIsEnforced(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantStatus  int
		wantCode    string
	}{
		{name: "missing", contentType: "", wantStatus: http.StatusUnsupportedMediaType, wantCode: apierror.CodeUnsupportedMediaType},
		{name: "plain text", contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType, wantCode: apierror.CodeUnsupportedMediaType},
		{name: "form encoded", contentType: "application/x-www-form-urlencoded", wantStatus: http.StatusUnsupportedMediaType, wantCode: apierror.CodeUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, http.MethodPost, calculatePath, `{"operation":"add","a":1,"b":2}`, tt.contentType)
			assertAPIError(t, rec, tt.wantStatus, tt.wantCode)
		})
	}
}

func TestCalculate_AcceptsContentTypeWithCharset(t *testing.T) {
	rec := request(t, http.MethodPost, calculatePath, `{"operation":"add","a":1,"b":2}`, "application/json; charset=utf-8")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCalculate_BodyTooLarge(t *testing.T) {
	oversized := `{"operation":"` + strings.Repeat("x", 8<<10) + `"}`

	assertAPIError(t, post(t, oversized), http.StatusRequestEntityTooLarge, apierror.CodePayloadTooLarge)
}
