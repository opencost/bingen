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
		t.Errorf(r)
		return false
	}

	return true
}

func TestVarNames(t *testing.T) {
	varNames := vars.NewAlphaVarNames()

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
		fmt.Println(current)
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
		fmt.Println(current)
		if !checkVarNames(t, current, "a") {
			return
		}
	}

	if current != "zzzz" {
		t.Errorf("Expected \"vvvv\", got: \"%s\"\n", current)
	}

}

func TestVarNamesIgnoreBack(t *testing.T) {
	varNames := vars.NewAlphaVarNames(vars.Z)

	current := ""
	for i := 0; i < 100; i++ {
		current = varNames.Next()
		fmt.Println(current)
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
		fmt.Println(current)
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
		fmt.Println(current)

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
		fmt.Println(current)

		if !checkVarNames(t, current, chars...) {
			return
		}
	}
}

func TestVarNamesIgnoreAllButOne(t *testing.T) {
	ignore := vars.AllAlphaRunes()
	varNames := vars.NewAlphaVarNames(ignore[:len(ignore)-1]...)

	current := ""
	for i := 0; i < 10; i++ {
		current = varNames.Next()
		fmt.Println(current)
	}
}

func TestVarAlphaStrings(t *testing.T) {
	for _, r := range vars.AllAlphaRunes() {
		fmt.Println(vars.AlphaCharFor(r))
	}
}
