package calculator

import "math"

// Evaluate applies op to its operands. Unary operations ignore b.
func Evaluate(op Operation, a, b float64) (float64, error) {
	result, err := apply(op, a, b)
	if err != nil {
		return 0, err
	}

	// Overflow yields ±Inf and undefined results yield NaN, both silently. Neither is expressible
	// in JSON, so they are reported rather than serialised into an invalid response body.
	if math.IsInf(result, 0) || math.IsNaN(result) {
		return 0, ErrNonFiniteResult
	}
	return result, nil
}

func apply(op Operation, a, b float64) (float64, error) {
	switch op {
	case Add:
		return a + b, nil
	case Subtract:
		return a - b, nil
	case Multiply:
		return a * b, nil
	case Divide:
		// -0.0 == 0 in Go, so this rejects a negative zero divisor too.
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		return a / b, nil
	case Power:
		return math.Pow(a, b), nil
	case SquareRoot:
		if a < 0 {
			return 0, ErrNegativeSquareRoot
		}
		return math.Sqrt(a), nil
	case Percentage:
		// Multiplying before dividing keeps common cases exact: 7% of 3 is 0.21, where dividing
		// first yields 0.21000000000000002. The finite guard catches the products that overflow.
		return a * b / 100, nil
	default:
		return 0, ErrUnsupportedOperation
	}
}
