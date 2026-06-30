package generator

import (
	"fmt"
	"strings"

	"github.com/opencost/bingen/internal/generator/vars"
	"github.com/opencost/bingen/internal/types"
)

// GeneratorContext is a stateful context that controls the flow of generation based on
// the scope of the generation.
type GeneratorContext interface {
	// VersionSetConst returns the version set constant used for this generator
	VersionSetConst() string

	// HandleError handles any inline errors that occur during reading or writing by returning the
	// generated error handling code for the provided error string
	HandleError(errString string) string

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
	WithErrorHandler(handler func(string) string) GeneratorContext
}

// GeneratorContextFactory creates new GeneratorContext instances
type GeneratorContextFactory interface {
	// NewContext creates a new GeneratorContext implementation
	NewContext(opts *types.GenerateTypeOpts) GeneratorContext
}

// genContextFactory is the default implementation of GeneratorContextFactory
type genContextFactory struct {
	errorHandler func(string) string
}

// NewContext creates a new GeneratorContext implementation
func (gcf *genContextFactory) NewContext(opts *types.GenerateTypeOpts) GeneratorContext {
	if opts == nil {
		opts = &types.GenerateTypeOpts{}
	}

	return &genContext{
		vsc:        fmt.Sprintf("%sCodecVersion", titleCaser.String(opts.SetName)),
		opts:       opts,
		loops:      vars.NewAlphaVarNames(vars.SkipAllBut(vars.I, vars.J)...),
		maps:       vars.NewAlphaVarNames(vars.SkipAllBut(vars.Z, vars.V)...),
		vars:       vars.NewAlphaVarNames(vars.I, vars.J, vars.K, vars.Z, vars.V),
		oks:        vars.NewAlphaVarNames(),
		errs:       vars.NewAlphaVarNames(),
		errHandler: gcf.errorHandler,
	}
}

// NewGeneratorContextFactory creates a default GeneratorContext with the default error handler
func NewGeneratorContextFactory(errorHandler func(string) string) GeneratorContextFactory {
	return &genContextFactory{
		errorHandler: errorHandler,
	}
}

// genContext is the default implementation of GeneratorContext
type genContext struct {
	vsc        string
	opts       *types.GenerateTypeOpts
	vars       vars.VarNames
	loops      vars.VarNames
	maps       vars.VarNames
	oks        vars.VarNames
	errs       vars.VarNames
	errHandler func(string) string
}

// WithErrorHandler allows the context to be copied with a new error handler override
func (gc *genContext) WithErrorHandler(handler func(string) string) GeneratorContext {
	return &genContext{
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

// HandleError handles any inline errors that occur during reading or writing by
// returning the generated code to be used in handling the errString provided.
//
// There are two straight-forward implementations here, where the non-streaming generators
// simply return the error, and the streaming variants set the error flag. The default implementation
// is the non-streaming error handler
func (gc *genContext) HandleError(errString string) string {
	return gc.errHandler(errString)
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
