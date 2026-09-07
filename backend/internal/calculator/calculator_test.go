package calculator

import (
	"errors"
	"math"
	"testing"
)

const epsilon = 1e-9

func almostEqual(got, want float64) bool {
	if got == want {
		return true
	}
	diff := math.Abs(got - want)
	if want == 0 {
		return diff < epsilon
	}
	return diff/math.Abs(want) < epsilon
}

func assertResult(t *testing.T, op Operation, a, b, got, want float64, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Evaluate(%q, %v, %v) unexpected error: %v", op, a, b, err)
	}
	if !almostEqual(got, want) {
		t.Errorf("Evaluate(%q, %v, %v) = %v, want %v", op, a, b, got, want)
	}
}

func TestEvaluate_Arithmetic(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		a, b float64
		want float64
	}{
		{name: "add positives", op: Add, a: 2, b: 3, want: 5},
		{name: "add negatives", op: Add, a: -7, b: -3, want: -10},
		{name: "add identity", op: Add, a: 42, b: 0, want: 42},
		{name: "add decimals", op: Add, a: 0.1, b: 0.2, want: 0.3},
		{name: "subtract to positive", op: Subtract, a: 10, b: 4, want: 6},
		{name: "subtract to negative", op: Subtract, a: 4, b: 10, want: -6},
		{name: "subtract a negative", op: Subtract, a: 5, b: -5, want: 10},
		{name: "multiply positives", op: Multiply, a: 6, b: 7, want: 42},
		{name: "multiply by zero", op: Multiply, a: 12345, b: 0, want: 0},
		{name: "multiply mixed signs", op: Multiply, a: -3, b: 4, want: -12},
		{name: "multiply decimals", op: Multiply, a: 2.5, b: 1.5, want: 3.75},
		{name: "divide evenly", op: Divide, a: 12, b: 4, want: 3},
		{name: "divide to a fraction", op: Divide, a: 10, b: 4, want: 2.5},
		{name: "divide mixed signs", op: Divide, a: -9, b: 3, want: -3},
		{name: "divide zero numerator", op: Divide, a: 0, b: 5, want: 0},
		{name: "power squared", op: Power, a: 2, b: 10, want: 1024},
		{name: "power identity", op: Power, a: 7, b: 1, want: 7},
		{name: "sqrt perfect square", op: SquareRoot, a: 81, want: 9},
		{name: "sqrt of zero", op: SquareRoot, a: 0, want: 0},
		{name: "sqrt irrational", op: SquareRoot, a: 2, want: math.Sqrt2},
		{name: "percentage of a value", op: Percentage, a: 15, b: 200, want: 30},
		{name: "percentage whole", op: Percentage, a: 100, b: 55, want: 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.op, tt.a, tt.b)
			assertResult(t, tt.op, tt.a, tt.b, got, tt.want, err)
		})
	}
}

func TestEvaluate_Percentage(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{name: "15% of 200", a: 15, b: 200, want: 30},
		{name: "0% of anything", a: 0, b: 999, want: 0},
		{name: "anything of 0", a: 25, b: 0, want: 0},
		{name: "fractional percent", a: 2.5, b: 400, want: 10},
		{name: "over one hundred percent", a: 150, b: 40, want: 60},
		{name: "negative percent", a: -10, b: 50, want: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(Percentage, tt.a, tt.b)
			assertResult(t, Percentage, tt.a, tt.b, got, tt.want, err)
		})
	}
}

func TestEvaluate_PowerEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{name: "zero to the zero is one", a: 0, b: 0, want: 1},
		{name: "anything to the zero", a: 12345, b: 0, want: 1},
		{name: "negative exponent", a: 2, b: -3, want: 0.125},
		{name: "fractional exponent", a: 9, b: 0.5, want: 3},
		{name: "negative base integer exponent", a: -2, b: 3, want: -8},
		{name: "one to a huge exponent", a: 1, b: 1e6, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(Power, tt.a, tt.b)
			assertResult(t, Power, tt.a, tt.b, got, tt.want, err)
		})
	}
}

func TestEvaluate_DivisionByZero(t *testing.T) {
	negativeZero := math.Copysign(0, -1)

	tests := []struct {
		name string
		a, b float64
	}{
		{name: "positive over zero", a: 10, b: 0},
		{name: "negative over zero", a: -10, b: 0},
		{name: "zero over zero", a: 0, b: 0},
		{name: "positive over negative zero", a: 10, b: negativeZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(Divide, tt.a, tt.b)
			if !errors.Is(err, ErrDivisionByZero) {
				t.Errorf("Evaluate(divide, %v, %v) error = %v, want ErrDivisionByZero", tt.a, tt.b, err)
			}
		})
	}
}

func TestEvaluate_NegativeSquareRoot(t *testing.T) {
	for _, a := range []float64{-1, -0.0001, -1e300} {
		_, err := Evaluate(SquareRoot, a, 0)
		if !errors.Is(err, ErrNegativeSquareRoot) {
			t.Errorf("Evaluate(sqrt, %v) error = %v, want ErrNegativeSquareRoot", a, err)
		}
	}

	if _, err := Evaluate(SquareRoot, 0, 0); err != nil {
		t.Errorf("Evaluate(sqrt, 0) unexpected error: %v", err)
	}
}

func TestEvaluate_NonFiniteResult(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		a, b float64
	}{
		{name: "multiplication overflows", op: Multiply, a: 1e308, b: 10},
		{name: "addition overflows", op: Add, a: math.MaxFloat64, b: math.MaxFloat64},
		{name: "power overflows", op: Power, a: 10, b: 400},
		{name: "multiplication overflows to negative infinity", op: Multiply, a: -1e308, b: 10},
		{name: "cube root of a negative is undefined", op: Power, a: -8, b: 1.0 / 3.0},
		{name: "zero to a negative exponent", op: Power, a: 0, b: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.op, tt.a, tt.b)
			if !errors.Is(err, ErrNonFiniteResult) {
				t.Errorf("Evaluate(%q, %v, %v) error = %v, want ErrNonFiniteResult", tt.op, tt.a, tt.b, err)
			}
		})
	}
}

func TestEvaluate_UnaryIgnoresB(t *testing.T) {
	for _, b := range []float64{0, 5, -3, 1e10} {
		got, err := Evaluate(SquareRoot, 16, b)
		if err != nil {
			t.Fatalf("Evaluate(sqrt, 16, %v) unexpected error: %v", b, err)
		}
		if got != 4 {
			t.Errorf("Evaluate(sqrt, 16, %v) = %v, want 4", b, got)
		}
	}
}

func TestEvaluate_UnsupportedOperation(t *testing.T) {
	if _, err := Evaluate(Operation("modulo"), 10, 3); !errors.Is(err, ErrUnsupportedOperation) {
		t.Errorf("Evaluate(modulo) error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestEvaluate_ReturnsZeroAlongsideAnError(t *testing.T) {
	got, err := Evaluate(Divide, 10, 0)
	if err == nil {
		t.Fatal("expected an error")
	}
	if got != 0 {
		t.Errorf("Evaluate returned %v alongside an error, want 0", got)
	}
}
