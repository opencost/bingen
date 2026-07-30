package golang

import (
	"fmt"

	"github.com/opencost/bingen/internal/types"
)

// GeneratorContext exposes the per-type flags and constants the marshaller,
// unmarshaller, and streamer method skeletons need.
type GeneratorContext interface {
	// VersionSetConst returns the version set constant used for this generator
	VersionSetConst() string

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
}

// GeneratorContextFactory creates new GeneratorContext instances
type GeneratorContextFactory interface {
	// NewContext creates a new GeneratorContext implementation
	NewContext(opts *types.GenerateTypeOpts) GeneratorContext
}

// genContextFactory is the default implementation of GeneratorContextFactory
type genContextFactory struct{}

// NewContext creates a new GeneratorContext implementation
func (gcf *genContextFactory) NewContext(opts *types.GenerateTypeOpts) GeneratorContext {
	if opts == nil {
		opts = &types.GenerateTypeOpts{}
	}

	return &genContext{
		vsc:  fmt.Sprintf("%sCodecVersion", titleCaser.String(opts.SetName)),
		opts: opts,
	}
}

// NewGeneratorContextFactory creates the default GeneratorContextFactory.
func NewGeneratorContextFactory() GeneratorContextFactory {
	return &genContextFactory{}
}

// genContext is the default implementation of GeneratorContext
type genContext struct {
	vsc  string
	opts *types.GenerateTypeOpts
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
