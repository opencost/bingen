// Package java implements the Java exporter for bingen. It reuses the
// language-neutral type intermediate representation produced by
// internal/types and emits Java 21 records, split encoder/decoder types, and
// the supporting runtime for the annotated Go structs.
package java

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Language-specific option keys accepted via the CLI (-opt key=value).
const (
	// OptBasePackage is the required base Java package. The go package name is
	// appended to it so that the final package's last segment always equals the
	// source go package (ie:  base "com.opencost" + go "container" ->
	// "com.opencost.container").
	OptBasePackage = "basePackage"

	// OptOutputDir is the optional root output directory. Generated sources are
	// written beneath it following the java package path. Defaults to ".".
	OptOutputDir = "outputDir"

	// OptGoImport is the optional fully-qualified import path of the source go
	// package. It is used to build the interface type strings written on the
	// wire. When omitted it is resolved from the source directory at generation
	// time.
	OptGoImport = "goImport"

	// OptRuntimePackage is the optional java package the shared runtime (buffers,
	// codec contexts, registry) is emitted into. It is kept separate from the
	// data package so it is not duplicated when several go packages are generated
	// under the same base package. Defaults to DefaultRuntimePackage.
	OptRuntimePackage = "runtimePackage"
)

// DefaultRuntimePackage is the java package the shared bingen runtime is
// emitted into when OptRuntimePackage is not provided.
const DefaultRuntimePackage = "com.bingen"

// Options is the resolved, validated configuration for a single Java
// generation invocation.
type Options struct {
	// BasePackage is the validated base java package (ie:  "com.opencost").
	BasePackage string

	// GoPackage is the source go package name (ie:  "container").
	GoPackage string

	// JavaPackage is BasePackage + "." + GoPackage. Its last segment always
	// equals GoPackage.
	JavaPackage string

	// OutputDir is the root directory generated sources are written beneath.
	OutputDir string

	// GoImport is the fully-qualified import path of the source go package, or
	// the empty string when it should be resolved at generation time.
	GoImport string

	// RuntimePackage is the java package the shared runtime is emitted into.
	RuntimePackage string
}

// FromConfig validates the language-specific config map for the provided go
// package and returns the resolved Options.
func FromConfig(goPackage string, cfg map[string]string) (*Options, error) {
	base := strings.TrimSpace(cfg[OptBasePackage])
	if base == "" {
		return nil, fmt.Errorf("java: required option %q is not set", OptBasePackage)
	}
	if err := validateJavaPackage(base); err != nil {
		return nil, fmt.Errorf("java: invalid %q %q: %w", OptBasePackage, base, err)
	}

	outputDir := strings.TrimSpace(cfg[OptOutputDir])
	if outputDir == "" {
		outputDir = "."
	}

	runtimePkg := strings.TrimSpace(cfg[OptRuntimePackage])
	if runtimePkg == "" {
		runtimePkg = DefaultRuntimePackage
	}
	if err := validateJavaPackage(runtimePkg); err != nil {
		return nil, fmt.Errorf("java: invalid %q %q: %w", OptRuntimePackage, runtimePkg, err)
	}

	return &Options{
		BasePackage:    base,
		GoPackage:      goPackage,
		JavaPackage:    base + "." + goPackage,
		OutputDir:      outputDir,
		GoImport:       strings.TrimSpace(cfg[OptGoImport]),
		RuntimePackage: runtimePkg,
	}, nil
}

// PackageDir returns the directory generated data .java files are written to,
// resolving the java package into nested directories beneath OutputDir.
func (o *Options) PackageDir() string {
	return o.packageDir(o.JavaPackage)
}

// RuntimePackageDir returns the directory the shared runtime .java files are
// written to.
func (o *Options) RuntimePackageDir() string {
	return o.packageDir(o.RuntimePackage)
}

// packageDir resolves a java package into nested directories beneath OutputDir.
func (o *Options) packageDir(pkg string) string {
	segments := strings.Split(pkg, ".")
	return filepath.Join(append([]string{o.OutputDir}, segments...)...)
}
