package calculator

// Operation identifies a supported arithmetic operation.
type Operation string

const (
	Add        Operation = "add"
	Subtract   Operation = "subtract"
	Multiply   Operation = "multiply"
	Divide     Operation = "divide"
	Power      Operation = "power"
	SquareRoot Operation = "sqrt"
	Percentage Operation = "percentage"
)

// arities defines both which operations exist and how many operands each consumes.
var arities = map[Operation]int{
	Add:        2,
	Subtract:   2,
	Multiply:   2,
	Divide:     2,
	Power:      2,
	SquareRoot: 1,
	Percentage: 2,
}

// supported fixes the order used in documentation and error messages; the map alone would iterate
// randomly. TestSupportedCoversEveryOperation keeps the two in step.
var supported = []Operation{Add, Subtract, Multiply, Divide, Power, SquareRoot, Percentage}

// Supported returns every operation in a stable, human-friendly order.
func Supported() []Operation {
	names := make([]Operation, len(supported))
	copy(names, supported)
	return names
}

// ParseOperation converts raw input into an Operation, rejecting anything unsupported.
func ParseOperation(raw string) (Operation, error) {
	op := Operation(raw)
	if _, ok := arities[op]; !ok {
		return "", ErrUnsupportedOperation
	}
	return op, nil
}

// Arity reports how many operands the operation consumes. Unknown operations report 0.
func (o Operation) Arity() int {
	return arities[o]
}

func (o Operation) String() string {
	return string(o)
}
