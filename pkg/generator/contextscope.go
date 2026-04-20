package generator

// Scope represents a state within the GeneratorContext, generally pertaining to the number of indentions
// to prepend on any GeneratorContext::Write() calls.
type Scope interface {
	// Push writes the formatted output, and pushes a new scope and returns it
	Push(format string, rest ...interface{}) Scope

	// Pop pops the current scope, then writes the provided formatted output
	Pop(format string, rest ...interface{}) Scope

	// PopPush pops the current scope, writes the formatted output, then pushes new scope and returns it
	PopPush(format string, rest ...interface{}) Scope

	// Write writes the formatted string and prepends the current indention level and returns
	// the current scope
	Write(format string, rest ...interface{}) Scope

	// Writeln writes the formatted string and prepends the current indention level, and appends a newline,
	// and returns the current scope.
	Writeln(format string, rest ...interface{}) Scope
}

// contextScope is the default implementation of Scope
type contextScope struct {
	gc GeneratorContext
}

// scope creates a new Scope implementation for a GeneratorContext
func scope(gc GeneratorContext) Scope {
	return &contextScope{gc}
}

// Push writes the formatted output, and pushes a new scope and returns it
func (cs *contextScope) Push(format string, rest ...interface{}) Scope {
	cs.gc.Writeln(format, rest...)
	cs.gc.PushIndent()
	return cs
}

// Pop pops the current scope, then writes the provided formatted string with an appended newline
func (cs *contextScope) Pop(format string, rest ...interface{}) Scope {
	cs.gc.PopIndent()
	cs.gc.Writeln(format, rest...)
	return cs
}

// PopPush pops the current scope, writes the formatted output, then pushes new scope and returns it
func (cs *contextScope) PopPush(format string, rest ...interface{}) Scope {
	cs.Pop(format, rest...)
	return cs.Push("")
}

// Write writes the formatted string and prepends the current indention level.
func (cs *contextScope) Write(format string, rest ...interface{}) Scope {
	cs.gc.Write(format, rest...)
	return cs
}

// Writeln writes the formatted string and prepends the current indention level, and appends a newline.
func (cs *contextScope) Writeln(format string, rest ...interface{}) Scope {
	cs.gc.Writeln(format, rest...)
	return cs
}
