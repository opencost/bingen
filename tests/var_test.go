package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/opencost/bingen/pkg/generator/vars"
)

func checkVarNames(t *testing.T, current string, varName ...string) bool {
	var b strings.Builder
	for _, vn := range varName {
		if strings.Contains(current, vn) {
			fmt.Fprintf(&b, "Unexpected '%s' found in variable name.\n", vn)
		}
	}

	r := b.String()
	if r != "" {
		t.Error(r)
		return false
	}

	return true
}

func TestVarNames(t *testing.T) {
	varNames := vars.NewAlphaVarNames()

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
	}

	if current != "vvvv" {
		t.Errorf("Expected \"vvvv\", got: \"%s\"\n", current)
	}
}

func TestVarNamesIgnoreFront(t *testing.T) {
	varNames := vars.NewAlphaVarNames(vars.A)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
		if !checkVarNames(t, current, "a") {
			return
		}
	}

	if current != "zzzz" {
		t.Errorf("Expected \"zzzz\", got: \"%s\"\n", current)
	}

}

func TestVarNamesIgnoreBack(t *testing.T) {
	varNames := vars.NewAlphaVarNames(vars.Z)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
		if !checkVarNames(t, current, "z") {
			return
		}
	}

	if current != "yyyy" {
		t.Errorf("Expected \"yyyy\", got: \"%s\"\n", current)
	}
}

func TestVarNamesIgnoreFrontBack(t *testing.T) {
	varNames := vars.NewAlphaVarNames(vars.A, vars.Z)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
		if !checkVarNames(t, current, "a", "z") {
			return
		}
	}
}

func TestVarNamesIgnoreSequence(t *testing.T) {
	ignore := []rune{vars.A, vars.B, vars.C, vars.D}
	varNames := vars.NewAlphaVarNames(ignore...)
	chars := vars.ToAlphaChars(ignore...)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()

		if !checkVarNames(t, current, chars...) {
			return
		}
	}
}

func TestAllBut(t *testing.T) {
	ignore := vars.SkipAllBut(vars.I, vars.J, vars.K)
	varNames := vars.NewAlphaVarNames(ignore...)
	chars := vars.ToAlphaChars(ignore...)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()

		if !checkVarNames(t, current, chars...) {
			return
		}
	}
}

// TestVarNamesIgnoreAllButOne verifies that when all letters except one are
// skipped, the generator produces a deterministic sequence that is always the
// remaining letter repeated, with the repetition count increasing each call.
func TestVarNamesIgnoreAllButOne(t *testing.T) {
	ignore := vars.AllAlphaRunes()
	varNames := vars.NewAlphaVarNames(ignore[:len(ignore)-1]...)

	// The only allowed rune is Z (last in the alphabet).
	want := []string{
		"z",
		"zz",
		"zzz",
		"zzzz",
		"zzzzz",
		"zzzzzz",
		"zzzzzzz",
		"zzzzzzzz",
		"zzzzzzzzz",
		"zzzzzzzzzz",
	}

	for i, expected := range want {
		got := varNames.Next()
		if got != expected {
			t.Errorf("iteration %d: expected %q, got %q", i, expected, got)
		}
	}
}

func TestVarAlphaStrings(t *testing.T) {
	cases := []struct {
		r    rune
		want string
	}{
		{vars.A, "a"},
		{vars.B, "b"},
		{vars.C, "c"},
		{vars.I, "i"},
		{vars.J, "j"},
		{vars.M, "m"},
		{vars.Z, "z"},
	}

	for _, tc := range cases {
		got := vars.AlphaCharFor(tc.r)
		if got != tc.want {
			t.Errorf("AlphaCharFor(%v): expected %q, got %q", tc.r, tc.want, got)
		}
	}

	// Out-of-range runes should return empty string.
	if got := vars.AlphaCharFor(rune('0')); got != "" {
		t.Errorf("AlphaCharFor(non-alpha): expected empty string, got %q", got)
	}
	if got := vars.AlphaCharFor(rune('A')); got != "" {
		t.Errorf("AlphaCharFor(uppercase A): expected empty string, got %q", got)
	}

	// Confirm the full alphabet is produced in order.
	runes := vars.AllAlphaRunes()
	if len(runes) != 26 {
		t.Fatalf("AllAlphaRunes: expected 26 runes, got %d", len(runes))
	}
	if got := vars.AlphaCharFor(runes[0]); got != "a" {
		t.Errorf("expected first rune to map to \"a\", got %q", got)
	}
	if got := vars.AlphaCharFor(runes[len(runes)-1]); got != "z" {
		t.Errorf("expected last rune to map to \"z\", got %q", got)
	}
}
