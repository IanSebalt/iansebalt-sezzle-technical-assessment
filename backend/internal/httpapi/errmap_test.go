package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
	"github.com/iansebalt/sezzle-calculator/internal/calculator"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus int
	}{
		{
			name: "unsupported operation", err: calculator.ErrUnsupportedOperation,
			wantCode: apierror.CodeUnsupportedOperation, wantStatus: http.StatusBadRequest,
		},
		{
			name: "division by zero", err: calculator.ErrDivisionByZero,
			wantCode: apierror.CodeDivisionByZero, wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "negative square root", err: calculator.ErrNegativeSquareRoot,
			wantCode: apierror.CodeNegativeSquareRoot, wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "non finite result", err: calculator.ErrNonFiniteResult,
			wantCode: apierror.CodeNonFiniteResult, wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "wrapped sentinel", err: fmt.Errorf("evaluating: %w", calculator.ErrDivisionByZero),
			wantCode: apierror.CodeDivisionByZero, wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unrecognised error", err: errors.New("something unexpected"),
			wantCode: apierror.CodeInternal, wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapError(tt.err)
			if got.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", got.Code, tt.wantCode)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("status = %d, want %d", got.Status, tt.wantStatus)
			}
		})
	}
}

func TestMapErrorHidesUnrecognisedDetail(t *testing.T) {
	secret := "connection string leaked into the error"

	if got := mapError(errors.New(secret)).Message; strings.Contains(got, secret) {
		t.Errorf("message %q leaks the underlying error", got)
	}
}

func TestSupportedOperationsListsEveryOperation(t *testing.T) {
	got := supportedOperations()

	for _, op := range calculator.Supported() {
		if !strings.Contains(got, op.String()) {
			t.Errorf("supportedOperations() = %q, missing %q", got, op)
		}
	}
	if want := len(calculator.Supported()) - 1; strings.Count(got, ",") != want {
		t.Errorf("supportedOperations() = %q, want %d separators", got, want)
	}
}
