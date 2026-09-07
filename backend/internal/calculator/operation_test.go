package calculator

import (
	"errors"
	"testing"
)

func TestParseOperation(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Operation
		wantErr error
	}{
		{name: "add", raw: "add", want: Add},
		{name: "subtract", raw: "subtract", want: Subtract},
		{name: "multiply", raw: "multiply", want: Multiply},
		{name: "divide", raw: "divide", want: Divide},
		{name: "power", raw: "power", want: Power},
		{name: "sqrt", raw: "sqrt", want: SquareRoot},
		{name: "percentage", raw: "percentage", want: Percentage},
		{name: "unknown name", raw: "modulo", wantErr: ErrUnsupportedOperation},
		{name: "empty", raw: "", wantErr: ErrUnsupportedOperation},
		{name: "wrong case", raw: "ADD", wantErr: ErrUnsupportedOperation},
		{name: "surrounding whitespace", raw: " add ", wantErr: ErrUnsupportedOperation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOperation(tt.raw)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseOperation(%q) error = %v, want %v", tt.raw, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseOperation(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestOperationArity(t *testing.T) {
	tests := []struct {
		op   Operation
		want int
	}{
		{Add, 2},
		{Subtract, 2},
		{Multiply, 2},
		{Divide, 2},
		{Power, 2},
		{SquareRoot, 1},
		{Percentage, 2},
		{Operation("modulo"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.op.String(), func(t *testing.T) {
			if got := tt.op.Arity(); got != tt.want {
				t.Errorf("%q.Arity() = %d, want %d", tt.op, got, tt.want)
			}
		})
	}
}

func TestSupportedCoversEveryOperation(t *testing.T) {
	names := Supported()
	if len(names) != len(arities) {
		t.Fatalf("Supported() has %d entries, arities has %d", len(names), len(arities))
	}

	seen := make(map[Operation]bool, len(names))
	for _, op := range names {
		if _, ok := arities[op]; !ok {
			t.Errorf("Supported() lists %q, which has no arity", op)
		}
		if seen[op] {
			t.Errorf("Supported() lists %q twice", op)
		}
		seen[op] = true
	}
}

func TestSupportedReturnsACopy(t *testing.T) {
	first := Supported()
	first[0] = Operation("tampered")

	if second := Supported(); second[0] == Operation("tampered") {
		t.Error("Supported() exposes its backing array to callers")
	}
}

func TestOperationString(t *testing.T) {
	if got := Divide.String(); got != "divide" {
		t.Errorf("Divide.String() = %q, want %q", got, "divide")
	}
}
