package java

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencost/bingen/internal/generator/core"
	"github.com/opencost/bingen/internal/types"
)

// Generate emits the Java exporter output for the collected types into the
// directory derived from opts. The per-type records, split encoder/decoder
// types, and runtime are added in subsequent phases; this scaffolding
// establishes the end-to-end pipeline and output layout.
func Generate(sourceDir string, tc types.TypeCollection, opts *Options) error {
	if opts == nil {
		return fmt.Errorf("java: nil options")
	}

	pkgDir := opts.PackageDir()
	if err := os.MkdirAll(pkgDir, 0o750); err != nil {
		return fmt.Errorf("java: creating output dir %q: %w", pkgDir, err)
	}

	if err := writeFile(filepath.Join(pkgDir, "package-info.java"), PackageInfoTemplate, opts); err != nil {
		return err
	}

	if err := emitRuntime(opts); err != nil {
		return err
	}

	// Resolve the interface implementation graph via the go type checker when any
	// generated struct references a named interface.
	info := &interfaceInfo{
		Implements: map[string][]string{},
	}

	interfaces := neededInterfaces(tc)
	if len(interfaces) > 0 {
		loaded, err := loadInterfaceInfo(sourceDir, structNames(tc), interfaces)
		if err != nil {
			return err
		}
		
		info = loaded
		if err := emitInterfaces(pkgDir, opts, interfaces); err != nil {
			return err
		}
		if err := emitRegistry(pkgDir, opts, tc, info); err != nil {
			return err
		}
	}

	for _, t := range tc.Types() {
		st, ok := t.(*types.StructType)
		if !ok {
			continue
		}
		if err := emitRecord(pkgDir, opts, st, info); err != nil {
			return err
		}
		if err := emitEncoder(pkgDir, opts, st); err != nil {
			return err
		}
		if err := emitDecoder(pkgDir, opts, st); err != nil {
			return err
		}
		if st.Opts.IsStreamable {
			if err := emitStreamer(pkgDir, opts, st); err != nil {
				return err
			}
		}
	}

	fmt.Printf("java: generated package %s -> %s (%d types)\n", opts.JavaPackage, pkgDir, len(tc.Types()))
	return nil
}

// emitInterfaces writes an empty marker interface per referenced go interface.
func emitInterfaces(pkgDir string, opts *Options, interfaces []string) error {
	for _, name := range interfaces {
		params := InterfaceParams{Package: opts.JavaPackage, Name: name}
		if err := writeFile(filepath.Join(pkgDir, name+".java"), InterfaceTemplate, params); err != nil {
			return err
		}
	}
	return nil
}

// emitRegistry writes the wire-string -> codec dispatch table for every
// generated struct type.
func emitRegistry(pkgDir string, opts *Options, tc types.TypeCollection, info *interfaceInfo) error {
	var entries []RegistryEntry
	for _, name := range structNames(tc) {
		entries = append(entries, RegistryEntry{
			Wire: info.wireName(name),
			Type: name,
		})
	}

	params := RegistryParams{
		Package:        opts.JavaPackage,
		RuntimePackage: opts.RuntimePackage,
		Entries:        entries,
	}
	return writeFile(filepath.Join(pkgDir, "TypeRegistry.java"), RegistryTemplate, params)
}

// emitRecord writes the immutable record + builder for a single struct type.
func emitRecord(pkgDir string, opts *Options, st *types.StructType, info *interfaceInfo) error {
	params := RecordParams{
		Package:          opts.JavaPackage,
		Imports:          requiredImports(st),
		ImplementsClause: implementsClause(info.Implements[st.Name()]),
		Type:             st,
	}
	return writeFile(filepath.Join(pkgDir, st.Name()+".java"), RecordTemplate, params)
}

// implementsClause builds the java `implements A, B` clause, or "" when the
// struct satisfies no generated interfaces.
func implementsClause(interfaces []string) string {
	if len(interfaces) == 0 {
		return ""
	}
	return " implements " + strings.Join(interfaces, ", ")
}

// emitEncoder writes the BinEncoder implementation for a single struct type.
func emitEncoder(pkgDir string, opts *Options, st *types.StructType) error {
	body, imports, err := core.EncodeStruct(NewJavaBackend(opts.RuntimePackage), st)
	if err != nil {
		return fmt.Errorf("java: encoding %s: %w", st.Name(), err)
	}

	params := CodecParams{
		Package:        opts.JavaPackage,
		RuntimePackage: opts.RuntimePackage,
		Imports:        imports,
		Version:        st.Opts.SetVersion,
		StringTable:    st.Opts.IsGenerateStringTable,
		PreProcess:     st.Opts.IsPreProcess,
		Body:           body,
		Type:           st,
	}
	return writeFile(filepath.Join(pkgDir, st.Name()+"Encoder.java"), EncoderTemplate, params)
}

// emitDecoder writes the BinDecoder implementation for a single struct type.
func emitDecoder(pkgDir string, opts *Options, st *types.StructType) error {
	body, imports, err := core.DecodeStruct(NewJavaBackend(opts.RuntimePackage), st)
	if err != nil {
		return fmt.Errorf("java: decoding %s: %w", st.Name(), err)
	}

	params := CodecParams{
		Package:        opts.JavaPackage,
		RuntimePackage: opts.RuntimePackage,
		Imports:        imports,
		Version:        st.Opts.SetVersion,
		PostProcess:    st.Opts.IsPostProcess,
		Migration:      st.Opts.IsMigration,
		Body:           body,
		Type:           st,
	}
	return writeFile(filepath.Join(pkgDir, st.Name()+"Decoder.java"), DecoderTemplate, params)
}

// emitStreamer writes the push streamer for a streamable struct type.
func emitStreamer(pkgDir string, opts *Options, st *types.StructType) error {
	body, imports, err := core.StreamStruct(NewJavaBackend(opts.RuntimePackage), st)
	if err != nil {
		return fmt.Errorf("java: streaming %s: %w", st.Name(), err)
	}

	params := CodecParams{
		Package:        opts.JavaPackage,
		RuntimePackage: opts.RuntimePackage,
		Imports:        imports,
		Version:        st.Opts.SetVersion,
		Body:           body,
		Type:           st,
	}
	return writeFile(filepath.Join(pkgDir, st.Name()+"Streamer.java"), StreamerTemplate, params)
}

// emitRuntime writes the shared runtime (buffers, and in later phases the codec
// contexts and registry) into the configured runtime package. It is idempotent
// and safe to run for every generated package.
func emitRuntime(opts *Options) error {
	runtimeDir := opts.RuntimePackageDir()
	if err := os.MkdirAll(runtimeDir, 0o750); err != nil {
		return fmt.Errorf("java: creating runtime dir %q: %w", runtimeDir, err)
	}

	params := RuntimeParams{
		RuntimePackage: opts.RuntimePackage,
	}

	for fileName, tmpl := range runtimeTemplates {
		if err := writeFile(filepath.Join(runtimeDir, fileName), tmpl, params); err != nil {
			return err
		}
	}

	return nil
}

// writeFile renders the named template with data and writes it to path.
func writeFile(path string, template string, data any) error {
	var buf bytes.Buffer
	if err := writeTemplate(&buf, template, data); err != nil {
		return err
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("java: writing %q: %w", path, err)
	}
	return nil
}
