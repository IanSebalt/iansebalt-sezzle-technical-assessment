package apierror

import (
	"strings"
	"testing"
	"unicode"
)

// catalog lists one sample of every entry, so the invariants below cover the whole contract.
func catalog() []*Error {
	return []*Error{
		MethodNotAllowed(),
		NotFound(),
		UnsupportedMediaType(),
		PayloadTooLarge(),
		MalformedJSON(),
		UnknownField(),
		MissingOperation(),
		InvalidOperation(),
		InvalidOperand("a"),
		MissingOperand("b"),
		UnsupportedOperation("add, subtract"),
		DivisionByZero(),
		NegativeSquareRoot(),
		NonFiniteResult(),
		Internal(),
	}
}

func TestCatalogCodesAreUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, entry := range catalog() {
		if seen[entry.Code] {
			t.Errorf("code %q appears twice in the catalogue", entry.Code)
		}
		seen[entry.Code] = true
	}
}

func TestCatalogEntriesAreWellFormed(t *testing.T) {
	for _, entry := range catalog() {
		t.Run(entry.Code, func(t *testing.T) {
			if entry.Status < 400 || entry.Status > 599 {
				t.Errorf("status = %d, want a 4xx or 5xx", entry.Status)
			}
			if entry.Code == "" {
				t.Error("code is empty")
			}
			if entry.Code != strings.ToUpper(entry.Code) {
				t.Errorf("code %q is not upper snake case", entry.Code)
			}
			if entry.Message == "" {
				t.Fatal("message is empty")
			}
			if first := []rune(entry.Message)[0]; !unicode.IsUpper(first) {
				t.Errorf("message %q does not start with a capital letter", entry.Message)
			}
			if !strings.HasSuffix(entry.Message, ".") {
				t.Errorf("message %q does not end in a full stop", entry.Message)
			}
		})
	}
}

func TestErrorMessageCombinesCodeAndCopy(t *testing.T) {
	err := DivisionByZero()
	want := "DIVISION_BY_ZERO: Cannot divide by zero."

	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestInvalidOperandNamesTheOffendingField(t *testing.T) {
	if got := InvalidOperand("b").Message; !strings.Contains(got, `"b"`) {
		t.Errorf("message %q does not name field b", got)
	}
}

func TestMissingOperandNamesTheOffendingField(t *testing.T) {
	if got := MissingOperand("a").Message; !strings.Contains(got, `"a"`) {
		t.Errorf("message %q does not name field a", got)
	}
}

func TestUnsupportedOperationListsTheAllowedNames(t *testing.T) {
	if got := UnsupportedOperation("add, sqrt").Message; !strings.Contains(got, "add, sqrt") {
		t.Errorf("message %q does not list the allowed operations", got)
	}
}
