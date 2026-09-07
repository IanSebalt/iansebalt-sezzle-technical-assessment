package httpapi

import (
	"net/http"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
	"github.com/iansebalt/sezzle-calculator/internal/calculator"
)

func (a *api) calculate(w http.ResponseWriter, r *http.Request) {
	req, apiErr := decodeCalculateRequest(w, r)
	if apiErr != nil {
		a.writeError(w, apiErr)
		return
	}

	operation, err := calculator.ParseOperation(*req.Operation)
	if err != nil {
		a.writeError(w, mapError(err))
		return
	}

	if apiErr := requireOperands(operation, req); apiErr != nil {
		a.writeError(w, apiErr)
		return
	}

	var b float64
	if req.B != nil {
		b = *req.B
	}

	result, err := calculator.Evaluate(operation, *req.A, b)
	if err != nil {
		a.writeError(w, mapError(err))
		return
	}

	response := calculateResponse{
		Operation: operation.String(),
		A:         *req.A,
		Result:    result,
	}
	// Unary operations echo only the operand they actually used.
	if operation.Arity() == 2 {
		response.B = req.B
	}

	a.writeJSON(w, http.StatusOK, response)
}

func requireOperands(operation calculator.Operation, req calculateRequest) *apierror.Error {
	if req.A == nil {
		return apierror.MissingOperand("a")
	}
	if operation.Arity() == 2 && req.B == nil {
		return apierror.MissingOperand("b")
	}
	return nil
}
