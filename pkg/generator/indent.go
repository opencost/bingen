package generator

import (
	"fmt"
	"strings"
)

// Indent is an implementation prototype that controls the indention levels during
// code generation.
type Indent interface {
	// String returns the current indention string, and allows Indent to be formatted with %s
	fmt.Stringer

	// Out increases the indentation level.
	Out()

	// OutN increases the indentation level a specific number of times
	OutN(n int)

	// In decreases the indentation level.
	In()

	// InN decreases the indentation level a specific number of times.
	InN(n int)
}

// defaultIndent is the default implementation of Indent
type defaultIndent struct {
	current string
	indent  string
}

// NewIndention returns an Indent implementation which uses a specific number of spaces
// as an indention level.
func NewIndent(spaces int) Indent {
	return &defaultIndent{
		current: "",
		indent:  strings.Repeat(" ", spaces),
	}
}

// Out increases the indentation level.
func (i *defaultIndent) Out() {
	i.OutN(1)
}

// OutN increases the indentation level a specific number of times
func (i *defaultIndent) OutN(n int) {
	i.current += strings.Repeat(i.indent, n)
}

// In decreases the indentation level.
func (i *defaultIndent) In() {
	i.InN(1)
}

// InN decreases the indentation level a specific number of times.
func (i *defaultIndent) InN(n int) {
	l := len(i.current)
	il := len(i.indent) * n

	idx := l - il
	if idx < 0 {
		idx = 0
	}

	i.current = i.current[:idx]
}

// String returns the current indention string, and allows Indent to be formatted with %s
func (i *defaultIndent) String() string {
	return i.current
}
