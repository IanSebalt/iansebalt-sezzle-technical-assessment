package calculator

import "errors"

// Sentinel errors stating why an evaluation could not produce a number. They deliberately carry no
// status code or user-facing copy: translating them for a transport is that transport's job.
var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNegativeSquareRoot   = errors.New("square root of a negative number")
	ErrNonFiniteResult      = errors.New("result is not a finite number")
)
