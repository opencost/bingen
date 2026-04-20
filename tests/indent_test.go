package test

import (
	"strings"
	"testing"

	"github.com/opencost/bingen/pkg/generator"
)

func TestBasicIndent(t *testing.T) {
	i := generator.NewIndent(4)

	if got := i.String(); got != "" {
		t.Errorf("expected empty initial indent, got %q", got)
	}

	i.OutN(2)
	if got, want := i.String(), strings.Repeat(" ", 8); got != want {
		t.Errorf("after OutN(2) expected %q, got %q", want, got)
	}

	i.In()
	if got, want := i.String(), strings.Repeat(" ", 4); got != want {
		t.Errorf("after In expected %q, got %q", want, got)
	}

	i.In()
	if got := i.String(); got != "" {
		t.Errorf("after second In expected empty string, got %q", got)
	}

	// InN should not underflow below an empty string.
	i.InN(5)
	if got := i.String(); got != "" {
		t.Errorf("InN beyond zero should remain empty, got %q", got)
	}

	i.Out()
	if got, want := i.String(), strings.Repeat(" ", 4); got != want {
		t.Errorf("after Out expected %q, got %q", want, got)
	}
}

func TestIndentCustomSize(t *testing.T) {
	i := generator.NewIndent(2)
	i.Out()
	i.Out()
	if got, want := i.String(), "    "; got != want {
		t.Errorf("expected %q with 2-space indent at depth 2, got %q", want, got)
	}
}
