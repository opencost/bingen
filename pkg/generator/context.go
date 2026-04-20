package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kubecost/bingen/pkg/generator/vars"
	"github.com/kubecost/bingen/pkg/types"
)

// GeneratorContext is a stateful context that controls the flow of generation based on
// the scope of the generation.
type GeneratorContext interface {
	// VersionSetConst returns the version set constant used for this generator
	VersionSetConst() string

	// Write writes the formatted string and prepends the current indention level.
	Write(format string, rest ...interface{})

	// Writeln writes the formatted string and prepends the current indention level, and appends a newline.
	// Will not append newline to an empty format
	Writeln(format string, rest ...interface{})

	// WriteErrorHandler handles any inline errors that occur during reading or writing
	// errorVar is the error implementation to check for non-nil
	WriteErrorHandler(errCheck Scope, newError string)

	// PushScope pushes a new indention scope for the GeneratorContext after writing the provided format. The
	// new indention will prepend any GeneratorContext::Write() calls until Pop() is called on the returned scope.
	PushScope(format string, rest ...interface{}) Scope

	// AsScope returns the current context as its own scope
	AsScope() Scope

	// PushDebugWrite pushes a write debug scope for a specific generator type and encodable type. Calling
	// End() on the return value will pop the scope and write out the final state.
	PushDebugWrite(genType string, typeName string) DebugScope

	// PushDebugRead pushes a read debug scope for a specific generator type and encodable type. Calling
	// End() on the return value will pop the scope and write out the final state.
	PushDebugRead(genType string, typeName string) DebugScope

	// IsStreamable returns true if we should generate streaming implementations of the type
	IsStreamable() bool

	// IsStringTable returns true if we store/load a table of strings in the format
	IsStringTable() bool

	// IsPreProcess returns true if we should pre process the generated struct before encoding.
	IsPreProcess() bool

	// IsPostProcess returns true if we should post process the generated struct before returning.
	IsPostProcess() bool

	// IsMigration returns true if we should test for a migration scenario (version delta), and call
	// a migration func.
	IsMigration() bool

	// PushIndent advances the indentation level, which is prepended to any Write() or Writeln() output.
	PushIndent()

	// PopIndent reduces the indentation level, which is prepended to any Write() or Writeln() output.
	PopIndent()

	// NextVar returns the next available general variable name which can be used to avoid collision.
	NextVar() string

	// NextMapVar returns the next available map variable name which can be used to avoid collisions.
	NextMapVar() string

	// NextLoopVar returns the next available loop index variable name which can be used to avoid collisions.
	NextLoopVar() string

	// NextErrVar returns the next available ok variable name.
	NextOKVar() string

	// NextErrVar returns the next available error variable name.
	NextErrVar() string

	// WithErrorHandler allows the context to be copied with a new error handler override
	WithErrorHandler(handler func(Scope, string)) GeneratorContext
}

// GeneratorContextFactory creates new GeneratorContext instances
type GeneratorContextFactory interface {
	// NewContext creates a new GeneratorContext implementation
	NewContext(opts *types.GenerateTypeOpts) GeneratorContext
}

// genContextFactory is the default implementation of GeneratorContextFactory
type genContextFactory struct {
	version       uint8
	buffer        *bytes.Buffer
	indentionSize int
	errorHandler  func(Scope, string)
}

// NewContext creates a new GeneratorContext implementation
func (gcf *genContextFactory) NewContext(opts *types.GenerateTypeOpts) GeneratorContext {
	if opts == nil {
		opts = &types.GenerateTypeOpts{}
	}

	return &genContext{
		b:          gcf.buffer,
		i:          NewIndent(4),
		vsc:        fmt.Sprintf("%sCodecVersion", strings.Title(opts.SetName)),
		opts:       opts,
		loops:      vars.NewAlphaVarNames(vars.SkipAllBut(vars.I, vars.J)...),
		maps:       vars.NewAlphaVarNames(vars.SkipAllBut(vars.Z, vars.V)...),
		vars:       vars.NewAlphaVarNames(vars.I, vars.J, vars.K, vars.Z, vars.V),
		oks:        vars.NewAlphaVarNames(),
		errs:       vars.NewAlphaVarNames(),
		errHandler: gcf.errorHandler,
	}
}

// NewGeneratorContextFactory creates a default GeneratorContext implementation using
// a target byte buffer, and the size to use for indentation.
func NewGeneratorContextFactory(buffer *bytes.Buffer, indentionSize int, errorHandler func(Scope, string)) GeneratorContextFactory {
	return &genContextFactory{
		buffer:        buffer,
		indentionSize: indentionSize,
		errorHandler:  errorHandler,
	}
}

// genContext is the default implementation of GeneratorContext
type genContext struct {
	b          *bytes.Buffer
	i          Indent
	vsc        string
	opts       *types.GenerateTypeOpts
	vars       vars.VarNames
	loops      vars.VarNames
	maps       vars.VarNames
	oks        vars.VarNames
	errs       vars.VarNames
	errHandler func(Scope, string)
}

// WithErrorHandler allows the context to be copied with a new error handler override
func (gc *genContext) WithErrorHandler(handler func(Scope, string)) GeneratorContext {
	return &genContext{
		b:          gc.b,
		i:          gc.i,
		vsc:        gc.vsc,
		opts:       gc.opts,
		vars:       gc.vars,
		loops:      gc.loops,
		maps:       gc.maps,
		oks:        gc.oks,
		errs:       gc.errs,
		errHandler: handler,
	}
}

// Write writes the formatted string and prepends the current indention level.
func (gc *genContext) Write(format string, rest ...interface{}) {
	formatted := fmt.Sprintf(format, rest...)
	if formatted == "" {
		return
	}

	fmt.Fprintf(gc.b, "%s%s", gc.i, formatted)
}

// Writeln writes the formatted string and prepends the current indention level, and appends a newline.
func (gc *genContext) Writeln(format string, rest ...interface{}) {
	formatted := fmt.Sprintf(format, rest...)
	if formatted == "" {
		return
	}

	fmt.Fprintf(gc.b, "%s%s\n", gc.i, formatted)
}

// WriteErrorHandler handles any inline errors that occur during reading or writing
// errorVar is the error implementation to check for non-nil and sets the error usage
// to the newError parameter.
//
// There are two straight-forward implementations here, where the non-streaming generators
// simply return the error, and the streaming variants set the error flag. The default implementation
// is the non-streaming error handler
func (gc *genContext) WriteErrorHandler(errCheck Scope, newErr string) {
	gc.errHandler(errCheck, newErr)
}

// PushScope pushes a new indention scope for the genContext after writing the provided format. The
// new indention will prepend any genContext::Write() calls until Pop() is called on the returned scope.
func (gc *genContext) PushScope(format string, rest ...interface{}) Scope {
	s := scope(gc)
	return s.Push(format, rest...)
}

// AsScope returns the current context as its own scope
func (gc *genContext) AsScope() Scope {
	return scope(gc)
}

// PushDebugWrite pushes a write debug scope for a specific generator type and encodable type. Calling
// End() on the return value will pop the scope and write out the final state.
func (gc *genContext) PushDebugWrite(genType string, typeName string) DebugScope {
	return BeginDebugScope(gc, DebugEncodeTypeWrite, genType, typeName)
}

// PushDebugRead pushes a read debug scope for a specific generator type and encodable type. Calling
// End() on the return value will pop the scope and write out the final state.
func (gc *genContext) PushDebugRead(genType string, typeName string) DebugScope {
	return BeginDebugScope(gc, DebugEncodeTypeRead, genType, typeName)
}

func (gc *genContext) VersionSetConst() string {
	return gc.vsc
}

// IsStreamable returns true if we should generate streaming implementations of the type
func (gc *genContext) IsStreamable() bool {
	return gc.opts.IsStreamable
}

// IsStringTable returns true if we store/load a table of strings in the format
func (gc *genContext) IsStringTable() bool {
	return gc.opts.IsGenerateStringTable
}

// IsPreProcess returns true if we should pre process the generated struct before encoding.
func (gc *genContext) IsPreProcess() bool {
	return gc.opts.IsPreProcess
}

// IsPostProcess returns true if we should post process the generated struct before returning.
func (gc *genContext) IsPostProcess() bool {
	return gc.opts.IsPostProcess
}

// IsMigration returns true if we should test for a migration scenario (version delta), and call
// a migration func.
func (gc *genContext) IsMigration() bool {
	return gc.opts.IsMigration
}

func (gc *genContext) Out() *bytes.Buffer {
	return gc.b
}

// PushIndent advances the indentation level, which is prepended to any Write() or Writeln() output.
func (gc *genContext) PushIndent() {
	gc.i.Out()
}

// PopIndent reduces the indentation level, which is prepended to any Write() or Writeln() output.
func (gc *genContext) PopIndent() {
	gc.i.In()
}

// NextVar returns the next available general variable name which can be used to avoid collision.
func (gc *genContext) NextVar() string {
	return gc.vars.Next()
}

// NextMapVar returns the next available map variable name which can be used to avoid collisions.
func (gc *genContext) NextMapVar() string {
	return gc.maps.Next()
}

// NextLoopVar returns the next available loop index variable name which can be used to avoid collisions.
func (gc *genContext) NextLoopVar() string {
	return gc.loops.Next()
}

// NextErrVar returns the next available ok variable name.
func (gc *genContext) NextOKVar() string {
	n := gc.oks.Next()
	return fmt.Sprintf("ok%s", strings.ToUpper(n))
}

// NextErrVar returns the next available error variable name.
func (gc *genContext) NextErrVar() string {
	n := gc.errs.Next()
	return fmt.Sprintf("err%s", strings.ToUpper(n))
}
