// Package core is the language-neutral core of the bingen code generators. It
// walks the type IR and drives the binary format structure (field ordering, nil
// flags, string-table branching, version gating, reference/alias resolution,
// and recursion), delegating all target-language syntax to a Backend implementation.
package core

import (
	"fmt"
	"sort"
	"strings"
)

// Emit accumulates generated source lines with indentation, unique variable
// names, the set of required imports, and the first error encountered.
type Emit struct {
	sb      strings.Builder
	nextVar int
	indent  int
	imports map[string]bool
	err     error
}

// NewEmit returns an Emit that starts at the given indent level.
func NewEmit(indent int) *Emit {
	return &Emit{
		indent:  indent,
		imports: make(map[string]bool),
	}
}

// NextVar returns a new unique local variable name.
func (e *Emit) NextVar() string {
	v := fmt.Sprintf("v%d", e.nextVar)
	e.nextVar++
	return v
}

// Line writes a single indented line.
func (e *Emit) Line(s string) {
	e.sb.WriteString(strings.Repeat("    ", e.indent))
	e.sb.WriteString(s)
	e.sb.WriteByte('\n')
}

// Linef writes a single indented, formatted line.
func (e *Emit) Linef(format string, args ...any) {
	e.Line(fmt.Sprintf(format, args...))
}

// Indented runs body with one additional level of indentation.
func (e *Emit) Indented(body func()) {
	e.indent++
	body()
	e.indent--
}

// Import records a required import.
func (e *Emit) Import(imp string) {
	e.imports[imp] = true
}

// Errf records the first error encountered; later calls are ignored.
func (e *Emit) Errf(format string, args ...any) {
	if e.err == nil {
		e.err = fmt.Errorf(format, args...)
	}
}

// Result returns the rendered body (trailing newline trimmed), the sorted set of
// required imports, and any error captured.
func (e *Emit) Result() (string, []string, error) {
	if e.err != nil {
		return "", nil, e.err
	}

	imports := make([]string, 0, len(e.imports))
	for imp := range e.imports {
		imports = append(imports, imp)
	}
	sort.Strings(imports)

	return strings.TrimRight(e.sb.String(), "\n"), imports, nil
}
