package httpapi

import (
	"errors"
	"strings"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
	"github.com/iansebalt/sezzle-calculator/internal/calculator"
)

// mapError is the only translation point between domain failures and the wire error contract.
// Anything unrecognised is deliberately reported as an internal error rather than leaked.
func mapError(err error) *apierror.Error {
	switch {
	case errors.Is(err, calculator.ErrUnsupportedOperation):
		return apierror.UnsupportedOperation(supportedOperations())
	case errors.Is(err, calculator.ErrDivisionByZero):
		return apierror.DivisionByZero()
	case errors.Is(err, calculator.ErrNegativeSquareRoot):
		return apierror.NegativeSquareRoot()
	case errors.Is(err, calculator.ErrNonFiniteResult):
		return apierror.NonFiniteResult()
	default:
		return apierror.Internal()
	}
}

func supportedOperations() string {
	operations := calculator.Supported()
	names := make([]string, len(operations))
	for i, op := range operations {
		names[i] = op.String()
	}
	return strings.Join(names, ", ")
}
