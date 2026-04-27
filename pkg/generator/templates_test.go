package generator

import (
	"strings"
	"testing"

	"github.com/opencost/bingen/pkg/types"
)

// TestToDefaultValueRejectsCodeInjection ensures that a default value crafted to
// escape the generated source string and inject Go code is rejected (panics
// with a parse error) rather than silently splicing arbitrary tokens into
// generated code. Without validation, an annotation like
// `default=foo"; var X = exec.Command(...)` would compile-time eval at
// `go generate`.
func TestToDefaultValueRejectsCodeInjection(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic for malicious default value")
		}
	}()

	// Numeric type with a payload that's not a valid Go expression.
	ToDefaultValue(types.TypeInt, `0); var Pwned = "yes"; _ = (0`)
}

// TestToDefaultValueAcceptsLegitimateDefaults exercises the happy path so we
// don't accidentally regress accepted defaults during validation.
func TestToDefaultValueAcceptsLegitimateDefaults(t *testing.T) {
	cases := []struct {
		typeCode uint8
		value    string
		want     string
	}{
		{types.TypeString, "momma-container", `"momma-container"`},
		{types.TypeBool, "true", "true"},
		{types.TypeBool, "false", "false"},
		{types.TypeBool, "", "false"},
		{types.TypeInt, "42", "int(42)"},
		{types.TypeInt, "", "int(0)"},
		{types.TypeFloat64, "3.14", "float64(3.14)"},
		{types.TypeInt, "Foo(15)", "int(Foo(15))"},
	}

	for _, c := range cases {
		got := ToDefaultValue(c.typeCode, c.value)
		if got != c.want {
			t.Errorf("ToDefaultValue(%d, %q): got %q, want %q", c.typeCode, c.value, got, c.want)
		}
	}
}

// TestToDefaultValueQuotesMaliciousString verifies that a string default is
// fully escaped — the produced literal must round-trip through the Go parser
// without being interpretable as anything other than a string.
func TestToDefaultValueQuotesMaliciousString(t *testing.T) {
	in := `evil"; doBad(); //`
	got := ToDefaultValue(types.TypeString, in)
	if !strings.HasPrefix(got, `"`) || !strings.HasSuffix(got, `"`) {
		t.Errorf("string default not quoted: %q", got)
	}
	if strings.Contains(got, `doBad`) && !strings.Contains(got, `\"`) {
		t.Errorf("string default did not escape inner quotes: %q", got)
	}
}

// TestToRawDefaultRejectsInjection ensures the raw splice for non-basic types
// rejects anything that isn't a Go expression.
func TestToRawDefaultRejectsInjection(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for non-expression default")
		}
	}()
	ToRawDefault(`bad"; var X = 1; _ = "`)
}

// TestToRawDefaultAcceptsExpressions verifies legitimate Go expressions pass.
func TestToRawDefaultAcceptsExpressions(t *testing.T) {
	for _, in := range []string{`Foo(15)`, `[]int{1,2,3}`, `&Bar{Baz: 1}`} {
		if got := ToRawDefault(in); got != in {
			t.Errorf("ToRawDefault(%q) = %q", in, got)
		}
	}
}
